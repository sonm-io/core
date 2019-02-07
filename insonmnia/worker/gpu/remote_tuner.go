package gpu

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type remoteTuner struct {
	log *zap.Logger
	rt  sonm.RemoteGPUTunerClient
}

func newRemoteTuner(ctx context.Context, opts ...Option) (Tuner, error) {
	options := remoteDefaultOptions()
	for _, f := range opts {
		f(options)
	}

	cc, err := xgrpc.NewClient(ctx, options.RemoteSocket, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %v", err)
	}

	tun := sonm.NewRemoteGPUTunerClient(cc)
	if _, err := tun.Devices(ctx, &sonm.RemoteGPUDeviceRequest{}); err != nil {
		return nil, fmt.Errorf("failed to access remote GPU tuner via %s: %v", options.RemoteSocket, err)
	}

	return &remoteTuner{
		log: ctxlog.GetLogger(ctx),
		rt:  sonm.NewRemoteGPUTunerClient(cc),
	}, nil
}

func (m *remoteTuner) Devices() []*sonm.GPUDevice {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := m.rt.Devices(ctx, &sonm.RemoteGPUDeviceRequest{})
	if err != nil {
		return nil
	}

	return devices.GetDevices()
}

func (m *remoteTuner) deviceMap() map[GPUID]*sonm.GPUDevice {
	devMap := map[GPUID]*sonm.GPUDevice{}
	for _, dev := range m.Devices() {
		devMap[GPUID(dev.ID)] = dev
	}

	return devMap
}

func (m *remoteTuner) Tune(hostconfig *container.HostConfig, ids []GPUID) error {
	devMap := m.deviceMap()

	var cardsToBind = make(map[GPUID]*sonm.GPUDevice)
	for _, id := range ids {
		card, ok := devMap[id]
		if !ok {
			return fmt.Errorf("cannot allocate device: unknown id %s", id)
		}

		cardsToBind[id] = card
	}

	for _, card := range cardsToBind {
		for _, device := range card.GetDeviceFiles() {
			hostconfig.Devices = append(hostconfig.Devices, container.DeviceMapping{
				PathOnHost:        device,
				PathInContainer:   device,
				CgroupPermissions: "rwm",
			})
		}

		for name, pair := range card.GetDriverVolumes() {
			srcDst := strings.Split(pair, ":")
			if len(srcDst) != 2 {
				return fmt.Errorf("malformed driver mount-point `%s`", pair)
			}

			hostconfig.Mounts = append(hostconfig.Mounts, newVolumeMount(srcDst[0], srcDst[1], name))
		}
	}

	return nil
}

func (m *remoteTuner) Close() error { return nil }
