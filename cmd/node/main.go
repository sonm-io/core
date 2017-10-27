package main

import (
	"fmt"
	"os"

	log "github.com/noxiouz/zapctx/ctxlog"
	flag "github.com/ogier/pflag"
	"github.com/sonm-io/core/common"
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

	logger := logging.BuildLogger(cfg.LogLevel(), common.DevelopmentMode)
	ctx := log.WithLogger(context.Background(), logger)

	n, err := node.New(ctx, cfg)
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
