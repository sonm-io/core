package hardware

import (
	"testing"

	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	dev1, _ := gpu.NewDevice("", "", 1, 1, gpu.WithVendorId(4098))
	dev2, _ := gpu.NewDevice("", "", 1, 1, gpu.WithVendorId(4098))
	dev3, _ := gpu.NewDevice("", "", 1, 1, gpu.WithVendorId(1234))
	dev4, _ := gpu.NewDevice("", "", 1, 1, gpu.WithVendorId(4318))

	devs := []gpu.Device{dev1, dev2, dev3, dev4}

	assert.Len(t, FilterGPUByVendor(sonm.GPUVendorType_RADEON)(&Hardware{GPU: devs}).GPU, 2)
	assert.Len(t, FilterGPUByVendor(sonm.GPUVendorType_NVIDIA)(&Hardware{GPU: devs}).GPU, 1)
	assert.Len(t, FilterGPUByVendor(sonm.GPUVendorType_GPU_UNKNOWN)(&Hardware{GPU: devs}).GPU, 0)
}
