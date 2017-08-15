package hardware

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/sonm-io/core/insonmnia/hardware/gpu"

	pb "github.com/sonm-io/core/proto"
	"strconv"
	"strings"
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
			Name:   i.Model,
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
	}
}

func MemoryFromProto(m *pb.RAMDevice) (*mem.VirtualMemoryStat, error) {
	return &mem.VirtualMemoryStat{
		Total: m.Total,
	}, nil
}

func GPUIntoProto(g []*gpu.Device) []*pb.GPUDevice {
	result := make([]*pb.GPUDevice, 0)

	for _, i := range g {
		ext := make(map[string]string)
		ext["flags"] = strings.Join(i.Flags, ",")
		ext["max_clock_frequency"] = strconv.Itoa(int(i.MaxClockFrequency))
		ext["address_bits"] = strconv.Itoa(int(i.AddressBits))
		ext["cache_line_size"] = strconv.Itoa(int(i.CacheLineSize))

		device := &pb.GPUDevice{
			Name:   i.Name,
			Vendor: i.Vendor,
			Ext:    ext,
		}

		result = append(result, device)
	}

	return result
}

func GPUFromProto(g []*pb.GPUDevice) ([]*gpu.Device, error) {
	result := make([]*gpu.Device, 0)

	for _, i := range g {
		mhz, err := strconv.Atoi(i.Ext["max_clock_frequency"])
		if err != nil {
			mhz = 0
		}

		addressBits, err := strconv.Atoi(i.Ext["address_bits"])
		if err != nil {
			addressBits = 0.0
		}

		cacheSize, err := strconv.Atoi(i.Ext["cache_line_size"])
		if err != nil {
			cacheSize = 0
		}

		device := &gpu.Device{
			Name:              i.GetName(),
			Vendor:            i.GetVendor(),
			Flags:             strings.Split(i.Ext["flags"], " "),
			MaxClockFrequency: mhz,
			AddressBits:       addressBits,
			CacheLineSize:     cacheSize,
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
