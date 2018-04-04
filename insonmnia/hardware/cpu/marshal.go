package cpu

import (
	"github.com/sonm-io/core/proto"
)

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
