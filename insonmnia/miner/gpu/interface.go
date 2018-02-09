package gpu

import (
	"context"
	"net"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Tuner is responsible for preparing GPU-friendly environment and baking proper options in container.HostConfig
type Tuner interface {
	Tune(hostConfig *container.HostConfig) error
	Close() error
}

// New creates Tuner based on provided config.
// Currently supported type: "radeon" or "nvidia".
// if type is undefined, nilTuner will be returned.
func New(ctx context.Context, gpuType pb.GPUVendorType, opts ...Option) (Tuner, error) {
	switch gpuType {
	case pb.GPUVendorType_RADEON:
		return newRadeonTuner(opts...)
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
	options  *tunerOptions
	devices  []string
	handler  *volume.Handler
	listener net.Listener
}

func (vh *volumePluginHandler) tune(hostconfig *container.HostConfig) error {
	if vh.handler != nil {
		// we have plugin handler, so must mount the volume into the container
		hostconfig.Mounts = append(hostconfig.Mounts,
			makeVolumeMount(vh.options.volumeName(), vh.options.libsMountPoint, vh.options.VolumeDriverName))
	}

	for _, device := range vh.devices {
		hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   device,
			CgroupPermissions: "rwm",
		})
	}

	return nil
}

func makeVolumeMount(src, dst, name string) mount.Mount {
	return mount.Mount{
		Type:         mount.TypeVolume,
		Source:       src,
		Target:       dst,
		ReadOnly:     true,
		Consistency:  mount.ConsistencyDefault,
		BindOptions:  nil,
		TmpfsOptions: nil,
		VolumeOptions: &mount.VolumeOptions{
			NoCopy: false,
			Labels: map[string]string{},
			DriverConfig: &mount.Driver{
				Name:    name,
				Options: map[string]string{},
			},
		},
	}
}
