package gpu

import (
	"errors"
	"strings"

	"github.com/sonm-io/core/proto"
)

// hasGPUWithVendor uses OpenCL to check device existence on the Worker's system
func hasGPUWithVendor(v sonm.GPUVendorType, devices []*sonm.GPUDevice) error {
	found := false
	for _, dev := range devices {
		if dev.VendorType() == v {
			found = true
		}
	}

	if !found {
		return errors.New("cannot detect required GPU")
	}

	return nil
}

func GetVendorByName(vendor string) (sonm.GPUVendorType, error) {
	vendorName := strings.ToUpper(vendor)
	t, ok := sonm.GPUVendorType_value[vendorName]
	if !ok {
		return sonm.GPUVendorType_GPU_UNKNOWN, errors.New("unknown GPU vendor type")
	}

	return sonm.GPUVendorType(t), nil
}
