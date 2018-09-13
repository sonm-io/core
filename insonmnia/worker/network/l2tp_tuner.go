package network

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"

	"fmt"

	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/ipam"
	netDriver "github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type L2TPTuner struct {
	cfg        *L2TPConfig
	cli        *client.Client
	netDriver  *L2TPNetworkDriver
	ipamDriver *IPAMDriver
}

func NewL2TPTuner(ctx context.Context, cfg *L2TPConfig) (*L2TPTuner, error) {
	err := os.MkdirAll(cfg.ConfigDir, 0770)
	if err != nil {
		return nil, err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	state, err := newL2TPNetworkState(ctx, cfg.StatePath)
	if err != nil {
		return nil, err
	}

	tuner := &L2TPTuner{
		cfg:        cfg,
		cli:        cli,
		netDriver:  NewL2TPDriver(ctx, state),
		ipamDriver: NewIPAMDriver(ctx, state),
	}
	if err := tuner.Run(ctx); err != nil {
		return nil, err
	}

	return tuner, nil
}

func (t *L2TPTuner) GenerateInvitation(ID string) (*sonm.NetworkSpec, error) {
	return nil, errors.New("not supported")
}

func (t *L2TPTuner) Tuned(ID string) bool {
	return false
}

func (t *L2TPTuner) Run(ctx context.Context) error {
	syscall.Unlink(t.cfg.NetSocketPath)
	netListener, err := net.Listen("unix", t.cfg.NetSocketPath)
	if err != nil {
		log.G(ctx).Error("Failed to listen", zap.Error(err))
		return err
	}

	netHandler := netDriver.NewHandler(t.netDriver)

	syscall.Unlink(t.cfg.IPAMSocketPath)
	ipamListener, err := net.Listen("unix", t.cfg.IPAMSocketPath)
	if err != nil {
		log.G(ctx).Error("Failed to listen", zap.Error(err))
		return err
	}

	ipamHandler := ipam.NewHandler(t.ipamDriver)

	go func() {
		<-ctx.Done()
		log.G(ctx).Info("stopping l2tp socket listener")
		netListener.Close()
		ipamListener.Close()
	}()

	go func() {
		log.G(ctx).Info("l2tp ipam driver has been initialized")
		if err := ipamHandler.Serve(ipamListener); err != nil {
			log.G(ctx).Error("IPAM driver failed", zap.Error(err))
		}
	}()

	go func() {
		log.G(ctx).Info("l2tp network driver has been initialized")
		if err := netHandler.Serve(netListener); err != nil {
			log.G(ctx).Error("Network driver failed", zap.Error(err))
		}
	}()

	return nil
}

func (t *L2TPTuner) Tune(ctx context.Context, net *sonm.NetworkSpec, hostCfg *container.HostConfig, netCfg *network.NetworkingConfig) (Cleanup, error) {
	log.G(ctx).Info("tuning l2tp")
	configPath, err := t.writeConfig(net.GetID(), net.Options)
	if err != nil {
		return nil, err
	}

	driverOpts := map[string]string{"config": configPath}
	createOpts := types.NetworkCreate{
		Driver:  "l2tp_net",
		Options: driverOpts,
		IPAM:    &network.IPAM{Driver: "l2tp_ipam", Options: driverOpts},
	}

	response, err := t.cli.NetworkCreate(ctx, net.GetID(), createOpts)
	if err != nil {
		return nil, err
	}

	if netCfg.EndpointsConfig == nil {
		netCfg.EndpointsConfig = make(map[string]*network.EndpointSettings)
		netCfg.EndpointsConfig[response.ID] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: net.Addr},
			IPAddress:  net.Addr,
			NetworkID:  response.ID,
		}
	}

	return &L2TPCleaner{
		ctx:        ctx,
		cli:        t.cli,
		networkID:  response.ID,
		configPath: configPath,
	}, nil
}

func (t *L2TPTuner) GetCleaner(ctx context.Context, ID string) (Cleanup, error) {
	if _, ok := t.netDriver.Networks[ID]; !ok {
		return nil, errors.New("failed to find network with id " + ID)
	}
	configPath := t.cfg.ConfigDir + "/" + ID
	return &L2TPCleaner{
		ctx:        ctx,
		cli:        t.cli,
		networkID:  ID,
		configPath: configPath,
	}, nil
}

func (t *L2TPTuner) writeConfig(netID string, opts map[string]string) (string, error) {
	var data string
	for k, v := range opts {
		data += fmt.Sprintf("%s: %s\n", k, v)
	}

	path := t.cfg.ConfigDir + "/" + netID

	return path, ioutil.WriteFile(path, []byte(data), 0700)
}

type L2TPCleaner struct {
	ctx        context.Context
	cli        *client.Client
	networkID  string
	configPath string
}

func (t *L2TPCleaner) Close() error {
	log.G(t.ctx).Info("closing l2tp driver")
	if err := t.cli.NetworkRemove(t.ctx, t.networkID); err != nil {
		return err
	}

	return os.Remove(t.configPath)
}
