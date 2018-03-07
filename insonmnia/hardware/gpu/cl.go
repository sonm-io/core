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
	"strconv"
	"unsafe"

	"github.com/sonm-io/core/proto"
)

const (
	maxPlatforms              = 32
	maxDeviceCount            = 64
	CL_PLATFORM_NOT_FOUND_KHR = C.cl_int(-1001)
)

// GetGPUDevicesUsingOpenCL returns a list of available GPU devices on the machine using OpenCL API.
func GetGPUDevicesUsingOpenCL() ([]*sonm.GPUDevice, error) {
	platforms, err := getPlatforms()
	if err != nil {
		return nil, err
	}

	var result []*sonm.GPUDevice

	for _, platform := range platforms {
		devices, err := platform.getGPUDevices()
		if err != nil {
			continue
		}

		for _, d := range devices {
			name, err := d.deviceName()
			if err != nil {
				return nil, err
			}

			vendorName, err := d.vendorName()
			if err != nil {
				return nil, err
			}

			vendorId, err := d.vendorID()
			if err != nil {
				return nil, err
			}

			result = append(result, &sonm.GPUDevice{
				DeviceName: name,
				VendorID:   uint64(vendorId),
				VendorName: vendorName,
			})
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
		if err == CL_PLATFORM_NOT_FOUND_KHR {
			return []*platform{}, nil
		}

		return nil, fmt.Errorf("failed to obtain OpenCL platforms: %s", errorToString(err))
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
		return nil, fmt.Errorf("failed to obtain GPU devices for a platform: %s", errorToString(err))
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
		return "", fmt.Errorf("failed to convert device info into a string: %s", errorToString(err))
	}

	return C.GoStringN((*C.char)(unsafe.Pointer(&data)), C.int(size)-1), nil
}

func (d *clDevice) getInfoUint(param C.cl_device_info) (uint, error) {
	var val C.cl_uint

	if err := C.clGetDeviceInfo(d.id, param, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return 0, fmt.Errorf("failed to convert device info into an integer: %s", errorToString(err))
	}

	return uint(val), nil
}

func (d *clDevice) getInfoUint64(param C.cl_device_info) (uint64, error) {
	var val C.cl_ulong

	if err := C.clGetDeviceInfo(d.id, param, C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), nil); err != C.CL_SUCCESS {
		return 0, fmt.Errorf("failed to convert device info into an integer: %s", errorToString(err))
	}

	return uint64(val), nil
}

func (d *clDevice) deviceName() (string, error) {
	return d.getInfoString(C.CL_DEVICE_NAME)
}

func (d *clDevice) vendorName() (string, error) {
	return d.getInfoString(C.CL_DEVICE_VENDOR)
}

func (d *clDevice) vendorID() (uint, error) {
	return d.getInfoUint(C.CL_DEVICE_VENDOR_ID)
}

func errorToString(err C.cl_int) string {
	switch err {
	case C.CL_SUCCESS:
		return "success"
	case C.CL_DEVICE_NOT_FOUND:
		return "no OpenCL devices that matched device_type were found"
	case C.CL_DEVICE_NOT_AVAILABLE:
		return "device is currently not available"
	case C.CL_COMPILER_NOT_AVAILABLE:
		return "compiler not available"
	case C.CL_MEM_OBJECT_ALLOCATION_FAILURE:
		return "memory object allocation failure"
	case C.CL_OUT_OF_RESOURCES:
		return "out of resources"
	case C.CL_OUT_OF_HOST_MEMORY:
		return "out of host memory"
	case C.CL_PROFILING_INFO_NOT_AVAILABLE:
		return "profiling information not available"
	case C.CL_MEM_COPY_OVERLAP:
		return "memory copy overlap"
	case C.CL_IMAGE_FORMAT_MISMATCH:
		return "image format mismatch"
	case C.CL_IMAGE_FORMAT_NOT_SUPPORTED:
		return "image format not supported"
	case C.CL_BUILD_PROGRAM_FAILURE:
		return "program build failure"
	case C.CL_MAP_FAILURE:
		return "map failure"
	case C.CL_INVALID_VALUE:
		return "invalid value"
	case C.CL_INVALID_DEVICE_TYPE:
		return "invalid device type"
	case C.CL_INVALID_PLATFORM:
		return "invalid platform"
	case C.CL_INVALID_DEVICE:
		return "invalid device"
	case C.CL_INVALID_CONTEXT:
		return "invalid context"
	case C.CL_INVALID_QUEUE_PROPERTIES:
		return "invalid queue properties"
	case C.CL_INVALID_COMMAND_QUEUE:
		return "invalid command queue"
	case C.CL_INVALID_HOST_PTR:
		return "invalid host pointer"
	case C.CL_INVALID_MEM_OBJECT:
		return "invalid memory object"
	case C.CL_INVALID_IMAGE_FORMAT_DESCRIPTOR:
		return "invalid image format descriptor"
	case C.CL_INVALID_IMAGE_SIZE:
		return "invalid image size"
	case C.CL_INVALID_SAMPLER:
		return "invalid sampler"
	case C.CL_INVALID_BINARY:
		return "invalid binary"
	case C.CL_INVALID_BUILD_OPTIONS:
		return "invalid build options"
	case C.CL_INVALID_PROGRAM:
		return "invalid program"
	case C.CL_INVALID_PROGRAM_EXECUTABLE:
		return "invalid program executable"
	case C.CL_INVALID_KERNEL_NAME:
		return "invalid kernel name"
	case C.CL_INVALID_KERNEL_DEFINITION:
		return "invalid kernel definition"
	case C.CL_INVALID_KERNEL:
		return "invalid kernel"
	case C.CL_INVALID_ARG_INDEX:
		return "invalid argument index"
	case C.CL_INVALID_ARG_VALUE:
		return "invalid argument value"
	case C.CL_INVALID_ARG_SIZE:
		return "invalid argument size"
	case C.CL_INVALID_KERNEL_ARGS:
		return "invalid kernel arguments"
	case C.CL_INVALID_WORK_DIMENSION:
		return "invalid work dimension"
	case C.CL_INVALID_WORK_GROUP_SIZE:
		return "invalid work group size"
	case C.CL_INVALID_WORK_ITEM_SIZE:
		return "invalid work item size"
	case C.CL_INVALID_GLOBAL_OFFSET:
		return "invalid global offset"
	case C.CL_INVALID_EVENT_WAIT_LIST:
		return "invalid event wait list"
	case C.CL_INVALID_EVENT:
		return "invalid event"
	case C.CL_INVALID_OPERATION:
		return "invalid operation"
	case C.CL_INVALID_GL_OBJECT:
		return "invalid OpenGL object"
	case C.CL_INVALID_BUFFER_SIZE:
		return "invalid buffer size"
	case C.CL_INVALID_MIP_LEVEL:
		return "invalid mip-map level"
	case CL_PLATFORM_NOT_FOUND_KHR:
		return "no valid ICDs found"
	default:
		return "unknown OpenCL error: " + strconv.FormatInt(int64(err), 10)
	}
}
