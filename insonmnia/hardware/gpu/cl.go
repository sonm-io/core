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
	"fmt"
	"unsafe"
)

const (
	maxPlatforms   = 32
	maxDeviceCount = 64
)

// GetGPUDevicesUsingOpenCL returns a list of available GPU devices on the machine using OpenCL API.
func GetGPUDevicesUsingOpenCL() ([]Device, error) {
	platforms, err := getPlatforms()
	if err != nil {
		return nil, err
	}

	var result []Device

	for _, platform := range platforms {
		devices, err := platform.getGPUDevices()
		if err != nil {
			return nil, err
		}

		for _, d := range devices {
			options := []Option{}
			name, err := d.name()
			if err != nil {
				return nil, err
			}
			vendor, err := d.vendor()
			if err != nil {
				return nil, err
			}
			maxClockFrequency, err := d.deviceMaxClockFrequency()
			if err != nil {
				return nil, err
			}
			globalMemSize, err := d.globalMemSize()
			if err != nil {
				return nil, err
			}
			if vendorId, err := d.vendorId(); err == nil {
				options = append(options, WithVendorId(vendorId))
			}
			if deviceVersion, err := d.deviceVersion(); err == nil {
				options = append(options, WithOpenClDeviceVersion(deviceVersion))
			}

			device, err := NewDevice(name, vendor, uint64(maxClockFrequency), globalMemSize, options...)
			if err != nil {
				return nil, err
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
		return nil, fmt.Errorf("failed to obtain OpenCL platforms: %s", err)
	}

	platforms := make([]*platform, num)
	for i := 0; i < int(num); i++ {
		platforms[i] = &platform{id: ids[i]}
	}

	return platforms, nil
}

func (p *platform) getGPUDevices() ([]*clDevice, error) {
	var ids [maxDeviceCount]C.cl_device_id
	var num C.cl_uint

	if num > maxDeviceCount {
		num = maxDeviceCount
	}

	if err := C.clGetDeviceIDs(p.id, C.cl_device_type(C.CL_DEVICE_TYPE_GPU), C.cl_uint(maxDeviceCount), &ids[0], &num); err != C.CL_SUCCESS {
		return nil, fmt.Errorf("failed to obtain GPU devices for a platform: %s", err)
	}

	devices := make([]*clDevice, num)
	for i := 0; i < int(num); i++ {
		devices[i] = &clDevice{id: ids[i]}
	}

	return devices, nil
}

type clDevice struct {
	id C.cl_device_id
}

func (d *clDevice) getInfoString(param C.cl_device_info) (string, error) {
	var data [1024]C.char
	var size C.size_t

	if err := C.clGetDeviceInfo(d.id, param, 1024, unsafe.Pointer(&data), &size); err != C.CL_SUCCESS {
		return "", fmt.Errorf("failed to convert device info into a string: %s", err)
	}

	return C.GoStringN((*C.char)(unsafe.Pointer(&data)), C.int(size)-1), nil
}

func (d *clDevice) getInfoUint(param C.cl_device_info) (uint, error) {
	var val C.cl_uint

	if err := C.clGetDeviceInfo(d.id, param, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return 0, fmt.Errorf("failed to convert device info into an integer: %s", err)
	}

	return uint(val), nil
}

func (d *clDevice) getInfoUint64(param C.cl_device_info) (uint64, error) {
	var val C.cl_ulong

	if err := C.clGetDeviceInfo(d.id, param, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return 0, fmt.Errorf("failed to convert device info into an integer: %s", err)
	}

	return uint64(val), nil
}

func (d *clDevice) name() (string, error) {
	return d.getInfoString(C.CL_DEVICE_NAME)
}

func (d *clDevice) vendor() (string, error) {
	return d.getInfoString(C.CL_DEVICE_VENDOR)
}

func (d *clDevice) vendorId() (uint, error) {
	return d.getInfoUint(C.CL_DEVICE_VENDOR_ID)
}

func (d *clDevice) globalMemSize() (uint64, error) {
	return d.getInfoUint64(C.CL_DEVICE_GLOBAL_MEM_SIZE)
}

func (d *clDevice) driverVersion() (string, error) {
	return d.getInfoString(C.CL_DRIVER_VERSION)
}

func (d *clDevice) deviceVersion() (string, error) {
	return d.getInfoString(C.CL_DEVICE_VERSION)
}

func (d *clDevice) deviceMaxClockFrequency() (uint, error) {
	return d.getInfoUint(C.CL_DEVICE_MAX_CLOCK_FREQUENCY)
}

func (d *clDevice) deviceMaxComputeUnits() (uint, error) {
	return d.getInfoUint(C.CL_DEVICE_MAX_COMPUTE_UNITS)
}
