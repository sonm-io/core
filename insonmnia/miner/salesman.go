package miner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Salesman struct {
	state     *state.Storage
	resources *resource.Scheduler
	hardware  *hardware.Hardware
	eth       blockchain.API
	ethkey    *ecdsa.PrivateKey

	log      *zap.SugaredLogger
	dealChan chan *sonm.Deal
	tickLock sync.Mutex
}

func NewSalesman(
	ctx context.Context,
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
		log:       ctxlog.S(ctx).With("source", "salesman"),
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
	if err := askPlan.GetResources().GetGPU().Normalize(m.hardware); err != nil {
		return "", err
	}

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
	m.tickLock.Lock()
	defer m.tickLock.Unlock()
	ask, err := m.state.AskPlan(planID)
	if err != nil {
		return err
	}
	if !ask.GetDealID().IsZero() {
		return fmt.Errorf("ask plan %s is bound to deal %s", ask.ID, ask.DealID.String())
	}
	if !ask.GetOrderID().IsZero() {
		// background context is used  here because we need to get reply from blockchain
		err = <-m.eth.Market().CancelOrder(context.Background(), m.ethkey, ask.OrderID.Unwrap())
		if err != nil {
			return fmt.Errorf("could not cancel market order - %s", err)
		}
	}
	_, err = m.state.RemoveAskPlan(planID)
	if err != nil {
		return err
	}
	return m.resources.Release(planID)
}

func (m *Salesman) AskPlanByDeal(dealID string) (*sonm.AskPlan, error) {
	plans := m.state.AskPlans()
	for _, plan := range plans {
		if plan.DealID.String() == dealID {
			return plan, nil
		}
	}
	return nil, fmt.Errorf("ask plan for deal id %s is not found", dealID)
}

func (m *Salesman) syncRoutine(ctx context.Context) {
	m.log.Debugf("starting sync routine")
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
	m.log.Debugf("syncing salesman with blockchain")
	m.tickLock.Lock()
	defer m.tickLock.Unlock()
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
	m.log.Infof("checking deal %s for ask plan %s", plan.GetDealID().Unwrap().String(), plan.ID)
	panic("implement me")
}

func (m *Salesman) checkOrder(ctx context.Context, plan *sonm.AskPlan) {
	//TODO: validate deal that it is ours
	m.log.Infof("checking order %s for ask plan %s", plan.GetOrderID().Unwrap().String(), plan.ID)
	order, err := m.eth.Market().GetOrderInfo(ctx, plan.GetOrderID().Unwrap())
	if err != nil {
		m.log.Warnf("could not get order info for order %s - %s", plan.GetOrderID().Unwrap().String(), err)
		// TODO: log, what else can we do?
		return
	}

	//TODO: proper structs
	if len(order.DealID) != 0 {
		dealID, set := big.NewInt(0).SetString(order.DealID, 0)
		if !set {
			m.log.Warnf("could not parse order id from %s - %s", order.DealID, err)
			// TODO: log, what else can we do?
			return
		}

		deal, err := m.eth.Market().GetDealInfo(ctx, dealID)
		if err != nil {
			m.log.Warnf("could not get deal info from market - %s", err)
			// TODO: log, what else can we do?
			return
		}

		plan.DealID = sonm.NewBigInt(dealID)
		m.dealChan <- deal
		if err := m.state.SaveAskPlan(plan); err != nil {
			m.log.Warnf("could not get save ask plan with deal %s - %s", dealID.String(), err)
			// TODO: log, what else can we do?
			return
		}
	}
}

func (m *Salesman) placeOrder(ctx context.Context, plan *sonm.AskPlan) {
	benchmarks, err := m.hardware.ResourcesToBenchmarks(plan.GetResources())
	if err != nil {
		m.log.Warnf("could not get benchmarks for ask plan %s - %s", plan.ID, err)
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
	ordOrErr := <-m.eth.Market().PlaceOrder(ctx, m.ethkey, order)
	if ordOrErr.Err != nil {
		m.log.Warnf("could not place order on market for plan %s - %s", plan.ID, err)
		// TODO: log, what else can we do?
		return
	}
	orderId, set := big.NewInt(0).SetString(ordOrErr.Order.Id, 0)
	if !set {
		m.log.Warnf("could not parse order id string %s to big.Int - %s", ordOrErr.Order.Id, err)
		// TODO: log, what else can we do?
		return
	}
	plan.OrderID = sonm.NewBigInt(orderId)
	if err := m.state.SaveAskPlan(plan); err != nil {
		m.log.Warnf("could not save ask plan %s in storage - %s", plan.ID, err)
		// TODO: log, what else can we do?
		return
	}
}
