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
	EthPool                  = "stratum+tcp://eth-eu1.nanopool.org:9999"
	numberOfIterationForRH   = 5
	numberOfLives            = 5
	maximumDeviationOfWorker = 0.80
)

type PoolModule struct {
	c *Connor
}

func NewPoolModules(c *Connor) *PoolModule {
	return &PoolModule{
		c: c,
	}
}

type PoolType int32

const (
	PoolTypeReportedHashrate PoolType = 0
	PoolTypeAvgHashrate      PoolType = 1
)

type BanStatus int32

const (
	BanStatusBanned       BanStatus = 1
	BanStatusMasterBan    BanStatus = 2
	BanStatusWorkerInPool BanStatus = 6
)

func (p *PoolModule) DeployNewContainer(ctx context.Context, deal *sonm.Deal, image string) (*sonm.StartTaskReply, error) {
	env := map[string]string{
		"ETH_POOL": EthPool,
		"WORKER":   deal.Id.String(),
		"WALLET":   p.c.cfg.Pool.PoolAccount,
		"EMAIL":    p.c.cfg.Pool.EmailForPool,
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
		DealID: deal.GetID(),
		Spec:   spec,
	}

	reply, err := p.c.TaskClient.Start(ctx, startTaskRequest)
	// TODO(sshaman1101): retry on errors
	if err != nil {
		p.c.logger.Info("cannot start task on worker",
			zap.String("deal_id", deal.GetID().Unwrap().String()),
			zap.String("worker_eth", deal.GetSupplierID().Unwrap().Hex()))

		if err = p.c.db.UpdateDeployAndDealStatusDB(deal.Id.Unwrap().Int64(), DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
			return nil, err
		}

		dealStatus, err := p.c.DealClient.Status(ctx, deal.Id)
		if err != nil {
			return nil, err
		}

		switch dealStatus.Deal.Status {
		case sonm.DealStatus_DEAL_ACCEPTED:
			_, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{Id: deal.Id})
			if err != nil {
				return nil, fmt.Errorf("fail finish deal: %v", err)
			}
			p.c.logger.Info("fail start task deal finished", zap.Int64("deal", deal.Id.Unwrap().Int64()))

		case sonm.DealStatus_DEAL_CLOSED:
			p.c.logger.Info("deal already finished from worker", zap.Any("deal", deal))
		}
		return nil, err
	}
	return reply, nil
}

// Checks for a deal in the worker list. If it is not there, adds.
func (p *PoolModule) AddWorkerToPoolDB(ctx context.Context, deal *sonm.DealInfoReply, addr string) error {
	val, err := p.c.db.GetWorkerFromPoolDB(deal.Deal.Id.String())
	if err != nil {
		return err
	}
	if val == deal.Deal.Id.String() {
		return nil
	}
	if err := p.c.db.SavePoolIntoDB(&database.PoolDB{
		DealID:    deal.Deal.Id.Unwrap().Int64(),
		PoolID:    addr,
		TimeStart: time.Now()}); err != nil {
		return err
	}
	return nil
}

func (p *PoolModule) UpdatePoolData(ctx context.Context, pool watchers.PoolWatcher, addr string, typePool int64) error {
	if err := pool.Update(ctx); err != nil {
		return err
	}

	data, err := pool.GetData(addr)
	if err != nil {
		p.c.logger.Warn("cannot get data", zap.String("addr", addr), zap.Error(err))
		return err
	}

	for _, d := range data.Data {
		switch typePool {
		case int64(PoolTypeReportedHashrate):
			if err := p.c.db.UpdateReportedHashratePoolDB(d.Worker, d.Hashrate, time.Now()); err != nil {
				return err
			}
		case int64(PoolTypeAvgHashrate):
			if err := p.c.db.UpdateAvgPoolDB(d.Worker, d.Hashrate, time.Now()); err != nil {
				return err
			}
		}
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

// Check task status and make decision
func (p *PoolModule) CheckFatalTaskStatus(ctx context.Context, d *database.DealDB, taskStatus *sonm.TaskStatusReply) error {
	if taskStatus.Status == sonm.TaskStatusReply_BROKEN || taskStatus.Status == sonm.TaskStatusReply_FINISHED {
		return p.finishDeal(ctx, d, taskStatus.GetStatus())
	}
	return nil
}

func (p *PoolModule) finishDeal(ctx context.Context, deal *database.DealDB, status sonm.TaskStatusReply_Status) error {
	if _, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{Id: sonm.NewBigIntFromInt(deal.DealID)}); err != nil {
		return err
	}

	p.c.logger.Info("task status reply broken or finished => deal destroy", zap.Int64("deal", deal.DealID), zap.String("task", status.String()))

	return p.UpdateDestroyedDealDB(ctx, deal)
}

// Create deal finish request
func (p *PoolModule) DestroyDeal(ctx context.Context, dealInfo *sonm.DealInfoReply) error {

	if _, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{
		Id:            dealInfo.Deal.Id,
		BlacklistType: sonm.BlacklistType_BLACKLIST_MASTER,
	}); err != nil {
		p.c.logger.Info("couldn't finish deal", zap.Any("deal", dealInfo),
			zap.Error(err))
	}
	if err := p.c.db.UpdateDeployAndDealStatusDB(dealInfo.Deal.Id.Unwrap().Int64(), DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
		return err
	}

	err := p.c.db.UpdateBadGayStatusInPoolDB(dealInfo.Deal.Id.Unwrap().Int64(), int64(BanStatusWorkerInPool), time.Now())
	if err != nil {
		return err
	}

	p.c.logger.Info("destroyed deal", zap.String("deal", dealInfo.Deal.Id.Unwrap().String()))
	return nil
}

func (p *PoolModule) UpdateDestroyedDealDB(ctx context.Context, deal *database.DealDB) error {
	if err := p.c.db.UpdateDeployAndDealStatusDB(deal.DealID, DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
		return err
	}

	if err := p.c.db.UpdateBadGayStatusInPoolDB(deal.DealID, int64(BanStatusWorkerInPool), time.Now()); err != nil {
		return err
	}
	return nil
}

// Default hashrate tracking. Updates and evaluates hashrate by workers, depending on the iteration.
func (p *PoolModule) DefaultPoolHashrateTracking(ctx context.Context, reportedPool watchers.PoolWatcher, avgPool watchers.PoolWatcher) error {
	workers, err := p.c.db.GetWorkersFromDB()
	if err != nil {
		return fmt.Errorf("cannot get worker from pool DB: %v", err)
	}

	for _, w := range workers {
		if w.BadGuy > numberOfLives {
			continue
		}
		iteration := int64(w.Iterations + 1)

		dealInfo, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(w.DealID))
		if err != nil {
			return fmt.Errorf("cannot get deal %v from market: %v", w.DealID, err)
		}

		bidHashrate, err := p.ReturnBidHashrateForDeal(ctx, dealInfo)
		if err != nil {
			return err
		}

		if iteration < numberOfIterationForRH {
			if err = p.UpdatePoolData(ctx, reportedPool, p.c.cfg.Pool.PoolAccount, int64(PoolTypeReportedHashrate)); err != nil {
				return err
			}
			changePercentRHWorker := w.WorkerReportedHashrate * float64(hashes) / float64(bidHashrate)
			if err = p.DetectingDeviation(ctx, changePercentRHWorker, w, dealInfo); err != nil {
				return err
			}
		} else {
			err := p.UpdatePoolData(ctx, avgPool, p.c.cfg.Pool.PoolAccount+"/1", int64(PoolTypeAvgHashrate))
			if err != nil {
				return err
			}
			p.c.logger.Info("getting avg pool data for worker", zap.Int64("deal", w.DealID))
			changePercentAvgWorker := w.WorkerAvgHashrate * float64(hashes) / float64(bidHashrate)
			if err = p.DetectingDeviation(ctx, changePercentAvgWorker, w, dealInfo); err != nil {
				return err
			}
		}
		err = p.c.db.UpdateIterationPoolDB(iteration, w.DealID)
		if err != nil {
			return err
		}
	}
	return nil
}

//Detection of getting a lowered hashrate and sending to a blacklist (create deal finish request).
func (p *PoolModule) DetectingDeviation(ctx context.Context, changePercentDeviationWorker float64, worker *database.PoolDB, dealInfo *sonm.DealInfoReply) error {

	if changePercentDeviationWorker < p.c.cfg.Pool.WorkerLimitChangePercent {
		if worker.BadGuy < numberOfLives {
			newStatus := worker.BadGuy + 1
			err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, newStatus, time.Now())
			if err != nil {
				return err
			}
		} else {
			if err := p.DestroyDeal(ctx, dealInfo); err != nil {
				return err
			}
			err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWorkerInPool), time.Now())
			if err != nil {
				return err
			}
			p.c.logger.Info("destroy deal", zap.String("bad_status_in_pool", dealInfo.Deal.Id.String()))
		}
	} else if changePercentDeviationWorker < maximumDeviationOfWorker {
		err := p.DestroyDeal(ctx, dealInfo)
		if err != nil {
			return err
		}
		err = p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWorkerInPool), time.Now())
		if err != nil {
			return err
		}
		p.c.logger.Info("destroy deal", zap.String("bad_status_in_pool", dealInfo.Deal.Id.String()))
	}
	return nil
}
