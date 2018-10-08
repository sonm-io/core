package connor

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sonm-io/core/connor/antifraud"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/connor/types"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	concurrency      = 10
	maxRetryCount    = 5
	taskRestartCount = 3
)

type engine struct {
	log       *zap.Logger
	cfg       *Config
	antiFraud antifraud.AntiFraud
	ethAddr   common.Address

	market        sonm.MarketClient
	deals         sonm.DealManagementClient
	tasks         sonm.TaskManagementClient
	priceProvider price.Provider
	corderFactory types.CorderFactory
	dealFactory   types.DealFactory

	ordersCreateChan chan *types.Corder
	orderCancelChan  chan *types.CorderCancelTuple
	state            *state
}

var (
	activeDealsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sonm_deals_active",
		Help: "Number of currently processing deals",
	})

	activeOrdersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "sonm_orders_active",
		Help: "Number of currently processing orders",
	})

	createdOrdersCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sonm_orders_created",
		Help: "Number of orders that were sent to marker",
	})

	replacedOrdersCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sonm_orders_replaced",
		Help: "Number of orders that were re-created on marker because of price deviation",
	})

	adoptedOrdersCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sonm_orders_adopted",
		Help: "Number of orders restored in runtime",
	})

	adoptedDealsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sonm_deals_adopted",
		Help: "Number of deals restored in runtime",
	})
)

func init() {
	prometheus.MustRegister(activeDealsGauge)
	prometheus.MustRegister(activeOrdersGauge)
	prometheus.MustRegister(createdOrdersCounter)
	prometheus.MustRegister(replacedOrdersCounter)
	prometheus.MustRegister(adoptedOrdersCounter)
	prometheus.MustRegister(adoptedDealsCounter)
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*engine, error) {
	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("cannot load eth keys: %v", err)
	}

	benchList, err := benchmarks.NewBenchmarksList(ctx, cfg.BenchmarkList)
	if err != nil {
		return nil, fmt.Errorf("cannot load benchmark list: %v", err)
	}

	// perform extra config validation using external list of required benchmarks
	if err := cfg.validateBenchmarks(benchList); err != nil {
		return nil, fmt.Errorf("benchmarks validation failed: %v", err)
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cert TLS config: %v", err)
	}

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(key.PublicKey))
	cc, err := xgrpc.NewClient(ctx, cfg.Node.Endpoint.String(), creds)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to node: %v", err)
	}

	return &engine{
		cfg:     cfg,
		log:     log,
		state:   NewState(log),
		ethAddr: crypto.PubkeyToAddress(key.PublicKey),

		priceProvider: cfg.getBackends().priceProvider,
		corderFactory: cfg.getBackends().corderFactory,
		dealFactory:   cfg.getBackends().dealFactory,

		market:    sonm.NewMarketClient(cc),
		deals:     sonm.NewDealManagementClient(cc),
		tasks:     sonm.NewTaskManagementClient(cc),
		antiFraud: antifraud.NewAntiFraud(cfg.AntiFraud, log, cfg.getBackends().processorFactory, cfg.getBackends().dealFactory, cc),

		ordersCreateChan: make(chan *types.Corder, concurrency),
		orderCancelChan:  make(chan *types.CorderCancelTuple, concurrency),
	}, nil
}

func (e *engine) Serve(ctx context.Context) error {
	defer e.close()

	e.log.Info("starting engine",
		zap.Int("concurrency", concurrency),
		zap.Any("config", *e.cfg))

	// load initial state from external sources
	if err := e.loadInitialData(ctx); err != nil {
		return fmt.Errorf("failed to load initial data: %v", err)
	}

	e.log.Debug("price",
		zap.String("value", e.priceProvider.GetPrice().String()),
		zap.Float64("margin", e.cfg.Market.PriceControl.Marginality))

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return e.startPriceTracking(ctx)
	})
	wg.Go(func() error {
		return e.start(ctx)
	})
	wg.Go(func() error {
		return e.antiFraud.Run(ctx)
	})

	if err := e.restoreMarketState(ctx); err != nil {
		return fmt.Errorf("failed to restore market state: %v", err)
	}

	wg.Go(func() error {
		return e.waitForExternalUpdates(ctx)
	})

	return wg.Wait()
}

func (e *engine) CreateOrder(bid *types.Corder) {
	e.state.AddQueuedOrder(bid)
	e.ordersCreateChan <- bid
}

func (e *engine) CancelOrder(order *types.Corder) {
	e.orderCancelChan <- types.NewCorderCancelTuple(order)
}

func (e *engine) RestoreOrder(ctx context.Context, order *types.Corder) {
	e.log.Debug("restoring order", zap.String("id", order.Order.GetId().Unwrap().String()))
	go e.waitForDeal(ctx, order)
}

func (e *engine) RestoreDeal(ctx context.Context, deal *sonm.Deal) {
	e.log.Debug("restoring deal", zap.String("id", deal.GetId().Unwrap().String()))
	go e.processDeal(ctx, deal)
}

func (e *engine) sendOrderToMarket(ctx context.Context, bid *sonm.BidOrder) (*types.Corder, error) {
	e.log.Debug("creating order on market",
		zap.String("price", bid.GetPrice().GetPerSecond().Unwrap().String()),
		zap.Any("benchmarks", bid.Resources.GetBenchmarks()))

	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	order, err := e.market.CreateOrder(ctx, bid)
	if err != nil {
		return nil, err
	}

	return e.corderFactory.FromOrder(order), nil
}

func (e *engine) processOrderCreate(ctx context.Context) {
	for bid := range e.ordersCreateChan {
		// set actual order price just before sending it to the Market
		hashRate := big.NewInt(0).SetUint64(bid.GetHashrate())
		bid.Price = sonm.NewBigInt(big.NewInt(0).Mul(e.priceProvider.GetPrice(), hashRate))

		created, err := e.sendOrderToMarket(ctx, bid.AsBID())
		if err != nil {
			e.log.Warn("cannot place order, retrying", zap.Error(err))
			e.CreateOrder(bid)
			continue
		}
		e.log.Debug("order successfully created", zap.String("order_id", created.GetId().Unwrap().String()))
		createdOrdersCounter.Inc()
		e.state.DeleteQueuedOrder(created)
		go e.waitForDeal(ctx, created)
	}
}

func (e *engine) processOrderCancel(ctx context.Context) {
	for tuple := range e.orderCancelChan {
		// prometheus counter?
		e.log.Debug("cancelling order",
			zap.String("order_id", tuple.Corder.GetId().Unwrap().String()))

		if err := e.cancelOrder(ctx, tuple.Corder.GetId()); err != nil {
			e.log.Warn("cannot cancel order", zap.Error(err),
				zap.Duration("retry_after", tuple.Delay),
				zap.String("order_id", tuple.Corder.GetId().Unwrap().String()))

			go func(tup *types.CorderCancelTuple) {
				time.Sleep(tup.Delay)
				e.orderCancelChan <- tup.WithIncreasedDelay()
			}(tuple)

			continue
		}

		e.state.DeleteActiveOrder(tuple.Corder)
		e.log.Debug("order cancelled", zap.String("order_id", tuple.Corder.GetId().Unwrap().String()))
	}
}

func (e *engine) cancelOrder(ctx context.Context, id *sonm.BigInt) error {
	order, err := e.getOrderByID(ctx, id.Unwrap().String())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	if order.GetOrderStatus() == sonm.OrderStatus_ORDER_ACTIVE {
		_, err := e.market.CancelOrder(ctx, &sonm.ID{Id: id.Unwrap().String()})
		return err
	}

	return nil
}

func (e *engine) getOrderByID(ctx context.Context, id string) (*types.Corder, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	order, err := e.market.GetOrderByID(ctx, &sonm.ID{Id: id})
	if err != nil {
		return nil, err
	}

	return e.corderFactory.FromOrder(order), nil
}

func (e *engine) waitForDeal(ctx context.Context, order *types.Corder) {
	e.state.AddActiveOrder(order)
	activeOrdersGauge.Inc()

	id := order.GetId().Unwrap().String()
	log := e.log.With(zap.String("order_id", id))

	t := util.NewImmediateTicker(e.cfg.Engine.OrderWatchInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			actualPrice := e.priceProvider.GetPrice()
			if order.IsReplaceable(actualPrice, e.cfg.Market.PriceControl.OrderReplaceThreshold) {
				pricePerOrder := big.NewInt(0).Mul(actualPrice, big.NewInt(int64(order.GetHashrate())))
				log.Named("price-deviation").Info("we can replace order with more profitable one",
					zap.Uint64("benchmark", order.GetHashrate()),
					zap.String("actual_price", pricePerOrder.String()),
					zap.String("current_price", order.GetPrice().Unwrap().String()))

				e.CancelOrder(order)
				replacedOrdersCounter.Inc()
				e.CreateOrder(order)
				return
			}

			deal, err := e.checkOrderForDealOnce(ctx, log, id)
			if err != nil {
				continue
			}

			e.CreateOrder(order)
			if deal != nil {
				if ok := e.state.AddDeal(e.dealFactory.FromDeal(deal)); ok {
					go e.processDeal(ctx, deal)
				}
			}

			return
		}
	}
}

func (e *engine) checkOrderForDealOnce(ctx context.Context, log *zap.Logger, orderID string) (*sonm.Deal, error) {
	ord, err := e.getOrderByID(ctx, orderID)
	if err != nil {
		log.Warn("cannot get order info from market", zap.Error(err))
		return nil, err
	}

	if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
		activeOrdersGauge.Dec()
		e.state.DeleteActiveOrder(ord)
		// TODO: (in a separate PR) check for `one_shot` parameter, do not re-create order.
		log.Info("order becomes inactive, looking for related deal")

		if ord.GetDealID().IsZero() {
			log.Debug("order have no deal, probably order is cancelled by hand")
			return nil, nil
		}

		deal, err := e.deals.Status(ctx, ord.GetDealID())
		if err != nil {
			log.Warn("cannot get deal info from market", zap.Error(err),
				zap.String("deal_id", ord.GetDealID().Unwrap().String()))
			return nil, err
		}

		return deal.GetDeal(), nil
	}

	return nil, fmt.Errorf("order have no deal")
}

func (e *engine) processDeal(ctx context.Context, deal *sonm.Deal) {
	activeDealsGauge.Inc()
	defer activeDealsGauge.Dec()

	dealID := deal.GetId().Unwrap().String()
	log := e.log.Named("process-deal").With(
		zap.String("deal_id", dealID),
		zap.String("supplier_id", deal.GetSupplierID().Unwrap().Hex()))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Debug("start deal processing")
	defer log.Debug("stop deal processing")
	defer e.state.DeleteDeal(e.dealFactory.FromDeal(deal))

	e.antiFraud.DealOpened(deal)
	defer e.antiFraud.FinishDeal(ctx, deal, antifraud.AllChecks)

	taskID, err := e.restoreTasks(ctx, log, deal.GetId())
	if err != nil {
		log.Warn("cannot restore tasks", zap.Error(err))
		return
	}

	if len(taskID) == 0 {
		log.Debug("no tasks restored, starting new one")

		taskReply, err := e.startTaskWithRetry(ctx, log, deal)
		if err != nil {
			log.Warn("cannot start task", zap.Error(err))
			return
		}

		taskID = taskReply.GetId()
		log.Info("task started", zap.String("task_id", taskID))
	}

	try := 0
	for {
		ok := func() bool {
			taskCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			go e.antiFraud.TrackTask(taskCtx, deal, taskID)

			shouldRestartTask, err := e.trackTaskWithRetry(taskCtx, log, deal.GetId(), taskID)
			if err != nil {
				log.Warn("task tracking failed", zap.Error(err), zap.String("task_id", taskID))
			}

			if !shouldRestartTask {
				log.Warn("should not restarting the task", zap.Error(err), zap.Int("try", try))
				return false
			}

			if try >= taskRestartCount {
				log.Debug("stop task restarting: retry count exceeded")
				return false
			}

			log.Debug("going to restart a broken task")
			nextTask, err := e.startTaskWithRetry(taskCtx, log, deal)
			if err != nil {
				log.Warn("cannot start task", zap.Error(err), zap.Int("try", try))
				return false
			}

			taskID = nextTask.GetId()
			log.Debug("task restarted", zap.String("task_id", taskID), zap.Int("try", try))
			try++

			return true
		}()

		if !ok {
			return
		}
	}
}

func (e *engine) startTaskWithRetry(ctx context.Context, log *zap.Logger, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	t := util.NewImmediateTicker(e.cfg.Engine.TaskStartInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return nil, fmt.Errorf("cannot start task: retry count exceeded")
			}

			list, err := e.loadTasksOnce(ctx, log, deal.GetId())
			if err != nil {
				try++
				log.Warn("cannot obtain task list from worker", zap.Error(err), zap.Int("try", try))
				continue
			}

			// check for single task
			// because worker's task list sanitizing
			// already performed in `restoreTasks` method.
			if len(list) == 1 {
				log.Info("found already started task, continue tracking", zap.String("task_id", list[0].ID))
				return &sonm.StartTaskReply{Id: list[0].ID}, nil
			}

			taskReply, err := e.startTaskOnce(ctx, log, deal.GetId())
			if err != nil {
				try++
				log.Warn("task start failed", zap.Error(err), zap.Int("try", try))
				continue
			}

			return taskReply, nil
		}
	}
}

func (e *engine) startTaskOnce(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt) (*sonm.StartTaskReply, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.TaskStartTimeout)
	defer cancel()

	env := applyEnvTemplate(e.cfg.Container.Env, dealID)
	log.Info("starting task",
		zap.String("deal_id", dealID.Unwrap().String()),
		zap.Any("environment", env))

	taskReply, err := e.tasks.Start(ctx, &sonm.StartTaskRequest{
		DealID: dealID,
		Spec: &sonm.TaskSpec{
			Tag: e.cfg.Container.getTag(),
			Container: &sonm.Container{
				Image:  e.cfg.Container.Image,
				SshKey: e.cfg.Container.SSHKey,
				Env:    env,
			},
			Resources: &sonm.AskPlanResources{},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start task on worker: %v", err)
	}

	return taskReply, nil
}

func (e *engine) trackTaskWithRetry(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt, taskID string) (bool, error) {
	log = log.Named("task").With(zap.String("task_id", taskID))
	log.Info("start task status tracking")

	t := util.NewImmediateTicker(e.cfg.Engine.TaskTrackInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return false, fmt.Errorf("task tracking failed: retry count exceeded")
			}

			ok, err := e.checkDealStatus(ctx, log, dealID)
			if err != nil {
				try++
				log.Warn("cannot check deal status, increasing retry counter", zap.Error(err), zap.Int("try", try))
				continue
			}

			if !ok {
				log.Warn("deal is closed, finishing tracking")
				return false, fmt.Errorf("deal is closed")
			}

			log.Debug("deal status OK, checking tasks")
			shouldRetry, err := e.trackTaskOnce(ctx, log, dealID, taskID)
			if err != nil {
				if !shouldRetry {
					return true, err
				}

				try++
				log.Warn("cannot get task status, increasing retry counter", zap.Error(err), zap.Int("try", try))
				continue
			}

			log.Debug("task tracking OK, resetting failure counter")
			try = 0
		}
	}
}

func (e *engine) checkDealStatus(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	dealStatus, err := e.deals.Status(ctx, dealID)
	if err != nil {
		return false, err
	}

	if dealStatus.GetDeal().GetStatus() == sonm.DealStatus_DEAL_ACCEPTED {
		deal := e.dealFactory.FromDeal(dealStatus.GetDeal())
		actualPrice := e.priceProvider.GetPrice()
		if deal.IsReplaceable(actualPrice, e.cfg.Market.PriceControl.DealCancelThreshold) {
			log := log.Named("price-deviation")
			if len(e.orderCancelChan) > 0 {
				log.Warn("shouldn't finish deal, orders replacing in progress",
					zap.Int("cancel", len(e.orderCancelChan)),
					zap.Int("create", len(e.ordersCreateChan)))
				return true, nil
			}

			pricePerDeal := big.NewInt(0).Mul(actualPrice, big.NewInt(int64(deal.BenchmarkValue())))
			log.Info("too much price deviation detected: closing deal",
				zap.Uint64("benchmark", deal.BenchmarkValue()),
				zap.String("actual_price", pricePerDeal.String()),
				zap.String("current_price", deal.GetPrice().Unwrap().String()))

			if err := e.antiFraud.FinishDeal(ctx, deal.Unwrap(), antifraud.SkipBlacklisting); err != nil {
				log.Warn("failed to finish deal", zap.Error(err))
			}

			return false, nil
		}

		return true, nil
	}

	return false, nil
}

func (e *engine) trackTaskOnce(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt, taskID string) (bool, error) {
	log.Debug("checking task status")

	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	// 3. ping task
	status, err := e.tasks.Status(ctx, &sonm.TaskID{Id: taskID, DealID: dealID})
	if err != nil {
		return true, err
	}

	if status.GetStatus() == sonm.TaskStatusReply_FINISHED || status.GetStatus() == sonm.TaskStatusReply_BROKEN {
		log.Warn("task is failed by unknown reasons", zap.String("status", status.GetStatus().String()))
		return false, fmt.Errorf("task is finished by unknown reasons")
	}

	log.Debug("task status is OK")
	return true, nil
}

func (e *engine) restoreTasks(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt) (string, error) {
	log = log.Named("restore-tasks")
	log.Debug("restoring tasks")

	t := util.NewImmediateTicker(e.cfg.Engine.TaskRestoreInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return "", fmt.Errorf("restore tasks failed: retry count exceeded")
			}

			list, err := e.loadTasksOnce(ctx, log, dealID)
			if err != nil {
				try++
				log.Warn("cannot obtain task list from worker", zap.Error(err), zap.Int("try", try))
				continue
			}

			switch len(list) {
			case 0:
				return "", nil
			case 1:
				requiredTag := string(e.cfg.Container.getTag().GetData())
				givenTag := string(list[0].GetTag().GetData())
				if givenTag != requiredTag {
					log.Warn("unexpected tag assigned to the running on task",
						zap.String("running", givenTag), zap.String("expected", requiredTag))
					e.stopOneTask(ctx, log, dealID, list[0].ID)
					return "", nil
				}

				return list[0].ID, nil
			default:
				// weird case, we always starting only one task per deal
				log.Info("worker have more than one task running", zap.Int("count", len(list)))
				if err := e.stopAllTasks(ctx, log, list, dealID); err != nil {
					return "", err
				}

				return "", nil
			}
		}
	}
}

func (e *engine) loadTasksOnce(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt) ([]*types.TaskStatus, error) {
	log.Debug("loading tasks from worker")

	ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	taskList, err := e.tasks.List(ctx, &sonm.TaskListRequest{DealID: dealID})
	if err != nil {
		return nil, err
	}

	list := make([]*types.TaskStatus, 0)
	for id, task := range taskList.GetInfo() {
		if task.GetStatus() == sonm.TaskStatusReply_RUNNING {
			list = append(list, &types.TaskStatus{TaskStatusReply: task, ID: id})
		}
	}

	return list, nil
}

func (e *engine) stopAllTasks(ctx context.Context, log *zap.Logger, list []*types.TaskStatus, dealID *sonm.BigInt) error {
	log.Debug("stopping all tasks on worker")

	var failedIDs []string
	for _, task := range list {
		if err := e.stopOneTask(ctx, log, dealID, task.ID); err != nil {
			log.Warn("cannot stop task", zap.Error(err), zap.String("task_id", task.ID))
			failedIDs = append(failedIDs, task.ID)
		}
	}

	if len(failedIDs) > 0 {
		return fmt.Errorf("cannot stop tasks ids = %s", strings.Join(failedIDs, ","))
	}

	return nil
}

func (e *engine) stopOneTask(ctx context.Context, log *zap.Logger, dealID *sonm.BigInt, taskID string) error {
	log.Debug("stopping task", zap.String("task_id", taskID))

	for try := 0; try < maxRetryCount; try++ {
		err := func() error {
			ctx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
			defer cancel()

			if _, err := e.tasks.Stop(ctx, &sonm.TaskID{Id: taskID, DealID: dealID}); err != nil {
				log.Warn("cannot stop task", zap.Error(err), zap.Int("try", try))
				return err
			}

			return nil
		}()

		if err != nil {
			continue
		}
		return nil
	}

	return fmt.Errorf("cannot stop task: retry count exceeded")
}

func (e *engine) start(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	for i := 0; i < concurrency; i++ {
		wg.Go(func() error {
			e.processOrderCreate(ctx)
			return nil
		})
	}

	for i := 0; i < concurrency; i++ {
		wg.Go(func() error {
			e.processOrderCancel(ctx)
			return nil
		})
	}

	<-ctx.Done()
	return ctx.Err()
}

func (e *engine) loadInitialData(ctx context.Context) error {
	if p, ok := e.priceProvider.(price.Updateable); ok {
		if err := p.Update(ctx); err != nil {
			return fmt.Errorf("cannot update price: %v", err)
		}
	}

	return nil
}

func (e *engine) startPriceTracking(ctx context.Context) error {
	log := e.log.Named("token-price")

	provider, ok := e.priceProvider.(price.Updateable)
	if !ok {
		log.Info("price source shouldn't be updated")
		return nil
	}

	t := time.NewTicker(provider.Interval())
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("stop price tracking")
			return ctx.Err()

		case <-t.C:
			if err := provider.Update(ctx); err != nil {
				log.Warn("cannot update token price", zap.Error(err))
			} else {
				log.Debug("received new token price",
					zap.String("new_price", e.priceProvider.GetPrice().String()))
			}
		}
	}
}

func (e *engine) filterDeals(deals []*sonm.Deal) []*sonm.Deal {
	set := map[string]*sonm.Deal{}
	for _, deal := range deals {
		if deal.ConsumerID.Unwrap() == e.ethAddr {
			set[deal.GetId().Unwrap().String()] = deal
		}
	}
	filtered := make([]*sonm.Deal, 0, len(set))
	for _, deal := range set {
		filtered = append(filtered, deal)
	}
	return filtered
}

func (e *engine) restoreMarketState(ctx context.Context) error {
	existingOrders, err := e.market.GetOrders(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load orders from market: %v", err)
	}

	existingDeals, err := e.deals.List(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load deals from market: %v", err)
	}

	// use only deals where Connor is consumer
	dealsToRestore := e.filterDeals(existingDeals.GetDeal())

	existingCorders := e.corderFactory.FromSlice(existingOrders.GetOrders())
	targetCorders := e.getTargetCorders()

	// adding all existing orders as active, no matter if it should be canceled in the nearest future
	for _, ord := range existingCorders {
		e.state.AddActiveOrder(ord)
	}

	set := types.DivideOrdersSets(existingCorders, targetCorders)
	e.log.Debug("restoring existing entities",
		zap.Int("orders_restore", len(set.Restore)),
		zap.Int("orders_create", len(set.Create)),
		zap.Int("orders_cancel", len(set.Cancel)),
		zap.Int("deals_restore", len(dealsToRestore)))

	for _, deal := range dealsToRestore {
		e.state.AddDeal(e.dealFactory.FromDeal(deal))
		e.RestoreDeal(ctx, deal)
	}

	for _, ord := range set.Create {
		e.CreateOrder(ord)
	}

	for _, ord := range set.Restore {
		e.RestoreOrder(ctx, ord)
	}

	if e.cfg.Market.AdoptOrders {
		e.log.Debug("adopt_orders option is enabled, restoring excess orders")
		for _, ord := range set.Cancel {
			e.RestoreOrder(ctx, ord)
		}
	} else {
		e.log.Debug("adopt_orders option is disabled, cancelling excess orders")
		for _, ord := range set.Cancel {
			e.CancelOrder(ord)
		}
	}

	return nil
}

func (e *engine) waitForExternalUpdates(ctx context.Context) error {
	tk := util.NewImmediateTicker(time.Minute)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
			e.adoptExternalDeals(ctx)
			e.adoptExternalOrders(ctx)
			e.state.DumpToFile()
		}
	}
}

func (e *engine) adoptExternalDeals(ctx context.Context) {
	reqCtx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	deals, err := e.deals.List(reqCtx, &sonm.Count{Count: 1000})
	if err != nil {
		e.log.Warn("failed to load deals from DWH", zap.Error(err))
		return
	}

	for _, deal := range deals.GetDeal() {
		if ok := e.state.AddDeal(e.dealFactory.FromDeal(deal)); ok {
			adoptedDealsCounter.Inc()
			e.RestoreDeal(ctx, deal)
		}
	}
}

func (e *engine) adoptExternalOrders(ctx context.Context) {
	reqCtx, cancel := context.WithTimeout(ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	orders, err := e.market.GetOrders(reqCtx, &sonm.Count{Count: 1000})
	if err != nil {
		e.log.Warn("failed to load orders from DWH", zap.Error(err))
		return
	}

	for _, order := range orders.GetOrders() {
		cord := e.corderFactory.FromOrder(order)
		if !e.state.HasOrder(cord) {
			adoptedOrdersCounter.Inc()
			e.RestoreOrder(ctx, cord)
		}
	}
}

func (e *engine) getTargetCorders() []*types.Corder {
	v := make([]*types.Corder, 0)

	for hashrate := e.cfg.Market.From; hashrate <= e.cfg.Market.To; hashrate += e.cfg.Market.Step {
		// settings zero price is OK for now, we'll update it just before sending to the Marketplace.
		v = append(v, e.corderFactory.FromParams(big.NewInt(0), hashrate, e.cfg.getBaseBenchmarks()))
	}

	return v
}

func (e *engine) close() {
	e.log.Info("closing engine")

	close(e.ordersCreateChan)
	close(e.orderCancelChan)
}
