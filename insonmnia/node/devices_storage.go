package node

import (
	"context"

	"github.com/sonm-io/core/proto"
)

type devicesStorageAPI struct {
	remotes *remoteOptions
}

func newDevicesStorageAPI(opts *remoteOptions) sonm.DevicesStorageServer {
	return &devicesStorageAPI{remotes: opts}
}

func (m *devicesStorageAPI) Devices(ctx context.Context, address *sonm.EthAddress) (*sonm.StoredDevicesReply, error) {
	return m.remotes.eth.DeviceStorage().Devices(ctx, address.Unwrap())
}

func (m *devicesStorageAPI) RawDevices(ctx context.Context, address *sonm.EthAddress) (*sonm.RawDevicesReply, error) {
	rawDevices, ts, err := m.remotes.eth.DeviceStorage().RawDevices(ctx, address.Unwrap())
	if err != nil {
		return nil, err
	}
	return &sonm.RawDevicesReply{
		Data:      rawDevices,
		Timestamp: ts,
	}, nil
}
