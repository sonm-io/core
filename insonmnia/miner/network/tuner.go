package network

import (
	"github.com/docker/docker/api/types/network"
)

type Network interface {
	// ID returns a unique identifier that will be used as a new network name.
	ID() string
	// NetworkType returns a network driver name used to establish networking.
	NetworkType() string
	// NetworkOptions return configuration map, passed directly to network driver, this map should not be mutated.
	NetworkOptions() map[string]string
	// Returns network subnet in CIDR notation if applicable
	NetworkCIDR() string
	// Returns specified addr to join the network
	NetworkAddr() string
}

type Cleanup interface {
	Close() error
}

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	Tune(net Network, config *network.NetworkingConfig) (Cleanup, error)
}
