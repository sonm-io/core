package gpu

import (
	"errors"

	"github.com/sonm-io/core/proto"
)

var (
	ErrUnsupportedPlatform = errors.New("the platform is not currently supported to expose GPU devices")
)

// GetGPUDevices returns a list of available GPU devices on the machine.
func GetGPUDevices() ([]*sonm.GPUDevice, error) {
	devices, err := GetGPUDevicesUsingOpenCL()
	if err != nil {
		return nil, err
	}

	return devices, nil
}
