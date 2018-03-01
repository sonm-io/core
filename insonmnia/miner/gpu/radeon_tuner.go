// +build !darwin,cl

package gpu

import (
	"context"
	"fmt"
	"sync"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	pb "github.com/sonm-io/core/proto"
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

	if err := hasGPUWithVendor(sonm.GPUVendorType_RADEON); err != nil {
		return nil, err
	}

	cards, err := CollectDRICardDevices()
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		tun.devMap[GPUID(card.Path)] = card

		log.G(ctx).Debug("discovered gpu device ",
			zap.String("dev", card.Path),
			zap.Strings("driDevice", card.Devices))
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

func (tun radeonTuner) Devices() []*pb.GPUDevice {
	tun.m.Lock()
	defer tun.m.Unlock()

	var devices []*pb.GPUDevice
	for _, d := range tun.devMap {
		devices = append(devices, &pb.GPUDevice{
			ID:         d.PCIBusID,
			VendorName: "Radeon",
			VendorID:   d.VendorID,
			// note: name may be some awkward shit
			DeviceName:  d.Name,
			DeviceID:    d.DeviceID,
			MajorNumber: d.Major,
			MinorNumber: d.Minor,
		})
	}

	return devices
}

func (tun radeonTuner) Close() error {
	return nil
}
