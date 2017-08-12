// +build !darwin

package gpu

import "github.com/pkg/errors"

func GetGPUDevicesUsingOpenCL() ([]*Device, error) {
	return nil, errors.Errorf("the platform is not currently supported to expose GPU devices")
}
