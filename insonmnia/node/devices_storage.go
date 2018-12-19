package node

import (
	"context"

	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
)

type devicesStorageAPI struct {
	eth blockchain.API
}

func newDevicesStorageAPI(opts *remoteOptions) sonm.DevicesStorageServer {
	return &devicesStorageAPI{eth: opts.eth}
}

func (m *devicesStorageAPI) Devices(ctx context.Context, address *sonm.EthAddress) (*sonm.StoredDevicesReply, error) {
	return m.eth.DeviceStorage().Devices(ctx, address.Unwrap())
}

func (m *devicesStorageAPI) RawDevices(ctx context.Context, address *sonm.EthAddress) (*sonm.RawDevicesReply, error) {
	return m.eth.DeviceStorage().RawDevices(ctx, address.Unwrap())
}
