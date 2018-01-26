package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	configFlag  string
	versionFlag bool
	version     string
)

func main() {
	cmd.NewCmd("sonmminer", &configFlag, &versionFlag, run).Execute()
}

func run() {
	if versionFlag {
		fmt.Printf("SONM Worker %s\r\n", version)
		return
	}

	ctx := context.Background()

	cfg, err := miner.NewConfig(configFlag)
	if err != nil {
		log.G(ctx).Error("cannot load config", zap.Error(err))
		os.Exit(1)
	}

	key, err := cfg.ETH().LoadKey()
	if err != nil {
		log.GetLogger(ctx).Error("failed load private key", zap.Error(err))
		os.Exit(1)
	}

	if _, err := os.Stat(cfg.UUIDPath()); os.IsNotExist(err) {
		ioutil.WriteFile(cfg.UUIDPath(), []byte(uuid.New()), 0660)
	}

	uuidData, err := ioutil.ReadFile(cfg.UUIDPath())
	if err != nil {
		log.G(ctx).Error("cannot load uuid", zap.Error(err))
		os.Exit(1)
	}

	workerID := string(uuidData)
	logger := logging.BuildLogger(cfg.Logging().Level, true)
	ctx = log.WithLogger(ctx, logger)

	m, err := miner.NewMiner(cfg, miner.WithContext(ctx), miner.WithKey(key), miner.WithUUID(workerID))
	if err != nil {
		log.G(ctx).Error("cannot create worker instance", zap.Error(err))
		os.Exit(1)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		m.Close()
	}()

	go util.StartPrometheus(ctx, cfg.MetricsListenAddr())

	if err = m.Serve(); err != nil {
		log.G(ctx).Error("Server stop", zap.Error(err))
	}
}
