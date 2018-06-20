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
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

const defaultMaintenancePeriod = time.Hour * 24 * 365 * 100

const maintenanceGap = time.Minute * 10

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

	nextMaintenance time.Time
	dealsCh         chan *sonm.Deal
	mu              sync.Mutex
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
		options:         o,
		askPlanStorage:  state.NewKeyedStorage("ask_plans", o.storage),
		askPlanCGroups:  map[string]cgroups.CGroup{},
		deals:           map[string]*sonm.Deal{},
		orders:          map[string]*sonm.Order{},
		nextMaintenance: time.Now().Add(defaultMaintenancePeriod),
		dealsCh:         make(chan *sonm.Deal, 100),
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

func (m *Salesman) ScheduleMaintenance(timePoint time.Time) error {
	m.log.Infof("Scheduling next maintenance at %s", timePoint.String())
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextMaintenance = timePoint
	return m.storage.Save("next_maintenance", m.nextMaintenance)
}

func (m *Salesman) NextMaintenance() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.nextMaintenance
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
	if plan.Status != sonm.AskPlan_PENDING_DELETION {
		return nil
	}
	m.log.Debugf("trying to shut down ask plan %s", plan.GetID())
	if !plan.GetDealID().IsZero() {
		m.log.Debugf("ask plan %s is still bound to deal %s", plan.ID, plan.GetDealID().Unwrap().String())
		return nil
	}

	if !plan.GetOrderID().IsZero() {
		if err := m.eth.Market().CancelOrder(ctx, m.ethkey, plan.GetOrderID().Unwrap()); err != nil {
			return fmt.Errorf("could not cancel order: %s", err)
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
		return nil, fmt.Errorf("deal not found by %s", id)
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
		if !dealId.IsZero() {
			if err := m.loadCheckDeal(ctxWithTimeout, plan); err != nil {
				m.log.Warnf("could not check deal %s for plan %s: %s", dealId.Unwrap().String(), plan.ID, err)
			}
		} else if !orderId.IsZero() {
			if err := m.checkOrder(ctxWithTimeout, plan); err != nil {
				m.log.Warnf("could not check order %s for plan %s: %s", orderId.Unwrap().String(), plan.ID, err)
			}
		} else if plan.GetStatus() != sonm.AskPlan_PENDING_DELETION {
			order, err := m.placeOrder(ctxWithTimeout, plan)
			if err != nil {
				m.log.Warnf("could not place order for plan %s: %s", plan.ID, err)
			} else {
				go m.waitForDeal(ctx, order)
			}
		}
		if err := m.maybeShutdownAskPlan(ctxWithTimeout, plan); err != nil {
			m.log.Warnf("could not shutdown ask plan %s: %s", plan.ID, err)
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
	if err := m.storage.Load("next_maintenance", &m.nextMaintenance); err != nil {
		return fmt.Errorf("failed to load next maintenance: %s", err)
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

	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		if err := m.unregisterOrder(plan.ID); err != nil {
			return fmt.Errorf("failed to unregister order from ask plan %s: %s", plan.GetID(), err)
		}
		if err := m.unregisterDeal(plan.GetID(), deal); err != nil {
			return fmt.Errorf("failed to cleanup deal from ask plan %s: %s", plan.GetID(), err)
		}
		m.log.Debugf("succesefully removed closed deal %s from ask plan %s", deal.GetId().Unwrap().String(), plan.GetID())
		return nil
	} else {
		multi := multierror.NewMultiError()
		if err := m.registerDeal(plan.GetID(), deal); err != nil {
			multi = multierror.Append(multi, err)
		}
		if err := m.maybeCloseDeal(ctx, plan, deal); err != nil {
			multi = multierror.Append(multi, fmt.Errorf("could not close deal: %s", err))
		}
		if err := m.maybeBillDeal(ctx, deal); err != nil {
			multi = multierror.Append(multi, fmt.Errorf("could not bill deal: %s", err))
		}
		return multi.ErrorOrNil()
	}
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
		if err := m.eth.Market().Bill(ctx, m.ethkey, deal.GetId().Unwrap()); err != nil {
			return err
		}
		m.log.Infof("billed deal %s", deal.GetId().Unwrap().String())
	}
	return nil
}

func (m *Salesman) shouldCloseDeal(ctx context.Context, plan *sonm.AskPlan, deal *sonm.Deal) bool {
	if deal.GetDuration() != 0 {
		endTime := deal.GetStartTime().Unix().Add(time.Second * time.Duration(deal.GetDuration()))
		if time.Now().After(endTime) {
			return true
		}
	} else {
		if plan.Status == sonm.AskPlan_PENDING_DELETION {
			return true
		}
		if time.Now().After(m.NextMaintenance()) {
			return true
		}
	}
	return false
}

func (m *Salesman) maybeCloseDeal(ctx context.Context, plan *sonm.AskPlan, deal *sonm.Deal) error {
	if m.shouldCloseDeal(ctx, plan, deal) {
		// TODO: we will know about closed deal on next iteration for simplicicty,
		// but maybe we can optimize here.
		if err := m.eth.Market().CloseDeal(ctx, m.ethkey, deal.GetId().Unwrap(), sonm.BlacklistType_BLACKLIST_NOBODY); err != nil {
			return err
		}

		m.log.Infof("closed deal %s", deal.GetId().Unwrap().String())
	}
	return nil
}

func (m *Salesman) unregisterOrder(planID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("failed to drop order from plan %s: no such plan", planID)
	}
	orderID := plan.GetOrderID()
	if orderID.IsZero() {
		return fmt.Errorf("failed to drop order from plan %s: plan has zero order", planID)
	}
	idStr := orderID.Unwrap().String()
	delete(m.orders, idStr)
	plan.OrderID = nil
	m.log.Infof("unregistered order %s", idStr)
	return nil
}

func (m *Salesman) registerOrder(planID string, order *sonm.Order) error {
	if order.GetId().IsZero() {
		return fmt.Errorf("failed to register order: zero order id")
	}
	orderIDStr := order.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("could not assign order %s to plan %s: no such plan", orderIDStr, planID)
	}
	if plan.GetOrderID().Cmp(order.GetId()) == 0 {
		return nil
	}
	if !plan.GetOrderID().IsZero() {
		return fmt.Errorf("attempted to register order %s for plan %s with deal %s",
			orderIDStr, planID, plan.GetOrderID().Unwrap().String())
	}
	plan.OrderID = order.GetId()
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.orders[orderIDStr] = order
	m.log.Infof("assigned order %s to plan %s", orderIDStr, planID)
	return nil
}

func (m *Salesman) unregisterDeal(planID string, deal *sonm.Deal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("could not drop deal from plan %s: no such plan", planID)
	}
	dealID := plan.DealID
	if dealID.IsZero() {
		return nil
	}
	m.dealsCh <- deal
	delete(m.deals, dealID.Unwrap().String())
	plan.DealID = nil
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.log.Infof("dropped deal %s from plan %s", dealID.Unwrap().String(), planID)
	return nil
}

func (m *Salesman) registerDeal(planID string, deal *sonm.Deal) error {
	if deal.GetId().IsZero() {
		return fmt.Errorf("failed to register deal: zero deal id")
	}
	dealIDStr := deal.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		return fmt.Errorf("could not assign deal %s to plan %s: no such plan", dealIDStr, planID)
	}
	if plan.DealID.Cmp(deal.GetId()) != 0 && !plan.DealID.IsZero() {
		return fmt.Errorf("attempted to register deal %s for plan %s with deal %s",
			dealIDStr, planID, plan.DealID.Unwrap().String())
	}
	m.dealsCh <- deal
	m.deals[dealIDStr] = deal
	plan.DealID = deal.GetId()
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.log.Infof("assigned deal %s to plan %s", dealIDStr, planID)
	return nil
}

func (m *Salesman) checkOrder(ctx context.Context, plan *sonm.AskPlan) error {
	//TODO: validate deal that it is ours
	m.log.Debugf("checking order %s for ask plan %s", plan.GetOrderID().Unwrap().String(), plan.ID)
	order, err := m.eth.Market().GetOrderInfo(ctx, plan.GetOrderID().Unwrap())
	if err != nil {
		return fmt.Errorf("could not get order info for order %s: %s", plan.GetOrderID().Unwrap().String(), err)
	}

	if err := m.registerOrder(plan.GetID(), order); err != nil {
		return fmt.Errorf("could not register order %s: %s", plan.GetOrderID().Unwrap().String(), err)
	}

	if !order.DealID.IsZero() {
		plan.DealID = order.DealID
		return m.loadCheckDeal(ctx, plan)
	} else if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		return m.unregisterOrder(plan.ID)
	} else {
		maintenanceTime := m.NextMaintenance()
		orderEndTime := time.Now().Add(time.Second * time.Duration(order.Duration))
		if orderEndTime.After(maintenanceTime) {
			if err := m.eth.Market().CancelOrder(ctx, m.ethkey, plan.GetOrderID().Unwrap()); err != nil {
				return fmt.Errorf("could not cancel order for maintenance - %s", err)
			}
			return m.unregisterOrder(plan.ID)
		}
	}
	return nil
}

func (m *Salesman) placeOrder(ctx context.Context, plan *sonm.AskPlan) (*sonm.Order, error) {
	benchmarks, err := m.hardware.ResourcesToBenchmarks(plan.GetResources())
	if err != nil {
		return nil, fmt.Errorf("could not get benchmarks for ask plan %s: %s", plan.ID, err)
	}
	maintenanceTime := m.NextMaintenance()
	// we add some "gap" here to be ready for maintenance slightly before it occurs
	clearTime := maintenanceTime.Add(-maintenanceGap)
	now := time.Now()
	if now.After(clearTime) {
		return nil, fmt.Errorf("faiiled to place order: maintenance is scheduled at %s", maintenanceTime.String())
	}
	duration := plan.GetDuration().Unwrap()
	if duration != 0 && now.Add(duration).After(clearTime) {
		duration = clearTime.Sub(now)
		//rare case but still possible
		if uint64(duration) == 0 {
			return nil, fmt.Errorf("faiiled to place order: maintenance is scheduled at %s", maintenanceTime.String())
		}
		m.log.Infof("reducing order duration from %d to %d due to maintenance at %s",
			uint64(plan.GetDuration().Unwrap().Seconds()), uint64(duration.Seconds()), clearTime.String())
	}

	net := plan.GetResources().GetNetwork()
	order := &sonm.Order{
		OrderType:      sonm.OrderType_ASK,
		OrderStatus:    sonm.OrderStatus_ORDER_ACTIVE,
		AuthorID:       sonm.NewEthAddress(crypto.PubkeyToAddress(m.ethkey.PublicKey)),
		CounterpartyID: plan.GetCounterparty(),
		Duration:       uint64(duration.Seconds()),
		Price:          plan.GetPrice().GetPerSecond(),
		//TODO:refactor NetFlags in separqate PR
		Netflags:      net.GetNetFlags(),
		IdentityLevel: plan.GetIdentity(),
		Blacklist:     plan.GetBlacklist().Unwrap().Hex(),
		Tag:           plan.GetTag(),
		Benchmarks:    benchmarks,
	}
	order, err = m.eth.Market().PlaceOrder(ctx, m.ethkey, order)
	if err != nil {
		return nil, fmt.Errorf("could not place order on bc market: %s", err)
	}
	if err := m.registerOrder(plan.ID, order); err != nil {
		return nil, err
	}
	m.log.Infof("placed order %s on blockchain", plan.OrderID.Unwrap().String())
	return order, nil
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
			m.log.Infof("created deal %s for order %s", deal.Id.Unwrap().String(), order.Id.Unwrap().String())
			return nil
		}
	}
}
