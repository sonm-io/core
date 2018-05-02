package miner

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

type Salesman struct {
	state         *state.Storage
	resources     *resource.Scheduler
	hardware      *hardware.Hardware
	eth           blockchain.API
	dwh           sonm.DWHClient
	controlGroup  cgroups.CGroup
	cGroupManager cgroups.CGroupManager
	ethkey        *ecdsa.PrivateKey

	askPlanCGroups map[string]cgroups.CGroup

	matcher  matcher.Matcher
	log      *zap.SugaredLogger
	dealChan chan *sonm.Deal
	mu       sync.Mutex
}

func NewSalesman(
	ctx context.Context,
	state *state.Storage,
	resources *resource.Scheduler,
	hardware *hardware.Hardware,
	eth blockchain.API,
	matcher matcher.Matcher,
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
	if matcher == nil {
		return nil, errors.New("matcher is required for salesman")
	}

	s := &Salesman{
		state:     state,
		resources: resources,
		hardware:  hardware,
		eth:       eth,
		matcher:   matcher,
		ethkey:    ethkey,
		log:       ctxlog.S(ctx).With("source", "salesman"),
		dealChan:  make(chan *sonm.Deal, 1),
	}

	if err := s.restoreState(); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Salesman) Run(ctx context.Context) chan<- *sonm.Deal {
	go m.syncRoutine(ctx)
	return m.dealChan
}

func (m *Salesman) AskPlan(planID string) (*sonm.AskPlan, error) {
	return m.state.AskPlan(planID)
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

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.createCGroup(askPlan); err != nil {
		return "", err
	}

	if err := m.resources.Consume(askPlan); err != nil {
		m.dropCGroup(askPlan.ID)
		return "", err
	}

	if err := m.state.SaveAskPlan(askPlan); err != nil {
		m.dropCGroup(askPlan.ID)
		m.resources.Release(askPlan.ID)
		return "", err
	}
	return id, nil
}

func (m *Salesman) createCGroup(plan *sonm.AskPlan) error {
	cgroupResources := plan.GetResources().ToCgroupResources()
	cgroup, err := m.cGroupManager.Attach(plan.ID, cgroupResources)
	if err != nil {
		return err
	}
	m.askPlanCGroups[plan.ID] = cgroup
	return nil
}

func (m *Salesman) dropCGroup(planID string) error {
	cgroup, ok := m.askPlanCGroups[planID]
	if !ok {
		return fmt.Errorf("cgroup for ask plan %s not found, probably no such plan", planID)
	}
	err := cgroup.Delete()
	if err != nil {
		return fmt.Errorf("could not drop cgroup %s for ask plan %s - %s", cgroup.Suffix(), planID, err)
	}
	delete(m.askPlanCGroups, planID)
	return nil
}

func (m *Salesman) RemoveAskPlan(planID string) error {
	ask, err := m.state.AskPlan(planID)
	if err != nil {
		return err
	}
	if !ask.GetDealID().IsZero() {
		return fmt.Errorf("ask plan %s is bound to deal %s", ask.ID, ask.DealID.Unwrap().String())
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
	if err = m.resources.Release(planID); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dropCGroup(planID)
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
	defer ticker.Stop()
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

func (m *Salesman) restoreState() error {
	askPlans := m.state.AskPlans()
	for _, plan := range askPlans {
		m.resources.Consume(plan)
	}
	//TODO: restore tasks
	return nil
}

func (m *Salesman) checkDeal(ctx context.Context, plan *sonm.AskPlan) {
	m.log.Debugf("checking deal %s for ask plan %s and order %s",
		plan.DealID.Unwrap().String(), plan.ID, plan.GetOrderID().Unwrap().String())
	deal, err := m.eth.Market().GetDealInfo(ctx, plan.GetDealID().Unwrap())
	if err != nil {
		m.log.Warnf("could not get deal info for order %s, ask %s - %s",
			plan.GetOrderID().Unwrap().String(), plan.ID, err)
		// TODO: log, what else can we do?
		return
	}

	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		err = m.resources.Release(plan.ID)
		if err != nil {
			m.log.Warnf("could not release resources for deal %s, order %s, ask %s - %s",
				plan.DealID.Unwrap().String(), plan.OrderID.Unwrap().String(), plan.ID, err)
		}
		plan.DealID = nil
		plan.OrderID = nil
		m.state.SaveAskPlan(plan)
	}
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

	if !order.DealID.IsZero() {
		deal, err := m.eth.Market().GetDealInfo(ctx, order.DealID.Unwrap())
		if err != nil {
			m.log.Warnf("could not get deal info for ID %s from market - %s", order.DealID.Unwrap().String(), err)
			// TODO: log, what else can we do?
			return
		}
		m.dealChan <- deal

		plan.DealID = order.DealID

		if err := m.state.SaveAskPlan(plan); err != nil {
			m.log.Warnf("could not get save ask plan with deal %s - %s", order.DealID.Unwrap().String(), err)
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
		AuthorID:       sonm.NewEthAddress(crypto.PubkeyToAddress(m.ethkey.PublicKey)),
		CounterpartyID: plan.GetCounterparty(),
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
	plan.OrderID = ordOrErr.Order.Id
	if err := m.state.SaveAskPlan(plan); err != nil {
		m.log.Warnf("could not save ask plan %s in storage - %s", plan.ID, err)
		// TODO: log, what else can we do?
		return
	}
	go m.waitForDeal(ctx, ordOrErr.Order)
}

func (m *Salesman) waitForDeal(ctx context.Context, order *sonm.Order) {
	// TODO: make configurable
	ticker := util.NewImmediateTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			//TODO: we also need to do it on worker start
			deal, err := m.matcher.CreateDealByOrder(ctx, order)
			if err != nil {
				m.log.Warnf("could not wait for deal on order %s - %s", order.Id.Unwrap().String(), err)
				order, err := m.eth.Market().GetOrderInfo(ctx, order.Id.Unwrap())
				if err != nil {
					m.log.Warnf("could not get order info for order %s - %s", order.Id.Unwrap().String(), err)
					continue
				}

				if order.GetOrderStatus() != sonm.OrderStatus_ORDER_ACTIVE {
					return
				}
				continue
			}
			m.log.Infof("created deal %s for order %s", deal.Id.Unwrap().String(), order.Id.Unwrap().String())
			order.DealID = deal.Id
			m.dealChan <- deal
			return
		}
	}
}
