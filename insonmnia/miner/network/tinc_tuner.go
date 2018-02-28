package network

import (
	"context"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/ipam"
	netdriver "github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/structs"
)

type TincTuner struct {
	client     *client.Client
	netDriver  *TincNetworkDriver
	ipamDriver *TincIPAMDriver
}

type TincCleaner struct {
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
		return err
	}

	netHandle := netdriver.NewHandler(t.netDriver)
	ipamHandle := ipam.NewHandler(t.ipamDriver)

	go func() {
		<-ctx.Done()
		log.G(context.Background()).Info("stopping tinc socket listener")
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

//TODO: pass context from outside
func (t *TincTuner) Tune(net structs.Network, hostConfig *container.HostConfig, config *network.NetworkingConfig) (Cleanup, error) {
	opts := cloneOptions(net.NetworkOptions())
	opts["cgroup_parent"] = hostConfig.Resources.CgroupParent
	createOpts := types.NetworkCreate{
		Driver:  "tinc",
		Options: opts,
	}
	if len(net.NetworkCIDR()) != 0 {
		createOpts.IPAM = &network.IPAM{
			Driver: "default",
			Config: make([]network.IPAMConfig, 0),
		}
		createOpts.IPAM.Config = append(createOpts.IPAM.Config, network.IPAMConfig{Subnet: net.NetworkCIDR()})
	}
	response, err := t.client.NetworkCreate(context.Background(), net.ID(), createOpts)
	if err != nil {
		return nil, err
	}
	if config.EndpointsConfig == nil {
		config.EndpointsConfig = make(map[string]*network.EndpointSettings)
		config.EndpointsConfig[response.ID] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: net.NetworkAddr(),
			},
			IPAddress: net.NetworkAddr(),
			NetworkID: response.ID,
		}
	}

	return &TincCleaner{
		client:    t.client,
		networkID: response.ID,
	}, nil
}

func (t *TincCleaner) Close() error {
	return t.client.NetworkRemove(context.Background(), t.networkID)
}

func cloneOptions(from map[string]string) map[string]string {
	result := map[string]string{}
	for k, v := range from {
		result[k] = v
	}
	return result
}
