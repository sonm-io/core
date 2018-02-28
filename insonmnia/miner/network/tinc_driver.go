package network

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"

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

func NewTinc(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkDriver, *TincIPAMDriver, error) {
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

	netDr := &TincNetworkDriver{
		TincNetworkState: state,
		logger:           log.S(ctx).With("source", "tinc/network"),
	}

	ipamDr := &TincIPAMDriver{
		TincNetworkState: state,
		logger:           log.S(ctx).With("source", "tinc/ipam"),
	}
	return netDr, ipamDr, nil
}

type TincNetworkOptions struct {
	Invitation   string
	EnableBridge bool
	CgroupParent string
}

type TincNetwork struct {
	ID              string
	Options         *TincNetworkOptions
	IPv4Data        []*network.IPAMData
	IPv6Data        []*network.IPAMData
	ConfigPath      string
	cli             *client.Client
	TincContainerID string
	logger          *zap.SugaredLogger
}

type TincNetworkState struct {
	ctx      context.Context
	config   *TincNetworkConfig
	mu       sync.RWMutex
	cli      *client.Client
	networks map[string]*TincNetwork
	pools    map[string]*net.IPNet
	logger   *zap.SugaredLogger
}

type TincNetworkDriver struct {
	*TincNetworkState
	logger *zap.SugaredLogger
}

type TincIPAMDriver struct {
	*TincNetworkState
	logger *zap.SugaredLogger
}

func (t *TincIPAMDriver) GetCapabilities() (*ipam.CapabilitiesResponse, error) {
	t.logger.Info("received GetCapabilities request")
	return &ipam.CapabilitiesResponse{RequiresMACAddress: false}, nil
}

func (t *TincIPAMDriver) GetDefaultAddressSpaces() (*ipam.AddressSpacesResponse, error) {
	t.logger.Info("received GetDefaultAddressSpaces request")
	return nil, nil
}

func (t *TincIPAMDriver) RequestPool(request *ipam.RequestPoolRequest) (*ipam.RequestPoolResponse, error) {
	t.logger.Info("received RequestPool request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()
	id := uuid.New()
	_, n, err := net.ParseCIDR(request.Pool)
	if err != nil {
		t.logger.Errorf("invalid pool CIDR specified - %s", err)
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
	t.logger.Info("received ReleasePool request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.pools, request.PoolID)
	return nil
}

func (t *TincIPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	t.logger.Info("received RequestAddress request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()

	pool, ok := t.pools[request.PoolID]
	if !ok {
		t.logger.Errorf("pool %s not found", request.PoolID)
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
		addr := ip.String() + "/" + fmt.Sprint(mask)
		t.logger.Infof("providing gateway address %s", addr)
		return &ipam.RequestAddressResponse{
			Address: addr,
		}, nil
	}

	if mask == 0 {
		t.logger.Errorf("invalid subnet specified for pool %s", pool.String())
		return nil, errors.New("invalid subnet")
	}
	return &ipam.RequestAddressResponse{
		Address: request.Address + "/" + fmt.Sprint(mask),
	}, nil
}

func (t *TincIPAMDriver) ReleaseAddress(request *ipam.ReleaseAddressRequest) error {
	t.logger.Info("received ReleaseAddress request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) newTincNetwork(request *network.CreateNetworkRequest) (*TincNetwork, error) {
	opts, err := ParseNetworkOpts(request.Options)
	if err != nil {
		t.logger.Errorf("failed to parse network options - %s", err)
		return nil, err
	}
	containerConfig := &container.Config{
		Image: "antmat/tinc",
	}
	hostConfig := &container.HostConfig{
		Privileged:  true,
		NetworkMode: "host",
		Resources: container.Resources{
			CgroupParent: opts.CgroupParent,
		},
	}
	netConfig := &cnet.NetworkingConfig{}
	resp, err := t.cli.ContainerCreate(t.ctx, containerConfig, hostConfig, netConfig, "")
	if err != nil {
		t.logger.Errorf("failed to create tinc container - %s", err)
		return nil, err
	}
	err = t.cli.ContainerStart(t.ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		t.logger.Errorf("failed to start tinc container - %s", err)
		return nil, err
	}
	log.S(t.ctx).Infof("started container %s", resp.ID)

	return &TincNetwork{
		ID:              request.NetworkID,
		Options:         opts,
		IPv4Data:        request.IPv4Data,
		IPv6Data:        request.IPv6Data,
		ConfigPath:      t.config.ConfigDir + "/" + request.NetworkID,
		cli:             t.cli,
		TincContainerID: resp.ID,
		logger:          t.logger.With("source", "tinc/network/"+request.NetworkID, "container", resp.ID),
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
	_, enableBridge := generic["enable_bridge"]
	cgroupParent := ""

	cgroupParentI, ok := generic["cgroup_parent"]
	if ok {
		cgroupParent = cgroupParentI.(string)
	}

	return &TincNetworkOptions{Invitation: invitation.(string), EnableBridge: enableBridge, CgroupParent: cgroupParent}, nil
}

func (t *TincNetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	t.logger.Info("received GetCapabilities request")
	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (t *TincNetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	t.logger.Info("received CreateNetwork request", zap.Any("request", request))
	network, err := t.newTincNetwork(request)
	if err != nil {
		return err
	}

	err = network.Join(t.ctx)
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

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	net := t.popNetwork(request.NetworkID)
	if net == nil {
		return errors.Errorf("no network with id %s", request.NetworkID)
	}
	return net.Shutdown()
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	t.logger.Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	t.logger.Info("received CreateEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()
	net, ok := t.networks[request.NetworkID]
	if !ok {
		t.logger.Warn("no such network %s", request.NetworkID)
		return nil, errors.Errorf("no such network %s", request.NetworkID)
	}
	selfAddr := strings.Split(request.Interface.Address, "/")[0]
	err := net.Start(t.ctx, selfAddr)
	if err != nil {
		return nil, err
	}
	return &network.CreateEndpointResponse{}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	t.logger.Info("received DeleteEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()

	net, ok := t.networks[request.NetworkID]
	if !ok {
		return errors.Errorf("no such network %s", request.NetworkID)
	}
	return net.Stop(t.ctx)
	err := net.runCommand(t.ctx, "tinc", "--batch", "-n", net.ID, "-c", net.ConfigPath, "stop")
	if err != nil {
		return err
	}

	return nil
}

func (t *TincNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	t.logger.Info("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	return &network.InfoResponse{Value: val}, nil
}

func (t *TincNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	t.logger.Info("received Join request", zap.Any("request", request))
	t.mu.RLock()
	defer t.mu.RUnlock()
	iface := request.NetworkID[:15]
	return &network.JoinResponse{DisableGatewayService: false, InterfaceName: network.InterfaceName{SrcName: iface, DstPrefix: "tinc"}}, nil
}

func (t *TincNetworkDriver) Leave(request *network.LeaveRequest) error {
	t.logger.Info("received Leave request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	t.logger.Info("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	t.logger.Info("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	t.logger.Info("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	t.logger.Info("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetwork) Join(ctx context.Context) error {
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.ID, "-c", t.ConfigPath, "join", t.Options.Invitation)
	if err != nil {
		t.logger.Errorf("failed to join network - %s", err)
	} else {
		t.logger.Info("succesefully joined tinc network")
	}
	return err
}

func (t *TincNetwork) Start(ctx context.Context, addr string) error {
	iface := t.ID[:15]
	//TODO: each pool should be considered
	pool := t.IPv4Data[0].Pool

	err := t.runCommand(ctx, "tinc", "-n", t.ID, "-c", t.ConfigPath, "start",
		"-o", "Interface="+iface, "-o", "Subnet="+pool, "-o", "Subnet="+addr+"/32", "-o", "LogLevel=0")
	if err != nil {
		t.logger.Error("failed to start tinc - %s", err)
	} else {
		t.logger.Info("started tinc")
	}
	return err
}

func (t *TincNetwork) Shutdown() error {
	return nil
}

func (t *TincNetwork) Stop(ctx context.Context) error {
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.ID, "-c", t.ConfigPath, "stop")
	if err != nil {
		t.logger.Errorf("failed to stop tinc - %s", err)
		return err
	} else {
		t.logger.Info("successfully stoppped tinc")
	}
	return err
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
		t.logger.Warnf("ContainerExecCreate finished with error - %s", err)
		return err
	}

	conn, err := t.cli.ContainerExecAttach(ctx, execId.ID, cfg)
	if err != nil {
		t.logger.Warnf("ContainerExecAttach finished with error - %s", err)
	}

	byteOut, err := ioutil.ReadAll(conn.Reader)
	out := string(byteOut)
	if err != nil {
		t.logger.Warnf("failed to execute command - %s %s, output - %s", name, arg, out)
		return err
	}

	inspect, err := t.cli.ContainerExecInspect(ctx, execId.ID)
	if err != nil {
		t.logger.Warnf("failed to inspect command - %s", err)
		return err
	}

	if inspect.ExitCode != 0 {
		return errors.Errorf("failed to execute command %s %s, exit code %d, output: %s", name, arg, inspect.ExitCode, out)
	} else {
		t.logger.Debugf("finished command - %s %s, output - %s", name, arg, out)
		return nil
	}

}
