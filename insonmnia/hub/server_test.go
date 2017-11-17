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

func TestDevices(t *testing.T) {
	// GPU characteristics shared between miners.
	gpuDevice, err := gpu.NewDevice("a", "b", 1488, 660)
	assert.NoError(t, err)

	hub := Hub{
		miners: map[string]*MinerCtx{
			"miner1": {
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 64}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
			"miner2": {
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 65}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
		},
	}

	devices, err := hub.Devices(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(devices.CPUs), 2)
	assert.Equal(t, len(devices.GPUs), 1)
}

func TestMinerDevices(t *testing.T) {
	id := &pb.ID{Id: "miner1"}

	gpuDevice, err := gpu.NewDevice("a", "b", 1488, 660)
	assert.NoError(t, err)

	hub := Hub{
		miners: map[string]*MinerCtx{
			id.Id: {
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 64}},
					GPU: []gpu.Device{gpuDevice},
				},
			},

			"miner2": {
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 65}},
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
