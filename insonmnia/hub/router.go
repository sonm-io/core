// TODO: I think this is the first candidate for package decomposition. Not going to do this right now to avoid unreviewable diff.

package hub

import (
	"errors"
	"syscall"

	"github.com/sonm-io/core/insonmnia/gateway"
)

type Route struct {
	ID          string
	Protocol    string
	Host        string
	Port        uint16
	BackendHost string
	BackendPort uint16
}

type Protocol struct {
	ty uint16
}

func NewProtocol(ty string) (Protocol, error) {
	switch ty {
	case "tcp":
		return TCPProtocol(), nil
	case "udp":
		return UDPProtocol(), nil
	default:
		return Protocol{}, errors.New("unknown IPVS underlying protocol")
	}
}

func TCPProtocol() Protocol {
	return Protocol{ty: syscall.IPPROTO_TCP}
}

func UDPProtocol() Protocol {
	return Protocol{ty: syscall.IPPROTO_UDP}
}

func (p Protocol) String() string {
	switch p.ty {
	case syscall.IPPROTO_TCP:
		return "tcp"
	case syscall.IPPROTO_UDP:
		return "udp"
	default:
		return "tcp"
	}
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
	Register(ID string, protocol Protocol) (VirtualService, error)
	// Deregister deregisters a virtual service specified by the given ID.
	Deregister(ID string) error
	// GetMetrics returns gateway-specific metrics.
	GetMetrics() (*gateway.Metrics, error)
	// Close closes the router, freeing all associated resources.
	Close() error
}

type directRouter struct {
}

func newDirectRouter() Router {
	return &directRouter{}
}

func (r *directRouter) Register(ID string, protocol Protocol) (VirtualService, error) {
	return &directVirtualService{id: ID, protocol: protocol}, nil
}

func (r *directRouter) Deregister(ID string) error {
	return nil
}

func (r *directRouter) GetMetrics() (*gateway.Metrics, error) {
	return &gateway.Metrics{}, nil
}

func (r *directRouter) Close() error {
	return nil
}

type directVirtualService struct {
	id       string
	protocol Protocol
}

func (s *directVirtualService) ID() string {
	return s.id
}

func (s *directVirtualService) AddReal(ID string, host string, port uint16) (*Route, error) {
	route := &Route{
		ID:          ID,
		Protocol:    s.protocol.String(),
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
