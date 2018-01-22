package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sonm-io/core/insonmnia/hub"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
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
		log.GetLogger(ctx).Error("failed to load config", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Logging.Level, true)
	ctx = log.WithLogger(ctx, logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.GetLogger(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		log.GetLogger(ctx).Error("failed to create cert rotator", zap.Error(err))
		os.Exit(1)
	}
	creds := util.NewTLS(TLSConfig)

	h, err := hub.New(ctx, cfg, version, hub.WithVersion(version), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(creds), hub.WithCertRotator(certRotator))
	if err != nil {
		log.GetLogger(ctx).Error("failed to create a new Hub", zap.Error(err))
		os.Exit(1)
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		h.Close()
	}()

	go func() {
		log.GetLogger(ctx).Info(
			"starting metrics server", zap.String("metrics_addr", cfg.MetricsListenAddr))
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(cfg.MetricsListenAddr, nil)
	}()

	// TODO: check error type
	if err = h.Serve(); err != nil {
		log.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
