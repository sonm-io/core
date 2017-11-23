package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/insonmnia/locator"
	"github.com/sonm-io/core/insonmnia/logging"
	"go.uber.org/zap"
)

var (
	listenAddr  = flag.String("addr", ":9090", "Locator service listen addr")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Locator %s\r\n", version)
		return
	}

	cfg := locator.DefaultConfig(*listenAddr)

	logger := logging.BuildLogger(-1, true)
	ctx := log.WithLogger(context.Background(), logger)

	lc := locator.NewLocator(ctx, cfg)
	log.G(ctx).Info("starting Locator service", zap.String("bind_addr", cfg.ListenAddr))
	if err := lc.Serve(); err != nil {
		log.G(ctx).Error("cannot start Locator service", zap.Error(err))
		os.Exit(1)
	}
}
