package gpu

import (
	"context"
	"errors"

	"github.com/docker/docker/api/types/container"
)

type factory func(ctx context.Context, args Args) (Tuner, error)

var (
	modes = map[string]factory{
		"nil": nilFactory,
	}

	// ErrUnsupportedGPUType means that requeted GPU/OpenCL isolation type is not supported
	ErrUnsupportedGPUType = errors.New("requeted GPU type is not supported")
)

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	Tune(hostconfig *container.HostConfig) error
	Close() error
}

// New creates Tuner based on provided config.
// It may return ErrUnsupportedGPUType if the platform does not support requested iso type
// Currently supported type: nil, nvidiadocker, embedded
func New(ctx context.Context, cfg *Config) (Tuner, error) {
	switch cfg.Type {
	case "":
		return NilTuner{}, nil
	default:
		factory, ok := modes[cfg.Type]
		if !ok {
			return nil, ErrUnsupportedGPUType
		}

		return factory(ctx, cfg.Args)
	}
}

// NilTuner is just a null pattern
type NilTuner struct{}

func nilFactory(context.Context, Args) (Tuner, error) {
	return NilTuner{}, nil
}

func (NilTuner) Tune(hostconfig *container.HostConfig) error {
	return nil
}

func (NilTuner) Close() error { return nil }
