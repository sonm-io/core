package main

import (
	"fmt"
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"github.com/sonm-io/core/insonmnia/hub"
	"github.com/sonm-io/core/insonmnia/logging"

	log "github.com/noxiouz/zapctx/ctxlog"
)

var (
	configPath  = flag.String("config", "hub.yaml", "Path to hub config file")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Hub %s\r\n", version)
		return
	}

	ctx := context.Background()

	cfg, err := hub.NewConfig(*configPath)
	if err != nil {
		ctxlog.GetLogger(ctx).Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Logging.Level, true)
	ctx = log.WithLogger(ctx, logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		ctxlog.GetLogger(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		ctxlog.GetLogger(ctx).Error("failed to create cert rotator", zap.Error(err))
		os.Exit(1)
	}
	creds := util.NewTLS(TLSConfig)

	h, err := hub.New(ctx, cfg, version, hub.WithVersion(version), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(creds), hub.WithCertRotator(certRotator))
	if err != nil {
		ctxlog.GetLogger(ctx).Error("failed to create a new Hub", zap.Error(err))
		os.Exit(1)
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		h.Close()
	}()

	// TODO: check error type
	if err = h.Serve(); err != nil {
		ctxlog.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
