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
}

// MarketCache is a communication bus between fetching market orders and its
// consumers.
// Required, because there are multiple workers can be targeted by Optimus.
type MarketCache struct {
	mu             sync.Mutex
	market         MarketScanner
	orders         []*MarketOrder
	updatedAt      time.Time
	updateInterval time.Duration
}

func newMarketCache(market MarketScanner, updateInterval time.Duration) *MarketCache {
	return &MarketCache{
		market:         market,
		updateInterval: updateInterval,
	}
}

func (m *MarketCache) ActiveOrders(ctx context.Context) ([]*MarketOrder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Since(m.updatedAt) >= m.updateInterval {
		orders, err := m.market.ActiveOrders(ctx)
		if err != nil {
			return nil, err
		}

		m.orders = orders
		m.updatedAt = time.Now()
	}

	return m.orders, nil
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
