// +build !darwin,cl

package gpu

import (
	"context"
	"net"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sshaman1101/nvidia-docker/nvidia"
	"go.uber.org/zap"
)

type radeonTuner struct {
	volumePluginHandler
}

func newRadeonTuner(ctx context.Context, opts *tunerOptions) (Tuner, error) {
	tun := radeonTuner{}
	tun.options = opts
	tun.devices = tun.getDevices()

	if _, err := os.Stat(openCLVendorDir); err == nil {
		tun.OpenCLVendorDir = openCLVendorDir
	}

	volInfo := []nvidia.VolumeInfo{
		{
			Name:         tun.options.volumeDriverName,
			Mountpoint:   "/opt/amdgpu-pro",
			MountOptions: "ro",
			Components: map[string][]string{
				"libraries": {
					"amdvlk64.so",
					"libEGL.so",
					"libGL.so",
					"libGLESv2.so",
					"libOpenCL.so",
					"libamdocl12cl64.so",
					"libamdocl64.so",
					// vdpau
					"libvdpau_amdgpu.so.1.0.0",
					// dri
					"radeonsi_drv_video.so",
					// gbm
					"gbm_amdgpu.so",

					// by noxiouz from prev impl
					"libMesaOpenCL.so",
					"pipe_vmwgfx.so",
					"pipe_r600.so",
					"pipe_r300.so",
					"pipe_radeonsi.so",
					"pipe_swrast.so",
					"pipe_nouveau.so",
				},
			},
		},
	}

	log.G(ctx).Info("Provisioning volumes", zap.String("at", tun.options.volumePath()))
	volumes, err := nvidia.LookupVolumes(tun.options.volumePath(), tun.options.driverVersion, volInfo)
	if err != nil {
		return nil, err
	}

	tun.handler = volume.NewHandler(NewPlugin(volumes))
	tun.listener, err = net.Listen("unix", tun.options.socketPath())
	if err != nil {
		log.G(ctx).Error("failed to create listening socket for to communicate with Docker as plugin",
			zap.String("path", tun.options.socketPath()), zap.Error(err))
		return nil, err
	}

	go func() {
		tun.handler.Serve(tun.listener)
	}()

	return tun, nil
}

func (radeonTuner) getDevices() []string {
	var dev []string
	if _, err := os.Stat("/dev/dri"); err == nil {
		dev = append(dev, "/dev/dri")
	}

	if _, err := os.Stat("/dev/kfd"); err == nil {
		dev = append(dev, "/dev/kfd")
	}

	return dev
}

func (tun radeonTuner) Tune(hostconfig *container.HostConfig) error {
	// NOTE: driver name depends on UNIX socket name which Docker uses to connect to a driver
	hostconfig.VolumeDriver = tun.options.volumeDriverName
	hostconfig.Binds = append(hostconfig.Binds, tun.options.volumeName()+":/usr/local/lib/amdgpu:ro")

	// put CL vendor into a container
	if tun.OpenCLVendorDir != "" {
		hostconfig.Binds = append(hostconfig.Binds, tun.OpenCLVendorDir+":"+tun.OpenCLVendorDir+":ro")
	}

	// put devices into a container
	for _, device := range tun.devices {
		hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   device,
			CgroupPermissions: "rwm",
		})
	}

	return nil
}

func (tun radeonTuner) Close() error {
	if err := tun.listener.Close(); err != nil {
		return err
	}

	return os.Remove(tun.options.socketPath())
}
