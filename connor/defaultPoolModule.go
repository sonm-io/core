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
	ZecPool                  = "stratum+tcp://zec-eu1.nanopool.org:6666"
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

const (
	PoolTypeReportedHashrate = 0
	PoolTypeAvgHashrate      = 1
)

const (
	BanStatusBanned       = 1
	BanStatusMasterBan    = 2
	BanStatusWorkerInPool = 6
)

func (p *PoolModule) DeployNewContainer(ctx context.Context, deal *sonm.Deal, image string) (*sonm.StartTaskReply, error) {
	env := map[string]string{
		"WORKER": deal.Id.String(),
		"WALLET": p.c.cfg.Pool.PoolAccount,
		"EMAIL":  p.c.cfg.Pool.EmailForPool,
	}

	switch p.c.cfg.UsingToken {
	case "ETH":
		env["ETH_POOL"] = EthPool
	case "ZEC":
		env["ZEC_POOL"] = ZecPool
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
		switch p.c.cfg.UsingToken {
		case "ETH":
			p.c.logger.Info("ETH container wasn't deployed", zap.String("deal_ID", deal.GetID().Unwrap().String()))
			if err := p.FinishDealAfterFailedStartTask(ctx, deal); err != nil {
				return nil, err
			}
		case "ZEC":
			reply, err := p.TryCreateAMDContainer(ctx, deal)
			if err != nil {
				p.c.logger.Info("AMD container for ZEC wasn't deployed", zap.String("deal_ID", deal.GetID().Unwrap().String()))
				if err := p.FinishDealAfterFailedStartTask(ctx, deal); err != nil {
					return nil, err
				}
			}

			p.c.logger.Info("AMD container was deployed successfully", zap.String("deal_id", deal.GetID().Unwrap().String()),
				zap.String("worker_zec", deal.GetSupplierID().Unwrap().Hex()))

			return reply, nil
		}
	}
	return reply, nil
}

// FinishDealAfterFailedStartTask using for both tasks
func (p *PoolModule) FinishDealAfterFailedStartTask(ctx context.Context, deal *sonm.Deal) error {
	if err := p.c.db.UpdateDeployAndDealStatusDB(deal.Id.Unwrap().Int64(), DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
		return err
	}

	dealStatus, err := p.c.DealClient.Status(ctx, deal.Id)
	if err != nil {
		return fmt.Errorf("cannot get deal status %v", err)
	}

	switch dealStatus.Deal.Status {
	case sonm.DealStatus_DEAL_ACCEPTED:
		p.c.logger.Info("failed start task. Deal finished", zap.Int64("deal_ID", deal.Id.Unwrap().Int64()))
		_, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{Id: deal.Id})
		if err != nil {
			p.c.logger.Warn("failed finish deal", zap.Error(err))
		}
	case sonm.DealStatus_DEAL_CLOSED:
		p.c.logger.Info("deal already finished from worker", zap.Int64("deal_ID", deal.Id.Unwrap().Int64()))
	}
	return nil
}

func (p *PoolModule) TryCreateAMDContainer(ctx context.Context, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	imageAMD := "sonm/zcash-amd:latest" // for AMD

	p.c.logger.Info("processing of deploy new AMD container", zap.Any("deal_ID", deal), zap.String("image_AMD", imageAMD))
	env := map[string]string{
		"ZEC_POOL": ZecPool,
		"WORKER":   deal.Id.String(),
		"WALLET":   p.c.cfg.Pool.PoolAccount,
		"EMAIL":    p.c.cfg.Pool.EmailForPool,
	}
	container := &sonm.Container{
		Image: imageAMD,
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
	if err != nil {
		p.c.logger.Warn("cannot start task for AMD and NVIDIA containers", zap.Error(err))
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
		//пытаюсь сделать на AMD
		deal, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(d.DealID))
		if err != nil {
			p.c.logger.Error("cannot get deal from market %v")
			return err
		}

		reply, err := p.TryCreateAMDContainer(ctx, deal.Deal)
		if err != nil {
			p.c.logger.Warn("cannot create AMD Container. Deal finished", zap.Error(err))
			p.c.logger.Info("task status is broken or finished, closing deal", zap.Int64("deal_id", d.DealID),
				zap.String("task_status", taskStatus.Status.String()))
			return p.FinishDeal(ctx, d)
		}
		p.c.logger.Info("Replying AMD container successful", zap.String("start_task_reply", reply.Id))
	}
	return nil
}

func (p *PoolModule) FinishDeal(ctx context.Context, deal *database.DealDB) error {
	if _, err := p.c.DealClient.Finish(ctx, &sonm.DealFinishRequest{Id: sonm.NewBigIntFromInt(deal.DealID)}); err != nil {
		return err
	}

	return p.UpdateDestroyedDealDB(ctx, deal)
}

// Create deal finish request
func (p *PoolModule) FinishDealWithBlacklist(ctx context.Context, dealInfo *sonm.DealInfoReply) error {

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
			if err := p.FinishDealWithBlacklist(ctx, dealInfo); err != nil {
				return err
			}
			err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWorkerInPool), time.Now())
			if err != nil {
				return err
			}
			p.c.logger.Info("destroy deal", zap.String("bad_status_in_pool", dealInfo.Deal.Id.String()))
		}
	} else if changePercentDeviationWorker < maximumDeviationOfWorker {
		err := p.FinishDealWithBlacklist(ctx, dealInfo)
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
