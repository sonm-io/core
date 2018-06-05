package salesman

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mohae/deepcopy"
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

type Config struct {
	Logger        zap.SugaredLogger
	Storage       *state.Storage
	Resources     *resource.Scheduler
	Hardware      *hardware.Hardware
	Eth           blockchain.API
	CGroupManager cgroups.CGroupManager
	Matcher       matcher.Matcher
	Ethkey        *ecdsa.PrivateKey
	Config        YAMLConfig
}

type YAMLConfig struct {
	RegularBillPeriod    time.Duration `yaml:"regular_deal_bill_period" default:"24h"`
	SpotBillPeriod       time.Duration `yaml:"spot_deal_bill_period" default:"1h"`
	SyncStepTimeout      time.Duration `yaml:"sync_step_timeout" default:"2m"`
	SyncInterval         time.Duration `yaml:"sync_interval" default:"10s"`
	MatcherRetryInterval time.Duration `yaml:"matcher_retry_interval" default:"10s"`
}

type Salesman struct {
	*options
	askPlanStorage *state.KeyedStorage

	askPlans       map[string]*sonm.AskPlan
	askPlanCGroups map[string]cgroups.CGroup
	deals          map[string]*sonm.Deal
	orders         map[string]*sonm.Order

	dealsCh chan *sonm.Deal
	mu      sync.Mutex
}

func NewSalesman(opts ...Option) (*Salesman, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}

	s := &Salesman{
		options:        o,
		askPlanStorage: state.NewKeyedStorage("ask_plans", o.storage),
		askPlanCGroups: map[string]cgroups.CGroup{},
		deals:          map[string]*sonm.Deal{},
		orders:         map[string]*sonm.Order{},
		dealsCh:        make(chan *sonm.Deal, 100),
	}

	if err := s.restoreState(); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Salesman) Close() {}

func (m *Salesman) Run(ctx context.Context) <-chan *sonm.Deal {
	go func() {
		for _, plan := range m.askPlans {
			orderID := plan.GetOrderID()
			dealID := plan.GetDealID()
			if dealID.IsZero() && !orderID.IsZero() {
				order, err := m.eth.Market().GetOrderInfo(ctx, orderID.Unwrap())
				if err != nil {
					m.log.Warnf("failed to get order info for order %s, stopping waiting for deal: %s", orderID.Unwrap().String(), err)
					continue
				}
				go m.waitForDeal(ctx, order)
			}
		}
		go m.syncRoutine(ctx)
	}()
	return m.dealsCh
}

func (m *Salesman) AskPlan(planID string) (*sonm.AskPlan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	askPlan, ok := m.askPlans[planID]
	if !ok {
		return nil, errors.New("specified ask-plan does not exist")
	}
	copy := deepcopy.Copy(askPlan).(*sonm.AskPlan)
	return copy, nil
}

func (m *Salesman) AskPlans() map[string]*sonm.AskPlan {
	m.mu.Lock()
	defer m.mu.Unlock()
	return deepcopy.Copy(m.askPlans).(map[string]*sonm.AskPlan)
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

	m.askPlans[askPlan.ID] = askPlan
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		m.dropCGroup(askPlan.ID)
		m.resources.Release(askPlan.ID)
		return "", err
	}
	return id, nil
}

func (m *Salesman) RemoveAskPlan(planID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ask, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("no such plan %s", planID)
	}

	ask.Status = sonm.AskPlan_PENDING_DELETION
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return fmt.Errorf("could not mark ask plan %s with active deal %s for deletion: %s",
			planID, ask.GetDealID().Unwrap().String(), err)
	}
	return nil
}

func (m *Salesman) maybeShutdownAskPlan(ctx context.Context, plan *sonm.AskPlan) error {
	m.log.Debugf("trying to shut down ask plan %s", plan.GetID())
	if !plan.GetDealID().IsZero() {
		dealInfo, err := m.eth.Market().GetDealInfo(ctx, plan.GetDealID().Unwrap())
		if err != nil {
			return err
		}
		if err := m.checkDeal(ctx, plan, dealInfo); err != nil {
			return err
		}
		if dealInfo.Status == sonm.DealStatus_DEAL_ACCEPTED {
			if dealInfo.GetDuration() == 0 {
				m.log.Infof("closing spot deal %s for ask plan %s", dealInfo.GetId(), plan.GetID())
				if err := <-m.eth.Market().CloseDeal(ctx, m.ethkey, plan.GetDealID().Unwrap(), false); err != nil {
					return err
				}
			} else {
				m.log.Debugf("ask plan %s is still bound to deal %s, checking deal", plan.ID, plan.GetDealID().Unwrap().String())
				return nil
			}
		}
	}

	if plan.GetDealID().IsZero() && !plan.GetOrderID().IsZero() {
		if err := <-m.eth.Market().CancelOrder(ctx, m.ethkey, plan.GetOrderID().Unwrap()); err != nil {
			m.log.Infof("could not cancel order - %s, checking order to update info", err)
			return m.checkOrder(ctx, plan)
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.resources.Release(plan.ID); err != nil {
		// We can not handle this error, because it is persistent so just log it and skip
		m.log.Errorf("inconsistency found - could not release resources from pool: %s", err)
	}

	delete(m.askPlans, plan.ID)
	m.askPlanStorage.Save(m.askPlans)
	return m.dropCGroup(plan.ID)

}

func (m *Salesman) AskPlanByDeal(dealID *sonm.BigInt) (*sonm.AskPlan, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, plan := range m.askPlans {
		if plan.DealID.Cmp(dealID) == 0 {
			return deepcopy.Copy(plan).(*sonm.AskPlan), nil
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
	ticker := util.NewImmediateTicker(m.config.SyncInterval)
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
	plans := m.AskPlans()
	for _, plan := range plans {
		orderId := plan.GetOrderID()
		dealId := plan.GetDealID()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.SyncStepTimeout)
		if plan.Status == sonm.AskPlan_PENDING_DELETION {
			if err := m.maybeShutdownAskPlan(ctxWithTimeout, plan); err != nil {
				m.log.Warnf("could not shutdown ask plan %s: %s", plan.ID, err)
			}
		} else if !dealId.IsZero() {
			if err := m.loadCheckDeal(ctxWithTimeout, plan); err != nil {
				m.log.Warnf("could not check deal %s for plan %s: %s", dealId.Unwrap().String(), plan.ID, err)
			}
		} else if !orderId.IsZero() {
			if err := m.checkOrder(ctxWithTimeout, plan); err != nil {
				m.log.Warnf("could not check order %s for plan %s: %s", orderId.Unwrap().String(), plan.ID, err)
			}
		} else {
			order, err := m.placeOrder(ctxWithTimeout, plan)
			if err != nil {
				m.log.Warnf("could not place order for plan %s: %s", plan.ID, err)
			} else {
				go m.waitForDeal(ctx, order)
			}
		}
		cancel()
	}
}

func (m *Salesman) restoreState() error {
	m.askPlans = map[string]*sonm.AskPlan{}
	if err := m.askPlanStorage.Load(&m.askPlans); err != nil {
		return fmt.Errorf("could not restore salesman state: %s", err)
	}
	for _, plan := range m.askPlans {
		if err := m.resources.Consume(plan); err != nil {
			m.log.Warnf("dropping ask plan due to resource changes")
			//Ignore error here, as resources that were not consumed can not be released.
			m.RemoveAskPlan(plan.ID)
		} else {
			m.log.Debugf("consumed resource for ask plan %s", plan.GetID())
			if err := m.createCGroup(plan); err != nil {
				m.log.Warnf("can not create cgroup for ask plan %s: %s", plan.ID, err)
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
	m.log.Infof("created cgroup %s for ask plan %s", cgroup.Suffix(), plan.ID)
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
		return fmt.Errorf("could not drop cgroup %s for ask plan %s: %s", cgroup.Suffix(), planID, err)
	}
	m.log.Debugf("dropped cgroup for ask plan %s", planID)
	return nil
}

func (m *Salesman) loadCheckDeal(ctx context.Context, plan *sonm.AskPlan) error {
	dealID := plan.DealID.Unwrap()

	deal, err := m.eth.Market().GetDealInfo(ctx, dealID)
	if err != nil {
		return fmt.Errorf("could not get deal info for ask plan  %s: %s", plan.ID, err)
	}
	return m.checkDeal(ctx, plan, deal)

}

func (m *Salesman) checkDeal(ctx context.Context, plan *sonm.AskPlan, deal *sonm.Deal) error {
	m.log.Debugf("checking deal %s for ask plan %s", deal.GetId().Unwrap().String(), plan.GetID())

	m.registerDeal(deal)

	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		if err := m.assignOrder(plan.ID, nil); err != nil {
			return fmt.Errorf("failed to cleanup order from ask plan %s: %s", plan.GetID(), err)
		}
		if err := m.assignDeal(plan.ID, nil); err != nil {
			return fmt.Errorf("failed to cleanup deal from ask plan %s: %s", plan.GetID(), err)
		}
		m.log.Debugf("succesefully removed closed deal %s from ask plan %s", deal.GetId().Unwrap().String(), plan.GetID())
		return nil
	} else {
		errClose := m.maybeCloseDeal(ctx, deal)
		errBill := m.maybeBillDeal(ctx, deal)
		if errBill != nil && errClose != nil {
			return fmt.Errorf("could not bill deal: %s, and close deal: %s", errBill, errClose)
		}
		if errBill != nil {
			return fmt.Errorf("could not bill deal: %s", errBill)
		}
		if errClose != nil {
			return fmt.Errorf("could not close deal: %s", errClose)
		}
	}
	return nil
}

func (m *Salesman) maybeBillDeal(ctx context.Context, deal *sonm.Deal) error {
	startTime := deal.GetStartTime().Unix()
	billTime := deal.GetLastBillTS().Unix()
	if billTime.Before(startTime) {
		billTime = startTime
	}
	var billPeriod time.Duration
	if deal.IsSpot() {
		billPeriod = m.config.SpotBillPeriod
	} else {
		billPeriod = m.config.RegularBillPeriod
	}

	if time.Now().Sub(billTime) > billPeriod {
		if err := <-m.eth.Market().Bill(ctx, m.ethkey, deal.GetId().Unwrap()); err != nil {
			return err
		}
		m.log.Infof("billed deal %s", deal.GetId().Unwrap().String())
	}
	return nil
}

func (m *Salesman) maybeCloseDeal(ctx context.Context, deal *sonm.Deal) error {
	if deal.GetDuration() != 0 {
		endTime := deal.GetStartTime().Unix().Add(time.Second * time.Duration(deal.GetDuration()))
		if time.Now().After(endTime) {
			if err := <-m.eth.Market().CloseDeal(ctx, m.ethkey, deal.GetId().Unwrap(), false); err != nil {
				return err
			}
			m.log.Infof("closed expired deal %s", deal.GetId().Unwrap().String())
		}
	}
	return nil
}

func (m *Salesman) registerOrder(order *sonm.Order) {
	id := order.Id.Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	_, has := m.orders[id]
	if !order.DealID.IsZero() || order.OrderStatus == sonm.OrderStatus_ORDER_ACTIVE {
		if !has {
			m.orders[id] = order
			m.log.Infof("registered order %s", order.GetId().Unwrap().String())
		}
	} else {
		if has {
			delete(m.orders, id)
			m.log.Infof("unregistered order %s", order.GetId().Unwrap().String())
		}
	}
}

func (m *Salesman) registerDeal(deal *sonm.Deal) {
	// Always send deal in case it was closed or changed via change request or smth.
	m.dealsCh <- deal
	id := deal.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	_, has := m.deals[id]
	if deal.Status == sonm.DealStatus_DEAL_ACCEPTED {
		if !has {
			m.deals[id] = deal
			m.log.Infof("registered deal %s", deal.GetId().Unwrap().String())
		}
	} else {
		if has {
			delete(m.deals, id)
			m.log.Infof("unregistered deal %s", deal.GetId().Unwrap().String())
		}
	}
}

func (m *Salesman) assignDeal(planID string, dealID *sonm.BigInt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("could not assign deal %s to plan %s: no such plan", dealID.Unwrap().String(), planID)
	}
	plan.DealID = dealID
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.log.Infof("assigned deal %s to plan %s", dealID.Unwrap().String(), planID)
	return nil
}

func (m *Salesman) assignOrder(planID string, orderID *sonm.BigInt) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("could not assign order %s to plan %s: no such plan", orderID.Unwrap().String(), planID)
	}
	plan.OrderID = orderID
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.log.Infof("assigned order %s to plan %s", orderID.Unwrap().String(), planID)
	return nil
}

func (m *Salesman) checkOrder(ctx context.Context, plan *sonm.AskPlan) error {
	//TODO: validate deal that it is ours
	m.log.Debugf("checking order %s for ask plan %s", plan.GetOrderID().Unwrap().String(), plan.ID)
	order, err := m.eth.Market().GetOrderInfo(ctx, plan.GetOrderID().Unwrap())
	if err != nil {
		return fmt.Errorf("could not get order info for order %s: %s", plan.GetOrderID().Unwrap().String(), err)
	}

	m.registerOrder(order)

	if !order.DealID.IsZero() {
		plan.DealID = order.DealID
		if err := m.assignDeal(plan.ID, order.DealID); err != nil {
			return err
		}
		return m.loadCheckDeal(ctx, plan)
	} else if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		return m.assignOrder(plan.ID, nil)
	}
	return nil
}

func (m *Salesman) placeOrder(ctx context.Context, plan *sonm.AskPlan) (*sonm.Order, error) {
	benchmarks, err := m.hardware.ResourcesToBenchmarks(plan.GetResources())
	if err != nil {
		return nil, fmt.Errorf("could not get benchmarks for ask plan %s: %s", plan.ID, err)
	}

	net := plan.GetResources().GetNetwork()
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
		return nil, fmt.Errorf("could not place order on bc market: %s", ordOrErr.Err)
	}
	if err := m.assignOrder(plan.ID, ordOrErr.Order.GetId()); err != nil {
		return nil, err
	}
	m.log.Infof("placed order %s on blockchain", plan.OrderID.Unwrap().String())
	return ordOrErr.Order, nil
}

func (m *Salesman) waitForDeal(ctx context.Context, order *sonm.Order) error {
	m.log.Infof("waiting for deal for %s", order.GetId().Unwrap().String())
	ticker := util.NewImmediateTicker(m.config.MatcherRetryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			//TODO: we also need to do it on worker start
			deal, err := m.matcher.CreateDealByOrder(ctx, order)

			if err != nil {
				m.log.Warnf("could not wait for deal on order %s: %s", order.Id.Unwrap().String(), err)
				id := order.Id.Unwrap()
				order, err := m.eth.Market().GetOrderInfo(ctx, id)
				if err != nil {
					m.log.Warnf("could not get order info for order %s: %s", id.String(), err)
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
