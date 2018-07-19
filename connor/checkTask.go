package connor

import (
	"context"
	"time"

	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	retryStatusChecks    = 3
	taskStatusRetryDelay = 60 * time.Second
)

// CheckTaskStatus check the status of the task to get rid of the loss of money (pool check).
func (p *PoolModule) CheckTaskStatus(ctx context.Context) error {
	deals, err := p.c.db.GetDealsFromDB()
	if err != nil {
		p.c.logger.Debug("cannot get deals from db", zap.Error(err))
		return err
	}

	group := errgroup.Group{}
	for _, d := range deals {
		if d.DeployStatus == DeployStatusDeployed && d.Status == int64(sonm.DealStatus_DEAL_ACCEPTED) {

			dealID := sonm.NewBigIntFromInt(d.DealID)

			checkDealStatus, err := p.c.DealClient.Status(ctx, dealID)
			if err != nil {
				p.c.logger.Debug("cannot get deal status", zap.Error(err), zap.Int64("deal_id", d.DealID))
				continue
			}

			switch checkDealStatus.Deal.Status {
			case sonm.DealStatus_DEAL_ACCEPTED:
				p.c.logger.Info("deal accepted, loading tasks list from a worker", zap.Int64("deal", d.DealID))

				tasksList, err := p.c.TaskClient.List(ctx, &sonm.TaskListRequest{DealID: dealID})
				if err != nil {
					p.c.logger.Warn("cannot get tasks from worker", zap.Error(err))
					continue
				}

				for taskID := range tasksList.GetInfo() {
					taskStatus, err := p.c.TaskClient.Status(ctx, &sonm.TaskID{
						Id:     taskID,
						DealID: dealID,
					})

					if err != nil {
						p.c.logger.Debug("cannot get task status from worker, retrying")
						group.Go(func() error {
							if err := p.RetryCheckTaskStatus(ctx, *d, taskID); err != nil {
								p.c.logger.Debug("cannot get task status after retrying")
							}
							return nil
						})

						continue
					}
					if err := p.CheckFatalTaskStatus(ctx, d, taskStatus); err != nil {
						p.c.logger.Debug("cannot close deal (via CheckFatalTaskStatus)", zap.Error(err))
						continue
					}
				}

			case sonm.DealStatus_DEAL_CLOSED:
				if err := p.c.db.UpdateDeployAndDealStatusDB(d.DealID, DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
					p.c.logger.Info("cannot save deal status into db", zap.Error(err))
					continue
				}

				p.c.logger.Info("deal closed on market, task tracking stop")
			}
		}
	}

	return group.Wait()
}

func (p *PoolModule) RetryCheckTaskStatus(ctx context.Context, deal database.DealDB, taskID string) error {
	p.c.logger.Debug("retrying check status", zap.Any("deal", deal))
	for i := 0; i < retryStatusChecks; i++ {
		time.Sleep(taskStatusRetryDelay)
		p.c.logger.Debug("try to check deal status", zap.Int("try", i))

		taskStatus, err := p.c.TaskClient.Status(ctx, &sonm.TaskID{
			Id:     taskID,
			DealID: sonm.NewBigIntFromInt(deal.DealID),
		})

		if err != nil {
			p.c.logger.Debug("failed to check task status", zap.Error(err))
			continue
		}
		return p.CheckFatalTaskStatus(ctx, &deal, taskStatus)
	}

	dealOnMarket, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(deal.DealID))
	if err != nil {
		return err
	}

	// this condition is necessary if the worker prematurely completes the transaction
	if dealOnMarket.Deal.Status != sonm.DealStatus_DEAL_CLOSED {
		if err := p.FinishDeal(ctx, &deal); err != nil {
			p.c.logger.Warn("cannot finish deal", zap.Error(err))
		}
	} else {
		p.c.logger.Info("the deal has already been closed by worker",
			zap.Int64("worker", deal.DealID))

		if err := p.UpdateDestroyedDealDB(ctx, &deal); err != nil {
			return err
		}
	}
	return nil
}
