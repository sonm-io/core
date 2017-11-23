package main

import (
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/marketplace"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	listenAddr  = flag.String("addr", ":9095", "Marketplace service listen address")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Marketplace %s\r\n", version)
		return
	}

	logger := logging.BuildLogger(-1, common.DevelopmentMode)
	ctx := log.WithLogger(context.Background(), logger)

	mp := marketplace.NewMarketplace(ctx, *listenAddr)
	log.G(ctx).Info("starting Marketplace service", zap.String("bind_addr", *listenAddr))
	if err := mp.Serve(); err != nil {
		log.G(ctx).Error("cannot start Marketplace service", zap.Error(err))
		os.Exit(-1)
	}
}
