package connor

import (
	"context"
	"fmt"
	"sync"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const concurrency = 10

type orderManager struct {
	ordersChan  chan *Corder
	resultsChan chan *sonm.Order
	market      sonm.MarketClient
	log         *zap.Logger
	ctx         context.Context
}

func NewOrderManager(ctx context.Context, log *zap.Logger, market *sonm.MarketClient) *orderManager {
	return &orderManager{
		ctx: ctx,
		// todo: use real marketClient
		// marketClient:      marketClient,
		market:      &FakeMarketClient{},
		log:         log.Named("order-manager"),
		ordersChan:  make(chan *Corder, concurrency),
		resultsChan: make(chan *sonm.Order, concurrency),
	}
}

func (w *orderManager) Create(bid *Corder) {
	w.ordersChan <- bid
}

func (w *orderManager) Restore(order *Corder) {
	w.resultsChan <- order.Order
}

func (w *orderManager) sendOrderToMarket(bid *sonm.BidOrder) (*sonm.Order, error) {
	// todo: warp with required parameters
	w.log.Debug("sending order to marketClient")
	return w.market.CreateOrder(w.ctx, bid)
}

func (w *orderManager) processCreate() {
	w.log.Debug("processCreate started")

	for bid := range w.ordersChan {
		ord, err := w.sendOrderToMarket(bid.AsBID())
		if err != nil {
			w.log.Warn("cannot place order, retrying", zap.Error(err))
			w.Create(bid)
			continue
		}

		w.resultsChan <- ord
	}
}

func (w *orderManager) processResult() {
	w.log.Debug("processResult started")

	for order := range w.resultsChan {
		w.log.Info("order created",
			zap.String("id", order.GetId().Unwrap().String()),
			zap.String("price", order.GetPrice().ToPriceString()))

		// todo: spawn watching goroutine right here.
	}
}

func (w *orderManager) start(ctx context.Context) {
	defer close(w.ordersChan)
	defer close(w.resultsChan)

	w.log.Info("starting order manager", zap.Int("concurrency", concurrency))
	defer w.log.Info("stopping order manager")

	for i := 0; i < concurrency; i++ {
		go w.processCreate()
	}

	for i := 0; i < concurrency; i++ {
		go w.processResult()
	}

	<-ctx.Done()
}

type FakeMarketClient struct {
	mu sync.Mutex
	i  int64
}

func (f *FakeMarketClient) CreateOrder(ctx context.Context, in *sonm.BidOrder, opts ...grpc.CallOption) (*sonm.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.i++
	return &sonm.Order{Id: sonm.NewBigIntFromInt(f.i)}, nil

}

func (f *FakeMarketClient) GetOrders(ctx context.Context, in *sonm.Count, opts ...grpc.CallOption) (*sonm.GetOrdersReply, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *FakeMarketClient) GetOrderByID(ctx context.Context, in *sonm.ID, opts ...grpc.CallOption) (*sonm.Order, error) {
	return nil, fmt.Errorf("not implemented")

}

func (f *FakeMarketClient) CancelOrder(ctx context.Context, in *sonm.ID, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *FakeMarketClient) Purge(ctx context.Context, in *sonm.Empty, opts ...grpc.CallOption) (*sonm.Empty, error) {
	return nil, fmt.Errorf("not implemented")
}
