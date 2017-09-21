// +build cl

package gpu

// #cgo darwin LDFLAGS: -framework OpenCL
// #cgo linux LDFLAGS:-lOpenCL
// #ifdef __APPLE__
//     #include "TargetConditionals.h"
//     #if TARGET_OS_MAC
//         #include <OpenCL/cl.h>
//     #else
//         #error "Non OS X Apple targets are not supported, sorry"
//     #endif
// #elif __linux__
//     #include "CL/cl.h"
// #else
//     #include "cl.h"
// #endif
import "C"

import (
	"github.com/pkg/errors"
	"strings"
	"unsafe"
)

const (
	maxPlatforms   = 32
	maxDeviceCount = 64
)

// GetGPUDevicesUsingOpenCL returns a list of available GPU devices on the machine using OpenCL API.
func GetGPUDevicesUsingOpenCL() ([]*Device, error) {
	platforms, err := getPlatforms()
	if err != nil {
		return nil, err
	}

	var result []*Device

	for _, platform := range platforms {
		devices, err := platform.getGPUDevices()
		if err != nil {
			continue
		}

		for _, d := range devices {
			device := &Device{
				Name:              d.name(),
				Vendor:            d.vendor(),
				Flags:             d.extensions(),
				MaxClockFrequency: d.maxClockFrequency(),
				AddressBits:       d.addressBits(),
				CacheLineSize:     d.globalMemCacheLineSize(),
			}

			result = append(result, device)
		}
	}

	return result, nil
}

type platform struct {
	id C.cl_platform_id
}

func getPlatforms() ([]*platform, error) {
	var ids [maxPlatforms]C.cl_platform_id
	var num C.cl_uint

	if err := C.clGetPlatformIDs(C.cl_uint(maxPlatforms), &ids[0], &num); err != C.CL_SUCCESS {
		return nil, errors.Errorf("failed to obtain OpenCL platforms: %s", err)
	}

	platforms := make([]*platform, num)
	for i := 0; i < int(num); i++ {
		platforms[i] = &platform{id: ids[i]}
	}

	return platforms, nil
}

func (p *platform) getGPUDevices() ([]*device, error) {
	var ids [maxDeviceCount]C.cl_device_id
	var num C.cl_uint

	if num > maxDeviceCount {
		num = maxDeviceCount
	}

	if err := C.clGetDeviceIDs(p.id, C.cl_device_type(C.CL_DEVICE_TYPE_GPU), C.cl_uint(maxDeviceCount), &ids[0], &num); err != C.CL_SUCCESS {
		return nil, errors.Errorf("failed to obtain GPU devices for a platform: %s", err)
	}

	devices := make([]*device, num)
	for i := 0; i < int(num); i++ {
		devices[i] = &device{id: ids[i]}
	}

	return devices, nil
}

type device struct {
	id C.cl_device_id
}

func (d *device) getInfoString(param C.cl_device_info) (string, error) {
	var data [1024]C.char
	var size C.size_t

	if err := C.clGetDeviceInfo(d.id, param, 1024, unsafe.Pointer(&data), &size); err != C.CL_SUCCESS {
		return "", errors.Errorf("failed to convert device info into a string: %s", err)
	}

	return C.GoStringN((*C.char)(unsafe.Pointer(&data)), C.int(size)-1), nil
}

func (d *device) getInfoUint(param C.cl_device_info) (uint, error) {
	var val C.cl_uint

	if err := C.clGetDeviceInfo(d.id, param, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return 0, errors.Errorf("failed to convert device info into an integer: %s", err)
	}

	return uint(val), nil
}

func (d *device) name() string {
	result, _ := d.getInfoString(C.CL_DEVICE_NAME)
	return result
}

func (d *device) vendor() string {
	result, _ := d.getInfoString(C.CL_DEVICE_VENDOR)
	return result
}

func (d *device) extensions() []string {
	result, _ := d.getInfoString(C.CL_DEVICE_EXTENSIONS)
	return strings.Split(result, " ")
}

func (d *device) version() string {
	result, _ := d.getInfoString(C.CL_DEVICE_VERSION)
	return result
}

func (d *device) driverVersion() string {
	result, _ := d.getInfoString(C.CL_DRIVER_VERSION)
	return result
}

func (d *device) addressBits() int {
	val, _ := d.getInfoUint(C.CL_DEVICE_ADDRESS_BITS)
	return int(val)
}

func (d *device) globalMemCacheLineSize() int {
	val, _ := d.getInfoUint(C.CL_DEVICE_GLOBAL_MEM_CACHELINE_SIZE)
	return int(val)
}

func (d *device) maxClockFrequency() int {
	val, _ := d.getInfoUint(C.CL_DEVICE_MAX_CLOCK_FREQUENCY)
	return int(val)
}
