package main

import (
	"context"
	"fmt"
	"net"
	"net/url"

	"github.com/jinzhu/configor"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/cmd"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/insonmnia/sysinit"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/unix"
)

type Config struct {
	Endpoint  string          `yaml:"endpoint" default:"unix:///var/run/qos.sock"`
	Logging   logging.Config  `yaml:"logging"`
	GPUVendor string          `yaml:"gpu_vendor"`
	SysInit   *sysinit.Config `yaml:"sysinit"`
}

func main() {
	_ = cmd.NewCmd(run).Execute()
}

func run(app cmd.AppContext) error {
	cfg := &Config{}
	if err := configor.Load(cfg, app.ConfigPath); err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
	}

	log, err := logging.BuildLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to build logger instance: %s", err)
	}

	ctx := ctxlog.WithLogger(context.Background(), log)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return cmd.WaitInterrupted(ctx)
	})
	wg.Go(func() error {
		remoteQOS, err := network.NewRemoteQOS()
		if err != nil {
			return err
		}

		remoteTuner, err := gpu.NewRemoteTuner(ctxlog.WithLogger(ctx, log), cfg.GPUVendor)
		if err != nil {
			return err
		}

		remoteInit := sysinit.NewInitService(cfg.SysInit, log.Sugar())

		uri, err := url.Parse(cfg.Endpoint)
		if err != nil {
			return err
		}

		if uri.Scheme == "unix" {
			uri.Host = uri.Path
			_ = unix.Unlink(uri.Path)
		}

		listener, err := net.Listen(uri.Scheme, uri.Host)
		if err != nil {
			return err
		}

		log.Sugar().Infof("exposing QOS server on %s %s", uri.Scheme, uri.Host)
		defer log.Sugar().Infof("stopped QOS server on %s %s", uri.Scheme, uri.Host)

		server := xgrpc.NewServer(log, xgrpc.RequestLogInterceptor([]string{}))
		sonm.RegisterQOSServer(server, remoteQOS)
		sonm.RegisterRemoteGPUTunerServer(server, remoteTuner)
		sonm.RegisterInitServer(server, remoteInit)

		wg.Go(func() error {
			return server.Serve(listener)
		})

		<-ctx.Done()
		return listener.Close()
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("remote QOS termination: %s", err)
	}

	return nil
}
