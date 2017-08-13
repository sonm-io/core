package hardware

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/sonm-io/core/insonmnia/hardware/gpu"

	pb "github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() pb.Hardware {
	return pb.Hardware{
		Cpu: CPUIntoProto(h.CPU),
		Mem: MemoryIntoProto(h.Memory),
		Gpu: GPUIntoProto(h.GPU),
	}
}

func CPUIntoProto(c []cpu.InfoStat) []*pb.CPUDevice {
	result := make([]*pb.CPUDevice, 0)

	for _, i := range c {
		device := &pb.CPUDevice{
			Name:   i.Model,
			Vendor: i.VendorID,
			Cores:  i.Cores,
			Mhz:    i.Mhz,
		}

		result = append(result, device)
	}

	return result
}

func MemoryIntoProto(m *mem.VirtualMemoryStat) *pb.RAMDevice {
	return &pb.RAMDevice{
		Total: m.Total,
	}
}

func GPUIntoProto(g []*gpu.Device) []*pb.GPUDevice {
	result := make([]*pb.GPUDevice, 0)

	for _, i := range g {
		device := &pb.GPUDevice{
			Name:   i.Name,
			Vendor: i.Vendor,
		}

		result = append(result, device)
	}

	return result
}
