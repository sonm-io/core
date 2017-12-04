package main

import (
	"fmt"
	"os"

	"crypto/ecdsa"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/node"
	"golang.org/x/net/context"
)

var (
	configPath  = flag.String("config", "node.yaml", "Local Node config path")
	showVersion = flag.BoolP("version", "v", false, "Show Node version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Locator %s\r\n", version)
		return
	}

	cfg, err := node.NewConfig(*configPath)
	if err != nil {
		fmt.Printf("Err: Cannot load config file: %s\r\n", err)
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.LogLevel(), true)
	ctx := log.WithLogger(context.Background(), logger)

	key, err := loadKeys(cfg)
	if err != nil {
		fmt.Printf("Cannot load Etherum keys: %s\r\n", err.Error())
		os.Exit(1)
	}

	n, err := node.New(ctx, cfg, key)
	if err != nil {
		fmt.Printf("Err: cannot build Node instance: %s\r\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting Local Node at %s...\r\n", cfg.ListenAddress())
	if err := n.Serve(); err != nil {
		fmt.Printf("Cannot start Local Node: %s\r\n", err.Error())
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
