package network

import (
	"context"

	"sync"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type L2TPNeworktDriver struct {
	// NOTE: `mu` is shared with IPAMDriver driver.
	mu     *sync.Mutex
	state  *l2tpState
	logger *zap.SugaredLogger
}

func NewL2TPDriver(ctx context.Context, mu *sync.Mutex, state *l2tpState) *L2TPNeworktDriver {
	return &L2TPNeworktDriver{
		mu:     mu,
		state:  state,
		logger: log.S(ctx).With("source", "l2tp/network"),
	}
}

func (d *L2TPNeworktDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	d.logger.Infow("received GetCapabilities request")
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (d *L2TPNeworktDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	d.logger.Infow("received CreateNetwork request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	opts, err := parseOptsNetwork(request)
	if err != nil {
		d.logger.Errorw("failed to parse options", zap.Error(err))
		return errors.Wrap(err, "failed to parse options")
	}

	l2tpNet, err := d.state.GetNetwork(opts.GetPoolID())
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.IPv4Data[0].Pool),
			zap.Error(err))
		return errors.Wrap(err, "failed to get network")
	}

	l2tpNet.ID = request.NetworkID
	d.state.AddNetworkAlias(l2tpNet.NetworkOpts.GetPoolID(), request.NetworkID)

	d.logger.Infow("successfully registered network", zap.String("pool_id", l2tpNet.PoolID),
		zap.String("network_id", l2tpNet.ID))

	return nil
}

func (d *L2TPNeworktDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	d.logger.Infow("received CreateEndpoint request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	l2tpNet, err := d.state.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	if l2tpNet.Endpoint == nil {
		d.logger.Errorw("network's endpoint is nil", zap.String("network_id", l2tpNet.ID))
		return nil, errors.Errorf("network %s endpoint is nil", l2tpNet.ID)
	}

	if len(l2tpNet.Endpoint.ID) > 0 {
		d.logger.Errorw("asked to create second endpoint for network, aborting",
			zap.String("network_id", l2tpNet.ID))
		return nil, errors.Errorf("asked to create second endpoint for network %s, aborting", l2tpNet.ID)
	}

	l2tpNet.Endpoint.ID = request.EndpointID

	return nil, nil
}

func (d *L2TPNeworktDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	d.logger.Infow("received Join request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	l2tpNet, err := d.state.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	d.logger.Infow("joined endpoint", zap.String("pool_id", l2tpNet.PoolID),
		zap.String("network_id", l2tpNet.ID), zap.String("endpoint_id", request.EndpointID),
		zap.String("ip", l2tpNet.Endpoint.AssignedIP))

	return &network.JoinResponse{
		InterfaceName: network.InterfaceName{SrcName: l2tpNet.Endpoint.PPPDevName, DstPrefix: "ppp"},
		StaticRoutes: []*network.StaticRoute{
			{Destination: l2tpNet.NetworkOpts.Subnet, RouteType: 1},
		},
	}, nil
}

func (d *L2TPNeworktDriver) Leave(request *network.LeaveRequest) error {
	d.logger.Infow("received Leave request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *L2TPNeworktDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	d.logger.Infow("received EndpointInfo request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	l2tpNet, err := d.state.GetNetwork(request.NetworkID)
	if err != nil {
		d.logger.Errorw("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	return &network.InfoResponse{
		Value: map[string]string{
			"ip":              l2tpNet.Endpoint.AssignedIP,
			"l2tpd_conn_name": l2tpNet.Endpoint.ConnName,
			"device_name":     l2tpNet.Endpoint.PPPDevName,
		},
	}, nil
}

func (d *L2TPNeworktDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	d.logger.Infow("received AllocateNetwork request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil, nil
}

func (d *L2TPNeworktDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	d.logger.Infow("received DeleteNetwork request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *L2TPNeworktDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	d.logger.Infow("received FreeNetwork request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *L2TPNeworktDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	d.logger.Infow("received DeleteEndpoint request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *L2TPNeworktDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	d.logger.Infow("received DiscoverNew request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *L2TPNeworktDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	d.logger.Infow("received DiscoverDelete request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()

	return nil
}

func (d *L2TPNeworktDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	d.logger.Infow("received ProgramExternalConnectivity request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}

func (d *L2TPNeworktDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	d.logger.Infow("received RevokeExternalConnectivity request", zap.Any("request", request))
	d.mu.Lock()
	defer d.mu.Unlock()
	defer d.state.Sync()

	return nil
}
