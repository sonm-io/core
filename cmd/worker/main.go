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
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("hub", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() error {
	ctx := context.Background()
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

	w, err := miner.NewMiner(cfg, miner.WithContext(ctx), miner.WithKey(key), miner.WithStateStorage(storage))
	if err != nil {
		return fmt.Errorf("failed to create Worker instance: %s", err)
	}

	h, err := hub.New(cfg, hub.WithVersion(appVersion), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(credentials), hub.WithCertRotator(certRotator), hub.WithWorker(w))
	if err != nil {
		return fmt.Errorf("failed to create new Hub: %s", err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		h.Close()
	}()

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	if err = h.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}

	return nil
}
