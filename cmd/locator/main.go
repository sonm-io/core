package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/locator"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap"
)

var (
	configPath  = flag.String("config", "locator.yaml", "Path to locator config file")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Locator %s\r\n", version)
		return
	}

	ctx := context.Background()

	cfg, err := locator.NewConfig(*configPath)
	if err != nil {
		log.GetLogger(ctx).Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	key, err := accounts.LoadKeys(cfg.Eth.Keystore, cfg.Eth.Passphrase)
	if err != nil {
		log.GetLogger(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(-1, true)
	ctx = log.WithLogger(context.Background(), logger)

	lc, err := locator.NewLocator(ctx, cfg, key)
	if err != nil {
		log.G(ctx).Error("cannot start Locator service", zap.Error(err))
		os.Exit(1)
	}
	log.G(ctx).Info("starting Locator service", zap.String("bind_addr", cfg.ListenAddr))
	if err := lc.Serve(); err != nil {
		log.G(ctx).Error("cannot start Locator service", zap.Error(err))
		os.Exit(1)
	}
}
