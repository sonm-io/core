package miner

import (
	"context"
	"crypto/ecdsa"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/proto"
)

type Salesman struct {
	state  *state.Storage
	eth    blockchain.API
	ethkey *ecdsa.PrivateKey
}

func (m *Salesman) AskPlans() map[string]*sonm.AskPlan {
	return m.state.AskPlans()
}

func (m *Salesman) CreateAskPlan(askPlan *sonm.AskPlan) (string, error) {
	return m.state.CreateAskPlan(askPlan)
}

func (m *Salesman) RemoveAskPlan(planID string) error {
	return m.state.RemoveAskPlan(planID)
}

func (m *Salesman) AskPlanByDeal(dealID string) (*sonm.AskPlan, error) {
	plans := m.state.AskPlans()
}

func (m *Salesman) syncRoutine() {
	panic("implement me")
}

func (m *Salesman) syncWithBlockchain(ctx context.Context) {
	plans := m.state.AskPlans()
	for _, plan := range plans {
		orderId := plan.GetOrderID()
		dealId := plan.GetDealID()
		if dealId.IsZero() && orderId.IsZero() {
			//TODO: async processing
			m.eth.PlaceOrder(ctx, m.ethkey, plan.ToOrder())
			info, err := m.eth.GetOrderInfo(ctx, orderId)
			info.OrderStatus
		}
	}
}
