package network

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	cnet "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func NewTincNetwork(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkDriver, *TincIPAMDriver, error) {
	err := os.MkdirAll(config.ConfigDir, 0770)
	if err != nil {
		return nil, nil, err
	}

	state := &TincNetworkState{
		ctx:      ctx,
		config:   config,
		cli:      client,
		networks: make(map[string]*TincNetwork),
		pools:    make(map[string]*net.IPNet),
	}

	return &TincNetworkDriver{state}, &TincIPAMDriver{state}, nil
}

type TincNetworkOptions struct {
	Invitation   string
	EnableBridge bool
}

type TincNetwork struct {
	ID              string
	Options         *TincNetworkOptions
	IPv4Data        []*network.IPAMData
	IPv6Data        []*network.IPAMData
	ConfigPath      string
	cli             *client.Client
	TincContainerID string
}

type TincNetworkState struct {
	ctx      context.Context
	config   *TincNetworkConfig
	mu       sync.RWMutex
	cli      *client.Client
	networks map[string]*TincNetwork
	pools    map[string]*net.IPNet
}

type TincNetworkDriver struct {
	*TincNetworkState
}

type TincIPAMDriver struct {
	*TincNetworkState
}

func (t *TincIPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	log.G(t.ctx).Info("ipam: received GetCapabilities request", zap.Any("request", nil))
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (t *TincIPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	log.G(t.ctx).Info("ipam: received GetDefaultAddressSpaces request", zap.Any("request", nil))
	return nil, nil
}

func (t *TincIPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	log.G(t.ctx).Info("ipam: received RequestPool request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()
	id := uuid.New()
	_, n, err := net.ParseCIDR(request.Pool)
	if err != nil {
		return nil, err
	}
	t.pools[id] = n
	return &ipam.RequestPoolResponse{
		PoolID: id,
		Pool:   request.Pool,
		Data:   request.Options,
	}, nil
}

func (t *TincIPAMDriver) ReleasePool(request *ipam.ReleasePoolRequest) error {
	log.G(t.ctx).Info("ipam: received ReleasePool request", zap.Any("request", nil))
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.pools, request.PoolID)
	return nil
}

func (t *TincIPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	log.G(t.ctx).Info("ipam: received RequestAddress request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()

	pool, ok := t.pools[request.PoolID]
	if !ok {
		return nil, errors.New("pool not found")
	}

	mask, _ := pool.Mask.Size()

	ty, ok := request.Options["RequestAddressType"]
	if ok && ty == "com.docker.network.gateway" {
		ip := make(net.IP, len(pool.IP))
		copy(ip, pool.IP)
		if len(ip) == 4 {
			ip[3]++
		} else {
			ip[15]++
		}
		return &ipam.RequestAddressResponse{
			Address: ip.String() + "/" + fmt.Sprint(mask),
		}, nil
	}

	if mask == 0 {
		return nil, errors.New("invalid subnet")
	}
	return &ipam.RequestAddressResponse{
		Address: request.Address + "/" + fmt.Sprint(mask),
	}, nil
}

func (t *TincIPAMDriver) ReleaseAddress(*ipam.ReleaseAddressRequest) error {
	log.G(t.ctx).Info("ipam: received ReleaseAddress request", zap.Any("request", nil))
	return nil
}

func (t *TincNetworkDriver) newTincNetwork(request *network.CreateNetworkRequest) (*TincNetwork, error) {
	opts, err := ParseNetworkOpts(request.Options)
	if err != nil {
		return nil, err
	}
	configPath := t.config.ConfigDir + "/" + request.NetworkID
	err = os.MkdirAll(configPath, 0770)
	if err != nil {
		return nil, err
	}
	containerConfig := &container.Config{
		Image: "antmat/tinc",
	}
	hostConfig := &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
	}
	netConfig := &cnet.NetworkingConfig{}
	resp, err := t.cli.ContainerCreate(t.ctx, containerConfig, hostConfig, netConfig, "")
	if err != nil {
		return nil, err
	}
	err = t.cli.ContainerStart(t.ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	log.S(t.ctx).Infof("started container %s", resp.ID)

	return &TincNetwork{
		ID:       request.NetworkID,
		Options:  opts,
		IPv4Data: request.IPv4Data,
		IPv6Data: request.IPv6Data,
		// TODO: configurable
		ConfigPath:      configPath,
		cli:             t.cli,
		TincContainerID: resp.ID,
	}, nil
}

func ParseNetworkOpts(data map[string]interface{}) (*TincNetworkOptions, error) {
	g, ok := data["com.docker.network.generic"]
	if !ok {
		return nil, errors.New("no options passed - invitation is required")
	}
	generic, ok := g.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid type of generic options")
	}
	invitation, ok := generic["invitation"]
	if !ok {
		return nil, errors.New("invitation is required")
	}
	_, ok = generic["enable_bridge"]
	return &TincNetworkOptions{Invitation: invitation.(string), EnableBridge: ok}, nil
}

func (t *TincNetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	log.G(t.ctx).Info("received GetCapabilities request", zap.Any("request", nil))
	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (t *TincNetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	log.G(t.ctx).Info("received CreateNetwork request", zap.Any("request", request))
	network, err := t.newTincNetwork(request)
	if err != nil {
		return err
	}

	time.Sleep(time.Second)

	err = network.runCommand(t.ctx, "tinc", "--batch", "-n", network.ID, "-c", network.ConfigPath, "join", network.Options.Invitation)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.networks[network.ID] = network

	return nil
}

func (t *TincNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(t.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (t *TincNetworkDriver) popNetwork(ID string) *TincNetwork {
	t.mu.Lock()
	defer t.mu.Unlock()
	net, ok := t.networks[ID]
	if !ok {
		return nil
	}
	delete(t.networks, ID)
	return net
}

func (t *TincNetworkDriver) shutdownNetwork(network *TincNetwork) error {
	err := os.RemoveAll(network.ConfigPath)
	if err != nil {
		return err
	}
	return nil
}

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	net := t.popNetwork(request.NetworkID)
	if net == nil {
		return errors.Errorf("no network with id %s", request.NetworkID)
	}
	t.shutdownNetwork(net)
	return nil
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	log.G(t.ctx).Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(t.ctx).Info("received CreateEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()
	net, ok := t.networks[request.NetworkID]
	if !ok {
		return nil, errors.Errorf("no such network %s", request.NetworkID)
	}
	iface := net.ID[:15]
	//TODO: each pool should be considered
	pool := net.IPv4Data[0].Pool
	selfAddr := strings.Split(request.Interface.Address, "/")[0]
	err := net.runCommand(t.ctx, "tinc", "-n", net.ID, "-c", net.ConfigPath, "start",
		"-o", "Interface="+iface, "-o", "Subnet="+pool, "-o", "Subnet="+selfAddr+"/32", "-o", "LogLevel=9")
	if err != nil {
		return nil, err
	}

	return &network.CreateEndpointResponse{}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	log.G(t.ctx).Info("received DeleteEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()

	net, ok := t.networks[request.NetworkID]
	if !ok {
		return errors.Errorf("no such network %s", request.NetworkID)
	}
	err := net.runCommand(t.ctx, "tinc", "--batch", "-n", net.ID, "-c", net.ConfigPath, "stop")
	if err != nil {
		return err
	}

	return nil
}

func (t *TincNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	log.G(t.ctx).Info("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	return &network.InfoResponse{Value: val}, nil
}

func (t *TincNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(t.ctx).Info("received Join request", zap.Any("request", request))
	t.mu.RLock()
	defer t.mu.RUnlock()
	iface := request.NetworkID[:15]
	return &network.JoinResponse{DisableGatewayService: false, InterfaceName: network.InterfaceName{SrcName: iface, DstPrefix: "tinc"}}, nil
}

func (t *TincNetworkDriver) Leave(request *network.LeaveRequest) error {
	log.G(t.ctx).Info("received Leave request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	log.G(t.ctx).Info("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	log.G(t.ctx).Info("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	log.G(t.ctx).Info("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	log.G(t.ctx).Info("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetwork) runCommand(ctx context.Context, name string, arg ...string) error {
	cmd := append([]string{name}, arg...)
	cfg := types.ExecConfig{
		User:         "root",
		Detach:       false,
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	execId, err := t.cli.ContainerExecCreate(ctx, t.TincContainerID, cfg)
	if err != nil {
		log.G(ctx).Warn("ContainerExecCreate finished with error", zap.Error(err))
		return err
	}

	conn, err := t.cli.ContainerExecAttach(ctx, execId.ID, cfg)
	if err != nil {
		log.G(ctx).Warn("ContainerExecAttach finished with error", zap.Error(err))
	}

	byteOut, err := ioutil.ReadAll(conn.Reader)
	out := string(byteOut)
	if err != nil {
		log.S(ctx).Errorf("failed to execute command - %s %s, output - %s", name, arg, out)
		return err
	}

	inspect, err := t.cli.ContainerExecInspect(ctx, execId.ID)
	if err != nil {
		log.S(ctx).Errorf("failed to inspect command - %s", err)
		return err
	}

	if inspect.ExitCode != 0 {
		return errors.Errorf("failed to execute command %s %s, exit code %d, output: %s", name, arg, inspect.ExitCode, out)
	} else {
		log.S(ctx).Infof("finished command - %s %s, output - %s", name, arg, out)
		return nil
	}

}
