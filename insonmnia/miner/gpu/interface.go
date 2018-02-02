package gpu

import (
	"context"
	"net"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	openCLVendorDir = "/etc/OpenCL/vendors"
)

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	Tune(hostconfig *container.HostConfig) error
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

func (NilTuner) Tune(hostconfig *container.HostConfig) error {
	return nil
}

func (NilTuner) Close() error {
	return nil
}

type volumePluginHandler struct {
	options         *tunerOptions
	devices         []string
	OpenCLVendorDir string
	handler         *volume.Handler
	listener        net.Listener
}
