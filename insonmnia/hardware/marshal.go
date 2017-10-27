package hardware

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	pb "github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *pb.Capabilities {
	return &pb.Capabilities{
		Cpu: CPUIntoProto(h.CPU),
		Mem: MemoryIntoProto(h.Memory),
		Gpu: GPUIntoProto(h.GPU),
	}
}

func CPUIntoProto(c []cpu.InfoStat) []*pb.CPUDevice {
	result := make([]*pb.CPUDevice, 0)

	for _, i := range c {
		ext := make(map[string]string)
		ext["family"] = i.Family
		ext["cache_size"] = strconv.Itoa(int(i.CacheSize))
		ext["flags"] = strings.Join(i.Flags, " ")

		device := &pb.CPUDevice{
			Name:   i.ModelName,
			Vendor: i.VendorID,
			Cores:  i.Cores,
			Mhz:    i.Mhz,
			Ext:    ext,
		}

		result = append(result, device)
	}

	return result
}

func CPUFromProto(c []*pb.CPUDevice) ([]cpu.InfoStat, error) {
	result := make([]cpu.InfoStat, 0)

	for _, i := range c {
		cacheSize, err := strconv.Atoi(i.Ext["cache_size"])
		if err != nil {
			cacheSize = 0
		}

		device := cpu.InfoStat{
			Model:     i.GetName(),
			VendorID:  i.GetVendor(),
			Cores:     i.Cores,
			Mhz:       i.Mhz,
			Family:    i.Ext["family"],
			CacheSize: int32(cacheSize),
			Flags:     strings.Split(i.Ext["flags"], " "),
		}

		result = append(result, device)
	}

	return result, nil
}

func MemoryIntoProto(m *mem.VirtualMemoryStat) *pb.RAMDevice {
	return &pb.RAMDevice{
		Total: m.Total,
		Used:  m.Used,
	}
}

func MemoryFromProto(m *pb.RAMDevice) (*mem.VirtualMemoryStat, error) {
	return &mem.VirtualMemoryStat{
		Total: m.Total,
	}, nil
}

func GPUIntoProto(g []gpu.Device) []*pb.GPUDevice {
	result := make([]*pb.GPUDevice, 0)

	for _, i := range g {
		device := &pb.GPUDevice{
			Name:          i.Name(),
			VendorName:    i.VendorName(),
			MaxMemorySize: i.MaxMemorySize(),
		}

		result = append(result, device)
	}

	return result
}

func GPUFromProto(g []*pb.GPUDevice) ([]gpu.Device, error) {
	result := []gpu.Device{}

	for _, i := range g {
		device, err := gpu.NewDevice(
			i.GetName(),
			i.GetVendorName(),
			i.GetMaxMemorySize(),
			gpu.WithVendorId(uint(i.GetVendorId())),
			gpu.WithOpenClDeviceVersionSpec(i.GetOpenCLVersionMajor(), i.GetOpenCLVersionMinor()),
		)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}

	return result, nil
}

func HardwareFromProto(cap *pb.Capabilities) (*Hardware, error) {
	c, err := CPUFromProto(cap.Cpu)
	if err != nil {
		return nil, err
	}

	m, err := MemoryFromProto(cap.Mem)
	if err != nil {
		return nil, err
	}

	g, err := GPUFromProto(cap.Gpu)
	if err != nil {
		return nil, err
	}

	h := &Hardware{
		CPU:    c,
		Memory: m,
		GPU:    g,
	}

	return h, nil
}
