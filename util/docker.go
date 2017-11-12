package util

import (
	"errors"
	"strconv"
	"strings"
)

type PortBinding interface {
	// Network shows name of the network (for example, "tcp", "udp").
	Network() string
	// Port describes a port.
	Port() uint16
}

type portBinding struct {
	network string
	port    uint16
}

func (p *portBinding) Network() string {
	return p.network
}

func (p *portBinding) Port() uint16 {
	return p.port
}

// ParsePortBinding parses the given Docker port binding into components.
// For example `8080/tcp`.
func ParsePortBinding(v string) (PortBinding, error) {
	mapping := strings.Split(v, "/")
	if len(mapping) != 2 {
		return nil, errors.New("failed to decode Docker port mapping")
	}

	port, err := strconv.ParseUint(mapping[0], 10, 16)
	if err != nil {
		return nil, errors.New("failed to convert Docker port to a `uint16`")
	}

	return &portBinding{
		network: mapping[1],
		port:    uint16(port),
	}, nil
}
