package hub

import (
	"context"
	"sync"

	"github.com/sonm-io/core/insonmnia/gateway"
)

type ipvsRouter struct {
	gateway  *gateway.Gateway
	pool     *gateway.PortPool
	metrics  map[string]*gateway.Metrics
	services map[string]VirtualService
	mu       sync.Mutex
}

func newIPVSRouter(ctx context.Context, gate *gateway.Gateway, pool *gateway.PortPool) Router {
	return &ipvsRouter{
		gateway:  gate,
		pool:     pool,
		metrics:  make(map[string]*gateway.Metrics, 0),
		services: make(map[string]VirtualService, 0),
	}
}

func (r *ipvsRouter) Register(ID string, protocol Protocol) (VirtualService, error) {
	host, err := gateway.GetOutboundIP()
	if err != nil {
		return nil, err
	}

	port, err := r.pool.Assign(ID)
	if err != nil {
		return nil, err
	}

	serviceOptions, err := gateway.NewServiceOptions(host.String(), port, protocol.String())
	if err != nil {
		return nil, err
	}

	if err := r.gateway.CreateService(ID, serviceOptions); err != nil {
		return nil, err
	}

	virtualService := &ipvsVirtualService{
		vsID:    ID,
		options: serviceOptions,
		gateway: r.gateway,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics[ID] = &gateway.Metrics{}
	r.services[ID] = virtualService

	return virtualService, nil
}

func (r *ipvsRouter) Deregister(ID string) error {
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

	delete(r.services, ID)

	return nil
}

// GetMetrics collects network specific metrics that are associated with this router.
func (r *ipvsRouter) GetMetrics() (*gateway.Metrics, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.getMetrics()
}

func (r *ipvsRouter) getMetrics() (*gateway.Metrics, error) {
	for id := range r.services {
		if err := r.updateServiceMetrics(id); err != nil {
			return nil, err
		}
	}

	metrics := &gateway.Metrics{}
	for _, current := range r.metrics {
		metrics.Add(current)
	}

	return metrics, nil
}

func (r *ipvsRouter) updateServiceMetrics(ID string) error {
	metrics, err := r.gateway.GetMetrics(ID)
	if err != nil {
		return err
	}

	r.metrics[ID] = metrics
	return nil
}

// Close deregisters all routes.
func (r *ipvsRouter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for ID := range r.services {
		r.deregisterRoute(ID)
	}

	return nil
}

type ipvsVirtualService struct {
	vsID    string
	options *gateway.ServiceOptions
	gateway *gateway.Gateway
}

func (s *ipvsVirtualService) ID() string {
	return s.vsID
}

func (s *ipvsVirtualService) AddReal(ID string, host string, port uint16) (*route, error) {
	realOptions, err := gateway.NewRealOptions(host, port, 100, s.vsID)
	if err != nil {
		return nil, err
	}

	if err := s.gateway.CreateBackend(s.vsID, ID, realOptions); err != nil {
		return nil, err
	}

	route := &route{
		ID:          ID,
		Protocol:    s.options.Protocol,
		Host:        s.options.Host,
		Port:        s.options.Port,
		BackendHost: host,
		BackendPort: port,
	}

	return route, nil
}

func (s *ipvsVirtualService) RemoveReal(ID string) error {
	_, err := s.gateway.RemoveBackend(s.vsID, ID)
	return err
}
