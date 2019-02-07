package gpu

import (
	"context"

	"github.com/sonm-io/core/proto"
)

type remoteTunerService struct {
	tun Tuner
}

func NewRemoteTuner(ctx context.Context, name string) (sonm.RemoteGPUTunerServer, error) {
	vendor, err := GetVendorByName(name)
	if err != nil {
		return nil, err
	}

	t, err := New(ctx, vendor)
	if err != nil {
		return nil, err
	}

	return &remoteTunerService{tun: t}, nil
}

func (m *remoteTunerService) Devices(context.Context, *sonm.RemoteGPUDeviceRequest) (*sonm.RemoteGPUDeviceReply, error) {
	return &sonm.RemoteGPUDeviceReply{
		Devices: m.tun.Devices(),
	}, nil
}
