package gpu

import (
	"context"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type GPUID string

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	// Tune is attaching GPUs with given IDs into a container
	Tune(hostconfig *container.HostConfig, ids []GPUID) error
	// Devices returns device list that this tuner can control
	Devices() []*pb.GPUDevice
	// Close closes the tuner and removes related files and sockets from the host system
	Close() error
}

// New creates Tuner based on provided config.
// Currently supported type: "radeon" or "nvidia".
// if type is undefined, nilTuner will be returned.
func New(ctx context.Context, gpuType pb.GPUVendorType, opts ...Option) (Tuner, error) {
	switch gpuType {
	case pb.GPUVendorType_RADEON:
		return newRadeonTuner(ctx, opts...)
	case pb.GPUVendorType_NVIDIA:
		return newNvidiaTuner(ctx, opts...)
	default:
		log.G(ctx).Debug("cannot detect gpu type, use nil tuner", zap.Int32("given_type", int32(gpuType)))
		return NilTuner{}, nil
	}
}

// NilTuner is just a null pattern
type NilTuner struct{}

func (NilTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error { return nil }

func (NilTuner) Devices() []*pb.GPUDevice { return nil }

func (NilTuner) Close() error { return nil }
