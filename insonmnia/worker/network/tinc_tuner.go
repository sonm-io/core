package network

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/ipam"
	netdriver "github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/structs"
	"go.uber.org/zap"
)

type TincTuner struct {
	client     *client.Client
	netDriver  *TincNetworkDriver
	ipamDriver *TincIPAMDriver
}

type TincCleaner struct {
	ctx       context.Context
	networkID string
	client    *client.Client
}

func NewTincTuner(ctx context.Context, config *TincNetworkConfig) (*TincTuner, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	netDriver, ipamDriver, err := NewTinc(ctx, cli, config)
	if err != nil {
		return nil, err
	}

	tuner := TincTuner{
		client:     cli,
		netDriver:  netDriver,
		ipamDriver: ipamDriver,
	}
	err = tuner.runDriver(ctx)
	if err != nil {
		return nil, err
	}
	return &tuner, nil
}

func (t *TincTuner) runDriver(ctx context.Context) error {
	pluginDir := filepath.Dir(t.netDriver.config.DockerNetPluginSockPath)
	err := os.MkdirAll(pluginDir, 0770)
	if err != nil {
		return err
	}

	ipamDir := filepath.Dir(t.ipamDriver.config.DockerIPAMPluginSockPath)
	err = os.MkdirAll(ipamDir, 0770)
	if err != nil {
		return err
	}

	netListener, err := sockets.NewUnixSocket(t.netDriver.config.DockerNetPluginSockPath, syscall.Getgid())
	if err != nil {
		return err
	}

	ipamListener, err := sockets.NewUnixSocket(t.ipamDriver.config.DockerIPAMPluginSockPath, syscall.Getgid())
	if err != nil {
		// cleanup
		netListener.Close()
		return err
	}

	netHandle := netdriver.NewHandler(t.netDriver)
	ipamHandle := ipam.NewHandler(t.ipamDriver)

	go func() {
		<-ctx.Done()
		log.G(ctx).Info("stopping tinc socket listener")
		netListener.Close()
		ipamListener.Close()
	}()
	go func() {
		log.G(ctx).Info("tinc ipam plugin has been initialized")
		ipamHandle.Serve(ipamListener)
	}()
	go func() {
		log.G(ctx).Info("tinc network plugin has been initialized")
		netHandle.Serve(netListener)
	}()
	return nil
}

func (t *TincTuner) Tune(ctx context.Context, net *structs.NetworkSpec, hostConfig *container.HostConfig, config *network.NetworkingConfig) (Cleanup, error) {
	tincNet, err := t.netDriver.InsertTincNetwork(net, hostConfig.Resources.CgroupParent)
	if err != nil {
		return nil, err
	}
	opts := map[string]string{"id": tincNet.NodeID}

	createOpts := types.NetworkCreate{
		Driver:  "tinc",
		Options: opts,
	}
	createOpts.IPAM = &network.IPAM{
		Driver: "tincipam",
		Config: []network.IPAMConfig{
			{
				Subnet: tincNet.Pool.String(),
			},
		},
		Options: opts,
	}

	response, err := t.client.NetworkCreate(ctx, net.NetID, createOpts)
	if err != nil {
		log.G(ctx).Warn("failed to create tinc network", zap.Error(err))
		return nil, err
	}
	//t.netDriver.RegisterNetworkMapping(response.ID, net.ID())
	if config.EndpointsConfig == nil {
		config.EndpointsConfig = make(map[string]*network.EndpointSettings)
		config.EndpointsConfig[response.ID] = &network.EndpointSettings{
			//IPAMConfig: &network.EndpointIPAMConfig{
			//	IPv4Address: net.NetworkAddr(),
			//},
			//IPAddress: net.NetworkAddr(),
			//NetworkID: response.ID,
			DriverOpts: opts,
		}
	}

	return &TincCleaner{
		ctx:       ctx,
		client:    t.client,
		networkID: response.ID,
	}, nil
}

func (t *TincTuner) GetCleaner(ctx context.Context, ID string) (Cleanup, error) {
	if _, ok := t.netDriver.Networks[ID]; !ok {
		return nil, errors.New("failed to find network with id " + ID)
	}
	return &TincCleaner{
		ctx:       ctx,
		client:    t.client,
		networkID: ID,
	}, nil
}

func (t *TincTuner) Tuned(ID string) bool {
	return t.netDriver.HasNetwork(ID)
}

func (t *TincTuner) GenerateInvitation(ID string) (*structs.NetworkSpec, error) {
	return t.netDriver.GenerateInvitation(ID)
}

func (t *TincCleaner) Close() (err error) {
	timeout := time.Millisecond * 100
	for i := 0; i < 10; i++ {
		err = t.client.NetworkRemove(t.ctx, t.networkID)
		if err == nil {
			return
		}
		log.S(t.ctx).Warnf("failed to remove network, retrying after %s", timeout)
		timeout = timeout * 2
		if timeout > time.Second*2 {
			timeout = time.Second * 2
		}
		time.Sleep(timeout)
	}
	return
}

func cloneOptions(from map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range from {
		result[k] = v
	}
	return result
}
