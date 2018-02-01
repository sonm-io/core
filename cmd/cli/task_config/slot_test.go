package task_config

import (
	"testing"

	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestSlotConfig_IntoSlot(t *testing.T) {
	s := &SlotConfig{
		Duration: "1h",
		Resources: ResourcesConfig{
			Network: NetworkConfig{
				Type: pb.NetworkType_INCOMING.String(),
			},
			Gpu: pb.GPUCount_NO_GPU.String(),
		},
	}

	slot, err := s.IntoSlot()
	assert.NoError(t, err)
	assert.Equal(t, slot.Unwrap().GetDuration(), uint64(3600))
}

func TestSlotConfig_IntoSlot_TooShort(t *testing.T) {
	s := &SlotConfig{
		Duration: "5m",
		Resources: ResourcesConfig{
			Network: NetworkConfig{
				Type: pb.NetworkType_INCOMING.String(),
			},
			Gpu: pb.GPUCount_NO_GPU.String(),
		},
	}

	slot, err := s.IntoSlot()
	assert.EqualError(t, err, structs.ErrDurationIsTooShort.Error())
	assert.Nil(t, slot)
}

func TestSlotConfig_IntoSlot_HumanReadableValues(t *testing.T) {
	s := &SlotConfig{
		Duration: "15m",
		Resources: ResourcesConfig{
			Ram:     "4 GB",
			Storage: "100TB",
			Network: NetworkConfig{
				In:   "1 Mb",
				Out:  "2kb",
				Type: pb.NetworkType_INCOMING.String(),
			},
			Gpu: pb.GPUCount_NO_GPU.String(),
		},
	}

	slot, err := s.IntoSlot()
	assert.NoError(t, err)

	assert.Equal(t, uint64(4294967296), slot.Unwrap().GetResources().GetRamBytes())
	assert.Equal(t, uint64(109951162777600), slot.Unwrap().GetResources().GetStorage())
	assert.Equal(t, uint64(1048576), slot.Unwrap().GetResources().GetNetTrafficIn())
	assert.Equal(t, uint64(2048), slot.Unwrap().GetResources().GetNetTrafficOut())
}

func TestSlotConfig_SingleGPU(t *testing.T) {
	s := &SlotConfig{
		Duration: "15m",
		Resources: ResourcesConfig{
			Ram:     "4 GB",
			Storage: "100TB",
			Network: NetworkConfig{
				In:   "1 Mb",
				Out:  "2kb",
				Type: pb.NetworkType_INCOMING.String(),
			},
			Gpu: pb.GPUCount_SINGLE_GPU.String(),
		},
	}

	_, err := s.IntoSlot()
	assert.EqualError(t, err, structs.ErrUnsupportedSingleGPU.Error())
}
