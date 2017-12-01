package main

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/accounts"
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

	key, err := loadKeys(cfg)
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

	builder := hub.NewBuilder()
	builder.
		WithVersion(version).
		WithContext(ctx).
		WithPrivateKey(key).
		WithCreds(creds).
		WithCertRotator(certRotator)

	// h, err := hub.New(ctx, cfg, version)
	h, err := builder.Build(cfg)
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

func loadKeys(c *hub.Config) (*ecdsa.PrivateKey, error) {
	p := accounts.NewFmtPrinter()
	ko, err := accounts.DefaultKeyOpener(p, c.Eth.Keystore, c.Eth.Passphrase)
	if err != nil {
		return nil, err
	}

	_, err = ko.OpenKeystore()
	if err != nil {
		return nil, err
	}

	return ko.GetKey()
}
