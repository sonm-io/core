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

// check the status of the task to get rid of the loss of money (pool check).
func (p *PoolModule) CheckTaskStatus(ctx context.Context) error {
	dealsDb, err := p.c.db.GetDealsFromDB()
	if err != nil {
		return err
	}
	group := errgroup.Group{}
	for _, d := range dealsDb {
		if d.DeployStatus == DeployStatusDeployed && d.Status == int64(sonm.DealStatus_DEAL_ACCEPTED) {
			checkDealStatus, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(d.DealID))
			if err != nil {
				return nil
			}
			p.c.logger.Info("deployed and accepted in DB!")

			switch checkDealStatus.Deal.Status {
			case sonm.DealStatus_DEAL_ACCEPTED:

				p.c.logger.Info("check task status ", zap.Int64("deal", d.DealID))

				tasksList, err := p.c.TaskClient.List(ctx, &sonm.TaskListRequest{
					DealID: sonm.NewBigIntFromInt(d.DealID),
				})
				if err != nil {
					return err
				}

				dealOnMarket, err := p.c.DealClient.Status(ctx, sonm.NewBigIntFromInt(d.DealID))
				if err != nil {
					return err
				}

				for taskID := range tasksList.GetInfo() {

					taskStatus, err := p.c.TaskClient.Status(ctx, &sonm.TaskID{
						Id:     taskID,
						DealID: dealOnMarket.Deal.Id,
					})

					if err != nil {
						p.c.logger.Info("cannot get tasksList status from worker ==> retry")
						group.Go(func() error {
							return p.RetryCheckTaskStatus(ctx, *d, taskID)
						})
						continue
					}
					err = p.CheckFatalTaskStatus(ctx, d, taskStatus)
					if err != nil {
						return err
					}
				}

			case sonm.DealStatus_DEAL_CLOSED:
				if err = p.c.db.UpdateDeployAndDealStatusDB(d.DealID, DeployStatusDestroyed, sonm.DealStatus_DEAL_CLOSED); err != nil {
					return err
				}
				p.c.logger.Info("deal closed on market, task tracking stop")
				continue
			}
		}
	}
	return group.Wait()
}
func (p *PoolModule) RetryCheckTaskStatus(ctx context.Context, deal database.DealDB, taskID string) error {
	p.c.logger.Debug("retrying check status", zap.Any("deal", deal))

	for i := 0; i < retryStatusChecks; i++ {
		time.Sleep(taskStatusRetryDelay)
		p.c.logger.Info("try to check deal status", zap.Int("try", i))

		taskStatus, err := p.c.TaskClient.Status(ctx, &sonm.TaskID{
			Id:     taskID,
			DealID: sonm.NewBigIntFromInt(deal.DealID),
		})
		if err != nil {
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
		if err := p.finishDeal(ctx, &deal, sonm.TaskStatusReply_BROKEN); err != nil {
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
