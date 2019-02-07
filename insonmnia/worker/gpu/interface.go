package gpu

import (
	"context"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type GPUID string

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	// Tune is attaching GPUs with given IDs into a container
	Tune(hostconfig *container.HostConfig, ids []GPUID) error
	// Devices returns device list that this tuner can control
	Devices() []*sonm.GPUDevice
	// Close closes the tuner and removes related files and sockets from the host system
	Close() error
}

// New creates Tuner based on provided config.
// Currently supported type: "radeon" or "nvidia".
// if type is undefined, nilTuner will be returned.
func New(ctx context.Context, gpuType sonm.GPUVendorType, opts ...Option) (Tuner, error) {
	switch gpuType {
	case sonm.GPUVendorType_RADEON:
		return newRadeonTuner(ctx, opts...)
	case sonm.GPUVendorType_NVIDIA:
		return newNvidiaTuner(ctx, opts...)
	case sonm.GPUVendorType_FAKE:
		return newFakeTuner(ctx, opts...)
	case sonm.GPUVendorType_REMOTE:
		return newRemoteTuner(ctx, opts...)
	default:
		log.G(ctx).Debug("cannot detect gpu type, use nil tuner", zap.Int32("given_type", int32(gpuType)))
		return NilTuner{}, nil
	}
}

// MetricsHandlers returns hardware-dependent implementation
// of GPU metrics interface.
type MetricsHandler interface {
	GetMetrics() (map[string]float64, error)
	Close() error
}

func NewMetricsHandler(gpuType sonm.GPUVendorType) (MetricsHandler, error) {
	switch gpuType {
	case sonm.GPUVendorType_RADEON:
		return newRadeonMetricsHandler()
	case sonm.GPUVendorType_NVIDIA:
		return newNvidiaMetricsHandler()
	default:
		return nilMetricsHandler{}, nil
	}
}

type nilMetricsHandler struct{}

func (nilMetricsHandler) GetMetrics() (map[string]float64, error) { return map[string]float64{}, nil }
func (nilMetricsHandler) Close() error                            { return nil }

// NilTuner is just a null pattern
type NilTuner struct{}

func (NilTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error { return nil }

func (NilTuner) Devices() []*sonm.GPUDevice { return nil }

func (NilTuner) Close() error { return nil }
