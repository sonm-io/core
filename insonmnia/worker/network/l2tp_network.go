package network

import (
	"context"
	"fmt"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type L2TPNetworkDriver struct {
	*l2tpState
	logger *zap.SugaredLogger
}

func NewL2TPDriver(ctx context.Context, state *l2tpState) *L2TPNetworkDriver {
	return &L2TPNetworkDriver{
		l2tpState: state,
		logger:    log.S(ctx).With("source", "l2tp/network"),
	}
}

func (d *L2TPNetworkDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	d.logger.Infow("received GetCapabilities request")
	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (d *L2TPNetworkDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	d.logger.Infow("received CreateNetwork request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	opts, err := parseOptsNetwork(request)
	if err != nil {
		d.logger.Errorw("failed to parse options", zap.Error(err))
		return fmt.Errorf("failed to parse options: %v", err)
	}

	n, err := d.GetNetwork(opts.PoolID())
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.IPv4Data[0].Pool),
			zap.Error(err))
		return fmt.Errorf("failed to get network: %v", err)
	}

	n.ID = request.NetworkID
	d.AddNetworkAlias(n.NetworkOpts.PoolID(), request.NetworkID)

	d.logger.Infow("successfully registered network", zap.String("pool_id", n.PoolID),
		zap.String("network_id", n.ID))

	return nil
}

func (d *L2TPNetworkDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	d.logger.Infow("received CreateEndpoint request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	n, err := d.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	if n.Endpoint == nil {
		d.logger.Errorw("network's endpoint is nil", zap.String("network_id", n.ID))
		return nil, fmt.Errorf("network %s endpoint is nil", n.ID)
	}

	if len(n.Endpoint.ID) > 0 {
		d.logger.Errorw("asked to create second endpoint for network, aborting",
			zap.String("network_id", n.ID))
		return nil, fmt.Errorf("asked to create second endpoint for network %s, aborting", n.ID)
	}

	n.Endpoint.ID = request.EndpointID

	return nil, nil
}

func (d *L2TPNetworkDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	d.logger.Infow("received Join request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	n, err := d.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	d.logger.Infow("joined endpoint", zap.String("pool_id", n.PoolID),
		zap.String("network_id", n.ID), zap.String("endpoint_id", request.EndpointID),
		zap.String("ip", n.Endpoint.AssignedIP))

	return &network.JoinResponse{
		InterfaceName: network.InterfaceName{SrcName: n.Endpoint.PPPDevName, DstPrefix: "ppp"},
		StaticRoutes: []*network.StaticRoute{
			{Destination: n.NetworkOpts.Subnet, RouteType: 1},
		},
	}, nil
}

func (d *L2TPNetworkDriver) Leave(request *network.LeaveRequest) error {
	d.logger.Infow("received Leave request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	d.logger.Infow("received EndpointInfo request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.sync()

	n, err := d.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get network: %v", err)
	}

	return &network.InfoResponse{
		Value: map[string]string{
			"ip":              n.Endpoint.AssignedIP,
			"l2tpd_conn_name": n.Endpoint.ConnName,
			"device_name":     n.Endpoint.PPPDevName,
		},
	}, nil
}

func (d *L2TPNetworkDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	d.logger.Infow("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (d *L2TPNetworkDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	d.logger.Infow("received DeleteNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	d.logger.Infow("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	d.logger.Infow("received DeleteEndpoint request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	d.logger.Infow("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	d.logger.Infow("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	d.logger.Infow("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (d *L2TPNetworkDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	d.logger.Infow("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}
