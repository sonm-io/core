package connor

import (
	"context"
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

func (w *engine) CreateOrder(bid *Corder) {
	w.ordersCreateChan <- bid
}

func (w *engine) RestoreOrder(order *Corder) {
	w.log.Debug("restoring order", zap.String("id", order.Order.GetId().Unwrap().String()))
	w.ordersResultsChan <- order
}

func (w *engine) RestoreDeal(deal *sonm.Deal) {
	w.log.Debug("restoring deal", zap.String("id", deal.GetId().Unwrap().String()))
	go w.processDeal(deal)
}

func (w *engine) sendOrderToMarket(bid *sonm.BidOrder) (*sonm.Order, error) {
	w.log.Debug("creating order on market",
		zap.String("price", bid.GetPrice().GetPerSecond().Unwrap().String()),
		zap.Any("benchmarks", bid.Resources.GetBenchmarks()))

	return w.market.CreateOrder(w.ctx, bid)
}

func (w *engine) processOrderCreate() {
	for bid := range w.ordersCreateChan {
		created, err := w.sendOrderToMarket(bid.AsBID())
		if err != nil {
			w.log.Warn("cannot place order, retrying", zap.Error(err))
			w.CreateOrder(bid)
			continue
		}

		w.ordersResultsChan <- NewCorderFromOrder(created, bid.token)
	}
}

func (w *engine) processOrderResult() {
	for order := range w.ordersResultsChan {
		go w.waitForDeal(order)
	}
}

func (w *engine) waitForDeal(order *Corder) {
	id := order.GetId().Unwrap().String()

	// todo: use named logger with parameter (look at `processDeal`).

	t := util.NewImmediateTicker(w.cfg.OrderWatchInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			w.log.Info("checking for deal with order", zap.String("id", id))

			ord, err := w.market.GetOrderByID(w.ctx, &sonm.ID{Id: id})
			if err != nil {
				w.log.Warn("cannot get order info from market", zap.Error(err), zap.String("id", id))
				continue
			}

			if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
				w.log.Info("order becomes inactive, looking for related deal", zap.String("id", id))

				deal, err := w.deals.Status(w.ctx, ord.GetDealID())
				if err != nil {
					w.log.Warn("cannot get deal info from market", zap.Error(err), zap.String("deal_id", ord.GetDealID().Unwrap().String()))
					continue
				}

				w.CreateOrder(order)
				go w.processDeal(deal.GetDeal())
				return
			}

			w.log.Debug("order still have no deal", zap.String("id", id))
		}
	}
}

func (w *engine) processDeal(deal *sonm.Deal) {
	dealID := deal.GetId().Unwrap().String()
	log := w.log.Named("dealer").With(zap.String("deal_id", dealID),
		zap.String("supplier", deal.GetSupplierID().Unwrap().Hex()))

	log.Debug("start deal processing")
	defer log.Debug("stop deal processing")

	var taskReply *sonm.StartTaskReply

	// todo: move retry settings to cfg
	for try := 0; try < 5; try++ {
		// todo: ctx with timeout?
		// 1. ping worker

		if try > 0 {
			time.Sleep(10 * time.Second)
		}

		dealReply, err := w.deals.Status(w.ctx, deal.GetId())
		if err != nil || dealReply.GetResources() == nil {
			log.Warn("cannot connect to worker", zap.Error(err), zap.Int("try", try))
			continue
		}

		log.Debug("successfully obtained resources from worker", zap.Any("res", *dealReply.GetResources()))

		// 2. start task
		taskReply, err = w.tasks.Start(w.ctx, &sonm.StartTaskRequest{
			DealID: deal.GetId(),
			Spec: &sonm.TaskSpec{
				Container: &sonm.Container{
					Image: w.miningCfg.Image,
					Env: map[string]string{
						"WALLET":    w.miningCfg.Wallet.Hex(),
						"POOL_ADDR": w.miningCfg.PoolReportURL,
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

	// task still not started after all retries
	if taskReply == nil {
		w.finishDeal(deal.GetId())
		return
	}

	taskID := taskReply.GetId()
	log = log.Named("task").With(zap.String("task_id", taskID))
	log.Info("task started")

	// todo: move to config
	for try := 0; try < 5; try++ {
		log.Debug("checking task status")

		if try > 0 {
			time.Sleep(10 * time.Second)
		}

		// 3. ping task
		status, err := w.tasks.Status(w.ctx, &sonm.TaskID{Id: taskID, DealID: deal.GetId()})
		if err != nil {
			log.Warn("cannot get task status, increasing retry counter", zap.Error(err))
			continue
		}

		if status.GetStatus() == sonm.TaskStatusReply_FINISHED || status.GetStatus() == sonm.TaskStatusReply_BROKEN {
			log.Warn("task is failed by unknown reasons, finishing deal",
				zap.String("status", status.GetStatus().String()))
			w.finishDeal(deal.GetId())
			return
		}

		try = 0
		log.Debug("task status OK, resetting retry counter")
	}

	log.Debug("task status retries exceeded, finishing deal")
	w.finishDeal(deal.GetId())
}

func (w *engine) finishDeal(id *sonm.BigInt) {
	// todo: how to decide that we should add worker to blacklist?
	if _, err := w.deals.Finish(w.ctx, &sonm.DealFinishRequest{Id: id}); err != nil {
		w.log.Warn("cannot finish deal", zap.Error(err), zap.String("id", id.Unwrap().String()))
	}

	w.log.Info("deal finished", zap.String("id", id.Unwrap().String()))
}

func (w *engine) start(ctx context.Context) {
	go func() {
		defer close(w.ordersCreateChan)
		defer close(w.ordersResultsChan)

		w.log.Info("starting engine", zap.Int("concurrency", concurrency))
		defer w.log.Info("stopping engine")

		for i := 0; i < concurrency; i++ {
			go w.processOrderCreate()
		}

		for i := 0; i < concurrency; i++ {
			go w.processOrderResult()
		}

		<-ctx.Done()
	}()
}
