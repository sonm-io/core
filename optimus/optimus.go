package optimus

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/debug"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/metadata"
)

type Optimus struct {
	cfg     *Config
	version string
	log     *zap.SugaredLogger
}

func NewOptimus(cfg *Config, options ...Option) (*Optimus, error) {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	m := &Optimus{
		cfg:     cfg,
		version: opts.Version,
		log:     opts.Log.With(zap.String("source", "optimus")),
	}

	m.log.Debugw("configuring Optimus", zap.Any("config", cfg))

	return m, nil
}

func (m *Optimus) Run(ctx context.Context) error {
	m.log.Infow("starting Optimus", zap.String("version", m.version))
	defer m.log.Info("Optimus has been stopped")

	registry := newRegistry()
	defer registry.Close()

	dwh, err := registry.NewDWH(ctx, m.cfg.Marketplace.Endpoint, m.cfg.Marketplace.PrivateKey.Unwrap())
	if err != nil {
		return err
	}

	marketCache := newMarketCache(newMarketScanner(m.cfg.Marketplace, dwh), m.cfg.Marketplace.Interval)

	benchmarkMapping, err := benchmarks.NewLoader(m.cfg.Benchmarks.URL).Load(context.Background())
	if err != nil {
		return fmt.Errorf("failed to load benchmarks: %v", err)
	}

	market, err := blockchain.NewAPI(ctx, blockchain.WithConfig(m.cfg.Blockchain))
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if m.cfg.Debug == nil {
			return nil
		}

		return debug.ServePProf(ctx, *m.cfg.Debug, m.log.Desugar())
	})

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

		// TODO: Well, 10 parameters seems to be WAT.
		workerClient := &workerManagementClientAPI{worker}
		control, err := newWorkerEngine(cfg, ethAddr, masterAddr, blacklist, workerClient, market, marketCache, benchmarkMapping, newTagger(m.version), m.log)
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
