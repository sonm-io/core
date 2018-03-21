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
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	configFlag   string
	workerConfig string
	versionFlag  bool
	appVersion   string
)

func main() {
	c := cmd.NewCmd("hub", appVersion, &configFlag, &versionFlag, run)
	c.PersistentFlags().StringVar(&workerConfig, "worker_config", "worker.yaml", "Path to the worker config file")
	c.Execute()
}

func run() {
	ctx := context.Background()

	cfg, err := hub.NewConfig(configFlag)
	if err != nil {
		fmt.Printf("failed to load config: %s\r\n", err)
		os.Exit(1)
	}
	wCfg, err := miner.NewConfig(workerConfig)
	if err != nil {
		fmt.Printf("failed to load worker config: %s\r\n", err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.LogLevel())
	ctx = log.WithLogger(ctx, logger)

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		log.G(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		log.G(ctx).Error("failed to create cert rotator", zap.Error(err))
		os.Exit(1)
	}
	creds := util.NewTLS(TLSConfig)

	opts := make([]miner.Option, 0)
	opts = append(opts, miner.WithContext(ctx), miner.WithKey(key))
	w, err := miner.NewMiner(wCfg, opts...)
	if err != nil {
		log.G(ctx).Error("cannot create worker instance", zap.Error(err))
		os.Exit(1)
	}

	h, err := hub.New(ctx, cfg, hub.WithVersion(appVersion), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(creds), hub.WithCertRotator(certRotator), hub.WithWorker(w))
	if err != nil {
		log.G(ctx).Error("failed to create a new Hub", zap.Error(err))
		os.Exit(1)
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
}
