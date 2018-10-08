package optimus

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
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

type priceChange struct {
	InitialPrice *sonm.Price
	ChangedPrice *sonm.Price
}

func (m *optimizationInput) UpdateDealPrices(ctx context.Context, market blockchain.MarketAPI) (map[string]*priceChange, error) {
	changes := map[string]*priceChange{}

	mu := sync.Mutex{}

	wg, ctx := errgroup.WithContext(ctx)
	for id, plan := range m.Plans {
		dealID := plan.GetDealID()
		if dealID.IsZero() {
			continue
		}

		id := id
		plan := plan
		wg.Go(func() error {
			deal, err := market.GetDealInfo(ctx, dealID.Unwrap())
			if err != nil {
				return fmt.Errorf("failed to get deal `%s` for `%s`: %v", dealID.Unwrap().String(), id, err)
			}

			if plan.Price.GetPerSecond().Cmp(deal.GetPrice()) == 0 {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()

			changes[id] = &priceChange{
				InitialPrice: plan.Price,
				ChangedPrice: &sonm.Price{PerSecond: deal.GetPrice()},
			}

			plan.Price = &sonm.Price{PerSecond: deal.GetPrice()}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return changes, nil
}

func (m *optimizationInput) Price() *sonm.Price {
	var plans []*sonm.AskPlan
	for _, plan := range m.Plans {
		plans = append(plans, plan)
	}

	return sonm.SumPrice(plans)
}

func (m *optimizationInput) freeDevices(removalVictims map[string]*sonm.AskPlan) (*sonm.DevicesReply, error) {
	devices := proto.Clone(m.Devices).(*sonm.DevicesReply)
	workerHardware := hardware.Hardware{
		CPU:     devices.CPU,
		GPU:     devices.GPUs,
		RAM:     devices.RAM,
		Network: devices.Network,
		Storage: devices.Storage,
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

type Blacklist interface {
	Update(ctx context.Context) error
	IsAllowed(addr common.Address) bool
}

type workerEngine struct {
	cfg *workerConfig
	log *zap.SugaredLogger

	addr             common.Address
	masterAddr       common.Address
	blacklist        Blacklist
	market           blockchain.MarketAPI
	marketCache      MarketScanner
	worker           WorkerManagementClientExt
	benchmarkMapping benchmarks.Mapping

	tagger *Tagger
}

func newWorkerEngine(cfg *workerConfig, addr, masterAddr common.Address, blacklist Blacklist, worker WorkerManagementClientAPI, market blockchain.MarketAPI, marketCache MarketScanner, benchmarkMapping benchmarks.Mapping, tagger *Tagger, log *zap.SugaredLogger) (*workerEngine, error) {
	if cfg.DryRun {
		log.Infof("activated dry-run mode for this worker")
		worker = NewReadOnlyWorker(worker)
		log = log.With(zap.String("mode", "dry-run"))
	}

	if cfg.Simulation != nil {
		log.Infof("activated simulation mode for this worker")

		var err error
		marketCache, err = NewPredefinedMarketCache(cfg.Simulation.Orders, market)
		if err != nil {
			return nil, err
		}

		worker = NewReadOnlyWorker(worker)
		log = log.With(zap.String("mode", "simulation"))
	}

	m := &workerEngine{
		cfg: cfg,
		log: log.With(zap.Stringer("addr", addr)),

		addr:             addr,
		masterAddr:       masterAddr,
		blacklist:        blacklist,
		market:           market,
		marketCache:      marketCache,
		worker:           &workerManagementClientExt{worker},
		benchmarkMapping: benchmarkMapping,

		tagger: tagger,
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
	maintenance, err := m.worker.NextMaintenance(ctx, &sonm.Empty{})
	if err != nil {
		return fmt.Errorf("failed to get maintenance: %v", err)
	}
	if time.Since(maintenance.Unix()) >= 0 {
		return fmt.Errorf("worker is on the maintenance")
	}

	if err := m.blacklist.Update(ctx); err != nil {
		return fmt.Errorf("failed to update blacklist: %v", err)
	}

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

	priceChanges, err := input.UpdateDealPrices(ctx, m.market)
	if err != nil {
		return fmt.Errorf("failed to update deal prices: %v", err)
	}

	m.log.Infow("synchronized actual prices with the marketplace", zap.Any("changes", priceChanges))

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
	// Note, that this can return error when some victim plans did not place
	// their orders on the marketplace.
	// Either this can be temporary or worker's critical failure, network for
	// example.
	// The best we can do here is to return and try again in the next epoch.
	virtualFreeOrders, err := m.ordersForPlans(ctx, victimPlans)
	if err != nil {
		return fmt.Errorf("failed to collect orders for victim plans: %v", err)
	}

	m.log.Debugw("pulled victim orders", zap.Any("orders", virtualFreeOrders))

	// Extended orders set, with added currently executed orders.
	extOrders := append(append([]*MarketOrder{}, input.Orders...), virtualFreeOrders...)

	var naturalKnapsack, virtualKnapsack *Knapsack

	wg := errgroup.Group{}
	wg.Go(func() error {
		m.log.Info("optimizing using natural free devices")
		knapsack, err := m.optimize(input.Devices, naturalFreeDevices, input.Orders, nil, m.log.With(zap.String("optimization", "natural")))
		if err != nil {
			return err
		}

		naturalKnapsack = knapsack
		return nil
	})
	wg.Go(func() error {
		m.log.Info("optimizing using virtual free devices")
		knapsack, err := m.optimize(input.Devices, virtualFreeDevices, extOrders, virtualFreeOrders, m.log.With(zap.String("optimization", "virtual")))
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

	// Compare total USD/s before and after. Remove some plans if the diff is
	// more than the threshold.
	swingTime := m.cfg.PriceThreshold.Exceeds(virtualKnapsack.Price().GetPerSecond().Unwrap(), input.Price().GetPerSecond().Unwrap())

	var winners []*sonm.AskPlan
	var victims []*sonm.AskPlan
	if swingTime {
		m.log.Info("using replacement strategy")

		create, remove, ignore := splitPlans(input.Plans, virtualKnapsack.Plans())
		m.log.Infow("ignoring already existing plans", zap.Any("plans", ignore))
		m.log.Infow("removing plans", zap.Any("plans", remove))
		m.log.Infow("creating plans", zap.Any("plans", create))

		winners = create
		victims = remove
	} else {
		m.log.Info("using appending strategy")
		winners = naturalKnapsack.Plans()
	}

	if len(winners) == 0 {
		return fmt.Errorf("no plans found")
	}

	victimIDs := make([]string, 0, len(victims))
	for _, plan := range victims {
		victimIDs = append(victimIDs, plan.ID)
	}
	if err := m.worker.RemoveAskPlans(ctx, victimIDs); err != nil {
		return err
	}

	for _, plan := range winners {
		// Extract the order ID for whose the selling plan is created.
		orderID := plan.GetOrderID()

		// Then we need to clean this, because otherwise worker rejects such request.
		plan.OrderID = nil
		plan.Identity = m.cfg.Identity
		plan.Tag = m.tagger.Tag()

		id, err := m.worker.CreateAskPlan(ctx, plan)
		if err != nil {
			m.log.Warnw("failed to create sell plan", zap.Any("plan", *plan), zap.Error(err))
			continue
		}

		m.log.Infof("created sell plan %s for %s order", id.Id, orderID.String())
	}

	return nil
}

func splitPlans(plans map[string]*sonm.AskPlan, candidates []*sonm.AskPlan) (create, remove, ignore []*sonm.AskPlan) {
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

	create, remove = removeDuplicates(create, remove)

	return create, remove, ignore
}

func removeDuplicates(create, remove []*sonm.AskPlan) ([]*sonm.AskPlan, []*sonm.AskPlan) {
	return removeDuplicatesL(create, remove)
}

// Here we find orders in the creation list that are equal with orders in
// the removal list. This is required to not to replace existing orders
// with the same ones somehow appeared in the marketplace.
//
// The algorithm works as the follows:
// Given: [c0, c1, c2, c3, c4] [r0, r1, r2, r3]
// Where: c1==r3, c3==r2, c4==r2.
// Then:
//	                                     [c0, c1, c2, c3, c4] [r0, r1, r2, r3]
//	c0!=r0, c0!=r1, c0!=r2, c0!=r3 ->    [c0, c1, c2, c3, c4] [r0, r1, r2, r3]
//	c1!=r0, c1!=r1, c1!=r2, c1==r3(!) -> [c0, c2, c3, c4]     [r0, r1, r2]
//	c2!=r0, c2!=r1, c2!=r2 ->            [c0, c2, c3, c4]     [r0, r1, r2]
//	c3!=r0, c3!=r1, c3==r2(!) ->         [c0, c2, c4]         [r0, r1]
//	c4!=r0, c4!=r1 ->                    [c0, c2, c4]         [r0, r1]
func removeDuplicatesL(create, remove []*sonm.AskPlan) ([]*sonm.AskPlan, []*sonm.AskPlan) {
	sort.Slice(create, func(i, j int) bool {
		return create[i].Price.PerSecond.Cmp(create[j].Price.PerSecond) < 0
	})
	sort.Slice(remove, func(i, j int) bool {
		return remove[i].Price.PerSecond.Cmp(remove[j].Price.PerSecond) < 0
	})

	type Eq struct {
		i, j int
	}

	i, j := 0, 0
	var eq []Eq
	for {
		if i >= len(create) {
			break
		}
		if j >= len(remove) {
			break
		}
		if planEq(create[i], remove[j]) {
			eq = append(eq, Eq{i: i, j: j})
			i++
			j++
			continue
		}
		if create[i].Price.PerSecond.Cmp(remove[j].Price.PerSecond) < 0 {
			i++
		} else {
			j++
		}
	}
	if len(eq) == 0 {
		return create, remove
	}

	eqIdx := 0
	newCreate := make([]*sonm.AskPlan, 0, len(create))
	for idx, plan := range create {
		if eqIdx < len(eq) {
			if idx == eq[eqIdx].i {
				eqIdx++
				continue
			}
		}
		newCreate = append(newCreate, plan)
	}

	eqIdx = 0
	newRemove := make([]*sonm.AskPlan, 0, len(remove))
	for idx, plan := range remove {
		if eqIdx < len(eq) {
			if idx == eq[eqIdx].j {
				eqIdx++
				continue
			}
		}
		newRemove = append(newRemove, plan)
	}

	return newCreate, newRemove
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
				return fmt.Errorf("failed to get order `%s` for `%s`: %v", plan.OrderID.Unwrap().String(), id, err)
			}

			mu.Lock()
			defer mu.Unlock()
			orders = append(orders, &MarketOrder{
				Order:     order,
				CreatedTS: sonm.CurrentTimestamp(),
			})

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (m *workerEngine) optimize(devices, freeDevices *sonm.DevicesReply, orders, extra []*MarketOrder, log *zap.SugaredLogger) (*Knapsack, error) {
	deviceManager, err := newDeviceManager(devices, freeDevices, m.benchmarkMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to construct device manager: %v", err)
	}

	matchedOrders := m.matchingOrders(deviceManager, devices, orders, extra, log)
	log.Infof("found %d/%d matching orders", len(matchedOrders), len(orders))

	if len(matchedOrders) == 0 {
		log.Infof("no matching orders found")
		return NewKnapsack(deviceManager), nil
	}

	now := time.Now()
	knapsack := NewKnapsack(deviceManager)
	if err := m.optimizationMethod(orders, matchedOrders, log).Optimize(knapsack, matchedOrders); err != nil {
		return nil, err
	}

	log.Infof("optimized %d orders in %s", len(matchedOrders), time.Since(now))

	return knapsack, nil
}

// MatchingOrders filters the given orders to have only orders that are subset
// of ours.
func (m *workerEngine) matchingOrders(deviceManager *DeviceManager, devices *sonm.DevicesReply, orders, extra []*MarketOrder, log *zap.SugaredLogger) []*MarketOrder {
	matchedOrders := make([]*MarketOrder, 0, len(orders))

	filter := FittingFunc{
		Filters: m.filtersErr(deviceManager, devices),
	}

	for _, order := range extra {
		matchedOrders = append(matchedOrders, order)
	}

	for _, order := range orders {
		if err := filter.Filter(order.GetOrder()); err != nil {
			if m.cfg.VerboseLog {
				log.Debugf("exclude order %s from matching: %v", order.GetOrder().GetId(), err)
			}
			continue
		}

		matchedOrders = append(matchedOrders, order)
	}

	return matchedOrders
}

func (m *workerEngine) filtersErr(deviceManager *DeviceManager, devices *sonm.DevicesReply) []func(order *sonm.Order) error {
	return []func(order *sonm.Order) error{
		func(order *sonm.Order) error {
			if order.OrderType == sonm.OrderType_BID {
				return nil
			}

			return fmt.Errorf("expected order type %s, actual: %s", sonm.OrderType_BID, order.OrderType)
		},
		func(order *sonm.Order) error {
			switch m.cfg.OrderPolicy {
			case PolicySpotOnly:
				if order.GetDuration() == 0 {
					return nil
				}
				return fmt.Errorf("expected order duration 0, actual: %d", order.GetDuration())
			}

			return fmt.Errorf("unknown order policy: %s", m.cfg.OrderPolicy.String())
		},
		func(order *sonm.Order) error {
			if m.blacklist.IsAllowed(order.GetAuthorID().Unwrap()) {
				return nil
			}

			return fmt.Errorf("order is in blacklist")
		},
		func(order *sonm.Order) error {
			if devices.GetNetwork().GetNetFlags().ConverseImplication(order.GetNetflags()) {
				return nil
			}

			return fmt.Errorf("netflags mismatch")
		},
		func(order *sonm.Order) error {
			counterpartyID := order.CounterpartyID.Unwrap()
			if (counterpartyID == common.Address{} || counterpartyID == m.addr || counterpartyID == m.masterAddr) {
				return nil
			}

			return fmt.Errorf("counterparty mismatch")
		},
		func(order *sonm.Order) error {
			if order.IdentityLevel <= m.cfg.Identity {
				return nil
			}

			return fmt.Errorf("expected minimum identity %s, actual %s", m.cfg.Identity, order.IdentityLevel)
		},
		func(order *sonm.Order) error {
			if deviceManager.Contains(*order.Benchmarks, *order.Netflags) {
				return nil
			}

			return fmt.Errorf("benchmarks mismatch: free %d, actual %d", deviceManager.freeBenchmarks, order.Benchmarks.GetValues())
		},
	}
}

func (m *workerEngine) optimizationMethod(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	return m.cfg.Optimization.Model.Create(orders, matchedOrders, log)
}

type OptimizationMethodFactory interface {
	Config() interface{}
	Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod
}

type defaultPredictionOptimizationMethodFactory struct{}

func (m *defaultPredictionOptimizationMethodFactory) Config() interface{} {
	return m
}

func (m *defaultPredictionOptimizationMethodFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	if len(matchedOrders) < 64 {
		return &BranchBoundModel{
			Log: log.With(zap.String("model", "BBM")),
		}
	}

	return &BatchModel{
		Methods: []OptimizationMethod{
			&GreedyLinearRegressionModel{
				orders: orders,
				regression: &regressionClassifier{
					model: &SCAKKTModel{
						MaxIterations: 1e7,
						Log:           log,
					},
				},
				exhaustionLimit: 128,
				log:             log.With(zap.String("model", "LLS")),
			},
			&GeneticModel{
				NewGenomeLab:   NewPackedOrdersNewGenome,
				PopulationSize: 256,
				MaxGenerations: 32,
				MaxAge:         5 * time.Minute,
				Log:            log.With(zap.String("model", "GMP")),
			},
			&GeneticModel{
				NewGenomeLab:   NewDecisionOrdersNewGenome,
				PopulationSize: 512,
				MaxGenerations: 24,
				MaxAge:         5 * time.Minute,
				Log:            log.With(zap.String("model", "GMD")),
			},
		},
		Log: log,
	}
}

type defaultOptimizationMethodFactory struct{}

func (m *defaultOptimizationMethodFactory) Config() interface{} {
	return m
}

func (m *defaultOptimizationMethodFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	if len(matchedOrders) < 64 {
		return &BranchBoundModel{
			Log: log.With(zap.String("model", "BBM")),
		}
	}

	return &BatchModel{
		Methods: []OptimizationMethod{
			&GreedyLinearRegressionModel{
				orders: orders,
				regression: &regressionClassifier{
					model: &SCAKKTModel{
						MaxIterations: 1e7,
						Log:           log,
					},
				},
				exhaustionLimit: 128,
				log:             log.With(zap.String("model", "LLS")),
			},
			&GeneticModel{
				NewGenomeLab:   NewPackedOrdersNewGenome,
				PopulationSize: 256,
				MaxGenerations: 128,
				MaxAge:         5 * time.Minute,
				Log:            log.With(zap.String("model", "GMP")),
			},
			&GeneticModel{
				NewGenomeLab:   NewDecisionOrdersNewGenome,
				PopulationSize: 512,
				MaxGenerations: 64,
				MaxAge:         5 * time.Minute,
				Log:            log.With(zap.String("model", "GMD")),
			},
		},
		Log: log,
	}
}

func optimizationFactory(ty string) OptimizationMethodFactory {
	switch ty {
	case "batch":
		return &BatchModelFactory{}
	case "greedy":
		return &GreedyLinearRegressionModelFactory{}
	case "genetic":
		return &GeneticModelFactory{}
	case "branch_bound":
		return &BranchBoundModelFactory{}
	default:
		return nil
	}
}

type optimizationMethodFactory struct {
	OptimizationMethodFactory
}

func (m *optimizationMethodFactory) MarshalYAML() (interface{}, error) {
	return m.Config(), nil
}

func (m *optimizationMethodFactory) UnmarshalYAML(unmarshal func(interface{}) error) error {
	ty, err := typeofInterface(unmarshal)
	if err != nil {
		return err
	}

	factory := optimizationFactory(ty)
	if factory == nil {
		return fmt.Errorf("unknown optimization model: %s", ty)
	}

	cfg := factory.Config()
	if err := unmarshal(cfg); err != nil {
		return err
	}

	m.OptimizationMethodFactory = factory

	return nil
}

type OptimizationMethod interface {
	Optimize(knapsack *Knapsack, orders []*MarketOrder) error
}

type FittingFunc struct {
	Filters []func(order *sonm.Order) error
}

func (m *FittingFunc) Filter(order *sonm.Order) error {
	for _, filter := range m.Filters {
		if err := filter(order); err != nil {
			return err
		}
	}

	return nil
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

func planEq(a, b *sonm.AskPlan) bool {
	return a.GetResources().Eq(b.GetResources()) &&
		a.GetPrice().GetPerSecond().Cmp(b.GetPrice().GetPerSecond()) == 0 &&
		a.GetDuration().Unwrap() == b.GetDuration().Unwrap()
}
