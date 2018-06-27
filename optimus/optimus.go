package optimus

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
)

// Watch for current worker's status. Collect its devices.
// 	Worker MUST provide its network capabilities somehow.
// Fetch all bids.
// Optimize matrix looking for price per each benchmark unit.
// Map each bid to functional price.
// Then by comparing the price with expected price we can see what's underestimated.
// Sort.
// Filter.
// Pick.
type Optimus struct {
	cfg Config
	log *zap.SugaredLogger
}

func NewOptimus(cfg Config, log *zap.Logger) (*Optimus, error) {
	m := &Optimus{
		cfg: cfg,
		log: log.With(zap.String("source", "optimus")).Sugar(),
	}

	m.log.Debugw("configuring Optimus", zap.Any("config", cfg))

	return m, nil
}

func (m *Optimus) Run(ctx context.Context) error {
	m.log.Info("starting Optimus")
	defer m.log.Info("Optimus has been stopped")

	registry := newRegistry()
	defer registry.Close()

	dwh, err := registry.NewDWH(ctx, m.cfg.Marketplace.Endpoint, m.cfg.Marketplace.PrivateKey.Unwrap())
	if err != nil {
		return err
	}

	ordersScanner, err := newOrderScanner(dwh)
	if err != nil {
		return err
	}

	ordersSet := newOrdersSet()
	ordersControl, err := newOrdersControl(ordersScanner, m.cfg.Optimization.ClassifierFactory(m.log.Desugar()), ordersSet, m.log.Desugar())
	if err != nil {
		return err
	}

	wg := errgroup.Group{}
	wg.Go(func() error { return newManagedWatcher(ordersControl, m.cfg.Marketplace.Interval).Run(ctx) })

	loader := benchmarks.NewLoader(m.cfg.Benchmarks.URL)

	market, err := blockchain.NewAPI()
	if err != nil {
		return err
	}

	for addr, cfg := range m.cfg.Workers {
		ethAddr, err := addr.ETH()
		if err != nil {
			return err
		}

		masterAddr, err := market.Market().GetMaster(ctx, ethAddr)
		if err != nil {
			return err
		}

		worker, err := registry.NewWorkerManagement(ctx, m.cfg.Node.Endpoint, m.cfg.Node.PrivateKey.Unwrap())
		if err != nil {
			return err
		}

		control, err := newWorkerControl(cfg, ethAddr, masterAddr, worker, market.Market(), ordersSet, loader, m.log)
		if err != nil {
			return err
		}

		md := metadata.MD{
			util.WorkerAddressHeader: []string{addr.String()},
		}

		wg.Go(func() error {
			return newManagedWatcher(control, cfg.Epoch).Run(metadata.NewOutgoingContext(ctx, md))
		})
	}

	return wg.Wait()
}

// OrdersControl represents the marketplace watcher.
//
// This will pull all currently active orders from the marketplace.
type ordersControl struct {
	scanner    OrderScanner
	classifier OrderClassifier
	ordersSet  *ordersState
	log        *zap.SugaredLogger
}

func newOrdersControl(scanner OrderScanner, classifier OrderClassifier, orders *ordersState, log *zap.Logger) (*ordersControl, error) {
	m := &ordersControl{
		scanner:    scanner,
		classifier: classifier,
		ordersSet:  orders,
		log:        log.Sugar(),
	}

	return m, nil
}

func (m *ordersControl) OnRun() {
	m.log.Info("managing orders")
}

func (m *ordersControl) OnShutdown() {
	m.log.Info("stop managing orders")
}

func (m *ordersControl) Execute(ctx context.Context) {
	m.log.Debugf("pulling orders from the marketplace")

	now := time.Now()
	orders, err := m.scanner.All(ctx)
	if err != nil {
		m.log.Warnw("failed to pull orders from the marketplace", zap.Error(err))
		return
	}

	m.log.Infof("successfully pulled %d orders from the marketplace in %s", len(orders), time.Since(now))

	now = time.Now()
	weightedOrders, err := m.classifier.ClassifyExt(orders)
	if err != nil {
		m.log.Warnw("failed to classify orders", zap.Error(err))
		return
	}

	m.log.Infof("successfully classified %d orders in %s", len(weightedOrders.WeightedOrders), time.Since(now))
	m.ordersSet.Set(weightedOrders)
}

type workerControl struct {
	cfg             workerConfig
	addr            common.Address
	masterAddr      common.Address
	worker          sonm.WorkerManagementClient
	market          blockchain.MarketAPI
	benchmarkLoader benchmarks.Loader
	ordersSet       *ordersState
	log             *zap.SugaredLogger
}

func newWorkerControl(cfg workerConfig, addr, masterAddr common.Address, worker sonm.WorkerManagementClient, market blockchain.MarketAPI, orders *ordersState, benchmarkLoader benchmarks.Loader, log *zap.SugaredLogger) (*workerControl, error) {
	m := &workerControl{
		cfg:             cfg,
		addr:            addr,
		masterAddr:      masterAddr,
		worker:          worker,
		market:          market,
		benchmarkLoader: benchmarkLoader,
		ordersSet:       orders,
		log:             log.With(zap.Stringer("addr", addr)),
	}

	return m, nil
}

func (m *workerControl) OnRun() {
	m.log.Info("managing worker")
}

func (m *workerControl) OnShutdown() {
	m.log.Info("stop managing worker")
}

func (m *workerControl) Execute(ctx context.Context) {
	ordersClassification := m.ordersSet.Get()
	if ordersClassification == nil {
		m.log.Warn("not enough orders to perform optimization")
		return
	}

	orders := ordersClassification.WeightedOrders
	if len(orders) == 0 {
		m.log.Warn("not enough orders to perform optimization")
		return
	}

	m.log.Debugf("pulling worker plans")
	currentPlans, err := m.worker.AskPlans(ctx, &sonm.Empty{})
	if err != nil {
		m.log.Warnw("failed to pull worker plans", zap.Error(err))
		return
	}

	currentTotalPrice := calculateWorkerPriceMap(currentPlans.AskPlans).GetPerSecond()
	m.log.Debugw("successfully pulled worker plans",
		zap.Any("plans", *currentPlans),
		zap.String("Σ USD/s", currentTotalPrice.ToPriceString()),
	)

	cancellationCandidates := m.collectCancelCandidates(currentPlans.AskPlans)
	m.log.Debugw("cancellation candidates", zap.Any("plans", cancellationCandidates))

	m.log.Debugf("pulling worker devices")
	devices, err := m.worker.Devices(ctx, &sonm.Empty{})
	if err != nil {
		m.log.Warnw("failed to pull worker devices", zap.Error(err))
		return
	}

	m.log.Debugw("successfully pulled worker devices", zap.Any("devices", *devices))

	workerHardware := hardware.Hardware{
		CPU:     devices.CPU,
		GPU:     devices.GPUs,
		RAM:     devices.RAM,
		Network: devices.Network,
		Storage: devices.Storage,
	}
	freeResources := workerHardware.AskPlanResources()
	// Subtract plans except cancellation candidates. Doing so produces us a
	// new free(!) devices list.
	for id, plan := range currentPlans.AskPlans {
		_, ok := cancellationCandidates[id]
		if !ok {
			if err := freeResources.Sub(plan.Resources); err != nil {
				m.log.Warnw("failed to virtualize resource releasing", zap.Error(err))
				return
			}
		}
	}

	freeWorkerHardware, err := workerHardware.LimitTo(freeResources)
	if err != nil {
		m.log.Warnw("failed to limit virtual free hardware", zap.Error(err))
	}
	freeDevices := freeWorkerHardware.IntoProto()

	m.log.Debugw("successfully virtualized worker free devices", zap.Any("devices", *freeDevices))

	// Convert worker free devices into benchmarks set.
	bm := newBenchmarksFromDevices(freeDevices)
	freeWorkerBenchmarks, err := sonm.NewBenchmarks(bm[:])
	if err != nil {
		m.log.Warnw("failed to collect worker benchmarks", zap.Error(err))
		return
	}

	m.log.Infof("worker benchmarks: %v", strings.Join(strings.Fields(fmt.Sprintf("%v", freeWorkerBenchmarks.ToArray())), ", "))

	// Here we append cancellation candidate's orders to "orders" from
	// marketplace to be able to track their profitability.
	cancellationOrders, err := m.planOrders(ctx, cancellationCandidates)
	if err != nil {
		m.log.Warnw("failed to collect cancellation orders", zap.Error(err))
		return
	}

	for id, order := range cancellationOrders {
		marketOrder := &sonm.DWHOrder{
			Order: order,
		}
		predictedPrice, err := ordersClassification.Predictor.PredictPrice(marketOrder)
		if err != nil {
			m.log.Warnw("failed to predict cancellation order price", zap.Any("order", *order), zap.Error(err))
			return
		}

		price, _ := new(big.Float).SetInt(marketOrder.Order.Price.Unwrap()).Float64()

		orders = append(orders, WeightedOrder{
			Order:          marketOrder,
			Price:          price * priceMultiplier,
			PredictedPrice: math.Max(priceMultiplier, predictedPrice*priceMultiplier),
			Weight:         1.0,
			ID:             id,
		})

		ordersClassification.RecalculateWeightsAndSort(orders)
	}

	// Filter orders to have only orders that are subset of ours.
	matchedOrders := make([]WeightedOrder, 0, len(orders))
	for _, order := range orders {
		if order.Order.Order.OrderType != sonm.OrderType_BID {
			continue
		}

		switch m.cfg.OrderPolicy {
		case PolicySpotOnly:
			if order.Order.GetOrder().GetDuration() != 0 {
				continue
			}
		}

		if !freeDevices.GetNetwork().GetNetFlags().ConverseImplication(order.Order.Order.GetNetflags()) {
			continue
		}

		// Ignore filled with counterparty orders that are not created for us.
		counterpartyID := order.Order.Order.CounterpartyID.Unwrap()
		if !(counterpartyID == common.Address{} || counterpartyID == m.addr || counterpartyID == m.masterAddr) {
			continue
		}

		if !freeWorkerBenchmarks.Contains(order.Order.Order.Benchmarks) {
			continue
		}

		// No more than a single order with incoming network requirement should
		// be selected.
		// For this purpose we explicitly disable incoming network if such
		// order is matched.
		if order.Order.GetOrder().GetNetflags().GetIncoming() {
			freeDevices.GetNetwork().GetNetFlags().SetIncoming(false)
		}

		matchedOrders = append(matchedOrders, order)
	}

	m.log.Infof("found %d/%d matching orders", len(matchedOrders), len(orders))

	if len(matchedOrders) == 0 {
		m.log.Infof("no matching orders found")
		return
	}

	mapping, err := m.benchmarkLoader.Load(ctx)
	if err != nil {
		m.log.Warnw("failed to load benchmarks", zap.Error(err))
		return
	}

	deviceManager, err := newDeviceManager(devices, freeDevices, mapping)
	if err != nil {
		m.log.Warnw("failed to construct device manager", zap.Error(err))
		return
	}

	// Cut sell plans.
	var plans []*sonm.AskPlan
	exhaustedCounter := 0
	for _, weightedOrder := range matchedOrders {
		order := weightedOrder.Order.Order

		m.log.Debugw("trying to combine order into resources pool",
			zap.Any("order", *weightedOrder.Order),
			zap.Float64("weight", weightedOrder.Weight),
			zap.String("price", order.Price.ToPriceString()),
			zap.Float64("predictedPrice", weightedOrder.PredictedPrice),
		)
		// TODO: Hardcode. Not the best approach.
		if exhaustedCounter >= 100 {
			break
		}

		plan, err := deviceManager.Consume(*order.Benchmarks)
		switch err {
		case nil:
		case errExhausted:
			exhaustedCounter += 1
			continue
		default:
			m.log.Warnw("failed to consume order", zap.Error(err))
			return
		}

		m.log.Debugw("success")

		plan.Network.NetFlags = order.GetNetflags()

		plans = append(plans, &sonm.AskPlan{
			ID:        weightedOrder.ID,
			Price:     &sonm.Price{PerSecond: order.Price},
			Duration:  &sonm.Duration{Nanoseconds: 1e9 * int64(order.Duration)},
			Identity:  m.cfg.Identity,
			Resources: plan,
		})
	}

	pendingTotalPrice := calculateWorkerPrice(plans).GetPerSecond()
	m.log.Infow("successfully cut the following selling plans",
		zap.Any("size", len(plans)),
		zap.Any("plans", plans),
		zap.String("Σ USD/s", pendingTotalPrice.ToPriceString()),
	)

	// Compare total USD/s before and after. Cancel if the diff is more than
	// the threshold.
	priceDiff := big.NewInt(0).Sub(pendingTotalPrice.Unwrap(), currentTotalPrice.Unwrap())
	if big.NewInt(0).Sub(priceDiff, big.NewInt(10)).Sign() >= 0 {
		cancellationCandidates = m.filterCancellationCandidates(plans)

		m.log.Infow("cancelling plans", zap.Any("candidates", cancellationCandidates))

		if err := m.cancelPlans(ctx, cancellationCandidates); err != nil {
			m.log.Warnw("failed to cancel some plans", zap.Any("err", err))
		}
	}

	if m.cfg.DryRun {
		m.log.Debug("skipping creating ask-plans, because dry-run mode is active")
	} else {
		// Tell worker to create sell plans.
		for _, plan := range plans {
			id, err := m.worker.CreateAskPlan(ctx, plan)
			if err != nil {
				m.log.Warnw("failed to create sell plan", zap.Any("plan", *plan), zap.Error(err))
				continue
			}

			m.log.Infof("created sell plan %s", id.Id)
		}
	}
}

func (m *workerControl) cancelStalePlans(plans map[string]*sonm.AskPlan) {
	// Unimplemented.
}

func (m *workerControl) collectCancelCandidates(plans map[string]*sonm.AskPlan) map[string]*sonm.AskPlan {
	candidates := map[string]*sonm.AskPlan{}
	for id, plan := range plans {
		// Currently we can cancel spot orders without regret.
		if plan.GetDuration().Unwrap() == 0 {
			candidates[id] = plan
		}
	}

	return candidates
}

func (m *workerControl) planOrders(ctx context.Context, plans map[string]*sonm.AskPlan) (map[string]*sonm.Order, error) {
	orders := map[string]*sonm.Order{}
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(plans))

	for id, plan := range plans {
		go func(id string, plan *sonm.AskPlan) {
			defer wg.Done()

			order, err := m.market.GetOrderInfo(ctx, plan.OrderID.Unwrap())
			if err != nil {
				m.log.Warn("failed to get order", zap.String("planId", id), zap.Error(err))
				return
			}

			mu.Lock()
			defer mu.Unlock()

			orders[id] = order
		}(id, plan)
	}

	wg.Wait()

	if len(orders) != len(plans) {
		return nil, fmt.Errorf("failed to collect plan orders")
	}

	return orders, nil
}

func (m *workerControl) filterCancellationCandidates(plans []*sonm.AskPlan) map[string]*sonm.AskPlan {
	filtered := map[string]*sonm.AskPlan{}
	for _, plan := range plans {
		if len(plan.ID) > 0 {
			filtered[plan.ID] = plan
		}
	}

	return filtered
}

func (m *workerControl) cancelPlans(ctx context.Context, plans map[string]*sonm.AskPlan) map[string]error {
	errs := map[string]error{}
	for id := range plans {
		_, err := m.worker.RemoveAskPlan(ctx, &sonm.ID{Id: id})
		if err != nil {
			errs[id] = err
		}
	}

	fillRemainingWithErr := func(err error) {
		for id := range plans {
			if _, ok := errs[id]; !ok {
				errs[id] = err
			}
		}
	}

	// Wait for ask plans be REALLY removed.
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			fillRemainingWithErr(ctx.Err())
			return errs
		case <-timer.C:
			currentPlans, err := m.worker.AskPlans(ctx, &sonm.Empty{})
			if err != nil {
				fillRemainingWithErr(err)
				return errs
			}

			foundPending := false
			for id := range currentPlans.AskPlans {
				// Continue to wait if there are ask plans left.
				if _, ok := plans[id]; ok {
					foundPending = true
					break
				}
			}

			if !foundPending {
				fillRemainingWithErr(nil)
				return errs
			}
		}
	}
}

func calculateWorkerPrice(plans []*sonm.AskPlan) *sonm.Price {
	sum := big.NewInt(0)
	for _, plan := range plans {
		sum.Add(sum, plan.Price.PerSecond.Unwrap())
	}

	return &sonm.Price{PerSecond: sonm.NewBigInt(sum)}
}

func calculateWorkerPriceMap(plans map[string]*sonm.AskPlan) *sonm.Price {
	sum := big.NewInt(0)
	for _, plan := range plans {
		sum.Add(sum, plan.Price.PerSecond.Unwrap())
	}

	return &sonm.Price{PerSecond: sonm.NewBigInt(sum)}
}
