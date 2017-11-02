// +build linux,embeddednvidia

package gpu

import (
	"context"
	"net"
	"os"

	"go.uber.org/zap"

	"github.com/NVIDIA/nvidia-docker/src/nvidia"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
)

func init() {
	modes["embedded"] = newEmbeddedTuner
}

var volPath = "/var/lib/nvidia-docker/volumes"

const (
	// we pin up driver  version and use patched NvidiaDocekr packages
	// otherwise it requires for ATI host to have NVIDIA driver installed
	// This name must be related to VolumeMap in NvidiaDocker itself.
	// This map contains names of libnraries to look for
	magicVolumeName = "nvidia_driver_300.0"

	dockerVolumeDriverName = "nvidia-docker"

	// TODO: configurable
	sockPath = "/run/docker/plugins/" + dockerVolumeDriverName + ".sock"
)

type embeddedTuner struct {
	devices         []string
	OpenCLVendorDir string
	handler         *volume.Handler
	pluginListener  net.Listener
}

func (g *embeddedTuner) Tune(hostconfig *container.HostConfig) error {
	// NOTE: driver name depends on UNIX socket name which Docker uses to connect to a driver
	hostconfig.VolumeDriver = dockerVolumeDriverName
	hostconfig.Binds = append(hostconfig.Binds, magicVolumeName+":"+"/usr/local/nvidia:ro")
	if g.OpenCLVendorDir != "" {
		hostconfig.Binds = append(hostconfig.Binds, g.OpenCLVendorDir+":"+g.OpenCLVendorDir+":ro")
	}
	for _, device := range g.devices {
		hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
			PathOnHost:        device,
			CgroupPermissions: "rwm",
		})
	}
	return nil
}

func (g *embeddedTuner) Close() error {
	if err := g.pluginListener.Close(); err != nil {
		return err
	}
	return os.Remove(sockPath)
}

func newEmbeddedTuner(ctx context.Context, _ Args) (Tuner, error) {
	var ovs embeddedTuner
	// Detect if we support NVIDIA
	log.G(ctx).Info("Loading NVIDIA unified memory")
	UVMErr := nvidia.LoadUVM()
	if UVMErr != nil {
		log.G(ctx).Warn("failed to load UVM %v. Seems NVIDIA is not installed on the host", zap.Error(UVMErr))
	}

	log.G(ctx).Info("Loading NVIDIA management library")
	initErr := nvidia.Init()
	if initErr == nil {
		defer func() { nvidia.Shutdown() }()
	}

	var nvidiaSupported = initErr == nil && UVMErr == nil
	if nvidiaSupported {
		log.G(ctx).Info("NVIDIA GPU supported by the host. Discovering GPU devices")
		devices, err := nvidia.LookupDevices()
		if err != nil {
			log.G(ctx).Error("failed to lookup GPU devices", zap.Error(err))
			return nil, err
		}
		cdevices, err := nvidia.GetControlDevicePaths()
		if err != nil {
			log.G(ctx).Error("failed to get contorl devices paths", zap.Error(err))
			return nil, err
		}
		ovs.devices = append(ovs.devices, cdevices...)
		for _, device := range devices {
			ovs.devices = append(ovs.devices, device.Path)
		}
	}

	if _, err := os.Stat("/dev/dri"); err == nil {
		ovs.devices = append(ovs.devices, "/dev/dri")
	}

	if _, err := os.Stat(openCLVendorDir); err == nil {
		ovs.OpenCLVendorDir = openCLVendorDir
	}

	// NOTE: Add libraries we need to discover
	nvolumes := nvidia.Volumes[0]
	components := nvolumes.Components
	components["libraries"] = append(components["libraries"],
		"libdrm.so",
		"libOpenCL.so",
		"libMesaOpenCL.so",
		"pipe_vmwgfx.so",
		"pipe_r600.so",
		"pipe_r300.so",
		"pipe_radeonsi.so",
		"pipe_swrast.so",
		"pipe_nouveau.so",
	)
	log.G(ctx).Info("Provisioning volumes", zap.String("at", volPath))
	volumes, err := nvidia.LookupVolumes(volPath)
	if err != nil {
		return nil, err
	}
	ovs.handler = volume.NewHandler(NewPlugin(volumes))

	ovs.pluginListener, err = net.Listen("unix", sockPath)
	if err != nil {
		log.G(ctx).Error("failed to create listening socket for to communicate with Docker as plugin", zap.String("path", sockPath), zap.Error(err))
		return nil, err
	}

	go func() {
		ovs.handler.Serve(ovs.pluginListener)
	}()

	return &ovs, nil
}
