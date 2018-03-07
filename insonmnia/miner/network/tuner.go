package network

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/sonm-io/core/insonmnia/structs"
)

type Cleanup interface {
	Close() error
}

// Tuner is responsible for preparing networking and baking proper options in container.HostConfig and
// network.NetworkingConfig.
type Tuner interface {
	Tune(net structs.Network, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig) (Cleanup, error)
}
