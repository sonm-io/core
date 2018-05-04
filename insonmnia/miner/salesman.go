package miner

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
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
	cGroupManager cgroups.CGroupManager
	matcher       matcher.Matcher
	ethkey        *ecdsa.PrivateKey

	askPlanCGroups map[string]cgroups.CGroup
	deals          map[string]*sonm.Deal
	orders         map[string]*sonm.Order

	log *zap.SugaredLogger
	mu  sync.Mutex
}

//TODO: make configurable
const (
	regularDealBillPeriod = time.Second * 3600
	spotDealBillPeriod    = time.Second * 3600 * 24
)

func NewSalesman(
	ctx context.Context,
	state *state.Storage,
	resources *resource.Scheduler,
	hardware *hardware.Hardware,
	eth blockchain.API,
	cGroupManager cgroups.CGroupManager,
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
	if cGroupManager == nil {
		return nil, errors.New("cGroup manager is required for salesman")
	}

	s := &Salesman{
		state:          state,
		resources:      resources,
		hardware:       hardware,
		eth:            eth,
		cGroupManager:  cGroupManager,
		matcher:        matcher,
		ethkey:         ethkey,
		askPlanCGroups: map[string]cgroups.CGroup{},
		deals:          map[string]*sonm.Deal{},
		orders:         map[string]*sonm.Order{},
		log:            ctxlog.S(ctx).With("source", "salesman"),
	}

	if err := s.restoreState(); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Salesman) Run(ctx context.Context) {
	go m.syncRoutine(ctx)
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

func (m *Salesman) AskPlanByDeal(dealID *sonm.BigInt) (*sonm.AskPlan, error) {
	plans := m.state.AskPlans()
	for _, plan := range plans {
		if plan.DealID.Cmp(dealID) == 0 {
			return plan, nil
		}
	}
	return nil, fmt.Errorf("ask plan for deal id %s is not found", dealID)
}

func (m *Salesman) Deal(dealID *sonm.BigInt) (*sonm.Deal, error) {
	id := dealID.Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()

	deal, ok := m.deals[id]
	if !ok {
		return nil, fmt.Errorf(" deal not found by %s", id)
	}
	return deal, nil
}

func (m *Salesman) CGroup(planID string) (cgroups.CGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cGroup, ok := m.askPlanCGroups[planID]
	if !ok {
		return nil, fmt.Errorf("cgroup for ask plan %s not found, probably no such plan", planID)
	}
	return cGroup, nil
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
			if err := m.checkDeal(ctx, plan); err != nil {
				m.log.Warnf("could not check deal %s for plan %s - %s", dealId.Unwrap().String(), plan.ID, err)
			}
		} else if !orderId.IsZero() {
			if err := m.checkOrder(ctx, plan); err != nil {
				m.log.Warnf("could not check order %s for plan %s - %s", dealId.Unwrap().String(), plan.ID, err)
			}
		} else {
			if err := m.placeOrder(ctx, plan); err != nil {
				m.log.Warnf("could not place order for plan %s - %s", plan.ID, err)
			}
		}
	}
}

func (m *Salesman) restoreState() error {
	askPlans := m.state.AskPlans()
	//TODO:  check if we do not lack resources after restart
	for _, plan := range askPlans {
		if err := m.resources.Consume(plan); err != nil {
			m.log.Warnf("dropping ask plan due to resource changes")
			//Ignore error here, as resources that were not consumed can not be released.
			m.RemoveAskPlan(plan.ID)
		} else {
			if err := m.createCGroup(plan); err != nil {
				m.log.Warnf("can not create cgroup for ask plan %s - %s", plan.ID, err)
				return err
			}
		}
	}
	//TODO: restore tasks
	return nil
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
	delete(m.askPlanCGroups, planID)
	if err := cgroup.Delete(); err != nil {
		return fmt.Errorf("could not drop cgroup %s for ask plan %s - %s", cgroup.Suffix(), planID, err)
	}
	return nil
}

func (m *Salesman) checkDeal(ctx context.Context, plan *sonm.AskPlan) error {
	m.log.Debugf("checking deal %s for ask plan %s and order %s",
		plan.DealID.Unwrap().String(), plan.ID, plan.GetOrderID().Unwrap().String())
	deal, err := m.eth.Market().GetDealInfo(ctx, plan.GetDealID().Unwrap())
	if err != nil {
		return fmt.Errorf("could not get deal info for order %s, ask %s - %s",
			plan.GetOrderID().Unwrap().String(), plan.ID, err)
	}
	m.registerDeal(deal)

	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		err = m.resources.Release(plan.ID)
		if err != nil {
			m.log.Warnf("could not release resources for deal %s, order %s, ask %s - %s",
				plan.DealID.Unwrap().String(), plan.OrderID.Unwrap().String(), plan.ID, err)
		}
		plan.DealID = nil
		plan.OrderID = nil
		return m.state.SaveAskPlan(plan)
	} else {
		errBill := m.maybeBillDeal(ctx, deal)
		errClose := m.maybeCloseDeal(ctx, deal)
		if errBill != nil && errClose != nil {
			return fmt.Errorf("could not bill deal - %s, and close deal - %s", errBill, errClose)
		}
		if errBill != nil {
			return fmt.Errorf("could not bill deal - %s", errBill)
		}
		if errClose != nil {
			return fmt.Errorf("could not close deal - %s", errClose)
		}
	}
	return nil
}

func (m *Salesman) maybeBillDeal(ctx context.Context, deal *sonm.Deal) error {
	start := deal.StartTime.Unix()
	var billPeriod time.Duration
	if deal.GetDuration() == 0 {
		billPeriod = spotDealBillPeriod
	} else {
		billPeriod = regularDealBillPeriod
	}

	if time.Now().Sub(start) > billPeriod {
		//TODO: fix ETH API and this
		return <-m.eth.Market().Bill(ctx, m.ethkey, deal.GetId().Unwrap())
	}
	return nil
}

func (m *Salesman) maybeCloseDeal(ctx context.Context, deal *sonm.Deal) error {
	if deal.GetDuration() != 0 {
		endTime := deal.GetStartTime().Unix().Add(time.Second * time.Duration(deal.GetDuration()))
		if time.Now().After(endTime) {
			return <-m.eth.Market().CloseDeal(ctx, m.ethkey, deal.GetId().Unwrap(), false)
		}
	}
	return nil
}

func (m *Salesman) registerOrder(order *sonm.Order) {
	id := order.Id.Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	if !order.DealID.IsZero() || order.OrderStatus == sonm.OrderStatus_ORDER_ACTIVE {
		m.orders[id] = order
	} else {
		delete(m.orders, id)
	}
}

func (m *Salesman) registerDeal(deal *sonm.Deal) {
	id := deal.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	if deal.Status == sonm.DealStatus_DEAL_ACCEPTED {
		m.deals[id] = deal
	} else {
		delete(m.deals, id)
	}
}

func (m *Salesman) checkOrder(ctx context.Context, plan *sonm.AskPlan) error {
	//TODO: validate deal that it is ours
	m.log.Infof("checking order %s for ask plan %s", plan.GetOrderID().Unwrap().String(), plan.ID)
	order, err := m.eth.Market().GetOrderInfo(ctx, plan.GetOrderID().Unwrap())
	if err != nil {
		return fmt.Errorf("could not get order info for order %s - %s", plan.GetOrderID().Unwrap().String(), err)
	}

	m.registerOrder(order)

	if !order.DealID.IsZero() {
		plan.DealID = order.DealID

		if err := m.state.SaveAskPlan(plan); err != nil {
			return fmt.Errorf("could not get save ask plan with deal %s - %s", order.DealID.Unwrap().String(), err)
		}
		return m.checkDeal(ctx, plan)
	} else if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		plan.OrderID = nil
		m.state.SaveAskPlan(plan)
	}
	return nil
}

func (m *Salesman) placeOrder(ctx context.Context, plan *sonm.AskPlan) error {
	benchmarks, err := m.hardware.ResourcesToBenchmarks(plan.GetResources())
	if err != nil {
		return fmt.Errorf("could not get benchmarks for ask plan %s - %s", plan.ID, err)
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
		Netflags:      sonm.NetflagsToUint([3]bool{net.GetOverlay(), net.GetOutbound(), net.GetIncoming()}),
		IdentityLevel: plan.GetIdentity(),
		Blacklist:     plan.GetBlacklist().Unwrap().Hex(),
		Tag:           plan.GetTag(),
		Benchmarks:    benchmarks,
	}
	ordOrErr := <-m.eth.Market().PlaceOrder(ctx, m.ethkey, order)
	if ordOrErr.Err != nil {
		return fmt.Errorf("could not place order on market for plan %s - %s", plan.ID, err)
	}
	plan.OrderID = ordOrErr.Order.Id
	if err := m.state.SaveAskPlan(plan); err != nil {
		return fmt.Errorf("could not save ask plan %s in storage - %s", plan.ID, err)
	}
	go m.waitForDeal(ctx, ordOrErr.Order)
	return nil
}

func (m *Salesman) waitForDeal(ctx context.Context, order *sonm.Order) error {
	// TODO: make configurable
	ticker := util.NewImmediateTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			//TODO: we also need to do it on worker start
			deal, err := m.matcher.CreateDealByOrder(ctx, order)

			if err != nil {
				m.log.Warnf("could not wait for deal on order %s - %s", order.Id.Unwrap().String(), err)
				id := order.Id.Unwrap()
				order, err := m.eth.Market().GetOrderInfo(ctx, id)
				if err != nil {
					m.log.Warnf("could not get order info for order %s - %s", id.String(), err)
					continue
				}

				if order.GetOrderStatus() != sonm.OrderStatus_ORDER_ACTIVE {
					return nil
				}
				continue
			}
			m.registerDeal(deal)
			m.log.Infof("created deal %s for order %s", deal.Id.Unwrap().String(), order.Id.Unwrap().String())
			order.DealID = deal.Id
			return nil
		}
	}
}
