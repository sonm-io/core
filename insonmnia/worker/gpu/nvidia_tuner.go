// +build !darwin,cl

package gpu

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/sshaman1101/nvidia-docker/nvidia"
	"go.uber.org/zap"
)

type nvidiaDevice struct {
	// "NVidia 1080 Ti"
	name string
	// "/dev/nvidia0"
	devicePath string
	// "/dev/dri/card0", "/dev/dri/renderD128"
	driDevice DRICard
	// "/dev/nvidiactl", "/dev/nvidia-uvm", "/dev/nvidia-uvm-tools"
	ctrlDevices []string
	mem         uint64
}

func (dev *nvidiaDevice) String() string {
	return fmt.Sprintf("%s (%s)", dev.name, dev.devicePath)
}

func (dev *nvidiaDevice) ID() GPUID {
	return GPUID(dev.driDevice.PCIBusID)
}

// Devices returns all device files that must be bound to container
func (dev *nvidiaDevice) Devices() []string {
	return append(append(dev.ctrlDevices, dev.driDevice.Devices...), dev.devicePath)
}

type nvidiaTuner struct {
	options  *tunerOptions
	handler  *volume.Handler
	listener net.Listener

	m      sync.Mutex
	devMap map[GPUID]nvidiaDevice
}

/*
	Note: card cllecting code into the Tune() method is almost similar to radeon's one
*/

func (g *nvidiaTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	g.m.Lock()
	defer g.m.Unlock()

	var cardsToBind = make(map[GPUID]nvidiaDevice)
	for _, id := range ids {
		card, ok := g.devMap[id]
		if !ok {
			return fmt.Errorf("cannot allocate device: unknown id %s", id)
		}

		// copy cards to the map (instead of slice) preventing us
		// from binding same card more than once
		cardsToBind[id] = card
	}

	for _, card := range cardsToBind {
		for _, device := range card.Devices() {
			hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
				PathOnHost:        device,
				PathInContainer:   device,
				CgroupPermissions: "rwm",
			})
		}
	}

	mnt := newVolumeMount(g.options.volumeName(), g.options.libsMountPoint, g.options.VolumeDriverName)
	hostconfig.Mounts = append(hostconfig.Mounts, mnt)

	return nil
}

func (g *nvidiaTuner) Devices() []*sonm.GPUDevice {
	g.m.Lock()
	defer g.m.Unlock()

	var devices []*sonm.GPUDevice
	for _, d := range g.devMap {
		dev := &sonm.GPUDevice{
			ID:          string(d.ID()),
			VendorName:  "Nvidia",
			VendorID:    d.driDevice.VendorID,
			DeviceName:  d.name,
			DeviceID:    d.driDevice.DeviceID,
			MajorNumber: d.driDevice.Major,
			MinorNumber: d.driDevice.Minor,
			Memory:      d.mem,
		}

		dev.FillHashID()
		devices = append(devices, dev)
	}

	return devices
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

	ovs := nvidiaTuner{devMap: make(map[GPUID]nvidiaDevice)}
	ovs.options = options

	// get devices list provided by openCL
	clDevices, err := gpu.GetGPUDevices()
	if err != nil {
		return nil, err
	}

	// check if there is any device with required type
	if err := hasGPUWithVendor(sonm.GPUVendorType_NVIDIA, clDevices); err != nil {
		return nil, err
	}

	// Detect if we support NVIDIA
	log.G(ctx).Info("loading NVIDIA unified memory")
	if err := nvidia.LoadUVM(); err != nil {
		log.G(ctx).Error("failed to load UVM, seems NVIDIA is not installed on the host", zap.Error(err))
		return nil, err
	}

	log.G(ctx).Info("loading NVIDIA management library")
	if err := nvidia.Init(); err != nil {
		log.G(ctx).Error("failed to init NVML", zap.Error(err))
		return nil, err
	}

	defer func() { nvidia.Shutdown() }()

	log.G(ctx).Info("NVIDIA GPU supported by the host, discovering GPU devices")
	devices, err := nvidia.LookupDevices()
	if err != nil {
		log.G(ctx).Error("failed to lookup GPU devices", zap.Error(err))
		return nil, err
	}

	ctrlDevices, err := nvidia.GetControlDevicePaths()
	if err != nil {
		log.G(ctx).Error("failed to get control devices paths", zap.Error(err))
		return nil, err
	}

	for _, d := range devices {
		card, err := newCardByDevicePath(d.Path)
		if err != nil {
			return nil, err
		}

		// d.Memory.Global is presented in Megabytes
		memBytes := uint64(*d.Memory.Global) * 1024 * 1024

		dev := nvidiaDevice{
			name:        *d.Model,
			devicePath:  d.Path,
			driDevice:   card,
			ctrlDevices: ctrlDevices,
			mem:         memBytes,
		}
		ovs.devMap[dev.ID()] = dev

		log.G(ctx).Debug("discovered gpu devices",
			zap.String("root", dev.String()),
			zap.Strings("ctrl", dev.ctrlDevices),
			zap.Strings("dri", card.Devices),
			zap.Uint64("mem", memBytes))
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
					"libnvidia-ml.so",              // Management library
					"libcuda.so",                   // CUDA driver library
					"libnvidia-ptxjitcompiler.so",  // PTX-SASS JIT compiler (used by libcuda)
					"libnvidia-fatbinaryloader.so", // fatbin loader (used by libcuda)
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

func newVolumeMount(src, dst, name string) mount.Mount {
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
