package npp

import (
	"net"

	"github.com/sonm-io/core/insonmnia/auth"
)

type RendezvousConfig struct {
	Endpoints []string
}

func (m *RendezvousConfig) ConvertEndpoints() ([]auth.Addr, error) {
	var endpoints []auth.Addr
	for _, endpoint := range m.Endpoints {
		addr, err := auth.NewAddr(endpoint)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, *addr)
	}

	return endpoints, nil
}

type RelayConfig struct {
	Endpoints []string
}

func (m *RelayConfig) ConvertEndpoints() ([]net.Addr, error) {
	var endpoints []net.Addr
	for _, endpoint := range m.Endpoints {
		addr, err := net.ResolveTCPAddr("tcp", endpoint)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, addr)
	}

	return endpoints, nil
}

type Config struct {
	Rendezvous RendezvousConfig
	Relay      RelayConfig
}
