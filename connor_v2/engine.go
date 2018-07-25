package connor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const concurrency = 10

type engine struct {
	log       *zap.Logger
	ctx       context.Context
	cfg       engineConfig
	miningCfg miningConfig

	market sonm.MarketClient
	deals  sonm.DealManagementClient
	tasks  sonm.TaskManagementClient

	ordersCreateChan  chan *Corder
	ordersResultsChan chan *Corder
}

func NewEngine(ctx context.Context, cfg engineConfig, miningCfg miningConfig, log *zap.Logger, cc *grpc.ClientConn) *engine {
	return &engine{
		ctx:               ctx,
		cfg:               cfg,
		miningCfg:         miningCfg,
		log:               log.Named("engine"),
		market:            sonm.NewMarketClient(cc),
		deals:             sonm.NewDealManagementClient(cc),
		tasks:             sonm.NewTaskManagementClient(cc),
		ordersCreateChan:  make(chan *Corder, concurrency),
		ordersResultsChan: make(chan *Corder, concurrency),
	}
}

func (e *engine) CreateOrder(bid *Corder) {
	e.ordersCreateChan <- bid
}

func (e *engine) RestoreOrder(order *Corder) {
	e.log.Debug("restoring order", zap.String("id", order.Order.GetId().Unwrap().String()))
	e.ordersResultsChan <- order
}

func (e *engine) RestoreDeal(deal *sonm.Deal) {
	e.log.Debug("restoring deal", zap.String("id", deal.GetId().Unwrap().String()))
	go e.processDeal(deal)
}

func (e *engine) sendOrderToMarket(bid *sonm.BidOrder) (*sonm.Order, error) {
	e.log.Debug("creating order on market",
		zap.String("price", bid.GetPrice().GetPerSecond().Unwrap().String()),
		zap.Any("benchmarks", bid.Resources.GetBenchmarks()))

	return e.market.CreateOrder(e.ctx, bid)
}

func (e *engine) processOrderCreate() {
	for bid := range e.ordersCreateChan {
		created, err := e.sendOrderToMarket(bid.AsBID())
		if err != nil {
			e.log.Warn("cannot place order, retrying", zap.Error(err))
			e.CreateOrder(bid)
			continue
		}

		e.ordersResultsChan <- NewCorderFromOrder(created, bid.token)
	}
}

func (e *engine) processOrderResult() {
	for order := range e.ordersResultsChan {
		go e.waitForDeal(order)
	}
}

func (e *engine) waitForDeal(order *Corder) {
	id := order.GetId().Unwrap().String()
	log := e.log.With(zap.String("order_id", id))

	t := util.NewImmediateTicker(e.cfg.OrderWatchInterval)
	defer t.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-t.C:
			log.Debug("checking for deal for order")

			deal, err := e.checkOrderForDealOnce(log, id)
			if err != nil {
				continue
			}

			e.CreateOrder(order)
			if deal != nil {
				e.processDeal(deal)
			}

			return
		}
	}
}

func (e *engine) checkOrderForDealOnce(log *zap.Logger, orderID string) (*sonm.Deal, error) {
	ord, err := e.market.GetOrderByID(e.ctx, &sonm.ID{Id: orderID})
	if err != nil {
		log.Warn("cannot get order info from market", zap.Error(err))
		return nil, err
	}

	if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
		log.Info("order becomes inactive, looking for related deal")

		if ord.GetDealID() == nil || ord.GetDealID().IsZero() {
			log.Debug("order have no deal, probably order is cancelled by hand")
			return nil, nil
		}

		deal, err := e.deals.Status(e.ctx, ord.GetDealID())
		if err != nil {
			log.Warn("cannot get deal info from market", zap.Error(err),
				zap.String("deal_id", ord.GetDealID().Unwrap().String()))
			return nil, err
		}

		return deal.GetDeal(), nil
	}

	return nil, fmt.Errorf("order have no deal")
}

func (e *engine) processDeal(deal *sonm.Deal) {
	dealID := deal.GetId().Unwrap().String()
	log := e.log.Named("process-deal").With(zap.String("deal_id", dealID))

	log.Debug("start deal processing")
	defer log.Debug("stop deal processing")
	defer e.finishDeal(deal.GetId())

	taskID, err := e.restoreTasks(log, deal.GetId())
	if err != nil {
		log.Warn("cannot restore tasks", zap.Error(err))
		return
	}

	if len(taskID) == 0 {
		log.Debug("no tasks restored, starting new one")

		taskReply, err := e.startTaskWithRetry(log, deal)
		if err != nil {
			log.Warn("cannot start task", zap.Error(err))
			return
		}

		taskID = taskReply.GetId()
		log.Info("task started", zap.String("task_id", taskID))
	}

	err = e.trackTaskWithRetry(log, deal.GetId(), taskID)
	if err != nil {
		log.Warn("task tracking failed", zap.Error(err))
	}
}

func (e *engine) startTaskWithRetry(log *zap.Logger, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	try := 0
	deadline := 5

	t := util.NewImmediateTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return nil, e.ctx.Err()
		case <-t.C:
			if try > deadline {
				return nil, fmt.Errorf("cannot start task: retry count exceeded")
			}
			try++

			taskReply, err := e.startTaskOnce(log, deal.GetId())
			if err != nil {
				log.Warn("task start failed", zap.Error(err), zap.Int("try", try))
				continue
			}

			return taskReply, nil
		}
	}
}

func (e *engine) startTaskOnce(log *zap.Logger, dealID *sonm.BigInt) (*sonm.StartTaskReply, error) {
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	taskReply, err := e.tasks.Start(ctx, &sonm.StartTaskRequest{
		DealID: dealID,
		Spec: &sonm.TaskSpec{
			Container: &sonm.Container{
				Image: e.miningCfg.Image,
				Env: map[string]string{
					"WALLET":    e.miningCfg.Wallet.Hex(),
					"POOL_ADDR": e.miningCfg.PoolReportURL,
					"WORKER":    "CONNOR_" + dealID.String(),
				},
			},
			Resources: &sonm.AskPlanResources{},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start task on worker: %v", err)
	}

	return taskReply, nil
}

func (e *engine) trackTaskWithRetry(log *zap.Logger, dealID *sonm.BigInt, taskID string) error {
	log = log.Named("task").With(zap.String("task_id", taskID))
	log.Info("start task status tracking")

	try := 0
	deadline := 5

	t := time.NewTicker(15 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return e.ctx.Err()
		case <-t.C:
			if try > deadline {
				return fmt.Errorf("task tracking failed: retry count exceeded")
			}

			ok, err := e.checkDealStatus(log, dealID)
			if err != nil {
				try++
				log.Warn("cannot check deal status, increasing retry counter", zap.Error(err), zap.Int("try", try))
				continue
			}

			if !ok {
				log.Warn("deal is closed, finishing tracking")
				return fmt.Errorf("deal is closed")
			}

			log.Debug("deal status OK, checking tasks")
			shouldRetry, err := e.trackTaskOnce(log, dealID, taskID)
			if err != nil {
				if !shouldRetry {
					return err
				}

				try++
				log.Warn("cannot get task status, increasing retry counter", zap.Error(err), zap.Int("try", try))
				continue
			}

			log.Debug("task tracking OK, resetting failure counter")
			try = 0
		}
	}
}

func (e *engine) checkDealStatus(log *zap.Logger, dealID *sonm.BigInt) (bool, error) {
	ctx, cancel := context.WithTimeout(e.ctx, 15*time.Second)
	defer cancel()

	dealStatus, err := e.deals.Status(ctx, dealID)
	if err != nil {
		return false, err
	}

	return dealStatus.GetDeal().GetStatus() == sonm.DealStatus_DEAL_ACCEPTED, nil
}

func (e *engine) trackTaskOnce(log *zap.Logger, dealID *sonm.BigInt, taskID string) (bool, error) {
	log.Debug("checking task status")

	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	// 3. ping task
	status, err := e.tasks.Status(ctx, &sonm.TaskID{Id: taskID, DealID: dealID})
	if err != nil {
		return true, err
	}

	if status.GetStatus() == sonm.TaskStatusReply_FINISHED || status.GetStatus() == sonm.TaskStatusReply_BROKEN {
		log.Warn("task is failed by unknown reasons, finishing deal",
			zap.String("status", status.GetStatus().String()))
		return false, fmt.Errorf("task is finished by unknown reasons")
	}

	log.Debug("task status is OK")
	return true, nil
}

func (e *engine) restoreTasks(log *zap.Logger, dealID *sonm.BigInt) (string, error) {
	log = log.Named("restore-tasks")
	log.Debug("restoring tasks")

	t := util.NewImmediateTicker(10 * time.Second)
	defer t.Stop()

	try := 0
	deadline := 5

	for {
		select {
		case <-e.ctx.Done():
			return "", e.ctx.Err()
		case <-t.C:
			if try > deadline {
				return "", fmt.Errorf("restore tasks failed: retry count exceeded")
			}

			list, err := e.loadTasksOnce(log, dealID)
			if err != nil {
				try++
				log.Warn("cannot obtain task list from worker", zap.Error(err))
				continue
			}

			switch len(list) {
			case 0:
				return "", nil
			case 1:
				// todo: find proper way to mark tasks
				if list[0].ImageName != e.miningCfg.Image {
					log.Warn("unexpected docker image is running on task", zap.String("running", list[0].ImageName), zap.String("expected", e.miningCfg.Image))
					e.stopOneTask(log, dealID, list[0].id)
					return "", nil
				}

				return list[0].id, nil
			default:
				// weird case, we always starting only one task per deal
				log.Info("worker have more than one task running", zap.Int("count", len(list)))
				if err := e.stopAllTasks(log, list, dealID); err != nil {
					return "", err
				}

				return "", nil
			}
		}
	}
}

func (e *engine) loadTasksOnce(log *zap.Logger, dealID *sonm.BigInt) ([]*taskStatus, error) {
	log.Debug("loading tasks from worker")

	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	taskList, err := e.tasks.List(ctx, &sonm.TaskListRequest{DealID: dealID})
	if err != nil {
		return nil, err
	}

	list := make([]*taskStatus, 0)
	for id, task := range taskList.GetInfo() {
		if task.GetStatus() == sonm.TaskStatusReply_RUNNING {
			list = append(list, &taskStatus{task, id})
		}
	}

	return list, nil
}

func (e *engine) stopAllTasks(log *zap.Logger, list []*taskStatus, dealID *sonm.BigInt) error {
	log.Debug("stopping all tasks on worker")

	var failedIDs []string
	for _, task := range list {
		if err := e.stopOneTask(log, dealID, task.id); err != nil {
			log.Warn("cannot stop task", zap.Error(err), zap.String("task_id", task.id))
			failedIDs = append(failedIDs, task.id)
		}
	}

	if len(failedIDs) > 0 {
		return fmt.Errorf("cannot stop tasks ids = %s", strings.Join(failedIDs, ","))
	}

	return nil
}

func (e *engine) stopOneTask(log *zap.Logger, dealID *sonm.BigInt, taskID string) error {
	log.Debug("stopping task", zap.String("task_id", taskID))

	for try := 0; try < 5; try++ {
		ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
		if _, err := e.tasks.Stop(ctx, &sonm.TaskID{Id: taskID, DealID: dealID}); err != nil {
			log.Warn("cannot stop task", zap.Error(err), zap.Int("try", try))
			cancel()
			continue
		}

		cancel()
		return nil
	}

	return fmt.Errorf("cannot stop task: retry count exceeded")
}

func (e *engine) finishDeal(id *sonm.BigInt) {
	// todo: how to decide that we should add worker to blacklist?
	// todo: move this to the anti-fraud module

	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	if _, err := e.deals.Finish(ctx, &sonm.DealFinishRequest{Id: id}); err != nil {
		e.log.Warn("cannot finish deal", zap.Error(err), zap.String("id", id.Unwrap().String()))
		return
	}

	e.log.Info("deal finished", zap.String("id", id.Unwrap().String()))
}

func (e *engine) start(ctx context.Context) {
	go func() {
		defer close(e.ordersCreateChan)
		defer close(e.ordersResultsChan)

		e.log.Info("starting engine", zap.Int("concurrency", concurrency))
		defer e.log.Info("stopping engine")

		for i := 0; i < concurrency; i++ {
			go e.processOrderCreate()
		}

		go e.processOrderResult()

		<-ctx.Done()
	}()
}
