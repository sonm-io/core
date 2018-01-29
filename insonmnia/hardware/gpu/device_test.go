package gpu

import (
	"testing"

	"fmt"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deviceIdsTable = []struct {
	id uint64
	t  sonm.GPUVendorType
}{
	{id: 4098, t: sonm.GPUVendorType_RADEON},
	{id: 16915456, t: sonm.GPUVendorType_RADEON},
	{id: 4318, t: sonm.GPUVendorType_NVIDIA},
	{id: 16925952, t: sonm.GPUVendorType_GPU_UNKNOWN},
}

func TestGpuTypeFromVendorID(t *testing.T) {
	for _, cc := range deviceIdsTable {
		gpuType := gpuTypeFromVendorID(cc.id)
		assert.Equal(t, gpuType, cc.t, fmt.Sprintf("required %v, given %v",
			cc.t.String(), gpuType.String()))
	}
}

func TestNewDeviceWithGPUType(t *testing.T) {
	for _, cc := range deviceIdsTable {
		dev, err := NewDevice("name", "vendor", 1488, 8814, WithVendorId(uint(cc.id)))
		require.NoError(t, err)

		assert.Equal(t, dev.VendorType(), cc.t, fmt.Sprintf("required %v, given %v",
			cc.t.String(), dev.VendorType().String()))
	}
}
