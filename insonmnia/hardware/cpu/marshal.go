package cpu

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/sonm-io/core/proto"
)

func Marshal(d []Device) []*sonm.CPUDevice {
	devices := make([]*sonm.CPUDevice, len(d))
	for _, device := range d {
		devices = append(devices, device.Marshal())
	}

	return devices
}

func (d *Device) Marshal() *sonm.CPUDevice {
	return &sonm.CPUDevice{
		Num:            d.CPU,
		VendorId:       d.VendorID,
		Model:          d.Model,
		ModelName:      d.ModelName,
		Cores:          d.Cores,
		ClockFrequency: d.Mhz,
		CacheSize:      d.CacheSize,
		Stepping:       d.Stepping,
		Flags:          d.Flags,
	}
}

func UnmarshalProto(d []*sonm.CPUDevice) ([]Device, error) {
	devices := make([]Device, len(d))
	for _, device := range d {
		dev, err := Unmarshal(device)
		if err != nil {
			return nil, err
		}
		devices = append(devices, dev)
	}

	return devices, nil
}

func Unmarshal(proto *sonm.CPUDevice) (Device, error) {
	info := cpu.InfoStat{
		CPU:       proto.GetNum(),
		VendorID:  proto.GetVendorId(),
		Model:     proto.GetModel(),
		ModelName: proto.GetModelName(),
		Cores:     proto.GetCores(),
		Mhz:       proto.GetClockFrequency(),
		CacheSize: proto.GetCacheSize(),
		Stepping:  proto.GetStepping(),
		Flags:     proto.GetFlags(),
	}

	return Device(info), nil
}
