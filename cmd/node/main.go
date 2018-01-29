package main

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
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
	cmd.NewCmd("sonmnode", appVersion, &configFlag, &versionFlag, run).Execute()
}

func run() {
	cfg, err := node.NewConfig(configFlag)
	if err != nil {
		fmt.Printf("cannot load config file: %s\r\n", err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.LogLevel(), true)
	ctx := log.WithLogger(context.Background(), logger)

	key, err := loadKeys(cfg)
	if err != nil {
		log.G(ctx).Error("cannot load Ethereum keys", zap.Error(err))
		os.Exit(1)
	}

	n, err := node.New(ctx, cfg, key)
	if err != nil {
		log.G(ctx).Error("cannot build node instance", zap.Error(err))
		os.Exit(1)
	}

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr())

	log.G(ctx).Error("starting node", zap.String("listen_addr", cfg.ListenAddress()))
	if err := n.Serve(); err != nil {
		log.G(ctx).Error("cannot start node", zap.Error(err))
		os.Exit(1)
	}
}

func loadKeys(c node.Config) (*ecdsa.PrivateKey, error) {
	p := accounts.NewFmtPrinter()
	ko, err := accounts.DefaultKeyOpener(p, c.KeyStore(), c.PassPhrase())
	if err != nil {
		return nil, err
	}

	_, err = ko.OpenKeystore()
	if err != nil {
		return nil, err
	}

	return ko.GetKey()
}
