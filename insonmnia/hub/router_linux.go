package hub

import (
	"context"
	"sync"

	"github.com/sonm-io/core/insonmnia/gateway"
)

type ipvsRouter struct {
	gateway *gateway.Gateway
	pool    *gateway.PortPool
	ids     map[string]bool
	mu      sync.Mutex
}

func newIPVSRouter(ctx context.Context, gate *gateway.Gateway, pool *gateway.PortPool) router {
	return &ipvsRouter{
		gateway: gate,
		pool:    pool,
		ids:     make(map[string]bool, 0),
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

	r.mu.Lock()
	defer r.mu.Unlock()
	r.ids[ID] = true

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
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.deregisterRoute(ID)
}

func (r *ipvsRouter) deregisterRoute(ID string) error {
	if _, err := r.gateway.RemoveService(ID); err != nil {
		return err
	}

	if err := r.pool.Retain(ID); err != nil {
		return err
	}

	delete(r.ids, ID)

	return nil
}

// Close deregisters all routes.
func (r *ipvsRouter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for ID := range r.ids {
		r.deregisterRoute(ID)
	}

	return nil
}
