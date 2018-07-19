package optimus

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
)

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

	marketCache := newMarketCache(newMarketScanner(m.cfg.Marketplace, dwh), m.cfg.Marketplace.Interval)

	wg := errgroup.Group{}
	benchmarkMapping, err := benchmarks.NewLoader(m.cfg.Benchmarks.URL).Load(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load benchmarks: %v", err)
	}

	market, err := blockchain.NewAPI(ctx, blockchain.WithConfig(m.cfg.Blockchain))
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

		blacklist := newMultiBlacklist(
			newBlacklist(ethAddr, dwh, m.log),
			newBlacklist(masterAddr, dwh, m.log),
		)

		worker, err := registry.NewWorkerManagement(ctx, m.cfg.Node.Endpoint, m.cfg.Node.PrivateKey.Unwrap())
		if err != nil {
			return err
		}

		control, err := newWorkerEngine(cfg, ethAddr, masterAddr, blacklist, worker, market.Market(), marketCache, benchmarkMapping, m.cfg.Optimization, m.log)
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
