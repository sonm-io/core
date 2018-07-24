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
		case <-t.C:
			log.Info("checking for deal with order")

			ord, err := e.market.GetOrderByID(e.ctx, &sonm.ID{Id: id})
			if err != nil {
				log.Warn("cannot get order info from market", zap.Error(err))
				continue
			}

			if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
				log.Info("order becomes inactive, looking for related deal")

				// todo: check that order have a deal
				deal, err := e.deals.Status(e.ctx, ord.GetDealID())
				if err != nil {
					log.Warn("cannot get deal info from market", zap.Error(err),
						zap.String("deal_id", ord.GetDealID().Unwrap().String()))
					continue
				}

				e.CreateOrder(order)
				e.processDeal(deal.GetDeal())
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
	taskReply, err := e.startTask(log, deal)
	if err != nil {
		// log
		return
	}

	log.Info("task started")
	err = e.trackTask(log, deal.GetId(), taskReply.GetId())
	if err != nil {
		log.Warn("task tracking failed", zap.Error(err))
	}
}

func (e *engine) startTask(log *zap.Logger, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	var taskReply *sonm.StartTaskReply

	// todo: move retry settings to cfg
	for try := 0; try < 5; try++ {
		if try > 0 {
			time.Sleep(10 * time.Second)
		}

		// todo: ctx with timeout?
		dealReply, err := e.deals.Status(e.ctx, deal.GetId())
		if err != nil || dealReply.GetResources() == nil {
			// todo: separate into two checks with different error messages
			log.Warn("cannot connect to worker", zap.Error(err), zap.Int("try", try))
			continue
		}

		log.Debug("successfully obtained resources from worker", zap.Any("res", *dealReply.GetResources()))

		// 2. start task
		taskReply, err = e.tasks.Start(e.ctx, &sonm.StartTaskRequest{
			DealID: deal.GetId(),
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
			log.Warn("cannot start task", zap.Error(err), zap.Int("try", try))
			continue
		}

		break
	}

	if taskReply == nil {
		return nil, fmt.Errorf("cannot start task: retry count exceeded")
	}

	return taskReply, nil
}

func (e *engine) trackTask(log *zap.Logger, dealID *sonm.BigInt, taskID string) error {
	log = log.Named("task").With(zap.String("task_id", taskID))
	log.Info("start task status tracking")

	// todo: move to config
	// todo: ticker with retry counter
	for try := 0; try < 5; try++ {
		if try > 0 {
			time.Sleep(10 * time.Second)
		}

		log.Debug("checking task status", zap.Int("try", try))

		// 3. ping task
		status, err := e.tasks.Status(e.ctx, &sonm.TaskID{Id: taskID, DealID: dealID})
		if err != nil {
			log.Warn("cannot get task status, increasing retry counter", zap.Error(err))
			continue
		}

		if status.GetStatus() == sonm.TaskStatusReply_FINISHED || status.GetStatus() == sonm.TaskStatusReply_BROKEN {
			log.Warn("task is failed by unknown reasons, finishing deal",
				zap.String("status", status.GetStatus().String()))
			return fmt.Errorf("task is finished by unknown reasons")
		}

		try = 0
		log.Debug("task status OK, resetting retry counter")
	}

	return fmt.Errorf("task tracking failed: retry count exceeded")
}

func (e *engine) finishDeal(id *sonm.BigInt) {
	// todo: how to decide that we should add worker to blacklist?
	if _, err := e.deals.Finish(e.ctx, &sonm.DealFinishRequest{Id: id}); err != nil {
		e.log.Warn("cannot finish deal", zap.Error(err), zap.String("id", id.Unwrap().String()))
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
