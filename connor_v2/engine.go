package connor

import (
	"context"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

const concurrency = 10

type engine struct {
	ordersCreateChan  chan *Corder
	ordersResultsChan chan *Corder

	market sonm.MarketClient
	deals  sonm.DealManagementClient
	log    *zap.Logger
	ctx    context.Context
	cfg    engineConfig
}

func NewEngine(ctx context.Context, cfg engineConfig, log *zap.Logger, market sonm.MarketClient, deals sonm.DealManagementClient) *engine {
	return &engine{
		ctx:               ctx,
		market:            market,
		deals:             deals,
		cfg:               cfg,
		log:               log.Named("engine"),
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
	// ping worker
	// start task
	// move task traction to another coroutine
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
