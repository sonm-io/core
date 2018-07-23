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

type engine struct {
	ordersCreateChan  chan *Corder
	ordersResultsChan chan *sonm.Order

	market sonm.MarketClient
	deals  sonm.DealManagementClient
	log    *zap.Logger
	ctx    context.Context
}

func NewEngine(ctx context.Context, log *zap.Logger, market sonm.MarketClient, deals sonm.DealManagementClient) *engine {
	return &engine{
		ctx:               ctx,
		market:            market,
		deals:             deals,
		log:               log.Named("engine"),
		ordersCreateChan:  make(chan *Corder, concurrency),
		ordersResultsChan: make(chan *sonm.Order, concurrency),
	}
}

func (w *engine) CreateOrder(bid *Corder) {
	w.ordersCreateChan <- bid
}

func (w *engine) RestoreOrder(order *Corder) {
	w.log.Debug("restoring order", zap.String("id", order.Order.GetId().Unwrap().String()))
	w.ordersResultsChan <- order.Order
}

func (w *engine) sendOrderToMarket(bid *sonm.BidOrder) (*sonm.Order, error) {
	w.log.Debug("creating order on market",
		zap.String("price", bid.GetPrice().GetPerSecond().Unwrap().String()),
		zap.Any("benchmarks", bid.Resources.GetBenchmarks()))

	return w.market.CreateOrder(w.ctx, bid)
}

func (w *engine) processOrderCreate() {
	for bid := range w.ordersCreateChan {
		ord, err := w.sendOrderToMarket(bid.AsBID())
		if err != nil {
			w.log.Warn("cannot place order, retrying", zap.Error(err))
			w.CreateOrder(bid)
			continue
		}

		w.ordersResultsChan <- ord
	}
}

func (w *engine) processOrderResult() {
	for order := range w.ordersResultsChan {
		w.log.Info("watching for deal with order",
			zap.String("id", order.GetId().Unwrap().String()),
			zap.String("price", order.GetPrice().ToPriceString()))

		// todo: spawn watching goroutine right here.
	}
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
