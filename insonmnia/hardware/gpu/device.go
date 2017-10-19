package gpu

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cnf/structhash"
	"github.com/sonm-io/core/proto"
)

var (
	errMalformedOpenCLVersion = errors.New("malformed OpenCL device version string")
)

// Device describes a GPU device.
type Device interface {
	// Name returns GPU model name.
	Name() string
	// VendorId returns an unique device vendor identifier. An example of a
	// unique device identifier could be the PCIe ID.
	VendorId() uint
	// VendorName returns GPU vendor name.
	VendorName() string
	// MaxMemorySize returns the total maximum memory size the device can hold
	// in bytes.
	MaxMemorySize() uint64
	// MaxClockFrequency returns maximum configured clock frequency of the
	// device in MHz.
	MaxClockFrequency() uint
	// OpenCLDeviceVersion returns the OpenCL major version supported by the
	// device.
	OpenCLDeviceVersionMajor() int
	// OpenCLDeviceVersion returns the OpenCL minor version supported by the
	// device.
	OpenCLDeviceVersionMinor() int

	Hash() []byte
}

type device struct {
	d sonm.GPUDevice
}

type Version struct {
	Major int
	Minor int
}

type Option func(*sonm.GPUDevice) error

func WithVendorId(id uint) func(*sonm.GPUDevice) error {
	return func(d *sonm.GPUDevice) error {
		d.VendorId = uint64(id)
		return nil
	}
}

func WithMaxClockFrequency(mhz uint) func(*sonm.GPUDevice) error {
	return func(d *sonm.GPUDevice) error {
		d.MaxClockFrequency = uint64(mhz)
		return nil
	}
}

// WithOpenClDeviceVersion option sets OpenCL version.
//
// The format must be: `OpenCL <major.minor> <vendor-specific information>`.
func WithOpenClDeviceVersion(version string) func(*sonm.GPUDevice) error {
	return func(d *sonm.GPUDevice) error {
		var vendor string
		n, err := fmt.Sscanf(version, "OpenCL %d.%d %s", &d.OpenCLDeviceVersionMajor, &d.OpenCLDeviceVersionMinor, &vendor)
		if n < 2 {
			return errMalformedOpenCLVersion
		}

		if n == 2 && err != nil {
			return nil
		}

		return nil
	}
}

func WithOpenClDeviceVersionSpec(major, minor int32) func(*sonm.GPUDevice) error {
	return func(d *sonm.GPUDevice) error {
		d.OpenCLDeviceVersionMajor = major
		d.OpenCLDeviceVersionMinor = minor
		return nil
	}
}

func NewDevice(name, vendorName string, maxMemorySize uint64, options ...Option) (Device, error) {
	d := sonm.GPUDevice{
		Name:          name,
		VendorName:    vendorName,
		MaxMemorySize: maxMemorySize,
	}

	for _, option := range options {
		if err := option(&d); err != nil {
			return nil, err
		}
	}

	return &device{d: d}, nil
}

func (d *device) Name() string {
	return d.d.GetName()
}

func (d *device) VendorId() uint {
	return uint(d.d.GetVendorId())
}

func (d *device) VendorName() string {
	return d.d.GetVendorName()
}

func (d *device) MaxMemorySize() uint64 {
	return d.d.GetMaxMemorySize()
}

func (d *device) MaxClockFrequency() uint {
	return uint(d.d.GetMaxClockFrequency())
}

func (d *device) OpenCLDeviceVersionMajor() int {
	return int(d.d.GetOpenCLDeviceVersionMajor())
}

func (d *device) OpenCLDeviceVersionMinor() int {
	return int(d.d.GetOpenCLDeviceVersionMinor())
}

func (d *device) Hash() []byte {
	return structhash.Md5(d.d, 1)
}

func (d *device) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":                     d.Name(),
		"vendorId":                 d.VendorId(),
		"vendorName":               d.VendorName(),
		"maxMemorySize":            d.MaxMemorySize(),
		"maxClockFrequency":        d.MaxClockFrequency(),
		"openCLDeviceVersionMajor": d.OpenCLDeviceVersionMajor(),
		"openCLDeviceVersionMinor": d.OpenCLDeviceVersionMinor(),
	})
}

// GetGPUDevices returns a list of available GPU devices on the machine.
func GetGPUDevices() ([]Device, error) {
	devices, err := GetGPUDevicesUsingOpenCL()
	if err != nil {
		return nil, err
	}

	return devices, nil
}
