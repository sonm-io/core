package network

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/sonm-io/core/proto"
)

type Cleanup interface {
	Close() error
}

// Tuner is responsible for preparing networking and baking proper options in container.HostConfig and
// network.NetworkingConfig.
type Tuner interface {
	Tune(ctx context.Context, net *sonm.NetworkSpec, hostConfig *container.HostConfig, netConfig *network.NetworkingConfig) (Cleanup, error)
	GenerateInvitation(ID string) (*sonm.NetworkSpec, error)
	GetCleaner(ctx context.Context, ID string) (Cleanup, error)
	Tuned(ID string) bool
}
