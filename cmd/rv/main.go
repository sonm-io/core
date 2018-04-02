package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/rendezvous"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
)

var (
	cfgPath     string
	versionFlag bool
	appVersion  string
)

func start() {
	cfg, err := rendezvous.NewServerConfig(cfgPath)
	if err != nil {
		fmt.Printf("failed to load config file: %s\r\n", err)
		os.Exit(1)
	}

	ctx := log.WithLogger(context.Background(), logging.BuildLogger(cfg.LogLevel()))

	certRotator, TLSConfig, err := util.NewHitlessCertRotator(ctx, cfg.PrivateKey)
	if err != nil {
		log.G(ctx).Error("failed to create certificate rotator", zap.Error(err))
		os.Exit(1)
	}
	defer certRotator.Close()

	credentials := util.NewTLS(TLSConfig)

	options := []rendezvous.Option{
		rendezvous.WithLogger(log.G(ctx)),
		rendezvous.WithCredentials(credentials),
	}
	server, err := rendezvous.NewServer(*cfg, options...)
	if err != nil {
		os.Exit(1)
	}

	go server.Run()
	defer server.Stop()

	waitInterrupted()
}

func waitInterrupted() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func main() {
	cmd.NewCmd("rendezvous", appVersion, &cfgPath, &versionFlag, start).Execute()
}
