package hub

import (
	"context"
	"testing"

	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestMinerDevices(t *testing.T) {
	id := &pb.ID{Id: "test"}

	gpuDevice, err := gpu.NewDevice("a", "b", 1488, 660)
	assert.NoError(t, err)

	hub := Hub{
		miners: map[string]*MinerCtx{
			id.Id: &MinerCtx{
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 64}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
		},
	}

	devices, err := hub.MinerDevices(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, len(devices.CPUs), 1)
	assert.Equal(t, len(devices.GPUs), 1)

	devices, err = hub.MinerDevices(context.Background(), &pb.ID{Id: "span"})
	assert.Error(t, err)
}
