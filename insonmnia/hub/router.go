package hub

import "github.com/sonm-io/core/insonmnia/gateway"

type route struct {
	ID          string
	Protocol    string
	Host        string
	Port        uint16
	BackendHost string
	BackendPort uint16
}

type router interface {
	RegisterRoute(ID string, protocol string, realIP string, realPort uint16) (*route, error)
	DeregisterRoute(ID string) error
	GetMetrics() (*gateway.Metrics, error)
	Close() error
}

type directRouter struct {
}

func newDirectRouter() router {
	return &directRouter{}
}

func (r *directRouter) RegisterRoute(ID string, protocol string, realIP string, realPort uint16) (*route, error) {
	route := &route{
		ID:          ID,
		Protocol:    protocol,
		Host:        realIP,
		Port:        realPort,
		BackendHost: realIP,
		BackendPort: realPort,
	}

	return route, nil
}

func (r *directRouter) DeregisterRoute(ID string) error {
	return nil
}

func (r *directRouter) GetMetrics() (*gateway.Metrics, error) {
	return &gateway.Metrics{}, nil
}

func (r *directRouter) Close() error {
	return nil
}
