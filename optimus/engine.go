package optimus

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	minNumOrders = sonm.MinNumBenchmarks
)

type optimizationInput struct {
	Orders  []*MarketOrder
	Devices *sonm.DevicesReply
	Plans   map[string]*sonm.AskPlan
}

// VictimPlans returns plans that can be removed to be replaced with another
// plans.
// This is useful to virtualize worker free devices, that are currently
// occupied by either nearly-to-expire or spot plans.
func (m *optimizationInput) VictimPlans() map[string]*sonm.AskPlan {
	victims := map[string]*sonm.AskPlan{}
	for id, plan := range m.Plans {
		// Currently we can remove spot orders without regret.
		if plan.GetDuration().Unwrap() == 0 {
			victims[id] = plan
		}
	}

	return victims
}

func (m *optimizationInput) FreeDevices() (*sonm.DevicesReply, error) {
	return m.freeDevices(map[string]*sonm.AskPlan{})
}

func (m *optimizationInput) VirtualFreeDevices() (*sonm.DevicesReply, error) {
	return m.freeDevices(m.VictimPlans())
}

func (m *optimizationInput) Price() *sonm.Price {
	var plans []*sonm.AskPlan
	for _, plan := range m.Plans {
		plans = append(plans, plan)
	}

	return sonm.SumPrice(plans)
}

func (m *optimizationInput) freeDevices(removalVictims map[string]*sonm.AskPlan) (*sonm.DevicesReply, error) {
	workerHardware := hardware.Hardware{
		CPU:     m.Devices.CPU,
		GPU:     m.Devices.GPUs,
		RAM:     m.Devices.RAM,
		Network: m.Devices.Network,
		Storage: m.Devices.Storage,
	}
	// All resources are free by default.
	freeResources := workerHardware.AskPlanResources()

	// Subtract plans except cancellation removalVictims. Doing so produces us a
	// new free(!) devices list.
	for id, plan := range m.Plans {
		if _, ok := removalVictims[id]; !ok {
			if err := freeResources.Sub(plan.Resources); err != nil {
				return nil, fmt.Errorf("failed to virtualize resource releasing: %v", err)
			}
		}
	}

	freeWorkerHardware, err := workerHardware.LimitTo(freeResources)
	if err != nil {
		return nil, fmt.Errorf("failed to limit virtual free hardware: %v", err)
	}

	return freeWorkerHardware.IntoProto(), nil
}

type workerEngine struct {
	cfg workerConfig
	log *zap.SugaredLogger

	addr             common.Address
	masterAddr       common.Address
	market           blockchain.MarketAPI
	marketCache      *MarketCache
	worker           WorkerManagementClientExt
	benchmarkMapping benchmarks.Mapping

	optimizationConfig optimizationConfig
}

func newWorkerEngine(cfg workerConfig, addr, masterAddr common.Address, worker sonm.WorkerManagementClient, market blockchain.MarketAPI, marketCache *MarketCache, benchmarkMapping benchmarks.Mapping, optimizationConfig optimizationConfig, log *zap.SugaredLogger) (*workerEngine, error) {
	m := &workerEngine{
		cfg: cfg,
		log: log.With(zap.Stringer("addr", addr)),

		addr:             addr,
		masterAddr:       masterAddr,
		market:           market,
		marketCache:      marketCache,
		worker:           &workerManagementClientExt{worker},
		benchmarkMapping: benchmarkMapping,

		optimizationConfig: optimizationConfig,
	}

	return m, nil
}

func (m *workerEngine) OnRun() {
	m.log.Info("managing worker")
}

func (m *workerEngine) OnShutdown() {
	m.log.Info("stop managing worker")
}

func (m *workerEngine) Execute(ctx context.Context) {
	m.log.Info("optimization epoch started")

	if err := m.execute(ctx); err != nil {
		m.log.Warn(err.Error())
	}
}

func (m *workerEngine) execute(ctx context.Context) error {
	input, err := m.optimizationInput(ctx)
	if err != nil {
		return err
	}

	m.log.Debugf("pulled %d orders from the marketplace", len(input.Orders))
	m.log.Debugw("pulled worker devices", zap.Any("devices", *input.Devices))
	m.log.Debugw("pulled worker plans", zap.Any("plans", input.Plans))

	removedPlans, err := m.tryRemoveUnsoldPlans(ctx, input.Plans)
	if err != nil {
		return err
	}

	if len(removedPlans) != 0 {
		return m.execute(ctx)
	}

	victimPlans := input.VictimPlans()
	m.log.Debugw("victim plans", zap.Any("plans", victimPlans))

	naturalFreeDevices, err := input.FreeDevices()
	if err != nil {
		return err
	}

	virtualFreeDevices, err := input.VirtualFreeDevices()
	if err != nil {
		return err
	}

	m.log.Debugw("virtualized worker natural free devices", zap.Any("devices", *naturalFreeDevices))
	m.log.Debugw("virtualized worker virtual free devices", zap.Any("devices", *virtualFreeDevices))

	// Here we append removal candidate's orders to "orders" from the
	// marketplace to be able to track their profitability.
	virtualFreeOrders, err := m.ordersForPlans(ctx, victimPlans)
	if err != nil {
		return fmt.Errorf("failed to collect orders for victim plans: %v", err)
	}

	// Extended orders set, with added currently executed orders.
	extOrders := append(append([]*MarketOrder{}, input.Orders...), virtualFreeOrders...)

	var naturalKnapsack, virtualKnapsack *Knapsack

	wg := errgroup.Group{}
	wg.Go(func() error {
		m.log.Info("optimizing using natural free devices")
		knapsack, err := m.optimize(input.Devices, naturalFreeDevices, input.Orders, m.log.With(zap.String("optimization", "natural")))
		if err != nil {
			return err
		}

		naturalKnapsack = knapsack
		return nil
	})
	wg.Go(func() error {
		m.log.Info("optimizing using virtual free devices")
		knapsack, err := m.optimize(input.Devices, virtualFreeDevices, extOrders, m.log.With(zap.String("optimization", "virtual")))
		if err != nil {
			return err
		}

		virtualKnapsack = knapsack
		return nil
	})
	if err := wg.Wait(); err != nil {
		return err
	}

	m.log.Infow("current worker price", zap.String("Σ USD/s", input.Price().GetPerSecond().ToPriceString()))
	m.log.Infow("optimizing using natural free devices done", zap.String("Σ USD/s", naturalKnapsack.Price().GetPerSecond().ToPriceString()), zap.Any("plans", naturalKnapsack.Plans()))
	m.log.Infow("optimizing using virtual free devices done", zap.String("Σ USD/s", virtualKnapsack.Price().GetPerSecond().ToPriceString()), zap.Any("plans", virtualKnapsack.Plans()))

	if m.cfg.DryRun {
		return fmt.Errorf("further worker management has been interrupted: dry-run mode is active")
	}

	// Compare total USD/s before and after. Remove some plans if the diff is
	// more than the threshold.
	priceThreshold := m.cfg.PriceThreshold.GetPerSecond()
	priceDiff := new(big.Int).Sub(virtualKnapsack.Price().GetPerSecond().Unwrap(), input.Price().GetPerSecond().Unwrap())
	swingTime := new(big.Int).Sub(priceDiff, priceThreshold.Unwrap()).Sign() >= 0

	var winners []*sonm.AskPlan
	if swingTime {
		m.log.Info("using replacement strategy")

		create, remove, ignore := m.splitPlans(input.Plans, virtualKnapsack.Plans())
		m.log.Infow("ignoring already existing plans", zap.Any("plans", ignore))
		m.log.Infow("removing plans", zap.Any("plans", remove))

		victims := make([]string, 0, len(remove))
		for _, plan := range remove {
			victims = append(victims, plan.ID)
		}
		if err := m.worker.RemoveAskPlans(ctx, victims); err != nil {
			return err
		}

		winners = create
	} else {
		m.log.Info("using appending strategy")
		winners = naturalKnapsack.Plans()
	}

	if len(winners) == 0 {
		return fmt.Errorf("no plans found")
	}

	for _, plan := range winners {
		plan.Identity = m.cfg.Identity

		id, err := m.worker.CreateAskPlan(ctx, plan)
		if err != nil {
			m.log.Warnw("failed to create sell plan", zap.Any("plan", *plan), zap.Error(err))
			continue
		}

		m.log.Infof("created sell plan %s", id.Id)
	}

	return nil
}

func (m *workerEngine) splitPlans(plans map[string]*sonm.AskPlan, candidates []*sonm.AskPlan) (create, remove, ignore []*sonm.AskPlan) {
	orders := map[string]struct{}{}
	for _, plan := range plans {
		orders[plan.GetOrderID().Unwrap().String()] = struct{}{}
	}

	newOrders := map[string]struct{}{}
	for _, plan := range candidates {
		newOrders[plan.GetOrderID().Unwrap().String()] = struct{}{}
	}

	for _, plan := range candidates {
		if _, ok := orders[plan.GetOrderID().Unwrap().String()]; ok {
			ignore = append(ignore, plan)
		} else {
			create = append(create, plan)
		}
	}

	for _, plan := range plans {
		if _, ok := newOrders[plan.GetOrderID().Unwrap().String()]; !ok {
			remove = append(remove, plan)
		}
	}

	return create, remove, ignore
}

func (m *workerEngine) optimizationInput(ctx context.Context) (*optimizationInput, error) {
	input := &optimizationInput{}

	ctx, cancel := context.WithTimeout(ctx, m.cfg.PreludeTimeout)
	defer cancel()

	// Concurrently fetch all required inputs, such as market orders, worker
	// devices and plans.
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		orders, err := m.marketCache.ActiveOrders(ctx)
		if err != nil {
			return fmt.Errorf("failed to pull market orders: %v", err)
		}
		if len(orders) == 0 {
			return fmt.Errorf("not enough orders to perform optimization")
		}

		input.Orders = orders
		return nil
	})
	wg.Go(func() error {
		devices, err := m.worker.Devices(ctx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to pull worker devices: %v", err)
		}

		input.Devices = devices
		return nil
	})
	wg.Go(func() error {
		plans, err := m.worker.AskPlans(ctx, &sonm.Empty{})
		if err != nil {
			return fmt.Errorf("failed to pull worker plans: %v", err)
		}

		input.Plans = plans.GetAskPlans()
		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return input, nil
}

func (m *workerEngine) tryRemoveUnsoldPlans(ctx context.Context, plans map[string]*sonm.AskPlan) ([]string, error) {
	victims := make([]string, 0, len(plans))
	for id, plan := range plans {
		if plan.UnsoldDuration() >= m.cfg.StaleThreshold {
			victims = append(victims, id)
		}
	}

	if len(victims) == 0 {
		m.log.Info("no unsold plans found")
		return victims, nil
	}

	m.log.Infow("removing unsold plans", zap.Duration("threshold", m.cfg.StaleThreshold), zap.Any("plans", victims))
	if err := m.worker.RemoveAskPlans(ctx, victims); err != nil {
		return nil, fmt.Errorf("failed to remove some unsold plans: %v", err)
	}

	return victims, nil
}

func (m *workerEngine) ordersForPlans(ctx context.Context, plans map[string]*sonm.AskPlan) ([]*MarketOrder, error) {
	var orders []*MarketOrder

	mu := sync.Mutex{}
	wg, ctx := errgroup.WithContext(ctx)

	for id, plan := range plans {
		id := id
		plan := plan

		wg.Go(func() error {
			order, err := m.market.GetOrderInfo(ctx, plan.OrderID.Unwrap())
			if err != nil {
				return fmt.Errorf("failed to get order for `%s`: %v", id, err)
			}

			mu.Lock()
			defer mu.Unlock()
			orders = append(orders, &MarketOrder{
				Order: order,
			})

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (m *workerEngine) optimize(devices, freeDevices *sonm.DevicesReply, orders []*MarketOrder, log *zap.SugaredLogger) (*Knapsack, error) {
	deviceManager, err := newDeviceManager(devices, freeDevices, m.benchmarkMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to construct device manager: %v", err)
	}

	matchedOrders := m.matchingOrders(deviceManager, devices, orders)
	log.Infof("found %d/%d matching orders", len(matchedOrders), len(orders))

	if len(matchedOrders) == 0 {
		log.Infof("no matching orders found")
		return NewKnapsack(deviceManager), nil
	}

	now := time.Now()
	knapsack := NewKnapsack(deviceManager)
	if err := m.optimizationMethod(len(matchedOrders), log).Optimize(knapsack, matchedOrders); err != nil {
		return nil, err
	}

	log.Infof("optimized %d orders in %s", len(matchedOrders), time.Since(now))

	return knapsack, nil
}

// MatchingOrders filters the given orders to have only orders that are subset
// of ours.
func (m *workerEngine) matchingOrders(deviceManager *DeviceManager, devices *sonm.DevicesReply, orders []*MarketOrder) []*MarketOrder {
	matchedOrders := make([]*MarketOrder, 0, len(orders))

	filter := FittingFunc{
		Filters: m.filters(deviceManager, devices),
	}

	for _, order := range orders {
		if filter.Filter(order.GetOrder()) {
			matchedOrders = append(matchedOrders, order)
		}
	}

	return matchedOrders
}

func (m *workerEngine) filters(deviceManager *DeviceManager, devices *sonm.DevicesReply) []func(order *sonm.Order) bool {
	return []func(order *sonm.Order) bool{
		func(order *sonm.Order) bool {
			return order.OrderType == sonm.OrderType_BID
		},
		func(order *sonm.Order) bool {
			switch m.cfg.OrderPolicy {
			case PolicySpotOnly:
				return order.GetDuration() == 0
			}
			return false
		},
		func(order *sonm.Order) bool {
			return devices.GetNetwork().GetNetFlags().ConverseImplication(order.GetNetflags())
		},
		func(order *sonm.Order) bool {
			counterpartyID := order.CounterpartyID.Unwrap()
			return counterpartyID == common.Address{} || counterpartyID == m.addr || counterpartyID == m.masterAddr
		},
		func(order *sonm.Order) bool {
			return deviceManager.Contains(*order.Benchmarks, *order.Netflags)
		},
	}
}

func (m *workerEngine) optimizationMethod(size int, log *zap.SugaredLogger) OptimizationMethod {
	return &GreedyLinearRegressionModel{
		classifier:      m.optimizationConfig.ClassifierFactory(m.log.Desugar()),
		exhaustionLimit: 128,
		log:             log,
	}
}

type OptimizationMethod interface {
	Optimize(knapsack *Knapsack, orders []*MarketOrder) error
}

type BruteForceModel struct{}

// GreedyLinearRegressionModel implements greedy knapsack optimization
// algorithm.
type GreedyLinearRegressionModel struct {
	classifier      OrderClassifier
	exhaustionLimit int
	log             *zap.SugaredLogger
}

func (m *GreedyLinearRegressionModel) Optimize(knapsack *Knapsack, orders []*MarketOrder) error {
	if len(orders) <= minNumOrders {
		return fmt.Errorf("not enough orders to perform optimization")
	}

	weightedOrders, err := m.classifier.Classify(orders)
	if err != nil {
		return fmt.Errorf("failed to classify orders: %v", err)
	}

	exhaustedCounter := 0
	for _, weightedOrder := range weightedOrders {
		if exhaustedCounter >= m.exhaustionLimit {
			break
		}

		order := weightedOrder.Order.Order

		m.log.Debugw("trying to put an order into resources pool",
			zap.Any("order", *weightedOrder.Order),
			zap.Float64("weight", weightedOrder.Weight),
			zap.String("price", order.Price.ToPriceString()),
			zap.Float64("predictedPrice", weightedOrder.PredictedPrice),
		)

		switch knapsack.Put(order) {
		case nil:
		case errExhausted:
			exhaustedCounter += 1
			continue
		default:
			return fmt.Errorf("failed to consume order: %v", err)
		}
	}

	return nil
}

type FittingFunc struct {
	Filters []func(order *sonm.Order) bool
}

func (m *FittingFunc) Filter(order *sonm.Order) bool {
	for _, filter := range m.Filters {
		if !filter(order) {
			return false
		}
	}

	return true
}

type Knapsack struct {
	manager *DeviceManager
	plans   []*sonm.AskPlan
}

func NewKnapsack(deviceManager *DeviceManager) *Knapsack {
	return &Knapsack{
		manager: deviceManager,
	}
}

func (m *Knapsack) Put(order *sonm.Order) error {
	resources, err := m.manager.Consume(*order.GetBenchmarks(), *order.GetNetflags())
	if err != nil {
		return err
	}

	resources.Network.NetFlags = order.GetNetflags()

	m.plans = append(m.plans, &sonm.AskPlan{
		Price:     &sonm.Price{PerSecond: order.Price},
		Duration:  &sonm.Duration{Nanoseconds: 1e9 * int64(order.Duration)},
		Resources: resources,
	})

	return nil
}
func (m *Knapsack) Price() *sonm.Price {
	return sonm.SumPrice(m.plans)
}

func (m *Knapsack) Plans() []*sonm.AskPlan {
	return m.plans
}
