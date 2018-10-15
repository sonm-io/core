package optimus

import (
	"github.com/sonm-io/core/proto"
)

type Knapsack struct {
	manager *DeviceManager
	plans   []*sonm.AskPlan
}

func NewKnapsack(deviceManager *DeviceManager) *Knapsack {
	return &Knapsack{
		manager: deviceManager,
	}
}

func (m *Knapsack) Clone() *Knapsack {
	plans := make([]*sonm.AskPlan, len(m.plans))
	copy(plans, m.plans)

	return &Knapsack{
		manager: m.manager.Clone(),
		plans:   plans,
	}
}

func (m *Knapsack) Put(order *sonm.Order) error {
	resources, err := m.manager.Consume(*order.GetBenchmarks(), *order.GetNetflags())
	if err != nil {
		return err
	}

	resources.Network.NetFlags = order.GetNetflags()

	m.plans = append(m.plans, &sonm.AskPlan{
		OrderID:   order.GetId(),
		Price:     &sonm.Price{PerSecond: order.Price},
		Duration:  &sonm.Duration{Nanoseconds: 1e9 * int64(order.Duration)},
		Resources: resources,
	})

	return nil
}

func (m *Knapsack) Price() *sonm.Price {
	return sonm.SumPrice(m.plans)
}

func (m *Knapsack) PPSf64() float64 {
	return float64(m.Price().GetPerSecond().Unwrap().Uint64()) * 1e-18
}

func (m *Knapsack) Plans() []*sonm.AskPlan {
	return m.plans
}
