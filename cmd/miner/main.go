package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	flag "github.com/ogier/pflag"
	"golang.org/x/net/context"

	"github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"

	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/miner"

	log "github.com/noxiouz/zapctx/ctxlog"

	_ "expvar"
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
		ctxlog.GetLogger(ctx).Error("Cannot load config", zap.Error(err))
		os.Exit(1)
	}

	logger := logging.BuildLogger(cfg.Logging().Level, common.DevelopmentMode)
	ctx = log.WithLogger(ctx, logger)

	builder := miner.NewMinerBuilder(cfg)
	builder.Context(ctx)
	m, err := builder.Build()
	if err != nil {
		ctxlog.GetLogger(ctx).Fatal("failed to create a new Miner", zap.Error(err))
	}

	var onSigInt = make(chan struct{})

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)
		<-c
		close(onSigInt)
	}()

	if debugSrvCfg := cfg.DebugServer(); debugSrvCfg != nil && debugSrvCfg.Endpoint != "" {
		go func() {
			addr := debugSrvCfg.Endpoint
			ctxlog.GetLogger(ctx).Info("start HTTP debug server", zap.String("addr", addr))
			ln, lerr := net.Listen("tcp", addr)
			if lerr != nil {
				ctxlog.GetLogger(ctx).Fatal("failed to create TCP listener", zap.String("addr", addr), zap.Error(lerr))
			}
			defer ln.Close()

			go func() {
				<-onSigInt
				ln.Close()
			}()

			// expvar registers handler in http.DefaultServeMux
			http.Serve(ln, http.DefaultServeMux)
		}()
	}

	go func() {
		<-onSigInt
		m.Close()
	}()

	// TODO: check error type
	if err = m.Serve(); err != nil {
		ctxlog.GetLogger(ctx).Error("Server stop", zap.Error(err))
	}
}
