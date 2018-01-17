// +build !cl

package gpu

func GetGPUDevicesUsingOpenCL() ([]Device, error) {
	return nil, ErrUnsupportedPlatform
}
