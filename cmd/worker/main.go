package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/hub"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("worker", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	ctx := context.Background()
	waiter, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	cfg, err := miner.NewConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config file: %s", err)
	}

	logger := logging.BuildLogger(cfg.Logging.LogLevel())
	ctx = log.WithLogger(ctx, logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return fmt.Errorf("failed to load private key: %s", err)
	}

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to create certificate rotator: %s", err)
	}
	credentials := util.NewTLS(TLSConfig)

	storage, err := state.NewState(ctx, &cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to create state storage: %s", err)
	}

	waiter.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		log.G(ctx).Info("closing worker by interrupt signal")
		cancel()

		return nil
	})

	w, err := miner.NewMiner(cfg, miner.WithContext(ctx), miner.WithKey(key), miner.WithStateStorage(storage), miner.WithCreds(credentials))
	if err != nil {
		return fmt.Errorf("failed to create Worker instance: %s", err)
	}

	h, err := hub.New(cfg, hub.WithVersion(appVersion), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(credentials), hub.WithCertRotator(certRotator), hub.WithWorker(w))
	if err != nil {
		return fmt.Errorf("failed to create new Hub: %s", err)
	}

	//TODO: fixme dangling goroutine
	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	if err = h.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}
	waiter.Wait()

	return nil
}
