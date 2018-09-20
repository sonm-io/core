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
	devMap map[GPUID]DRICard
}

func newRadeonTuner(ctx context.Context, opts ...Option) (Tuner, error) {
	options := radeonDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	tun := radeonTuner{
		devMap: make(map[GPUID]DRICard),
	}

	clDevices, err := gpu.GetGPUDevices()
	if err != nil {
		return nil, err
	}

	if err := hasGPUWithVendor(sonm.GPUVendorType_RADEON, clDevices); err != nil {
		return nil, err
	}

	cards, err := CollectDRICardDevices()
	if err != nil {
		return nil, err
	}

	// filter CL devices by known vendor IDs
	var radeonDevices []*sonm.GPUDevice
	for _, dev := range clDevices {
		if dev.VendorType() == sonm.GPUVendorType_RADEON {
			radeonDevices = append(radeonDevices, dev)
		}
	}

	// match DRI and CL devices by vendor ID
	var driRadeonDevices []DRICard
	for _, dev := range cards {
		for _, rid := range sonm.Radeons {
			if dev.VendorID == rid {
				driRadeonDevices = append(driRadeonDevices, dev)
			}
		}
	}

	// wow, so different, such weird
	if len(radeonDevices) != len(driRadeonDevices) {
		return nil, errors.New("cannot find matching CL device to extract vmem")
	}

	for i, card := range driRadeonDevices {
		card.Memory = radeonDevices[i].Memory
		tun.devMap[GPUID(card.PCIBusID)] = card

		log.G(ctx).Debug("discovered gpu device ",
			zap.String("dev", card.Path),
			zap.Strings("driDevice", card.Devices),
			zap.Uint64("vmem", card.Memory))
	}

	return tun, nil
}

func (tun radeonTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	tun.m.Lock()
	defer tun.m.Unlock()

	var cardsToBind = make(map[GPUID]DRICard)

	for _, id := range ids {
		card, ok := tun.devMap[id]
		if !ok {
			return fmt.Errorf("cannot allocate device: unknown id %s", id)
		}

		// copy cards to the map (instead of slice) preventing us
		// from binding same card more than once
		cardsToBind[id] = card
	}

	for _, card := range cardsToBind {
		for _, device := range card.Devices {
			hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
				PathOnHost:        device,
				PathInContainer:   device,
				CgroupPermissions: "rwm",
			})
		}
	}

	return nil
}

func (tun radeonTuner) Devices() []*sonm.GPUDevice {
	tun.m.Lock()
	defer tun.m.Unlock()

	var devices []*sonm.GPUDevice
	for _, d := range tun.devMap {
		dev := &sonm.GPUDevice{
			ID:          d.PCIBusID,
			VendorName:  "Radeon",
			VendorID:    d.VendorID,
			DeviceName:  d.Name,
			DeviceID:    d.DeviceID,
			MajorNumber: d.Major,
			MinorNumber: d.Minor,
			Memory:      d.Memory,
		}

		dev.FillHashID()
		devices = append(devices, dev)
	}

	return devices
}

func (tun radeonTuner) Close() error {
	return nil
}
