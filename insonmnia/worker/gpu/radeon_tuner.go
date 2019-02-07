// +build !darwin,cl

package gpu

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type radeonTuner struct {
	m      sync.Mutex
	devMap map[GPUID]*sonm.GPUDevice
}

// collectDRICardsWithOpenCL collects DRI devices and match it with openCL.
func collectDRICardsWithOpenCL() ([]*DRICard, error) {
	openclDevices, err := gpu.GetGPUDevices()
	if err != nil {
		return nil, err
	}

	if err := hasGPUWithVendor(sonm.GPUVendorType_RADEON, openclDevices); err != nil {
		return nil, err
	}

	cards, err := CollectDRICardDevices()
	if err != nil {
		return nil, err
	}

	// filter CL devices by known vendor IDs
	var radeonDevices []*sonm.GPUDevice
	for _, dev := range openclDevices {
		if dev.VendorType() == sonm.GPUVendorType_RADEON {
			radeonDevices = append(radeonDevices, dev)
		}
	}

	// match DRI and CL devices by vendor ID
	var driDevices []*DRICard
	for i, dev := range cards {
		for _, rid := range sonm.Radeons {
			if dev.VendorID == rid {
				driDevices = append(driDevices, &cards[i])
			}
		}
	}

	// wow, so different, such weird
	if len(radeonDevices) != len(driDevices) {
		return nil, errors.New("cannot find matching CL device to extract vmem")
	}

	for i, card := range driDevices {
		// copy card memory value from openCL device to DRI device
		card.Memory = openclDevices[i].Memory
	}

	return driDevices, nil
}

func newRadeonTuner(ctx context.Context, opts ...Option) (Tuner, error) {
	options := radeonDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	tun := &radeonTuner{
		devMap: make(map[GPUID]*sonm.GPUDevice),
	}

	devices, err := collectDRICardsWithOpenCL()
	if err != nil {
		return nil, err
	}

	for _, card := range devices {
		dev := &sonm.GPUDevice{
			ID:          card.PCIBusID,
			VendorID:    card.VendorID,
			VendorName:  "Radeon",
			DeviceID:    card.DeviceID,
			DeviceName:  card.Name,
			MajorNumber: card.Major,
			MinorNumber: card.Minor,
			Memory:      card.Memory,
			DeviceFiles: card.Devices,
		}
		dev.FillHashID()

		tun.devMap[GPUID(card.PCIBusID)] = dev
		log.G(ctx).Debug("discovered gpu device ",
			zap.String("dev", card.Path),
			zap.Strings("dri_devices", card.Devices),
			zap.Uint64("vmem", card.Memory))
	}

	return tun, nil
}

func (tun *radeonTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	tun.m.Lock()
	defer tun.m.Unlock()

	return tuneContainer(hostconfig, tun.devMap, ids)
}

func (tun *radeonTuner) Devices() []*sonm.GPUDevice {
	tun.m.Lock()
	defer tun.m.Unlock()

	var devices []*sonm.GPUDevice
	for _, d := range tun.devMap {
		devices = append(devices, d)
	}

	return devices
}

func (radeonTuner) Close() error {
	return nil
}
