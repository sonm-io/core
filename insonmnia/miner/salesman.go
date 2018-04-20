package miner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

type Salesman struct {
	state     *state.Storage
	resources *resource.Scheduler
	hardware  *hardware.Hardware
	eth       blockchain.API
	ethkey    *ecdsa.PrivateKey
	dealChan  chan *sonm.Deal
}

func NewSalesman(
	state *state.Storage,
	resources *resource.Scheduler,
	hardware *hardware.Hardware,
	eth blockchain.API,
	ethkey *ecdsa.PrivateKey,
) (*Salesman, error) {
	if state == nil {
		return nil, errors.New("storage is required for salesman")
	}
	if resources == nil {
		return nil, errors.New("resource scheduler is required for salesman")
	}
	if hardware == nil {
		return nil, errors.New("hardware is required for salesman")
	}
	if eth == nil {
		return nil, errors.New("blockchain API is required for salesman")
	}
	if ethkey == nil {
		return nil, errors.New("ethereum private key is required for salesman")
	}
	return &Salesman{
		state:     state,
		resources: resources,
		hardware:  hardware,
		eth:       eth,
		ethkey:    ethkey,
		dealChan:  make(chan *sonm.Deal, 1),
	}, nil
}

func (m *Salesman) Run(ctx context.Context) chan *sonm.Deal {
	go m.syncRoutine(ctx)
	return m.dealChan
}

func (m *Salesman) AskPlans() map[string]*sonm.AskPlan {
	return m.state.AskPlans()
}

func (m *Salesman) CreateAskPlan(askPlan *sonm.AskPlan) (string, error) {
	id := uuid.New()
	askPlan.ID = id
	if err := m.resources.Consume(askPlan); err != nil {
		return "", err
	}

	if err := m.state.SaveAskPlan(askPlan); err != nil {
		m.resources.Release(askPlan.ID)
		return "", err
	}
	return id, nil
}

func (m *Salesman) RemoveAskPlan(planID string) error {
	ask, err := m.state.AskPlan(planID)
	ask, err := m.state.RemoveAskPlan(planID)
	if err != nil {
		return err
	}
	if (!ask.DealID.IsZero())
	m.scheduledForDeletion <- ask
	return nil
}

func (m *Salesman) AskPlanByDeal(dealID string) (*sonm.AskPlan, error) {
	plans := m.state.AskPlans()
}

func (m *Salesman) syncRoutine(ctx context.Context) {
	ticker := util.NewImmediateTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			m.syncWithBlockchain(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (m *Salesman) syncWithBlockchain(ctx context.Context) {
	plans := m.state.AskPlans()
	for _, plan := range plans {
		orderId := plan.GetOrderID()
		dealId := plan.GetDealID()
		if !dealId.IsZero() {
			m.checkDeal(ctx, plan)
		} else if !orderId.IsZero() {
			m.checkOrder(ctx, plan)
		} else if dealId.IsZero() && orderId.IsZero() {
			m.placeOrder(ctx, plan)
		}
	}
}

func (m *Salesman) checkDeal(ctx context.Context, plan *sonm.AskPlan) {

}

func (m *Salesman) checkOrder(ctx context.Context, plan *sonm.AskPlan) {
	order, err := m.eth.GetOrderInfo(ctx, plan.GetOrderID().Unwrap())
	if err != nil {
		// TODO: log, what else can we do?
		return
	}

	//TODO: proper structs
	if len(order.DealID) != 0 {
		dealID, set := big.NewInt(0).SetString(order.DealID, 0)
		if !set {
			// TODO: log, what else can we do?
			return
		}

		deal, err := m.eth.GetDealInfo(ctx, dealID)
		if err != nil {
			// TODO: log, what else can we do?
			return
		}

		plan.DealID = sonm.NewBigInt(dealID)
		m.dealChan <- deal
		if err := m.state.SaveAskPlan(plan); err != nil {
			// TODO: log, what else can we do?
			return
		}
	}
}

func (m *Salesman) placeOrder(ctx context.Context, plan *sonm.AskPlan) {
	benchmarks, err := m.hardware.ResourcesToBenchmarks(plan.GetResources())
	if err != nil {
		// TODO: log, what else can we do?
		return
	}
	net := m.hardware.Network
	order := &sonm.Order{
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       crypto.PubkeyToAddress(m.ethkey.PublicKey).Hex(),
		CounterpartyID: plan.GetCounterparty().Unwrap().Hex(),
		Duration:       uint64(plan.GetDuration().Unwrap().Seconds()),
		Price:          plan.GetPrice().GetPerSecond(),
		//TODO:refactor NetFlags in separqate PR
		Netflags:      sonm.NetflagsToUint([3]bool{net.GetOverlay(), true, net.Incoming}),
		IdentityLevel: plan.GetIdentity(),
		Blacklist:     plan.GetBlacklist().Unwrap().Hex(),
		Tag:           plan.GetTag(),
		Benchmarks:    benchmarks,
	}
	ordOrErr := <-m.eth.PlaceOrder(ctx, m.ethkey, order)
	if ordOrErr.Err != nil {
		// TODO: log, what else can we do?
		return
	}
	orderId, set := big.NewInt(0).SetString(ordOrErr.Order.Id, 0)
	if !set {
		// TODO: log, what else can we do?
		return
	}
	plan.OrderID = sonm.NewBigInt(orderId)
	if err := m.state.SaveAskPlan(plan); err != nil {
		// TODO: log, what else can we do?
		return
	}
}
