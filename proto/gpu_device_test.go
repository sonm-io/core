package sonm

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

var deviceIdsTable = []struct {
	id uint64
	t  GPUVendorType
}{
	{id: 4098, t: GPUVendorType_RADEON},
	{id: 16915456, t: GPUVendorType_RADEON},
	{id: 4318, t: GPUVendorType_NVIDIA},
	{id: 16925952, t: GPUVendorType_GPU_UNKNOWN},
}

func TestGpuTypeFromVendorID(t *testing.T) {
	for _, cc := range deviceIdsTable {
		gpuType := TypeFromVendorID(cc.id)
		assert.Equal(t, gpuType, cc.t, fmt.Sprintf("required %v, given %v",
			cc.t.String(), gpuType.String()))
	}
}
