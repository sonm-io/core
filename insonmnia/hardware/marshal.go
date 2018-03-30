package hardware

import (
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/mem"
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.Capabilities {
	cpus := []cpu.Device{}
	for _, c := range h.CPU {
		cpus = append(cpus, c.Device)
	}

	gpus := []*sonm.GPUDevice{}
	for _, g := range h.GPU {
		gpus = append(gpus, g.Device)
	}

	return &sonm.Capabilities{
		Cpu: cpu.MarshalDevices(cpus),
		Mem: MemoryIntoProto(h.Memory.Device),
		Gpu: gpus,
	}
}

func MemoryIntoProto(m *mem.Device) *sonm.RAMDevice {
	return &sonm.RAMDevice{
		Total: m.Total,
	}
}
