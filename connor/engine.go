package connor

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sonm-io/core/connor/antifraud"
	"github.com/sonm-io/core/connor/price"
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
	ctx       context.Context
	cfg       *Config
	antiFraud antifraud.AntiFraud

	market        sonm.MarketClient
	deals         sonm.DealManagementClient
	tasks         sonm.TaskManagementClient
	priceProvider price.Provider
	corderFactory CorderFactoriy

	ordersCreateChan  chan *Corder
	ordersResultsChan chan *Corder
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
)

func init() {
	prometheus.MustRegister(activeDealsGauge)
	prometheus.MustRegister(activeOrdersGauge)
	prometheus.MustRegister(createdOrdersCounter)
	prometheus.MustRegister(replacedOrdersCounter)
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*engine, error) {
	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("cannot load eth keys: %v", err)
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cert TLS config: %v", err)
	}

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(key.PublicKey))
	cc, err := xgrpc.NewClient(ctx, cfg.Node.Endpoint.String(), creds)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to node: %v", err)
	}

	return &engine{
		// todo: remove context from struct
		ctx: ctx,
		cfg: cfg,
		log: log,

		priceProvider: cfg.getTokenParams().priceProvider,
		corderFactory: cfg.getTokenParams().corderFactory,

		market:    sonm.NewMarketClient(cc),
		deals:     sonm.NewDealManagementClient(cc),
		tasks:     sonm.NewTaskManagementClient(cc),
		antiFraud: antifraud.NewAntiFraud(cfg.AntiFraud, log.Named("anti-fraud"), cc),

		ordersCreateChan:  make(chan *Corder, concurrency),
		ordersResultsChan: make(chan *Corder, concurrency),
	}, nil
}

func (e *engine) Serve(ctx context.Context) error {
	defer e.close()

	e.log.Info("starting engine", zap.Int("concurrency", concurrency))

	// perform extra config validation using external list of required benchmarks
	if err := e.validateBenchmarks(ctx); err != nil {
		return fmt.Errorf("benchmarks validation failed: %v", err)
	}

	// load initial state from external sources
	if err := e.loadInitialData(ctx); err != nil {
		return fmt.Errorf("failed to load initial data: %v", err)
	}

	e.log.Debug("price",
		zap.String(e.cfg.Mining.Token, e.priceProvider.GetPrice().String()),
		zap.Float64("margin", e.cfg.Market.PriceMarginality))

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

	return wg.Wait()
}

func (e *engine) CreateOrder(bid *Corder) {
	e.ordersCreateChan <- bid
}

func (e *engine) RestoreOrder(order *Corder) {
	e.log.Debug("restoring order", zap.String("id", order.Order.GetId().Unwrap().String()))
	e.ordersResultsChan <- order
}

func (e *engine) RestoreDeal(deal *sonm.Deal) {
	e.log.Debug("restoring deal", zap.String("id", deal.GetId().Unwrap().String()))
	go e.processDeal(deal)
}

func (e *engine) sendOrderToMarket(bid *sonm.BidOrder) (*sonm.Order, error) {
	e.log.Debug("creating order on market",
		zap.String("price", bid.GetPrice().GetPerSecond().Unwrap().String()),
		zap.Any("benchmarks", bid.Resources.GetBenchmarks()))

	return e.market.CreateOrder(e.ctx, bid)
}

func (e *engine) processOrderCreate() {
	for bid := range e.ordersCreateChan {
		created, err := e.sendOrderToMarket(bid.AsBID())
		if err != nil {
			e.log.Warn("cannot place order, retrying", zap.Error(err))
			e.CreateOrder(bid)
			continue
		}

		createdOrdersCounter.Inc()
		e.ordersResultsChan <- e.corderFactory.FromOrder(created)
	}
}

func (e *engine) processOrderResult() {
	for order := range e.ordersResultsChan {
		go e.waitForDeal(order)
	}
}

func (e *engine) waitForDeal(order *Corder) {
	activeOrdersGauge.Inc()

	id := order.GetId().Unwrap().String()
	log := e.log.With(zap.String("order_id", id))

	t := util.NewImmediateTicker(e.cfg.Engine.OrderWatchInterval)
	defer t.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-t.C:
			actualPrice := e.priceProvider.GetPrice()
			if order.isReplaceable(actualPrice, e.cfg.Mining.TokenPrice.Threshold) {
				log.Info("we can replace order with more profitable one",
					zap.String("actual_price", actualPrice.String()),
					zap.String("current_price", order.restorePrice().String()))

				e.cancelOrder(log, order.GetId())

				hashRate := big.NewInt(0).SetUint64(order.GetHashrate())
				order.Price = sonm.NewBigInt(big.NewInt(0).Mul(actualPrice, hashRate))

				replacedOrdersCounter.Inc()
				e.CreateOrder(order)
				return
			}

			deal, err := e.checkOrderForDealOnce(log, id)
			if err != nil {
				continue
			}

			e.CreateOrder(order)
			if deal != nil {
				activeOrdersGauge.Dec()
				e.processDeal(deal)
			}

			return
		}
	}
}

func (e *engine) cancelOrder(log *zap.Logger, id *sonm.BigInt) {
	log.Info("cancelling order")

	for try := 0; try < maxRetryCount; try++ {
		ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
		if _, err := e.market.CancelOrder(ctx, &sonm.ID{Id: id.Unwrap().String()}); err != nil {
			cancel()
			log.Warn("cannot cancel order", zap.Error(err), zap.Int("try", try))
			continue
		}

		cancel()
		log.Info("order cancelled")
		return
	}

	log.Warn("order cancellation failed: retry count exceeded")
}

func (e *engine) checkOrderForDealOnce(log *zap.Logger, orderID string) (*sonm.Deal, error) {
	ord, err := e.market.GetOrderByID(e.ctx, &sonm.ID{Id: orderID})
	if err != nil {
		log.Warn("cannot get order info from market", zap.Error(err))
		return nil, err
	}

	if ord.GetOrderStatus() == sonm.OrderStatus_ORDER_INACTIVE {
		log.Info("order becomes inactive, looking for related deal")

		if ord.GetDealID().IsZero() {
			log.Debug("order have no deal, probably order is cancelled by hand")
			return nil, nil
		}

		deal, err := e.deals.Status(e.ctx, ord.GetDealID())
		if err != nil {
			log.Warn("cannot get deal info from market", zap.Error(err),
				zap.String("deal_id", ord.GetDealID().Unwrap().String()))
			return nil, err
		}

		return deal.GetDeal(), nil
	}

	return nil, fmt.Errorf("order have no deal")
}

func (e *engine) processDeal(deal *sonm.Deal) {
	activeDealsGauge.Inc()
	defer activeDealsGauge.Dec()

	dealID := deal.GetId().Unwrap().String()
	log := e.log.Named("process-deal").With(zap.String("deal_id", dealID))

	e.antiFraud.DealOpened(deal)

	log.Debug("start deal processing")
	defer log.Debug("stop deal processing")
	defer e.antiFraud.FinishDeal(deal)

	taskID, err := e.restoreTasks(log, deal.GetId())
	if err != nil {
		log.Warn("cannot restore tasks", zap.Error(err))
		return
	}

	if len(taskID) == 0 {
		log.Debug("no tasks restored, starting new one")

		taskReply, err := e.startTaskWithRetry(log, deal)
		if err != nil {
			log.Warn("cannot start task", zap.Error(err))
			return
		}

		taskID = taskReply.GetId()
		log.Info("task started", zap.String("task_id", taskID))
	}

	ctx, cancel := context.WithCancel(e.ctx)
	defer cancel()
	go e.antiFraud.TrackTask(ctx, deal, taskID)

	try := 0
	for {
		shouldRestartTask, err := e.trackTaskWithRetry(log, deal.GetId(), taskID)
		if err != nil {
			log.Warn("task tracking failed", zap.Error(err))
		}

		if !shouldRestartTask {
			log.Warn("should not restarting the task", zap.Error(err), zap.Int("try", try))
			return
		}

		if try >= taskRestartCount {
			log.Debug("stop task restarting: retry count exceeded")
			return
		}

		log.Debug("going to restart a broken task")
		nextTask, err := e.startTaskWithRetry(log, deal)
		if err != nil {
			log.Warn("cannot start task", zap.Error(err), zap.Int("try", try))
			return
		}

		taskID = nextTask.GetId()
		log.Debug("task restarted", zap.String("task_id", taskID), zap.Int("try", try))
		try++
	}
}

func (e *engine) startTaskWithRetry(log *zap.Logger, deal *sonm.Deal) (*sonm.StartTaskReply, error) {
	t := util.NewImmediateTicker(e.cfg.Engine.TaskStartInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-e.ctx.Done():
			return nil, e.ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return nil, fmt.Errorf("cannot start task: retry count exceeded")
			}

			taskReply, err := e.startTaskOnce(log, deal.GetId())
			if err != nil {
				try++
				log.Warn("task start failed", zap.Error(err), zap.Int("try", try))
				continue
			}

			return taskReply, nil
		}
	}
}

func (e *engine) startTaskOnce(log *zap.Logger, dealID *sonm.BigInt) (*sonm.StartTaskReply, error) {
	ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	workerID := "c" + dealID.Unwrap().String()
	ethID := strings.ToLower(e.cfg.Mining.Wallet.Hex())
	poolAddr := fmt.Sprintf("%s/%s/%s", e.cfg.Mining.PoolReportURL, ethID, workerID)
	wallet := fmt.Sprintf("%s/%s", ethID, workerID)

	env := map[string]string{
		"WALLET": wallet,
		"POOL":   poolAddr,
	}

	e.log.Debug("starting task", zap.Any("environment", env))

	taskReply, err := e.tasks.Start(ctx, &sonm.StartTaskRequest{
		DealID: dealID,
		Spec: &sonm.TaskSpec{
			Tag: e.cfg.Mining.getTag(),
			Container: &sonm.Container{
				Image: e.cfg.Mining.Image,
				Env:   env,
			},
			Resources: &sonm.AskPlanResources{},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start task on worker: %v", err)
	}

	return taskReply, nil
}

func (e *engine) trackTaskWithRetry(log *zap.Logger, dealID *sonm.BigInt, taskID string) (bool, error) {
	log = log.Named("task").With(zap.String("task_id", taskID))
	log.Info("start task status tracking")

	t := util.NewImmediateTicker(e.cfg.Engine.TaskTrackInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-e.ctx.Done():
			return false, e.ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return false, fmt.Errorf("task tracking failed: retry count exceeded")
			}

			ok, err := e.checkDealStatus(log, dealID)
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
			shouldRetry, err := e.trackTaskOnce(log, dealID, taskID)
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

func (e *engine) checkDealStatus(log *zap.Logger, dealID *sonm.BigInt) (bool, error) {
	ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	dealStatus, err := e.deals.Status(ctx, dealID)
	if err != nil {
		return false, err
	}

	return dealStatus.GetDeal().GetStatus() == sonm.DealStatus_DEAL_ACCEPTED, nil
}

func (e *engine) trackTaskOnce(log *zap.Logger, dealID *sonm.BigInt, taskID string) (bool, error) {
	log.Debug("checking task status")

	ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
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

func (e *engine) restoreTasks(log *zap.Logger, dealID *sonm.BigInt) (string, error) {
	log = log.Named("restore-tasks")
	log.Debug("restoring tasks")

	t := util.NewImmediateTicker(e.cfg.Engine.TaskRestoreInterval)
	defer t.Stop()

	try := 0
	for {
		select {
		case <-e.ctx.Done():
			return "", e.ctx.Err()
		case <-t.C:
			if try > maxRetryCount {
				return "", fmt.Errorf("restore tasks failed: retry count exceeded")
			}

			list, err := e.loadTasksOnce(log, dealID)
			if err != nil {
				try++
				log.Warn("cannot obtain task list from worker", zap.Error(err))
				continue
			}

			switch len(list) {
			case 0:
				return "", nil
			case 1:
				requiredTag := string(e.cfg.Mining.getTag().GetData())
				givenTag := string(list[0].GetTag().GetData())
				if givenTag != requiredTag {
					log.Warn("unexpected tag assigned to the running on task",
						zap.String("running", givenTag), zap.String("expected", requiredTag))
					e.stopOneTask(log, dealID, list[0].id)
					return "", nil
				}

				return list[0].id, nil
			default:
				// weird case, we always starting only one task per deal
				log.Info("worker have more than one task running", zap.Int("count", len(list)))
				if err := e.stopAllTasks(log, list, dealID); err != nil {
					return "", err
				}

				return "", nil
			}
		}
	}
}

func (e *engine) loadTasksOnce(log *zap.Logger, dealID *sonm.BigInt) ([]*taskStatus, error) {
	log.Debug("loading tasks from worker")

	ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
	defer cancel()

	taskList, err := e.tasks.List(ctx, &sonm.TaskListRequest{DealID: dealID})
	if err != nil {
		return nil, err
	}

	list := make([]*taskStatus, 0)
	for id, task := range taskList.GetInfo() {
		if task.GetStatus() == sonm.TaskStatusReply_RUNNING {
			list = append(list, &taskStatus{task, id})
		}
	}

	return list, nil
}

func (e *engine) stopAllTasks(log *zap.Logger, list []*taskStatus, dealID *sonm.BigInt) error {
	log.Debug("stopping all tasks on worker")

	var failedIDs []string
	for _, task := range list {
		if err := e.stopOneTask(log, dealID, task.id); err != nil {
			log.Warn("cannot stop task", zap.Error(err), zap.String("task_id", task.id))
			failedIDs = append(failedIDs, task.id)
		}
	}

	if len(failedIDs) > 0 {
		return fmt.Errorf("cannot stop tasks ids = %s", strings.Join(failedIDs, ","))
	}

	return nil
}

func (e *engine) stopOneTask(log *zap.Logger, dealID *sonm.BigInt, taskID string) error {
	log.Debug("stopping task", zap.String("task_id", taskID))

	for try := 0; try < maxRetryCount; try++ {
		ctx, cancel := context.WithTimeout(e.ctx, e.cfg.Engine.ConnectionTimeout)
		if _, err := e.tasks.Stop(ctx, &sonm.TaskID{Id: taskID, DealID: dealID}); err != nil {
			log.Warn("cannot stop task", zap.Error(err), zap.Int("try", try))
			cancel()
			continue
		}

		cancel()
		return nil
	}

	return fmt.Errorf("cannot stop task: retry count exceeded")
}

func (e *engine) start(ctx context.Context) error {
	for i := 0; i < concurrency; i++ {
		go e.processOrderCreate()
	}

	go e.processOrderResult()

	<-ctx.Done()
	return ctx.Err()
}

func (e *engine) loadInitialData(ctx context.Context) error {
	if err := e.priceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update %s price: %v", e.cfg.Mining.Token, err)
	}

	return nil
}

func (e *engine) validateBenchmarks(ctx context.Context) error {
	benchList, err := benchmarks.NewBenchmarksList(ctx, e.cfg.BenchmarkList)
	if err != nil {
		return fmt.Errorf("cannot load benchmark list: %v", err)
	}

	return e.cfg.validateBenchmarks(benchList)
}

func (e *engine) startPriceTracking(ctx context.Context) error {
	log := e.log.Named("token-price")
	t := time.NewTicker(e.cfg.Mining.TokenPrice.UpdateInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("stop price tracking")
			return ctx.Err()

		case <-t.C:
			if err := e.priceProvider.Update(ctx); err != nil {
				log.Warn("cannot update token price", zap.Error(err))
			} else {
				log.Debug("received new token price",
					zap.String("new_price", e.priceProvider.GetPrice().String()))
			}
		}
	}
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

	existingCorders := e.corderFactory.FromSlice(existingOrders.GetOrders())
	targetCorders := e.getTargetCorders()

	set := divideOrdersSets(existingCorders, targetCorders)
	e.log.Debug("restoring existing entities",
		zap.Int("orders_restore", len(set.toRestore)),
		zap.Int("orders_create", len(set.toCreate)),
		zap.Int("deals_restore", len(existingDeals.GetDeal())))

	for _, deal := range existingDeals.GetDeal() {
		e.RestoreDeal(deal)
	}

	for _, ord := range set.toCreate {
		e.CreateOrder(ord)
	}

	for _, ord := range set.toRestore {
		e.RestoreOrder(ord)
	}

	return nil
}

func (e *engine) getTargetCorders() []*Corder {
	v := make([]*Corder, 0)

	for hashrate := e.cfg.Market.FromHashRate; hashrate <= e.cfg.Market.ToHashRate; hashrate += e.cfg.Market.Step {
		bigHashrate := big.NewInt(int64(hashrate))
		p := big.NewInt(0).Mul(bigHashrate, e.priceProvider.GetPrice())
		v = append(v, e.corderFactory.FromParams(p, hashrate, e.cfg.getBaseBenchmarks()))
	}

	return v
}

func (e *engine) close() {
	e.log.Info("closing engine")

	close(e.ordersCreateChan)
	close(e.ordersResultsChan)
}
