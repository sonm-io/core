package hub

import (
	"context"
	"github.com/sonm-io/core/insonmnia/gateway"
)

type ipvsRouter struct {
	gateway *gateway.Gateway
	pool    *gateway.PortPool
}

func newIPVSRouter(ctx context.Context, gate *gateway.Gateway, pool *gateway.PortPool) router {
	return &ipvsRouter{
		gateway: gate,
		pool:    pool,
	}
}

func (r *ipvsRouter) RegisterRoute(ID string, protocol string, realIP string, realPort uint16) (*route, error) {
	host, err := gateway.GetOutboundIP()
	if err != nil {
		return nil, err
	}

	port, err := r.pool.Assign(ID)
	if err != nil {
		return nil, err
	}

	serviceOptions, err := gateway.NewServiceOptions(host.String(), port, protocol)
	if err != nil {
		return nil, err
	}

	if err := r.gateway.CreateService(ID, serviceOptions); err != nil {
		return nil, err
	}

	realOptions, err := gateway.NewRealOptions(realIP, realPort, 100, ID)
	if err != nil {
		return nil, err
	}

	if err := r.gateway.CreateBackend(ID, ID, realOptions); err != nil {
		return nil, err
	}

	route := &route{
		ID:          ID,
		Protocol:    protocol,
		Host:        host.String(),
		Port:        port,
		BackendHost: realIP,
		BackendPort: realPort,
	}

	return route, nil
}

func (r *ipvsRouter) DeregisterRoute(ID string) error {
	if _, err := r.gateway.RemoveService(ID); err != nil {
		return err
	}

	if err := r.pool.Retain(ID); err != nil {
		return err
	}

	return nil
}

func (r *ipvsRouter) Close() error {
	return nil
}
