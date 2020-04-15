// +build !darwin,cl

package gpu

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/sshaman1101/nvidia-docker/nvidia"
	"go.uber.org/zap"
)

type nvidiaTuner struct {
	options  *tunerOptions
	handler  *volume.Handler
	listener net.Listener

	m      sync.Mutex
	devMap map[GPUID]*sonm.GPUDevice
}

func (g *nvidiaTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	g.m.Lock()
	defer g.m.Unlock()

	return tuneContainer(hostconfig, g.devMap, ids)
}

func (g *nvidiaTuner) Devices() []*sonm.GPUDevice {
	g.m.Lock()
	defer g.m.Unlock()

	var devices []*sonm.GPUDevice
	for _, d := range g.devMap {
		devices = append(devices, d)
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

	ovs := nvidiaTuner{
		devMap:  make(map[GPUID]*sonm.GPUDevice),
		options: options,
	}

	ovs.options.DriverVersion, err = nvidia.GetDriverVersion()
	if err != nil {
		log.G(ctx).Error("failed to get NVIDIA driver version", zap.Error(err))
		return nil, err
	}

	for _, d := range devices {
		card, err := newCardByDevicePath(d.Path)
		if err != nil {
			return nil, err
		}

		// d.Memory.Global is presented in Megabytes
		memBytes := uint64(*d.Memory.Global) * 1024 * 1024

		dev := &sonm.GPUDevice{
			ID:          card.PCIBusID,
			VendorName:  "NVidia",
			DeviceName:  *d.Model,
			VendorID:    card.VendorID,
			DeviceID:    card.DeviceID,
			MajorNumber: card.Major,
			MinorNumber: card.Minor,
			Memory:      memBytes,
			DeviceFiles: append(append(ctrlDevices, card.Devices...), d.Path),
			DriverVolumes: map[string]string{
				options.VolumeDriverName: fmt.Sprintf("%s:%s", options.volumeName(), options.libsMountPoint),
			},
		}

		dev.FillHashID()
		ovs.devMap[GPUID(card.PCIBusID)] = dev
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
					// content of closed-source drivers blob,
					// without ".version" and ".0", if any.
					"libEGL_nvidia.so",
					"libEGL.so",
					"libGLdispatch.so",
					"libGLESv1_CM_nvidia.so",
					"libGLESv1_CM.so",
					"libGLESv2_nvidia.so",
					"libGLESv2.so",
					"libGL.so",
					"libGLX_nvidia.so",
					"libglxserver_nvidia.so",
					"libGLX.so",
					"libnvcuvid.so",
					"libnvidia-allocator.so",
					"libnvidia-cbl.so",
					"libnvidia-cfg.so",
					"libnvidia-compiler.so",
					"libnvidia-eglcore.so",
					"libnvidia-encode.so",
					"libnvidia-fbc.so",
					"libnvidia-glcore.so",
					"libnvidia-glsi.so",
					"libnvidia-glvkspirv.so",
					"libnvidia-ifr.so",
					"libnvidia-opencl.so",
					"libnvidia-opticalflow.so",
					"libnvidia-rtcore.so",
					"libnvidia-tls.so",
					"libnvoptix.so",
					"libOpenCL.so",
					"libOpenGL.so",
					"libvdpau_nvidia.so",
					"nvidia_drv.so",
				},
			},
		},
	}

	log.G(ctx).Info("provisioning volumes",
		zap.String("at", ovs.options.VolumePath),
		zap.String("version", ovs.options.DriverVersion))

	volumes, err := nvidia.LookupVolumes(ovs.options.VolumePath, ovs.options.DriverVersion, volInfo)
	if err != nil {
		return nil, err
	}

	ovs.handler = volume.NewHandler(NewPlugin(volumes))
	ovs.listener, err = sockets.NewUnixSocket(ovs.options.SocketPath, syscall.Getgid())
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
