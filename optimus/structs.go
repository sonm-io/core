package optimus

import (
	"sync"
	"time"
)

type ordersState struct {
	mu sync.Mutex

	orders    *OrderClassification
	updatedAt time.Time
}

func newOrdersSet() *ordersState {
	return &ordersState{}
}

func (m *ordersState) Set(orders *OrderClassification) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.orders = orders
	m.updatedAt = time.Now()
}

func (m *ordersState) Get() *OrderClassification {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.orders
}

func (m *ordersState) UpdatedAt() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.updatedAt
}
