package gpu

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
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
// Currently supported type: radeon, nvidia, nvidia-docker-v1,
// if type is undefined, nilTuner will be returned.
func New(ctx context.Context, gpuType pb.GPUVendorType) (Tuner, error) {
	switch gpuType {
	case pb.GPUVendorType_RADEON:
		return newRadeonTuner(ctx, newRadeonOptions())
	case pb.GPUVendorType_NVIDIA:
		return newNvidiaTuner(ctx, newNvidiaOptions())
	default:
		log.G(ctx).Debug("cannot detect gpu type, use nil tuner")
		return NilTuner{}, nil
	}
}

// NilTuner is just a null pattern
type NilTuner struct{}

func (NilTuner) Tune(hostconfig *container.HostConfig) error { return nil }

func (NilTuner) Close() error { return nil }

// Config contains options related to NVIDIA GPU support
type Config struct {
	Type string `yaml:"type"`
}

// tunerOptions contains various options for embedded GPU tuners
type tunerOptions struct {
	volumeDriverName string
	driverVersion    string

	// extra option for nvidia-docker v1 support
	nvidiaDockerEndpoint string
}

func newNvidiaOptions() *tunerOptions {
	return &tunerOptions{
		volumeDriverName: "nvidia-docker",
		driverVersion:    "300.0",
	}
}

func newRadeonOptions() *tunerOptions {
	return &tunerOptions{
		volumeDriverName: "amd-docker",
		driverVersion:    "2482.3",
	}
}

func newNvidiaDockerV1Options() *tunerOptions {
	return &tunerOptions{
		// FIXME(sshaman1101): do not hardcode the addr
		nvidiaDockerEndpoint: "localhost:3476",
	}
}

func (opts *tunerOptions) volumePath() string {
	return fmt.Sprintf("/var/lib/%s/volumes", opts.volumeDriverName)
}

func (opts *tunerOptions) socketPath() string {
	return fmt.Sprintf("/run/docker/plugins/%s.sock", opts.volumeDriverName)
}

func (opts *tunerOptions) volumeName() string {
	return fmt.Sprintf("%s_%s", opts.volumeDriverName, opts.driverVersion)
}

type volumePluginHandler struct {
	options         *tunerOptions
	devices         []string
	OpenCLVendorDir string
	handler         *volume.Handler
	listener        net.Listener
}
