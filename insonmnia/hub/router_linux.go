package hub

import (
	"context"
	"github.com/sonm-io/core/insonmnia/gateway"
)

type ipvsRouter struct {
	gateway *gateway.Gateway
	pool    *gateway.PortPool
}

func newIPVSRouter(ctx context.Context, gate *gateway.Gateway) router {
	// TODO (3Hren): Make configurable.
	return &ipvsRouter{
		gateway: gate,
		pool:    gateway.NewPortPool(32768, 1024),
	}
}

func (r *ipvsRouter) RegisterRoute(ID string, protocol string, realIP string, realPort uint16) error {
	host, err := gateway.GetOutboundIP()
	if err != nil {
		return err
	}

	port, err := r.pool.Assign(ID)
	if err != nil {
		return err
	}

	serviceOptions, err := gateway.NewServiceOptions(host.String(), port, protocol)
	if err != nil {
		return err
	}

	if err := r.gateway.CreateService(ID, serviceOptions); err != nil {
		return err
	}

	realOptions, err := gateway.NewRealOptions(realIP, realPort, 100, ID)
	if err != nil {
		return err
	}

	if err := r.gateway.CreateBackend(ID, ID, realOptions); err != nil {
		return err
	}

	return nil
}
