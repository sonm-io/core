package salesman

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mohae/deepcopy"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xconcurrency"
	"go.uber.org/zap"
)

const defaultMaintenancePeriod = time.Hour * 24 * 365 * 100

const maintenanceGap = time.Minute * 10

const blockchainProcessConcurrency = 16

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

	askPlans        map[string]*sonm.AskPlan
	askPlanCGroups  map[string]cgroups.CGroup
	askPlanNetworks map[string]*network.Network
	deals           map[string]*sonm.Deal

	networkManager *network.NetworkManager

	nextMaintenance time.Time
	mu              sync.Mutex
}

func NewSalesman(ctx context.Context, opts ...Option) (*Salesman, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}

	askPlansKey := o.eth.ContractRegistry().MarketAddress().Hex() + "/ask_plans"

	networkManager, err := network.NewNetworkManagerWithConfig(network.NetworkManagerConfig{
		Log: o.log,
	})
	if err != nil {
		return nil, err
	}

	s := &Salesman{
		options:         o,
		askPlanStorage:  state.NewKeyedStorage(askPlansKey, o.storage),
		askPlanCGroups:  map[string]cgroups.CGroup{},
		askPlanNetworks: map[string]*network.Network{},
		deals:           map[string]*sonm.Deal{},
		nextMaintenance: time.Now().Add(defaultMaintenancePeriod),
		networkManager:  networkManager,
	}

	if err := s.restoreState(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Salesman) Close() error {
	return m.networkManager.Close()
}

func (m *Salesman) Run(ctx context.Context) {
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
}

func (m *Salesman) DebugDump() *sonm.SalesmanData {
	m.mu.Lock()
	defer m.mu.Unlock()
	reply := &sonm.SalesmanData{
		AskPlanCGroups: map[string]string{},
		Deals:          map[string]*sonm.Deal{},
		Orders:         map[string]*sonm.Order{},
	}
	for askID, cgroup := range m.askPlanCGroups {
		reply.AskPlanCGroups[askID] = cgroup.Suffix()
	}
	for askID, deal := range m.deals {
		reply.Deals[askID] = deal
	}
	return reply
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
	askPlan.CreateTime = sonm.CurrentTimestamp()
	if err := askPlan.GetResources().GetGPU().Normalize(m.hardware); err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.createCGroup(askPlan); err != nil {
		return "", err
	}

	if err := m.createNetwork(askPlan); err != nil {
		m.dropCGroup(askPlan.ID)
		return "", err
	}

	if err := m.resources.Consume(askPlan); err != nil {
		m.dropNetwork(askPlan.ID)
		m.dropCGroup(askPlan.ID)
		return "", err
	}

	m.askPlans[askPlan.ID] = askPlan
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		m.dropNetwork(askPlan.ID)
		m.dropCGroup(askPlan.ID)
		m.resources.Release(askPlan.ID)
		return "", err
	}
	return id, nil
}

func (m *Salesman) PurgeAskPlans(ctx context.Context) (*sonm.ErrorByStringID, error) {
	status := sonm.NewTSErrorByStringID()
	xconcurrency.Run(blockchainProcessConcurrency, m.AskPlans(), func(elem interface{}) {
		id := elem.(*sonm.AskPlan).ID
		err := m.RemoveAskPlan(ctx, id)
		status.Append(id, err)
	})

	return status.Unwrap(), nil
}

func (m *Salesman) RemoveAskPlan(ctx context.Context, planID string) error {

	ask, err := m.AskPlan(planID)
	if err != nil {
		return err
	}

	wasEmpty := false
	if !ask.DealID.IsZero() {
		if err := m.closeDeal(ctx, ask.DealID); err != nil {
			return fmt.Errorf("failed to remove ask plan %s: failed to close deal: %s", planID, err)
		}
	} else if !ask.OrderID.IsZero() {
		if err := m.cancelOrder(ctx, ask.OrderID); err != nil {
			return fmt.Errorf("failed to remove ask plan %s: failed to cancel order: %s", planID, err)
		}
	} else {
		wasEmpty = true
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	ask, ok := m.askPlans[planID]
	if !ok {
		// There is a race between  external call of this function and call from blockchain syncing routine,
		// so we check again if the plan was removed already
		return nil
	}
	if wasEmpty && (!ask.GetDealID().IsZero() || !ask.GetOrderID().IsZero()) {
		return fmt.Errorf("failed to remove ask plan %s: concurrent order or deal was placed", planID)
	}
	if err := m.resources.Release(planID); err != nil {
		// We can not handle this error, because it is persistent so just log it and skip
		m.log.Errorf("inconsistency found - could not release resources from pool: %s", err)
	}

	if err := m.dropNetwork(planID); err != nil {
		m.log.Errorf("failed to remove ask plan %s: failed to remove network: %s", planID, err)
	}
	if err := m.dropCGroup(planID); err != nil {
		m.log.Errorf("failed to remove ask plan %s: failed to remove cgroup: %s", planID, err)
	}
	delete(m.askPlans, planID)
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return fmt.Errorf("failed to remove ask plan %s: failed to save ask plans state in storage", planID)
	}
	return nil
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

func (m *Salesman) Network(planID string) (*network.Network, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	net, ok := m.askPlanNetworks[planID]
	if !ok {
		return nil, fmt.Errorf("network for ask plan %s not found, probably no such plan", planID)
	}
	return net, nil
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

func (m *Salesman) syncPlanWithBlockchain(ctx context.Context, plan *sonm.AskPlan) {
	orderId := plan.GetOrderID()
	dealId := plan.GetDealID()
	ctxWithTimeout, cancel := context.WithTimeout(ctx, m.config.SyncStepTimeout)
	defer cancel()
	if !dealId.IsZero() {
		if err := m.checkDeal(ctxWithTimeout, plan); err != nil {
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
}

func (m *Salesman) syncWithBlockchain(ctx context.Context) {
	m.log.Debugf("syncing salesman with blockchain")
	xconcurrency.Run(blockchainProcessConcurrency, m.AskPlans(), func(elem interface{}) {
		plan := elem.(*sonm.AskPlan)
		m.syncPlanWithBlockchain(ctx, plan)
	})
}

func (m *Salesman) restoreState(ctx context.Context) error {
	m.askPlans = map[string]*sonm.AskPlan{}
	if _, err := m.askPlanStorage.Load(&m.askPlans); err != nil {
		return fmt.Errorf("could not restore salesman state: %s", err)
	}

	pruneReply, err := m.networkManager.Prune(ctx, &network.PruneRequest{})
	if err != nil {
		m.log.Warnw("failed to prune unused networks", zap.Error(err))
		return err
	}
	m.log.Infow("removed no longer used networks", zap.Any("networks", *pruneReply))

	for _, plan := range m.askPlans {
		if err := m.resources.Consume(plan); err != nil {
			m.log.Warnf("dropping ask plan %s due to resource changes", plan.ID)
			//Ignore error here, as resources that were not consumed can not be released.
			m.RemoveAskPlan(ctx, plan.ID)
		} else {
			m.log.Debugf("consumed resource for ask plan %s", plan.GetID())
			if err := m.createCGroup(plan); err != nil {
				m.log.Warnf("can not create cgroup for ask plan %s: %s", plan.ID, err)
				return err
			}

			if err := m.createNetwork(plan); err != nil {
				m.log.Warnf("failed to restore network for ask plan %s: %s", plan.ID, err)
				return err
			}
		}
	}
	if _, err := m.storage.Load("next_maintenance", &m.nextMaintenance); err != nil {
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

func (m *Salesman) createNetwork(plan *sonm.AskPlan) error {
	net, err := m.networkManager.CreateNetwork(context.Background(), &network.CreateNetworkRequest{
		ID:               plan.ID,
		RateLimitIngress: plan.GetResources().GetNetwork().GetThroughputIn().GetBitsPerSecond(),
		RateLimitEgress:  plan.GetResources().GetNetwork().GetThroughputOut().GetBitsPerSecond(),
	})
	if err != nil {
		return fmt.Errorf("failed to create network - more detailed information can be found in worker logs")
	}

	m.log.Infof("created network %s for ask plan %s", net.Name, plan.ID)
	m.askPlanNetworks[plan.ID] = net

	return nil
}

func (m *Salesman) dropNetwork(planID string) error {
	net, ok := m.askPlanNetworks[planID]
	if !ok {
		return fmt.Errorf("network not found")
	}

	delete(m.askPlanNetworks, planID)

	if err := m.networkManager.RemoveNetwork(net); err != nil {
		return err
	}

	m.log.Infof("dropped network for ask plan %s", planID)
	return nil
}

func (m *Salesman) closeDeal(ctx context.Context, dealID *sonm.BigInt) error {
	deal, err := m.eth.Market().GetDealInfo(ctx, dealID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to fetch deal %s from market: %s", dealID, err)
	}

	if deal.Status == sonm.DealStatus_DEAL_ACCEPTED && deal.Duration != 0 {
		return fmt.Errorf("failed to close deal %s: deal is forward", dealID)
	}
	if deal.Status != sonm.DealStatus_DEAL_CLOSED {
		if err := m.eth.Market().CloseDeal(ctx, m.ethkey, deal.Id.Unwrap(), sonm.BlacklistType_BLACKLIST_NOBODY); err != nil {
			return fmt.Errorf("failed to close spot deal %s: %s", dealID, err)
		}
	}

	if err := m.dealDestroyer.CancelDealTasks(dealID); err != nil {
		return fmt.Errorf("failed to cancel deal's %s tasks: %s", dealID, err)
	}
	m.mu.Lock()
	delete(m.deals, dealID.Unwrap().String())
	m.mu.Unlock()
	return nil
}

func (m *Salesman) cancelOrder(ctx context.Context, orderID *sonm.BigInt) error {
	order, err := m.eth.Market().GetOrderInfo(ctx, orderID.Unwrap())
	if err != nil {
		return fmt.Errorf("failed to fetch order %s from market: %s", orderID, err)
	}

	if !order.DealID.IsZero() {
		return fmt.Errorf("failed to cancel order %s: order already have deal", orderID)
	}
	if order.OrderStatus == sonm.OrderStatus_ORDER_ACTIVE {
		if err := m.eth.Market().CancelOrder(ctx, m.ethkey, orderID.Unwrap()); err != nil {
			return fmt.Errorf("failed to cancel order %s: %s", orderID, err)
		}
	}
	return nil
}

func (m *Salesman) checkDeal(ctx context.Context, plan *sonm.AskPlan) error {
	dealID := plan.DealID.Unwrap()

	deal, err := m.eth.Market().GetDealInfo(ctx, dealID)
	if err != nil {
		return fmt.Errorf("could not get deal info for ask plan  %s: %s", plan.ID, err)
	}
	m.log.Debugf("checking deal %s for ask plan %s", deal.GetId().Unwrap().String(), plan.GetID())

	if err := m.registerDeal(ctx, plan.ID, deal); err != nil {
		return err
	}
	if deal.Status == sonm.DealStatus_DEAL_CLOSED {
		if err := m.RemoveAskPlan(ctx, plan.GetID()); err != nil {
			return fmt.Errorf("failed to remove ask plan %s with closed deal: %s", plan.GetID(), err)
		}
	} else {
		multi := multierror.NewMultiError()

		if err := m.maybeCloseDeal(ctx, plan, deal); err != nil {
			multi = multierror.Append(multi, fmt.Errorf("could not close deal: %s", err))
		}
		if err := m.maybeBillDeal(ctx, deal); err != nil {
			multi = multierror.Append(multi, fmt.Errorf("could not bill deal: %s", err))
		}
		return multi.ErrorOrNil()
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

func (m *Salesman) registerOrder(ctx context.Context, planID string, order *sonm.Order) error {
	if order.GetId().IsZero() {
		return fmt.Errorf("failed to register order: zero order id")
	}
	orderIDStr := order.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		m.log.Warnf("could not assign order %s to plan %s: no such plan", orderIDStr, planID)
		return m.cancelOrder(ctx, order.GetId())
	}
	if plan.GetOrderID().Cmp(order.GetId()) == 0 {
		return nil
	}
	if !plan.GetOrderID().IsZero() {
		return fmt.Errorf("attempted to register order %s for plan %s with deal %s",
			orderIDStr, planID, plan.GetOrderID().Unwrap().String())
	}
	plan.OrderID = order.GetId()
	plan.LastOrderPlacedTime = sonm.CurrentTimestamp()
	if err := m.askPlanStorage.Save(m.askPlans); err != nil {
		return err
	}
	m.log.Infof("assigned order %s to plan %s", orderIDStr, planID)
	return nil
}

func (m *Salesman) registerDeal(ctx context.Context, planID string, deal *sonm.Deal) error {
	if deal.GetId().IsZero() {
		return fmt.Errorf("failed to register deal: zero deal id")
	}
	dealIDStr := deal.GetId().Unwrap().String()
	m.mu.Lock()
	defer m.mu.Unlock()
	plan, ok := m.askPlans[planID]
	if !ok {
		// Looks like this should never happen
		m.closeDeal(ctx, deal.GetId())
		return fmt.Errorf("could not assign deal %s to plan %s: no such plan", dealIDStr, planID)
	}
	if plan.DealID.Cmp(deal.GetId()) != 0 && !plan.DealID.IsZero() {
		return fmt.Errorf("attempted to register deal %s for plan %s with deal %s",
			dealIDStr, planID, plan.DealID.Unwrap().String())
	}
	m.deals[dealIDStr] = deal
	plan.DealID = deal.GetId()

	ejectedPlans, err := m.resources.MakeRoomAndCommit(plan)
	if err != nil {
		m.log.Errorf("failed to make room and commit plan %s with new deal %s: %s", planID, deal.GetId(), err)
	}
	// Anyway check if any plans were ejected
	for _, planID := range ejectedPlans {
		plan, ok := m.askPlans[planID]
		if !ok {
			m.log.Errorf("ejected ask plan with ID %s is not found", planID)
			continue
		}
		if !plan.GetDealID().IsZero() {
			plan.Status = sonm.AskPlan_PENDING_DELETION
			m.closeDeal(ctx, plan.DealID)
		} else {
			m.log.Errorf("ejected ask plan with ID %s has no deal", planID)
		}
	}

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

	if !order.DealID.IsZero() {
		plan.DealID = order.DealID
		return m.checkDeal(ctx, plan)
	} else if order.OrderStatus != sonm.OrderStatus_ORDER_ACTIVE {
		return m.RemoveAskPlan(ctx, plan.ID)
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
	if err := m.registerOrder(ctx, plan.ID, order); err != nil {
		return nil, err
	}
	m.log.Infof("placed order %s on blockchain", order.GetId().Unwrap().String())
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
