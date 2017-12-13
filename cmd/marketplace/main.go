package main

import (
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/marketplace"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	configPath  = flag.String("config", "marketplace.yaml", "Path to marketplace config file")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Marketplace %s\r\n", version)
		return
	}

	ctx := context.Background()

	cfg, err := marketplace.NewConfig(*configPath)
	if err != nil {
		log.GetLogger(ctx).Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.GetLogger(ctx).Error("failed to load private key", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(-1, true)
	ctx = log.WithLogger(context.Background(), logger)

	mp, err := marketplace.NewMarketplace(ctx, cfg, key)
	if err != nil {
		log.GetLogger(ctx).Error("failed to instantiate marketplace", zap.Error(err))
	}

	log.G(ctx).Info("starting Marketplace service", zap.String("bind_addr", cfg.ListenAddr))
	if err := mp.Serve(); err != nil {
		log.G(ctx).Error("cannot start Marketplace service", zap.Error(err))
		os.Exit(-1)
	}
}
