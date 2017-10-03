// +build !cl

package gpu

import "errors"

var ErrUnsupportedPlatform = errors.New("the platform is not currently supported to expose GPU devices")

func GetGPUDevicesUsingOpenCL() ([]Device, error) {
	return nil, ErrUnsupportedPlatform
}
