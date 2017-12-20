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
