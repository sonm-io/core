package gateway

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

const (
	DefaultProtocol         = "tcp"
	DefaultSchedulingMethod = "wrr"
)

// Possible validation errors.
var (
	ErrMissingEndpoint = errors.New("endpoint information is missing")
	ErrUnknownProtocol = errors.New("specified protocol is unknown")
)

// ServiceOptions describe a virtual service.
type ServiceOptions struct {
	Host       string
	Port       uint16
	Protocol   string
	Method     string
	Persistent bool

	// Host string resolved to an IP, including DNS lookup.
	host net.IP

	// Protocol string converted to a protocol number.
	protocol uint16
}

// NewServiceOptions constructs new virtual service options.
func NewServiceOptions(host string, port uint16, protocol string) (*ServiceOptions, error) {
	options := &ServiceOptions{
		Host:       host,
		Port:       port,
		Protocol:   protocol,
		Method:     DefaultSchedulingMethod,
		Persistent: true,
	}

	if len(host) != 0 {
		if addr, err := net.ResolveIPAddr("ip", host); err == nil {
			options.host = addr.IP
		} else {
			return nil, err
		}
	} else {
		return nil, ErrMissingEndpoint
	}

	if port == 0 {
		return nil, ErrMissingEndpoint
	}

	if len(protocol) == 0 {
		options.Protocol = DefaultProtocol
	}

	options.Protocol = strings.ToLower(options.Protocol)

	switch options.Protocol {
	case "tcp":
		options.protocol = syscall.IPPROTO_TCP
	case "udp":
		options.protocol = syscall.IPPROTO_UDP
	default:
		return nil, ErrUnknownProtocol
	}

	return options, nil
}

// RealOptions describe a virtual service real.
type RealOptions struct {
	Host   string
	Port   uint16
	Weight int32
	VsID   string

	// Host string resolved to an IP, including DNS lookup.
	host net.IP

	// Forwarding method string converted to a forwarding method number.
	methodID uint32
}

// NewRealOptions constructs new real service options.
func NewRealOptions(host string, port uint16, weight int32, vsID string) (*RealOptions, error) {
	if len(host) == 0 || port == 0 {
		return nil, ErrMissingEndpoint
	}

	options := &RealOptions{
		Host:     host,
		Port:     port,
		Weight:   weight,
		VsID:     vsID,
		methodID: 0,
	}

	if addr, err := net.ResolveIPAddr("ip", options.Host); err == nil {
		options.host = addr.IP
	} else {
		return nil, err
	}

	if options.Weight <= 0 {
		options.Weight = 100
	}

	return options, nil
}
