package optimus

import (
	"context"
	"sync"
	"time"
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
