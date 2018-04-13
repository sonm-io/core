package optimus

import (
	"sync"
	"time"
)

type ordersSet struct {
	mu        sync.Mutex
	orders    []WeightedOrder
	updatedAt time.Time
}

func newOrdersSet() *ordersSet {
	return &ordersSet{}
}

func (m *ordersSet) Set(orders []WeightedOrder) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders = orders
	m.updatedAt = time.Now()
}

func (m *ordersSet) Get() []WeightedOrder {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orders
}

func (m *ordersSet) UpdatedAt() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.updatedAt
}
