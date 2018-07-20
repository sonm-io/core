package connor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

/* This constant is used when nanopool.org begins to give average hashrate instead of reported hashrate.
Now for this worker the value is average hashrate only. */
const iterationForAccessAvg = 4

//Tracking hashrate with using Connor's blacklist. Get data for 1 hour and another time => Detecting deviation.
func (p *PoolModule) AdvancedPoolHashrateTracking(ctx context.Context, reportedPool watchers.PoolWatcher, avgPool watchers.PoolWatcher) error {
	if err := p.UpdatePoolData(ctx, reportedPool, p.c.cfg.Pool.PoolAccount, int64(PoolTypeReportedHashrate)); err != nil {
		return err
	}
	if err := p.UpdatePoolData(ctx, avgPool, p.c.cfg.Pool.PoolAccount+"/1", int64(PoolTypeAvgHashrate)); err != nil {
		return err
	}

	workers, err := p.c.db.GetWorkersFromDB()
	if err != nil {
		p.c.logger.Error("cannot get worker from pool DB", zap.Error(err))
		return err
	}

	for _, w := range workers {
		if w.BadGuy > numberOfLives {
			continue
		}
		if w.Iterations == 0 {
			newIteration := w.Iterations + 1

			if err := p.c.db.UpdateIterationPoolDB(newIteration, w.DealID); err != nil {
				return err
			}
			continue
		}

		p.c.logger.Info("update iteration worker less than 4", zap.Int64("worker", w.DealID))

		newIteration := w.Iterations + 1
		if err := p.c.db.UpdateIterationPoolDB(newIteration, w.DealID); err != nil {
			return err
		}

		dealInfo, err := p.c.DealClient.Status(ctx, sonm.NewBigInt(big.NewInt(0).SetInt64(w.DealID)))
		if err != nil {
			return fmt.Errorf("cannot get deal from market: %v", err)
		}

		dealHashrate, err := p.ReturnBidHashrateForDeal(ctx, dealInfo)
		if err != nil {
			return err
		}

		if w.Iterations < numberOfIterationForRH {

			if w.WorkerReportedHashrate == 0.0 {

				if w.WorkerAvgHashrate == 0.0 {
					p.c.logger.Info("worker reported and average hashrate = 0 => send to Connor's blacklist", zap.Int64("worker_ID :", w.DealID))
					if err := p.SendToConnorBlackList(ctx, dealInfo); err != nil {
						return err
					}
					continue
				} else {
					if err := p.c.db.UpdateIterationPoolDB(iterationForAccessAvg, w.DealID); err != nil {
						return err
					}
					continue
				}
			}

			workerReportedHashrate := uint64(w.WorkerReportedHashrate * hashes)
			p.c.logger.Info("update reported hashrate", zap.Uint64("reported_hashrate", workerReportedHashrate), zap.Int64("worker", w.DealID))

			if err := p.ComparisonWithDealHashrate(ctx, workerReportedHashrate, dealHashrate, w, dealInfo); err != nil {
				return err
			}
		} else {
			if w.WorkerAvgHashrate == 0.0 {
				p.c.logger.Info("worker average hashrate = 0 send to connor's blacklist", zap.Int64("worker_ID :", w.DealID))
				if err := p.SendToConnorBlackList(ctx, dealInfo); err != nil {
					return err
				}
				continue
			}

			workerAvgHashrate := uint64(w.WorkerAvgHashrate * hashes)
			p.c.logger.Info("update average hashrate", zap.Uint64("average_hashrate", workerAvgHashrate), zap.Int64("worker", w.DealID))

			if err := p.ComparisonWithDealHashrate(ctx, workerAvgHashrate, dealHashrate, w, dealInfo); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *PoolModule) ComparisonWithDealHashrate(ctx context.Context, workerHashrate uint64, dealHashrate uint64, worker *database.PoolDB, dealInfo *sonm.DealInfoReply) error {
	if workerHashrate > dealHashrate {
		p.c.logger.Info("change response hashrate worker > deal hashrate. It's ok.",
			zap.Int64("worker", worker.DealID), zap.Uint64("hashrate", workerHashrate), zap.Uint64("deal_hashrate", dealHashrate))
	} else {
		changeHashrate := float64(workerHashrate) / float64(dealHashrate)
		p.c.logger.Info("worker deviation", zap.Int64("iteration", worker.Iterations), zap.Float64("change_low_percent", changeHashrate),
			zap.Uint64("worker_hashrate", workerHashrate), zap.Uint64("deal_hashrate", dealHashrate))

		if err := p.AdvancedDetectingDeviation(ctx, changeHashrate, worker, dealInfo); err != nil {
			return err
		}
	}
	return nil
}

//Detects the percentage of deviation of the hashrate and save SupplierID (by MasterID) to Connor's blacklist .
func (p *PoolModule) AdvancedDetectingDeviation(ctx context.Context, changePercentDeviationWorker float64, worker *database.PoolDB, dealInfo *sonm.DealInfoReply) error {

	if changePercentDeviationWorker < maximumDeviationOfWorker {
		p.c.logger.Info("worker gives too low hashrate. Send worker (supplier ID) to Connor's blacklist", zap.Int64("worker", worker.DealID),
			zap.Float64("deviation_percent", changePercentDeviationWorker))

		if err := p.FinishDealWithBlacklist(ctx, dealInfo); err != nil {
			return err
		}

		if err := p.c.db.UpdateBadGayStatusInPoolDB(worker.DealID, int64(BanStatusWorkerInPool), time.Now()); err != nil {
			return err
		}

	} else if changePercentDeviationWorker < p.c.cfg.Pool.WorkerLimitChangePercent {
		p.c.logger.Info("send SupplierID to Connor's blacklist", zap.Int64("worker", worker.DealID))

		if err := p.SendToConnorBlackList(ctx, dealInfo); err != nil {
			return err
		}
	}
	return nil
}

// Send to Connor's blacklist failed worker. If percent of failed workers more than "cleaner" workers => send Master to blacklist and destroy deal.
func (p *PoolModule) SendToConnorBlackList(ctx context.Context, failedDeal *sonm.DealInfoReply) error {
	workerList, err := p.c.MasterClient.WorkersList(ctx, failedDeal.Deal.MasterID)
	if err != nil {
		return err
	}

	for _, wM := range workerList.Workers {
		val, err := p.c.db.GetFailSupplierFromBlacklistDb(wM.SlaveID.Unwrap().Hex())
		if err != nil {
			return err
		}

		if val == wM.SlaveID.Unwrap().Hex() {
			continue
		} else {
			if err := p.c.db.SaveBlacklistIntoDB(&database.BlackListDb{
				MasterID:       wM.MasterID.Unwrap().Hex(),
				FailSupplierId: wM.SlaveID.Unwrap().Hex(),
				BanStatus:      int64(BanStatusBanned),
				DealID:         failedDeal.Deal.Id.Unwrap().Int64(),
			}); err != nil {
				return err
			}
		}
	}

	amountFailWorkers, err := p.c.db.GetCountFailSupplierFromDb(failedDeal.Deal.MasterID.String())
	if err != nil {
		return err
	}

	if amountFailWorkers == 0 {
		p.c.logger.Info("amount failed workers is 0")
		return fmt.Errorf("no failed workers in Blacklist")
	}

	clearWorkers := int64(len(workerList.Workers)) - amountFailWorkers
	percentFailWorkers := float64(amountFailWorkers / int64(len(workerList.Workers)))

	p.c.logger.Info("check failed workers in master", zap.String("worker_ID", failedDeal.Deal.MasterID.String()),
		zap.String("deal", failedDeal.Deal.Id.String()), zap.Float64("percent_failed", percentFailWorkers),
		zap.Int64("amount_failed", amountFailWorkers), zap.Int64("clear_workers", clearWorkers),
	)

	if percentFailWorkers > p.c.cfg.Pool.BadWorkersPercent {
		p.c.logger.Info("the deal destroyed due to the excessive number of banned workers in master", zap.String("deal", failedDeal.Deal.Id.Unwrap().String()),
			zap.Float64("percent_failed_workers", percentFailWorkers), zap.String("Master_ID", failedDeal.Deal.MasterID.String()),
		)
		if err := p.FinishDealWithBlacklist(ctx, failedDeal); err != nil {
			return err
		}
		if err := p.c.db.UpdateBanStatusBlackListDB(failedDeal.Deal.MasterID.Unwrap().Hex(), int64(BanStatusMasterBan)); err != nil {
			return err
		}
		if err := p.c.db.UpdateBadGayStatusInPoolDB(failedDeal.Deal.Id.Unwrap().Int64(), int64(BanStatusWorkerInPool), time.Now()); err != nil {
			return err
		}
	}
	return nil
}
