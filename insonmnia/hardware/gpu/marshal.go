package gpu

import (
	"encoding/json"

	"github.com/sonm-io/core/proto"
)

func MarshalDevices(d []Device) []*sonm.GPUDevice {
	devices := make([]*sonm.GPUDevice, 0, len(d))
	for _, device := range d {
		devices = append(devices, Marshal(device))
	}

	return devices
}

func Marshal(d Device) *sonm.GPUDevice {
	return &sonm.GPUDevice{
		Name:                     d.Name(),
		VendorId:                 uint64(d.VendorId()),
		VendorName:               d.VendorName(),
		MaxMemorySize:            d.MaxMemorySize(),
		MaxClockFrequency:        uint64(d.MaxClockFrequency()),
		OpenCLDeviceVersionMajor: int32(d.OpenCLDeviceVersionMajor()),
		OpenCLDeviceVersionMinor: int32(d.OpenCLDeviceVersionMinor()),
	}
}

func UnmarshalDevices(d []*sonm.GPUDevice) ([]Device, error) {
	devices := make([]Device, 0, len(d))
	for _, device := range d {
		dev, err := Unmarshal(device)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}

	return devices, nil
}

func Unmarshal(proto *sonm.GPUDevice) (Device, error) {
	device, err := NewDevice(
		proto.GetName(),
		proto.GetVendorName(),
		proto.GetMaxClockFrequency(),
		proto.GetMaxMemorySize(),
		WithVendorId(uint(proto.GetVendorId())),
		WithOpenClDeviceVersionSpec(proto.GetOpenCLDeviceVersionMajor(), proto.GetOpenCLDeviceVersionMinor()),
	)
	if err != nil {
		return nil, err
	}

	return device, nil
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
