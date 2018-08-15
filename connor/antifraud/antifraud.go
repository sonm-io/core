package antifraud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	blacklistedDealCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sonm_deals_blacklisted",
		Help: "Number of deals that were closed with blacklisting",
	})
)

func init() {
	prometheus.MustRegister(blacklistedDealCounter)
}

type AntiFraud interface {
	Run(ctx context.Context) error
	DealOpened(deal *sonm.Deal) error
	TrackTask(ctx context.Context, deal *sonm.Deal, taskID string) error
	FinishDeal(deal *sonm.Deal) error
}

type dealMeta struct {
	deal          *sonm.Deal
	logProcessor  Processor
	poolProcessor Processor
}

func lifeTime(deal *sonm.Deal) time.Duration {
	return time.Since(deal.GetStartTime().Unix())
}

type antiFraud struct {
	mu                sync.RWMutex
	meta              map[string]*dealMeta
	blacklistWatchers map[common.Address]*blacklistWatcher
	processorFactory  ProcessorFactory

	cfg            Config
	nodeConnection *grpc.ClientConn
	deals          sonm.DealManagementClient
	log            *zap.Logger
}

func NewAntiFraud(cfg Config, log *zap.Logger, processors ProcessorFactory, nodeConnection *grpc.ClientConn) AntiFraud {
	return &antiFraud{
		processorFactory:  processors,
		meta:              make(map[string]*dealMeta),
		blacklistWatchers: map[common.Address]*blacklistWatcher{},
		nodeConnection:    nodeConnection,
		deals:             sonm.NewDealManagementClient(nodeConnection),
		log:               log.Named("anti-fraud"),
		cfg:               cfg,
	}
}

// Run blocks until context is cancelled or unrecoverable error met
func (m *antiFraud) Run(ctx context.Context) error {
	m.log.Info("starting antifraud")

	ticker := time.NewTicker(m.cfg.QualityCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			m.checkDeals(ctx)
		}
	}
}

//TODO: async
func (m *antiFraud) checkDeals(ctx context.Context) error {
	m.log.Debug("checking deals")
	defer m.log.Debug("stop checking deals")

	for _, dealMeta := range m.meta {
		log := m.log.With(zap.String("deal_id", dealMeta.deal.GetId().Unwrap().String()))

		if dealMeta.logProcessor == nil {
			m.log.Debug("skipping deal without task")
			continue
		}

		// attach task id to logger
		log = log.With(zap.String("task_id", dealMeta.logProcessor.TaskID()))
		watcher, ok := m.blacklistWatchers[dealMeta.deal.SupplierID.Unwrap()]
		if !ok {
			log.Warn("cannot obtain blacklist watcher for deal, skipping")
			continue
		}

		shouldClose := false
		accurateByLogs, qualityByLogs := dealMeta.logProcessor.TaskQuality()
		if !accurateByLogs {
			// always skip tasks without actual logs because log
			// analyzer starts lot more early than pool reports processor.
			continue
		}

		// anti-fraud should close deal if reported hashrate by logs
		// is not fit into required value.
		if qualityByLogs < m.cfg.TaskQuality {
			shouldClose = true
			log.Debug("task quality is less that required: detected by logs")
		}

		accurateByPool, qualityByPool := dealMeta.poolProcessor.TaskQuality()
		if accurateByPool && qualityByPool < m.cfg.TaskQuality {
			shouldClose = true
			log.Debug("task quality is less that required: detected by pool reports")
		}

		logQualityMetrics := []zapcore.Field{
			zap.Float64("by_logs", qualityByLogs),
			zap.Float64("by_pool", qualityByPool),
			zap.Float64("required", m.cfg.TaskQuality),
		}

		if shouldClose {
			log.Warn("task quality is less that required, closing deal", logQualityMetrics...)

			blacklistedDealCounter.Inc()
			if err := m.finishDeal(dealMeta.deal, sonm.BlacklistType_BLACKLIST_WORKER); err != nil {
				log.Warn("cannot finish deal", zap.Error(err))
			}

			watcher.Failure()
		} else {
			log.Debug("task quality is fit into required required value", logQualityMetrics...)
			watcher.Success()
		}
	}

	//TODO: save this in DB, load on start
	for _, watcher := range m.blacklistWatchers {
		watcher.TryUnblacklist(ctx)
	}
	return nil
}

func (m *antiFraud) TrackTask(ctx context.Context, deal *sonm.Deal, taskID string) error {
	m.mu.Lock()
	meta, ok := m.meta[deal.Id.Unwrap().String()]
	if !ok {
		return fmt.Errorf("could not register spawned task %s, no deal with id %s", taskID, deal.Id.Unwrap().String())
	}

	meta.poolProcessor = m.processorFactory.PoolProcessor(deal, taskID, WithLogger(m.log))
	meta.logProcessor = m.processorFactory.LogProcessor(deal, taskID, WithLogger(m.log), WithClientConn(m.nodeConnection))
	m.mu.Unlock()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return meta.poolProcessor.Run(ctx)
	})
	g.Go(func() error {
		return meta.logProcessor.Run(ctx)
	})

	return g.Wait()
}

func (m *antiFraud) DealOpened(deal *sonm.Deal) error {
	m.log.Info("registering deal", zap.String("deal_id", deal.GetId().Unwrap().String()))

	meta := &dealMeta{
		deal: deal,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.meta[deal.GetId().Unwrap().String()] = meta
	if _, ok := m.blacklistWatchers[deal.GetSupplierID().Unwrap()]; !ok {
		w := NewBlacklistWatcher(deal.GetSupplierID().Unwrap(), m.nodeConnection, m.log)
		m.blacklistWatchers[deal.GetSupplierID().Unwrap()] = w
	}

	return nil
}

func (m *antiFraud) FinishDeal(deal *sonm.Deal) error {
	return m.finishDeal(deal, sonm.BlacklistType_BLACKLIST_NOBODY)
}

func (m *antiFraud) finishDeal(deal *sonm.Deal, blacklistType sonm.BlacklistType) error {
	m.log.Info("finishing deal", zap.String("deal_id", deal.GetId().Unwrap().String()),
		zap.Duration("lifetime", lifeTime(deal)))

	ctx, cancel := context.WithTimeout(context.Background(), m.cfg.ConnectionTimeout)
	defer cancel()

	m.mu.Lock()
	delete(m.meta, deal.GetId().Unwrap().String())
	m.mu.Unlock()
	_, err := m.deals.Finish(ctx, &sonm.DealFinishRequest{
		Id:            deal.GetId(),
		BlacklistType: blacklistType,
	})

	return err
}
