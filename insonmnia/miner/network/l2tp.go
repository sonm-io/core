// +build linux

package network

import (
	"context"

	"github.com/docker/go-plugins-helpers/network"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	pppOptsDir = "/etc/ppp/"
)

type L2TPDriver struct {
	ctx   context.Context
	store *networkInfoStore
}

func NewL2TPDriver(ctx context.Context, store *networkInfoStore) *L2TPDriver {
	return &L2TPDriver{
		ctx:   ctx,
		store: store,
	}
}

func (d *L2TPDriver) GetCapabilities() (*network.CapabilitiesResponse, error) {
	log.G(d.ctx).Info("received GetCapabilities request")

	return &network.CapabilitiesResponse{
		Scope:             "local",
		ConnectivityScope: "local",
	}, nil
}

func (d *L2TPDriver) CreateNetwork(request *network.CreateNetworkRequest) error {
	log.G(d.ctx).Info("received CreateNetwork request", zap.Any("request", request))

	opts, err := parseOptsNetwork(request)
	if err != nil {
		log.G(d.ctx).Error("failed to parse options", zap.Error(err))
		return errors.Wrap(err, "failed to parse options")
	}

	netInfo, err := d.store.GetNetwork(opts.GetHash())
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.IPv4Data[0].Pool),
			zap.Error(err))
		return errors.Wrap(err, "failed to get network")
	}

	netInfo.ID = request.NetworkID
	d.store.AddNetworkAlias(netInfo.networkOpts.GetHash(), request.NetworkID)

	log.G(d.ctx).Info("successfully registered network", zap.String("pool_id", netInfo.PoolID),
		zap.String("network_id", netInfo.ID))

	return nil
}

func (d *L2TPDriver) CreateEndpoint(request *network.CreateEndpointRequest) (*network.CreateEndpointResponse, error) {
	log.G(d.ctx).Info("received CreateEndpoint request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.NetworkID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	addr, _ := getAddrFromCIDR(request.Interface.Address)
	eptInfo, err := netInfo.store.GetEndpoint(addr)
	if err != nil {
		log.G(d.ctx).Error("failed to get endpoint", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get endpoint")
	}

	eptInfo.ID = request.EndpointID
	netInfo.store.AddEndpointAlias(addr, request.EndpointID)

	return nil, nil
}

func (d *L2TPDriver) Join(request *network.JoinRequest) (*network.JoinResponse, error) {
	log.G(d.ctx).Info("received Join request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.NetworkID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	eptInfo, err := netInfo.store.GetEndpoint(request.EndpointID)
	if err != nil {
		log.G(d.ctx).Error("failed to get endpoint", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get endpoint")
	}

	log.G(d.ctx).Info("joined endpoint", zap.String("pool_id", netInfo.PoolID),
		zap.String("network_id", netInfo.ID), zap.String("endpoint_id", request.EndpointID),
		zap.String("ip", eptInfo.AssignedIP))

	return &network.JoinResponse{
		InterfaceName: network.InterfaceName{SrcName: eptInfo.PPPDevName, DstPrefix: "ppp"},
		StaticRoutes: []*network.StaticRoute{
			{Destination: netInfo.networkOpts.Subnet, RouteType: 1},
		},
	}, nil
}

func (d *L2TPDriver) Leave(request *network.LeaveRequest) error {
	log.G(d.ctx).Info("received Leave request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) EndpointInfo(request *network.InfoRequest) (*network.InfoResponse, error) {
	log.G(d.ctx).Info("received EndpointInfo request", zap.Any("request", request))

	netInfo, err := d.store.GetNetwork(request.NetworkID)
	if err != nil {
		log.G(d.ctx).Error("failed to get network", zap.String("pool_id", request.NetworkID),
			zap.Error(err))
		return nil, errors.Wrap(err, "failed to get network")
	}

	eptInfo, err := netInfo.store.GetEndpoint(request.EndpointID)
	if err != nil {
		log.G(d.ctx).Error("failed to get endpoint", zap.String("pool_id", netInfo.PoolID),
			zap.String("network_id", netInfo.ID), zap.Error(err))
		return nil, errors.Wrap(err, "failed to get endpoint")
	}

	return &network.InfoResponse{
		Value: map[string]string{
			"ip":              eptInfo.AssignedIP,
			"l2tpd_conn_name": eptInfo.ConnName,
			"device_name":     eptInfo.PPPDevName,
		},
	}, nil
}

func (d *L2TPDriver) AllocateNetwork(request *network.AllocateNetworkRequest) (*network.AllocateNetworkResponse, error) {
	log.G(d.ctx).Info("received AllocateNetwork request", zap.Any("request", request))
	return nil, nil
}

func (d *L2TPDriver) DeleteNetwork(request *network.DeleteNetworkRequest) error {
	log.G(d.ctx).Info("received DeleteNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) FreeNetwork(request *network.FreeNetworkRequest) error {
	log.G(d.ctx).Info("received FreeNetwork request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) DeleteEndpoint(request *network.DeleteEndpointRequest) error {
	log.G(d.ctx).Info("received DeleteEndpoint request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) DiscoverNew(request *network.DiscoveryNotification) error {
	log.G(d.ctx).Info("received DiscoverNew request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) DiscoverDelete(request *network.DiscoveryNotification) error {
	log.G(d.ctx).Info("received DiscoverDelete request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) ProgramExternalConnectivity(request *network.ProgramExternalConnectivityRequest) error {
	log.G(d.ctx).Info("received ProgramExternalConnectivity request", zap.Any("request", request))
	return nil
}

func (d *L2TPDriver) RevokeExternalConnectivity(request *network.RevokeExternalConnectivityRequest) error {
	log.G(d.ctx).Info("received RevokeExternalConnectivity request", zap.Any("request", request))
	return nil
}
