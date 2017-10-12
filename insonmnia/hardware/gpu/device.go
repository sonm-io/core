package gpu

import (
	"encoding/json"

	pb "github.com/sonm-io/core/proto"
)

// Device describes a GPU device.
type Device interface {
	// Name returns GPU model name.
	Name() string
	// VendorName returns GPU vendor name.
	VendorName() string
	// MaxMemorySize returns the total maximum memory size the device can hold
	// in bytes.
	MaxMemorySize() uint64
	// OpenCLDeviceVersion returns the OpenCL version supported by the device.
	OpenCLDeviceVersion() string
}

type device struct {
	d pb.GPUDevice
}

type Option func(*pb.GPUDevice) error

func WithOpenClDeviceVersion(version string) func(*pb.GPUDevice) error {
	return func(d *pb.GPUDevice) error {
		d.OpenCLVersion = version
		return nil
	}
}

func NewDevice(name, vendorName string, maxMemorySize uint64, options ...Option) (Device, error) {
	d := pb.GPUDevice{
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

func (d *device) VendorName() string {
	return d.d.GetVendorName()
}

func (d *device) MaxMemorySize() uint64 {
	return d.d.GetMaxMemorySize()
}

func (d *device) OpenCLDeviceVersion() string {
	return d.d.GetOpenCLVersion()
}

func (d *device) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":                d.Name(),
		"vendorName":          d.VendorName(),
		"maxMemorySize":       d.MaxMemorySize(),
		"openCLDeviceVersion": d.OpenCLDeviceVersion(),
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
