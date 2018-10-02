package optimus

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"golang.org/x/sync/errgroup"
)

type MarketScanner interface {
	ActiveOrders(ctx context.Context) ([]*MarketOrder, error)
	ExecutedOrders(ctx context.Context, orderType sonm.OrderType) ([]*MarketOrder, error)
}

type cache struct {
	mu             sync.Mutex
	orders         []*MarketOrder
	updatedAt      time.Time
	updateInterval time.Duration
}

func newCache(updateInterval time.Duration) *cache {
	return &cache{
		updateInterval: updateInterval,
	}
}

func (m *cache) get(ctx context.Context, fn func(ctx context.Context) ([]*MarketOrder, error)) ([]*MarketOrder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Since(m.updatedAt) >= m.updateInterval {
		orders, err := fn(ctx)
		if err != nil {
			return nil, err
		}

		m.orders = orders
		m.updatedAt = time.Now()
	}

	return m.orders, nil
}

// MarketCache is a communication bus between fetching market orders and its
// consumers.
// Required, because there are multiple workers can be targeted by Optimus.
type MarketCache struct {
	market MarketScanner

	activeOrdersCache   *cache
	executedOrdersCache *cache
}

func newMarketCache(market MarketScanner, updateInterval time.Duration) *MarketCache {
	return &MarketCache{
		market:              market,
		activeOrdersCache:   newCache(updateInterval),
		executedOrdersCache: newCache(updateInterval),
	}
}

func (m *MarketCache) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	return m.activeOrdersCache.get(ctx, func(ctx context.Context) ([]*MarketOrder, error) {
		return m.market.ActiveOrders(ctx)
	})
}

func (m *MarketCache) ExecutedOrders(ctx context.Context, orderType sonm.OrderType) ([]*MarketOrder, error) {
	return m.executedOrdersCache.get(ctx, func(ctx context.Context) ([]*MarketOrder, error) {
		return m.market.ExecutedOrders(ctx, orderType)
	})
}

type PredefinedMarketCache struct {
	Orders []*MarketOrder
}

func NewPredefinedMarketCache(orders []*sonm.BigInt, market blockchain.MarketAPI) (*PredefinedMarketCache, error) {
	ctx := context.Background()
	wg, ctx := errgroup.WithContext(ctx)

	marketOrders := make([]*MarketOrder, len(orders))
	for id := range orders {
		id := id
		wg.Go(func() error {
			order, err := market.GetOrderInfo(ctx, orders[id].Unwrap())
			if err != nil {
				return err
			}
			marketOrders[id] = &MarketOrder{
				Order: order,
			}
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch market orders for simulation: %v", err)
	}

	return &PredefinedMarketCache{Orders: marketOrders}, nil
}

func (m *PredefinedMarketCache) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	return m.Orders, nil
}

func (m *PredefinedMarketCache) ExecutedOrders(ctx context.Context, orderType sonm.OrderType) ([]*MarketOrder, error) {
	return m.Orders, nil
}
