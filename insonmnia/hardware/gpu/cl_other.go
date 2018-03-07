// +build !cl

package gpu

import (
	"github.com/sonm-io/core/proto"
)

func GetGPUDevicesUsingOpenCL() ([]*sonm.GPUDevice, error) {
	return nil, ErrUnsupportedPlatform
}
