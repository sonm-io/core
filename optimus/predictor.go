package optimus

import (
	"context"
	"fmt"
	"sync"

	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type PredictorConfig struct {
	Marketplace marketplaceConfig
}

type PredictorService struct {
	cfg *PredictorConfig
	log *zap.SugaredLogger

	mu             sync.RWMutex
	benchmarks     benchmarks.BenchList
	regression     *regressionClassifier
	classification *OrderClassification
}

// NewPredictorService constructs a new order price predictor service.
// Returns nil when nil "cfg" is passed.
func NewPredictorService(cfg *PredictorConfig, benchmarks benchmarks.BenchList, log *zap.SugaredLogger) *PredictorService {
	if cfg == nil {
		return nil
	}

	regression := &regressionClassifier{
		model: &SCAKKTModel{
			MaxIterations: 1e7,
			Log:           log,
		},
	}

	m := &PredictorService{
		cfg:        cfg,
		log:        log,
		benchmarks: benchmarks,
		regression: regression,
	}

	return m
}

func (m *PredictorService) Serve(ctx context.Context) error {
	return m.serve(ctx)
}

func (m *PredictorService) serve(ctx context.Context) error {
	m.log.Info("serving order price predictor")
	defer m.log.Info("stopped serving order price predictor")

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return m.serveMarketplace(ctx)
	})

	return wg.Wait()
}

func (m *PredictorService) serveMarketplace(ctx context.Context) error {
	registry := newRegistry()
	defer registry.Close()

	dwh, err := registry.NewDWH(ctx, m.cfg.Marketplace.Endpoint, m.cfg.Marketplace.PrivateKey.Unwrap())
	if err != nil {
		return err
	}

	marketCache := newMarketCache(newMarketScanner(m.cfg.Marketplace, dwh), m.cfg.Marketplace.Interval)

	timer := util.NewImmediateTicker(m.cfg.Marketplace.Interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			if err := m.executeRegression(ctx, marketCache); err != nil {
				m.log.Warnw("failed to perform regression analysis", zap.Error(err))
			}
		}
	}
}

func (m *PredictorService) executeRegression(ctx context.Context, marketCache *MarketCache) error {
	m.log.Info("performing regression analysis")

	orders, err := marketCache.ExecutedOrders(ctx, sonm.OrderType_BID)
	if err != nil {
		return fmt.Errorf("failed to fetch active orders: %v", err)
	}

	// This is the hack, which mathematicians call "tuning regression parameters".
	// We have some benchmark cross-correlated, which results in bad fitting, for
	// example GPU count, correlated to hashrate etc. To avoid this we just
	// reset them to zero, which forces the model to exclude them from
	// training.
	for _, order := range orders {
		order.GetOrder().GetBenchmarks().SetCPUCores(0)
		order.GetOrder().GetBenchmarks().SetGPUCount(0)
	}

	classification, err := m.regression.ClassifyExt(orders)
	if err != nil {
		m.log.Warnw("failed to classify active orders", zap.Error(err))
		return err
	}

	m.updateClassification(classification)
	return nil
}

func (m *PredictorService) updateClassification(classification *OrderClassification) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.classification = classification
}

func (m *PredictorService) Classification() *OrderClassification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.classification
}
