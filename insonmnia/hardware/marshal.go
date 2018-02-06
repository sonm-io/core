package hardware

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.Capabilities {
	return &sonm.Capabilities{
		Cpu: cpu.MarshalDevices(h.CPU),
		Mem: MemoryIntoProto(h.Memory),
		Gpu: h.GPU,
	}
}

func MemoryIntoProto(m *mem.VirtualMemoryStat) *sonm.RAMDevice {
	return &sonm.RAMDevice{
		Total: m.Total,
		Used:  m.Used,
	}
}

func MemoryFromProto(m *sonm.RAMDevice) (*mem.VirtualMemoryStat, error) {
	return &mem.VirtualMemoryStat{
		Total: m.Total,
	}, nil
}

func HardwareFromProto(cap *sonm.Capabilities) (*Hardware, error) {
	c, err := cpu.UnmarshalDevices(cap.Cpu)
	if err != nil {
		return nil, err
	}

	m, err := MemoryFromProto(cap.Mem)
	if err != nil {
		return nil, err
	}

	h := &Hardware{
		CPU:    c,
		Memory: m,
		GPU:    cap.Gpu,
	}

	return h, nil
}
