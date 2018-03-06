package network

import (
	"context"
	"net"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	cnet "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/network"
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

func NewTinc(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkDriver, *TincIPAMDriver, error) {
	err := os.MkdirAll(config.ConfigDir, 0770)
	if err != nil {
		return nil, nil, err
	}

	state, err := newTincNetworkState(ctx, client, config)
	if err != nil {
		return nil, nil, err
	}

	netDr := &TincNetworkDriver{
		TincNetworkState: state,
		logger:           log.S(ctx).With("source", "tinc/network"),
	}

	ipamDr, err := NewTincIPAMDriver(ctx, state, config)
	if err != nil {
		return nil, nil, err
	}
	//ipamDr := &TincIPAMDriver{
	//	TincNetworkState: state,
	//	logger:           log.S(ctx).With("source", "tinc/ipam"),
	//}
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
	//libnetwork.IpamConf
}

type TincNetworkDriver struct {
	*TincNetworkState
	logger *zap.SugaredLogger
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
