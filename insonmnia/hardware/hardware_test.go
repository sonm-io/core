package hardware

import (
	"testing"

	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHardwareGpuType(t *testing.T) {
	hw := &Hardware{
		GPU: []gpu.Device{},
	}

	assert.Equal(t, sonm.GPUVendorType_GPU_UNKNOWN, hw.GPUType(), "no GPU, type must be resolved as UNKNOWN")

	// must be nvidia
	dev1, err := gpu.NewDevice("1", "", 100, 200, gpu.WithVendorId(4318))
	require.NoError(t, err)

	// must be unknown
	dev2, err := gpu.NewDevice("2", "", 100, 200, gpu.WithVendorId(1234))
	require.NoError(t, err)

	hw.GPU = append(hw.GPU, dev1, dev2)
	assert.Equal(t, sonm.GPUVendorType_NVIDIA, hw.GPUType(), "one of GPUs has an NVIDIA type, must be return nvidia type")

	// must be radeon
	dev3, err := gpu.NewDevice("3", "", 100, 200, gpu.WithVendorId(16915456))
	require.NoError(t, err)

	hw.GPU = []gpu.Device{dev2, dev3}
	assert.Equal(t, sonm.GPUVendorType_RADEON, hw.GPUType(), "one of GPUs has an RADEON type, must be return radeon type")

	// all GPUs have unknown type
	hw.GPU = []gpu.Device{dev2, dev2, dev2}
	assert.Equal(t, sonm.GPUVendorType_GPU_UNKNOWN, hw.GPUType(), "one of GPUs has an UNKNOWN type, must be return UNKNOWN type")
}
