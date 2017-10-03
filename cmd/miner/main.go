package main

import (
	"fmt"
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"golang.org/x/net/context"

	"go.uber.org/zap"

	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"io/ioutil"
)

var (
	configPath  = flag.String("config", "miner.yaml", "Path to miner config file")
	showVersion = flag.BoolP("version", "v", false, "Show Hub version and exit")
	version     string
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("SONM Miner %s\r\n", version)
		return
	}

	ctx := context.Background()

	cfg, err := miner.NewConfig(*configPath)
	if err != nil {
		log.G(ctx).Error("Cannot load config", zap.Error(err))
		os.Exit(1)
	}

	if _, err := os.Stat(cfg.UUIDPath()); os.IsNotExist(err) {
		ioutil.WriteFile(cfg.UUIDPath(), []byte(uuid.New()), 0660)
	}
	uuidData, err := ioutil.ReadFile(cfg.UUIDPath())
	if err != nil {
		log.G(ctx).Error("Cannot load uuid", zap.Error(err))
		os.Exit(1)
	}
	uuid := string(uuidData)

	logger := logging.BuildLogger(cfg.Logging().Level, common.DevelopmentMode)
	ctx = log.WithLogger(ctx, logger)

	builder := miner.NewMinerBuilder(cfg)
	builder.Context(ctx)
	builder.UUID(uuid)
	m, err := builder.Build()
	if err != nil {
		log.G(ctx).Fatal("failed to create a new Miner", zap.Error(err))
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		m.Close()
	}()

	// TODO: check error type
	if err = m.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}
}
