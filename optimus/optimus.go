package optimus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
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
		log: log.Sugar(),
	}

	m.log.Debugw("configuring Optimus", zap.Any("config", cfg))

	return m, nil
}

func (m *Optimus) Run(ctx context.Context) error {
	m.log.Info("starting Optimus")
	defer m.log.Info("Optimus has been stopped")

	certificate, TLSConfig, err := util.NewHitlessCertRotator(ctx, m.cfg.PrivateKey.Unwrap())
	if err != nil {
		return err
	}
	defer certificate.Close()

	credentials := util.NewTLS(TLSConfig)

	newWorker := func(ctx context.Context, addr auth.Addr) (sonm.WorkerManagementClient, error) {
		conn, err := xgrpc.NewClient(ctx, addr.String(), credentials)
		if err != nil {
			return nil, err
		}

		return sonm.NewWorkerManagementClient(conn), nil
	}

	ordersSet := newOrdersSet()

	conn, err := xgrpc.NewClient(ctx, m.cfg.Marketplace.Endpoint.String(), credentials)
	if err != nil {
		return err
	}

	ordersScanner, err := newOrderScanner(sonm.NewDWHClient(conn))
	if err != nil {
		return err
	}

	ordersControl, err := newOrdersControl(ordersScanner, m.cfg.Optimization.Classifier(), ordersSet, m.log.Desugar())
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

	for _, addr := range m.cfg.Workers {
		ethAddr, err := addr.ETH()
		if err != nil {
			return err
		}

		masterAddr, err := market.Market().GetMaster(ctx, ethAddr)
		if err != nil {
			return err
		}

		worker, err := newWorker(ctx, addr)
		if err != nil {
			return err
		}

		control, err := newWorkerControl(ethAddr, masterAddr, worker, ordersSet, loader, m.log)
		if err != nil {
			return err
		}

		wg.Go(func() error {
			return newManagedWatcher(control, 60*time.Second).Run(ctx)
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
	ordersSet  *ordersSet
	log        *zap.SugaredLogger
}

func newOrdersControl(scanner OrderScanner, classifier OrderClassifier, orders *ordersSet, log *zap.Logger) (*ordersControl, error) {
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
	weightedOrders, err := m.classifier.Classify(orders)
	if err != nil {
		m.log.Warnw("failed to classify orders", zap.Error(err))
		return
	}

	m.log.Infof("successfully classified %d orders in %s", len(weightedOrders), time.Since(now))
	m.ordersSet.Set(weightedOrders)
}

type workerControl struct {
	addr            common.Address
	masterAddr      common.Address
	worker          sonm.WorkerManagementClient
	benchmarkLoader benchmarks.Loader
	ordersSet       *ordersSet
	log             *zap.SugaredLogger
}

func newWorkerControl(addr, masterAddr common.Address, worker sonm.WorkerManagementClient, orders *ordersSet, benchmarkLoader benchmarks.Loader, log *zap.SugaredLogger) (*workerControl, error) {
	m := &workerControl{
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

	m.log.Infof("worker benchmarks: %s", strings.Join(strings.Fields(fmt.Sprintf("%v", freeWorkerBenchmarks)), ", "))

	orders := m.ordersSet.Get()
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

		if order.Order.Order.GetNetflags().GetOverlay() && !devices.Network.GetNetFlags().GetOverlay() {
			continue
		}
		if order.Order.Order.GetNetflags().GetOutbound() && !devices.Network.GetNetFlags().GetOutbound() {
			continue
		}
		if order.Order.Order.GetNetflags().GetIncoming() && !devices.GetNetwork().GetNetFlags().GetIncoming() {
			continue
		}

		// Ignore filled with counterparty orders that are not created for us.
		counterpartyID := order.Order.Order.CounterpartyID.Unwrap()
		if !(counterpartyID == common.Address{} || counterpartyID == m.addr || counterpartyID == m.masterAddr) {
			continue
		}

		if freeWorkerBenchmarks.Contains(order.Order.Order.Benchmarks) {
			matchedOrders = append(matchedOrders, order)
		}
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
	for _, order := range matchedOrders {
		m.log.Debugw("trying to combine order into resources pool", zap.Any("order", *order.Order.Order))
		// TODO: Hardcode. Not the best approach.
		if exhaustedCounter >= 100 {
			break
		}

		plan, err := deviceManager.Consume(*order.Order.Order.Benchmarks)
		switch err {
		case nil:
		case errExhausted:
			exhaustedCounter += 1
			continue
		default:
			m.log.Warnw("failed to consume order", zap.Error(err))
			return
		}

		plan.Network.NetFlags = order.Order.Order.GetNetflags()

		plans = append(plans, &sonm.AskPlan{
			Price:     &sonm.Price{PerSecond: order.Order.Order.Price},
			Duration:  &sonm.Duration{Nanoseconds: 1e9 * int64(order.Order.Order.Duration)},
			Resources: plan,
		})
	}

	m.log.Infof("cut the following selling plans: %v", plans)

	// Tell worker to create sell plans.
	for _, plan := range plans {
		id, err := m.worker.CreateAskPlan(ctx, plan)
		if err != nil {
			m.log.Warnw("failed to create sell plan", zap.Error(err))
			continue
		}

		m.log.Infof("created sell plan %s", id.Id)
	}
}
