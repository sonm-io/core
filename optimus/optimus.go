package optimus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
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

		control, err := newWorkerControl(cfg, ethAddr, masterAddr, worker, ordersSet, loader, m.log)
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
	benchmarkLoader benchmarks.Loader
	ordersSet       *ordersState
	log             *zap.SugaredLogger
}

func newWorkerControl(cfg workerConfig, addr, masterAddr common.Address, worker sonm.WorkerManagementClient, orders *ordersState, benchmarkLoader benchmarks.Loader, log *zap.SugaredLogger) (*workerControl, error) {
	m := &workerControl{
		cfg:             cfg,
		addr:            addr,
		masterAddr:      masterAddr,
		worker:          worker,
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
	m.log.Debugf("pulling worker devices")
	devices, err := m.worker.Devices(ctx, &sonm.Empty{})
	if err != nil {
		m.log.Warnw("failed to pull worker devices", zap.Error(err))
		return
	}

	freeDevices, err := m.worker.FreeDevices(ctx, &sonm.Empty{})
	if err != nil {
		m.log.Warnw("failed to pull free worker devices", zap.Error(err))
		return
	}

	m.log.Debugw("successfully pulled worker devices", zap.Any("devices", *devices), zap.Any("freeDevices", *freeDevices))

	// Convert worker free devices into benchmarks set.
	bm := newBenchmarksFromDevices(freeDevices)
	freeWorkerBenchmarks, err := sonm.NewBenchmarks(bm[:])
	if err != nil {
		m.log.Warnw("failed to collect worker benchmarks", zap.Error(err))
		return
	}

	m.log.Infof("worker benchmarks: %v", strings.Join(strings.Fields(fmt.Sprintf("%v", freeWorkerBenchmarks.ToArray())), ", "))

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

		plan.Network.NetFlags = order.GetNetflags()

		plans = append(plans, &sonm.AskPlan{
			Price:     &sonm.Price{PerSecond: order.Price},
			Duration:  &sonm.Duration{Nanoseconds: 1e9 * int64(order.Duration)},
			Resources: plan,
		})
	}

	m.log.Infow("successfully cut the following selling plans", zap.Any("plans", plans))

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
