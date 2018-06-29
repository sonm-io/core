package connor

import (
	"context"
	"fmt"
	"time"

	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	EthPool                 = "stratum+tcp://eth-eu1.nanopool.org:9999"
	numberOfIterationsForH1 = 5
	numberOfLives           = 5
)

type PoolModule struct {
	c *Connor
}

func NewPoolModules(c *Connor) *PoolModule {
	return &PoolModule{
		c: c,
	}
}

type BanStatus int32

const (
	BanStatusBANNED       BanStatus = 1
	BanStatusMASTERBAN    BanStatus = 2
	BanStatusWORKERINPOOL BanStatus = 6
)

func (p *PoolModule) DeployNewContainer(ctx context.Context, cfg *Config, deal *sonm.Deal, image string) (*sonm.StartTaskReply, error) {
	env := map[string]string{
		"ETH_POOL": EthPool,
		"WALLET":   cfg.PoolAddress.EthPoolAddr,
		"WORKER":   deal.Id.String(),
		"EMAIL":    p.c.cfg.OtherParameters.EmailForPool,
	}
	container := &sonm.Container{
		Image: image,
		Env:   env,
	}
	spec := &sonm.TaskSpec{
		Container: container,
		Registry:  &sonm.Registry{},
		Resources: &sonm.AskPlanResources{},
	}
	startTaskRequest := &sonm.StartTaskRequest{
		DealID: deal.GetId(),
		Spec:   spec,
	}
	reply, err := p.c.TaskClient.Start(ctx, startTaskRequest)
	if err != nil {
		return nil, fmt.Errorf("cannot create start task request %s", err)
	}
	return reply, nil
}

// Checks for a deal in the worker list. If it is not there, adds.
func (p *PoolModule) AddWorkerToPoolDB(ctx context.Context, deal *sonm.DealInfoReply, addr string) error {
	val, err := p.c.db.GetWorkerFromPoolDb(deal.Deal.Id.String())
	if err != nil {
		return err
	}
	if val == deal.Deal.Id.String() {
		return nil
	}
	if err := p.c.db.SavePoolIntoDB(&database.PoolDb{
		DealID:    deal.Deal.Id.Unwrap().Int64(),
		PoolID:    addr,
		TimeStart: time.Now()}); err != nil {
		return err
	}
	return nil
}

// Updates and evaluates hashrate by workers, depending on the iteration.
func (p *PoolModule) DefaultPoolHashrateTracking(ctx context.Context, reportedPool watchers.PoolWatcher, avgPool watchers.PoolWatcher) error {
	workers, err := p.c.db.GetWorkersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get worker from pool DB :: %v", err)
	}

	for _, w := range workers {
		// FIXME: change value BadGuy in Db
		if w.BadGuy > numberOfLives {
			continue
		}
		iteration := int64(w.Iterations + 1)

		dealInfo, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(w.DealID))
		if err != nil {
			return fmt.Errorf("cannot get deal from market %v\r\n", w.DealID)
		}

		bidHashrate, err := p.ReturnBidHashrateForDeal(ctx, dealInfo)
		if err != nil {
			return err
		}

		if iteration < numberOfIterationsForH1 {
			if err = p.UpdateRHPoolData(ctx, reportedPool, p.c.cfg.PoolAddress.EthPoolAddr); err != nil {
				return err
			}
			changePercentRHWorker := 100 - ((uint64(w.WorkerReportedHashrate*hashes) * 100) / bidHashrate)
			if err = p.DetectingDeviation(ctx, changePercentRHWorker, w, dealInfo); err != nil {
				return err
			}
		} else {
			err := p.UpdateAvgPoolData(ctx, avgPool, p.c.cfg.PoolAddress.EthPoolAddr+"/1");
			if err != nil {
				return err
			}
			p.c.logger.Info("getting avg pool data for worker", zap.Int64("deal", w.DealID))
			changeAvgWorker := 100 - ((uint64(w.WorkerAvgHashrate*hashes) * 100) / bidHashrate)
			if err = p.DetectingDeviation(ctx, changeAvgWorker, w, dealInfo); err != nil {
				return err
			}
		}
		err = p.c.db.UpdateIterationPoolDB(iteration, w.DealID);
		if err != nil {
			return err
		}
	}
	return nil
}

//Detection of getting a lowered hashrate and sending to a blacklist (create deal finish request).
func (p *PoolModule) DetectingDeviation(ctx context.Context, changePercentDeviationWorker uint64, worker *database.PoolDb, dealInfo *sonm.DealInfoReply) error {

	if changePercentDeviationWorker >= uint64(p.c.cfg.Sensitivity.WorkerLimitChangePercent) {
		if worker.BadGuy < numberOfLives {
			newStatus := worker.BadGuy + 1
			err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, newStatus, time.Now());
			if err != nil {
				return err
			}
		} else {
			if err := p.DestroyDeal(ctx, dealInfo); err != nil {
				return err
			}
			err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWORKERINPOOL), time.Now());
			if err != nil {
				return err
			}
			p.c.logger.Info("Destroy deal", zap.String("bad status in pool", dealInfo.Deal.Id.String()))
		}
	} else if changePercentDeviationWorker >= 20 {
		err := p.DestroyDeal(ctx, dealInfo);
		if err != nil {
			return err
		}
		err = p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWORKERINPOOL), time.Now());
		if err != nil {
			return err
		}
		p.c.logger.Info("Destroy deal", zap.String("bad status in pool", dealInfo.Deal.Id.String()))
	}
	return nil
}

// Update pool data for first hour (use reported hashrate without shares)
func (p *PoolModule) UpdateRHPoolData(ctx context.Context, poolRHData watchers.PoolWatcher, addr string) error {
	if err := poolRHData.Update(ctx); err != nil {
		return err
	}
	dataRH, err := poolRHData.GetData(addr)
	if err != nil {
		p.c.logger.Warn("cannot get reported hashrate data", zap.Error(err))
		return err
	}

	for _, rh := range dataRH.Data {
		p.c.db.UpdateReportedHashratePoolDB(rh.Worker, rh.Hashrate, time.Now())
	}
	return nil
}

// Update pool data for another time (use average hashrate with shares)
func (p *PoolModule) UpdateAvgPoolData(ctx context.Context, poolAvgData watchers.PoolWatcher, addr string) error {
	if err := poolAvgData.Update(ctx); err != nil {
		return err
	}

	dataRH, err := poolAvgData.GetData(addr)
	if err != nil {
		p.c.logger.Error("cannot get average pool data", zap.Error(err))
		return err
	}

	for _, rh := range dataRH.Data {
		p.c.db.UpdateAvgPoolDB(rh.Worker, rh.Hashrate, time.Now())
	}

	return nil
}

func (p *PoolModule) ReturnBidHashrateForDeal(ctx context.Context, dealInfo *sonm.DealInfoReply) (uint64, error) {
	bidOrder, err := p.c.Market.GetOrderByID(ctx, &sonm.ID{Id: dealInfo.Deal.BidID.Unwrap().String()})
	if err != nil {
		p.c.logger.Error("cannot get order from market by ID", zap.Error(err))
		return 0, err
	}
	return bidOrder.GetBenchmarks().GPUEthHashrate(), nil
}

// Create deal finish request
func (p *PoolModule) DestroyDeal(ctx context.Context, dealInfo *sonm.DealInfoReply) error {
	if _, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{
		Id:            dealInfo.Deal.Id,
		BlacklistType: sonm.BlacklistType_BLACKLIST_MASTER,
	}); err != nil {
		p.c.logger.Error("couldn't finish deal", zap.String("deal", dealInfo.Deal.Id.String()), zap.Error(err))
		return err
	}
	err := p.c.db.SetDestroyDealPoolDB(int64(DeployStatusDESTROYED), dealInfo.Deal.Id.Unwrap().Int64())
	if err != nil {
		return err
	}
	err = p.c.db.UpdateBadGayStatusInPoolDB(dealInfo.Deal.Id.Unwrap().Int64(), int64(BanStatusWORKERINPOOL), time.Now())
	if err != nil {
		return err
	}

	p.c.logger.Info("destroyed deal", zap.String("deal", dealInfo.Deal.Id.Unwrap().String()))
	return nil
}
