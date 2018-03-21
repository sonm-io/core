package network

import (
	"context"
	"net"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

func NewTinc(ctx context.Context, client *client.Client, config *TincNetworkConfig) (*TincNetworkDriver, *TincIPAMDriver, error) {
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
	return netDr, ipamDr, nil
}

type TincNetwork struct {
	NodeID       string
	DockerID     string
	Pool         *net.IPNet
	Invitation   string
	EnableBridge bool
	CgroupParent string

	ConfigPath      string
	TincContainerID string

	cli    *client.Client
	logger *zap.SugaredLogger
}

type TincNetworkDriver struct {
	*TincNetworkState
	logger *zap.SugaredLogger
}

func (t *TincNetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	t.logger.Info("received GetCapabilities request")
	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (t *TincNetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	t.logger.Infow("received CreateNetwork request", zap.Any("request", request))
	n, err := t.netByOptions(request.Options)
	n.DockerID = request.NetworkID
	if err != nil {
		return err
	}

	if len(n.Invitation) != 0 {
		err = n.Join(t.ctx)
	} else {
		err = n.Init(t.ctx)
	}
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.sync()

	return nil
}

func (t *TincNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	t.logger.Infow("received AllocateNetwork request", zap.Any("request", request))
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
	t.sync()
	return n
}

func (t *TincNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	t.logger.Infow("received DeleteNetwork request", zap.Any("request", request))
	n, err := t.netByDockerID(request.NetworkID)
	if err != nil {
		return err
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.Networks, n.NodeID)
	return n.Shutdown(t.ctx)
}

func (t *TincNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	t.logger.Infow("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	t.logger.Infow("received CreateEndpoint request", zap.Any("request", request))

	n, err := t.netByOptions(request.Options)
	if err != nil {
		t.logger.Warnw("no such network", zap.Error(err))
		return nil, err
	}
	selfAddr := strings.Split(request.Interface.Address, "/")[0]
	err = n.Start(t.ctx, selfAddr)
	if err != nil {
		return nil, err
	}
	return &network.CreateEndpointResponse{}, nil
}

func (t *TincNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	t.logger.Infow("received DeleteEndpoint request", zap.Any("request", request))

	t.mu.RLock()
	defer t.mu.RUnlock()

	n, ok := t.Networks[request.NetworkID]
	if !ok {
		return errors.Errorf("no such network %s", request.NetworkID)
	}
	return n.Stop(t.ctx)
}

func (t *TincNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	t.logger.Infow("received EndpointInfo request", zap.Any("request", request))
	val := make(map[string]string)
	return &network.InfoResponse{Value: val}, nil
}

func (t *TincNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	t.logger.Infow("received Join request", zap.Any("request", request))
	n, err := t.netByDockerID(request.NetworkID)
	if err != nil {
		t.logger.Warnw("no such network", zap.Error(err))
		return nil, err
	}
	iface := n.NodeID[:15]
	return &network.JoinResponse{DisableGatewayService: false, InterfaceName: network.InterfaceName{SrcName: iface, DstPrefix: "tinc"}}, nil
}

func (t *TincNetworkDriver) Leave(request *network.LeaveRequest) error {
	t.logger.Infow("received Leave request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	t.logger.Infow("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	t.logger.Infow("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	t.logger.Infow("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	t.logger.Infow("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (t *TincNetworkDriver) HasNetwork(NodeID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.Networks[NodeID]
	return ok
}

func (t *TincNetworkDriver) GenerateInvitation(NodeID string) (structs.Network, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	n, ok := t.Networks[NodeID]
	if !ok {
		return nil, errors.Errorf("no such network %s", NodeID)
	}

	//TODO: Check this
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
