package main

import (
	"context"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/locator"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

var (
	configFlag  string
	versionFlag bool
	appVersion  string
)

func main() {
	cmd.NewCmd("locator", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() {
	logger := logging.BuildLogger(-1, true)
	ctx := log.WithLogger(context.Background(), logger)

	cfg, err := locator.NewConfig(configFlag)
	if err != nil {
		log.G(ctx).Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.G(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	lc, err := locator.NewLocator(ctx, cfg, key)
	if err != nil {
		log.G(ctx).Error("cannot start Locator service", zap.Error(err))
		os.Exit(1)
	}

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr)

	log.G(ctx).Info("starting Locator service", zap.String("bind_addr", cfg.ListenAddr))
	if err := lc.Serve(); err != nil {
		log.G(ctx).Error("cannot start Locator service", zap.Error(err))
		os.Exit(1)
	}
}
