package connor

import (
	"context"
	"fmt"
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

func (e *engine) CreateOrder(bid *Corder, reason string) {
	e.log.Debug("creating order", zap.String("reason", reason))
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

// todo: restore tasks

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
			e.CreateOrder(bid, "cannot place initial order")
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
		case <-t.C:
			log.Debug("checking for deal for order")

			ord, err := e.market.GetOrderByID(e.ctx, &sonm.ID{Id: id})
			if err != nil {
				log.Warn("cannot get order info from market", zap.Error(err))
				continue
			}

			if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
				log.Info("order becomes inactive, looking for related deal")

				if ord.GetDealID() == nil {
					log.Debug("order have no deal, probably order is cancelled by hand")
					e.CreateOrder(order, "order have no deal, probably closed by hand")
					return
				}

				deal, err := e.deals.Status(e.ctx, ord.GetDealID())
				if err != nil {
					log.Warn("cannot get deal info from market", zap.Error(err),
						zap.String("deal_id", ord.GetDealID().Unwrap().String()))
					continue
				}

				e.CreateOrder(order, "order is turned into deal")
				e.processDeal(deal.GetDeal())
				return
			}

			log.Debug("order still have no deal")
		}
	}
}

func (e *engine) processDeal(deal *sonm.Deal) {
	dealID := deal.GetId().Unwrap().String()
	log := e.log.Named("process-deal").With(zap.String("deal_id", dealID))

	log.Debug("start deal processing")
	defer log.Debug("stop deal processing")
	defer e.finishDeal(deal.GetId())

	// TODO(sshaman1101): check for deal status
	taskReply, err := e.startTaskWithRetry(log, deal)
	if err != nil {
		log.Warn("cannot start task", zap.Error(err))
		return
	}

	log.Info("task started")
	err = e.trackTaskWithRetry(log, deal.GetId(), taskReply.GetId())
	if err != nil {
		log.Warn("task tracking failed", zap.Error(err))
	}
}

func (e *engine) startTaskWithRetry(log *zap.Logger, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	// todo: move retry settings to cfg
	try := 0
	deadline := 5
	t := util.NewImmediateTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
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
	// todo: configure timeout
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	dealReply, err := e.deals.Status(ctx, dealID)
	if err != nil {
		return nil, fmt.Errorf("cannot get deal status: %v", err)
	}

	if dealReply.GetResources() == nil {
		return nil, fmt.Errorf("cannot connect to worker: no resources info into deal status reply")
	}

	log.Debug("successfully obtained resources from worker")
	taskReply, err := e.tasks.Start(ctx, &sonm.StartTaskRequest{
		DealID: dealID,
		Spec: &sonm.TaskSpec{
			Container: &sonm.Container{
				Image: e.miningCfg.Image,
				Env: map[string]string{
					"WALLET":    e.miningCfg.Wallet.Hex(),
					"POOL_ADDR": e.miningCfg.PoolReportURL,
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
	// todo: move to config
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if try > deadline {
				return fmt.Errorf("task tracking failed: retry count exceeded")
			}

			shouldRetry, err := e.trackTaskOnce(log, dealID, taskID)
			if err != nil {
				if shouldRetry {
					log.Warn("cannot get task status, increasing retry counter", zap.Error(err), zap.Int("try", try))
					try++
				} else {
					return err
				}
			}
			log.Debug("task tracking OK, resetting failure counter")
			try = 0
		}
	}
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

func (e *engine) finishDeal(id *sonm.BigInt) {
	// todo: how to decide that we should add worker to blacklist?

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
