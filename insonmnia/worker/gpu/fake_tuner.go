package gpu

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type fakeGPUTuner struct {
	log     *zap.Logger
	devices []*sonm.GPUDevice
}

func newFakeTuner(ctx context.Context, opts ...Option) (Tuner, error) {
	options := fakeDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	var devices []*sonm.GPUDevice
	for i := 0; i < options.DeviceCount; i++ {
		dev := &sonm.GPUDevice{
			ID:          fmt.Sprintf("PCI:%04d", i),
			VendorID:    uint64(sonm.GPUVendorType_FAKE),
			VendorName:  "FAKE",
			DeviceID:    uint64(i),
			DeviceName:  fmt.Sprintf("Fake GPU slot%d", i),
			MajorNumber: uint64(100 + i),
			MinorNumber: uint64(200 + i),
			Memory:      4294967296,
		}

		dev.FillHashID()
		devices = append(devices, dev)
	}

	return &fakeGPUTuner{
		log:     ctxlog.GetLogger(ctx),
		devices: devices,
	}, nil
}

func (f *fakeGPUTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	f.log.Debug("tuning container with fake GPU driver", zap.Any("device_ids", ids))
	return nil
}

func (f *fakeGPUTuner) Devices() []*sonm.GPUDevice {
	f.log.Debug("requesting devices from fake GPU driver")
	return f.devices
}

func (f *fakeGPUTuner) Close() error {
	f.log.Debug("closing fake GPU driver")
	return nil
}
