// +build !darwin,cl

package gpu

import (
	"context"
	"net"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"github.com/sshaman1101/nvidia-docker/nvidia"
	"go.uber.org/zap"
)

type nvidiaTuner struct {
	volumePluginHandler
}

func (g *nvidiaTuner) Tune(hostconfig *container.HostConfig) error {
	return g.tune(hostconfig)
}

func (g *nvidiaTuner) Close() error {
	if err := g.listener.Close(); err != nil {
		return err
	}
	return os.Remove(g.options.SocketPath)
}

func newNvidiaTuner(ctx context.Context, opts ...Option) (Tuner, error) {
	options := nvidiaDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	ovs := nvidiaTuner{}
	ovs.options = options

	if err := hasGPUWithVendor(sonm.GPUVendorType_NVIDIA); err != nil {
		return nil, err
	}

	// Detect if we support NVIDIA
	log.G(ctx).Info("loading NVIDIA unified memory")
	if err := nvidia.LoadUVM(); err != nil {
		log.G(ctx).Error("failed to load UVM. Seems NVIDIA is not installed on the host", zap.Error(err))
		return nil, err
	}

	log.G(ctx).Info("loading NVIDIA management library")
	if err := nvidia.Init(); err != nil {
		log.G(ctx).Error("failed to init NVML", zap.Error(err))
		return nil, err
	}

	defer func() { nvidia.Shutdown() }()

	log.G(ctx).Info("NVIDIA GPU supported by the host. Discovering GPU devices")
	devices, err := nvidia.LookupDevices()
	if err != nil {
		log.G(ctx).Error("failed to lookup GPU devices", zap.Error(err))
		return nil, err
	}

	cdevices, err := nvidia.GetControlDevicePaths()
	if err != nil {
		log.G(ctx).Error("failed to get control devices paths", zap.Error(err))
		return nil, err
	}

	ovs.devices = append(ovs.devices, cdevices...)
	for _, device := range devices {
		ovs.devices = append(ovs.devices, device.Path)
	}

	if _, err := os.Stat("/dev/dri"); err == nil {
		ovs.devices = append(ovs.devices, "/dev/dri")
	}

	volInfo := []nvidia.VolumeInfo{
		{
			Name:         ovs.options.VolumeDriverName,
			Mountpoint:   "/usr/local/nvidia",
			MountOptions: "ro",
			Components: map[string][]string{
				"binaries": {
					"nvidia-smi", // System management interface
				},
				"libraries": {
					"libnvidia-ml.so", // Management library
				},
			},
		},
	}

	log.G(ctx).Info("provisioning volumes", zap.String("at", ovs.options.VolumePath))
	volumes, err := nvidia.LookupVolumes(ovs.options.VolumePath, ovs.options.DriverVersion, volInfo)
	if err != nil {
		return nil, err
	}

	ovs.handler = volume.NewHandler(NewPlugin(volumes))
	ovs.listener, err = net.Listen("unix", ovs.options.SocketPath)

	if err != nil {
		log.G(ctx).Error("failed to create listening socket for to communicate with Docker as plugin",
			zap.String("path", ovs.options.SocketPath), zap.Error(err))
		return nil, err
	}

	go func() {
		ovs.handler.Serve(ovs.listener)
	}()

	return &ovs, nil
}
