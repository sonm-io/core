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

func run() {
	ctx := context.Background()

	cfg, err := hub.NewConfig(configFlag)
	if err != nil {
		fmt.Printf("failed to load config: %s\r\n", err)
		os.Exit(1)
	}

	logLevel, err := cfg.LogLevel()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(logLevel)
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

	h, err := hub.New(ctx, cfg, hub.WithVersion(appVersion), hub.WithContext(ctx),
		hub.WithPrivateKey(key), hub.WithCreds(creds), hub.WithCertRotator(certRotator))
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
