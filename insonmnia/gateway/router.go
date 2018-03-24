// TODO: I think this is the first candidate for package decomposition. Not going to do this right now to avoid unreviewable diff.

package gateway

type Route struct {
	ID          string
	Protocol    string
	Host        string
	Port        uint16
	BackendHost string
	BackendPort uint16
}

// VirtualService describes a virtual service.
type VirtualService interface {
	// ID returns this virtual service's ID.
	ID() string
	// AddReal registers a new real service endpoint under the current virtual
	// service.
	// Host parameter may be both FQDN and IP address.
	AddReal(ID string, host string, port uint16) (*Route, error)
	// RemoveReal removes the real service specified by the ID from the
	// current virtual service.
	// Established connections won't be affected.
	RemoveReal(ID string) error
}

type Router interface {
	// Register registers a new virtual service specified by the given ID.
	Register(ID string, protocol string) (VirtualService, error)
	// Deregister deregisters a virtual service specified by the given ID.
	Deregister(ID string) error
	// GetMetrics returns gateway-specific metrics.
	GetMetrics() (*Metrics, error)
	// Close closes the router, freeing all associated resources.
	Close() error
}

type directRouter struct {
}

func newDirectRouter() Router {
	return &directRouter{}
}

func (r *directRouter) Register(ID string, protocol string) (VirtualService, error) {
	return &directVirtualService{id: ID, protocol: protocol}, nil
}

func (r *directRouter) Deregister(ID string) error {
	return nil
}

func (r *directRouter) GetMetrics() (*Metrics, error) {
	return &Metrics{}, nil
}

func (r *directRouter) Close() error {
	return nil
}

type directVirtualService struct {
	id       string
	protocol string
}

func (s *directVirtualService) ID() string {
	return s.id
}

func (s *directVirtualService) AddReal(ID string, host string, port uint16) (*Route, error) {
	route := &Route{
		ID:          ID,
		Protocol:    s.protocol,
		Host:        host,
		Port:        port,
		BackendHost: host,
		BackendPort: port,
	}

	return route, nil
}

func (s *directVirtualService) RemoveReal(ID string) error {
	return nil
}
