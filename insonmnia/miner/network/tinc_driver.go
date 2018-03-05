package network

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	cnet "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-plugins-helpers/ipam"
	"github.com/docker/go-plugins-helpers/network"
	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/docker/libnetwork"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	defaultNetwork = "10.20.30.0/24"
)

func makeStore(ctx context.Context, path string) (store.Store, error) {
	boltdb.Register()
	s := store.Backend(store.BOLTDB)
	endpoints := []string{path}
	config := store.Config{
		Bucket: "sonm_tinc_driver_state",
	}
	return libkv.NewStore(s, endpoints, &config)
}

func NewTinc(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkDriver, *TincIPAMDriver, error) {
	err := os.MkdirAll(config.ConfigDir, 0770)
	if err != nil {
		return nil, nil, err
	}
	storage, err := makeStore(ctx, config.StatePath)

	state := &TincNetworkState{
		ctx:             ctx,
		config:          config,
		cli:             client,
		Networks:        map[string]*TincNetwork{},
		networkNameToId: map[string]string{},
		Pools:           map[string]*net.IPNet{},
		storage:         storage,
		logger:          log.S(ctx).With("source", "tinc/state"),
	}
	err = state.load()
	if err != nil {
		return nil, nil, err
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
	Name            string
	ID              string
	Options         *TincNetworkOptions
	IPv4Data        []*network.IPAMData
	IPv6Data        []*network.IPAMData
	ConfigPath      string
	cli             *client.Client
	TincContainerID string
	logger          *zap.SugaredLogger
}

type Pool struct {
	Net *net.IPNet
	libnetwork.IpamConf
}

type TincNetworkState struct {
	ctx             context.Context
	config          *TincNetworkConfig
	mu              sync.RWMutex
	cli             *client.Client
	Networks        map[string]*TincNetwork
	networkNameToId map[string]string
	Pools           map[string]*net.IPNet
	logger          *zap.SugaredLogger
	storage         store.Store
}

func (t *TincNetworkState) RegisterNetworkMapping(id string, name string) error {
	if len(name) == 0 || len(id) == 0 {
		return errors.Errorf("invalid network mapping arguments: \"%s\" \"%s\"", id, name)
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.Networks[id]
	if !ok {
		return errors.Errorf("no network with id %s", id)
	}
	t.networkNameToId[name] = id
	return nil
}

func (t *TincNetworkState) load() (err error) {
	defer func() {
		if err == store.ErrKeyNotFound {
			err = nil
		}
		if err != nil {
			t.logger.Errorf("could not load tinc network state - %s; erasing key", err)
			delErr := t.storage.Delete("state")
			if delErr != nil {
				t.logger.Errorf("could not cleanup storage for tinc network - %s", delErr)
			}
		}
	}()

	exists, err := t.storage.Exists("state")
	if err != nil || !exists {
		return
	}

	data, err := t.storage.Get("state")
	if err != nil {
		return
	}

	err = json.Unmarshal(data.Value, t)
	if err != nil {
		return
	}
	for _, n := range t.Networks {
		n.cli = t.cli
		n.logger = t.logger.With("source", "tinc/network/"+n.ID, "container", n.TincContainerID)
		t.RegisterNetworkMapping(n.ID, n.Name)
	}
	return
}

func (t *TincNetworkState) sync() error {
	var err error
	defer func() {
		if err != nil {
			t.logger.Errorf("could not sync network state - %s", err)
		}
	}()

	marshalled, err := json.Marshal(t)
	if err != nil {
		return err
	}
	err = t.storage.Put("state", marshalled, &store.WriteOptions{})
	return err
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
	t.Pools[id] = n
	t.sync()
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
	delete(t.Pools, request.PoolID)
	t.sync()
	return nil
}

func (t *TincIPAMDriver) RequestAddress(request *ipam.RequestAddressRequest) (*ipam.RequestAddressResponse, error) {
	t.logger.Info("received RequestAddress request", zap.Any("request", request))
	t.mu.Lock()
	defer t.mu.Unlock()

	pool, ok := t.Pools[request.PoolID]
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
	options := &TincNetworkOptions{}
	g, ok := data["com.docker.network.generic"]
	if !ok {
		return options, nil
		//return nil, errors.New("no options passed - invitation is required")
	}
	generic, ok := g.(map[string]interface{})
	if !ok {
		//return nil, errors.New("invalid type of generic options")
		return options, nil
	}
	invitation, ok := generic["invitation"]
	if ok {
		options.Invitation = invitation.(string)
		//return nil, errors.New("invitation is required")
	}
	_, enableBridge := generic["enable_bridge"]
	options.EnableBridge = enableBridge

	cgroupParent, ok := generic["cgroup_parent"]
	if ok {
		options.CgroupParent = cgroupParent.(string)
	}

	return options, nil
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
	n, err := t.newTincNetwork(request)
	if err != nil {
		return err
	}

	if len(n.Options.Invitation) != 0 {
		err = n.Join(t.ctx)
	} else {
		err = n.Init(t.ctx)
	}
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.Networks[n.ID] = n
	t.sync()

	return nil
}

func (t *TincNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(t.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (t *TincNetworkDriver) popNetwork(ID string) *TincNetwork {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, ok := t.Networks[ID]
	if !ok {
		return nil
	}
	delete(t.Networks, ID)
	delete(t.networkNameToId, n.Name)
	t.sync()
	return n
}

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	n := t.popNetwork(request.NetworkID)
	if n == nil {
		return errors.Errorf("no network with id %s", request.NetworkID)
	}
	return n.Shutdown()
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	t.logger.Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	t.logger.Info("received CreateEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()
	n, ok := t.Networks[request.NetworkID]
	if !ok {
		t.logger.Warn("no such network %s", request.NetworkID)
		return nil, errors.Errorf("no such network %s", request.NetworkID)
	}
	selfAddr := strings.Split(request.Interface.Address, "/")[0]
	err := n.Start(t.ctx, selfAddr)
	if err != nil {
		return nil, err
	}
	return &network.CreateEndpointResponse{}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	t.logger.Info("received DeleteEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()
	time.Timer{}

	n, ok := t.Networks[request.NetworkID]
	if !ok {
		return errors.Errorf("no such network %s", request.NetworkID)
	}
	return n.Stop(t.ctx)
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

func (t *TincNetworkDriver) HasNetwork(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	id, ok := t.networkNameToId[name]
	if !ok {
		return ok
	}
	_, ok = t.Networks[id]
	return ok
}

func (t *TincNetworkDriver) GenerateInvitation(name string) (structs.Network, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	id, ok := t.networkNameToId[name]
	if !ok {
		return nil, errors.Errorf("no such network %s", name)
	}
	n := t.Networks[id]

	inviteeID := strings.Replace(uuid.New(), "-", "_", -1)
	invitation, err := n.Invite(t.ctx, inviteeID)
	spec := structs.NetworkSpec{
		NetworkSpec: &sonm.NetworkSpec{
			Type:    "tinc",
			Options: map[string]string{"invitation": invitation},
		},
	}
	return &spec, err
}

func (t *TincNetwork) Init(ctx context.Context) error {
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.ID, "-c", t.ConfigPath, "init", "initial_node_"+t.ID)
	if err != nil {
		t.logger.Errorf("failed to init network - %s", err)
	} else {
		t.logger.Info("succesefully initialized tinc network")
	}
	return err
}

func (t *TincNetwork) Join(ctx context.Context) error {
	if len(t.Options.Invitation) == 0 {
		return errors.New("can not join to network without invitation")
	}
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

func (t *TincNetwork) Invite(ctx context.Context, inviteeID string) (string, error) {
	out, _, err := t.runCommandWithOutput(ctx, "tinc", "--batch", "-n", t.ID, "-c", t.ConfigPath, "invite", inviteeID)
	out = strings.Trim(out, "\n")
	return out, err
}

func (t *TincNetwork) runCommand(ctx context.Context, name string, arg ...string) error {
	_, _, err := t.runCommandWithOutput(ctx, name, arg...)
	return err
}
func (t *TincNetwork) runCommandWithOutput(ctx context.Context, name string, arg ...string) (string, string, error) {
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
		return "", "", err
	}

	conn, err := t.cli.ContainerExecAttach(ctx, execId.ID, cfg)
	if err != nil {
		t.logger.Warnf("ContainerExecAttach finished with error - %s", err)
	}
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	stdcopy.StdCopy(&stdoutBuf, &stderrBuf, conn.Reader)
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if err != nil {
		t.logger.Warnf("failed to execute command - %s %s, stdout - %s, stderr - %s", name, arg, stdout, stderr)
		return stdout, stderr, err
	}

	inspect, err := t.cli.ContainerExecInspect(ctx, execId.ID)
	if err != nil {
		t.logger.Warnf("failed to inspect command - %s", err)
		return stdout, stderr, err
	}

	if inspect.ExitCode != 0 {
		return stdout, stderr, errors.Errorf("failed to execute command %s %s, exit code %d, stdout - %s, stderr - %s", name, arg, inspect.ExitCode, stdout, stderr)
	} else {
		t.logger.Debugf("finished command - %s %s, stdout - %s, stderr - %s", name, arg, stdout, stderr)
		return stdout, stderr, err
	}
}
