package hardware

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.Capabilities {
	return &sonm.Capabilities{
		Cpu: CPUIntoProto(h.CPU),
		Mem: MemoryIntoProto(h.Memory),
		Gpu: GPUIntoProto(h.GPU),
	}
}

func CPUIntoProto(c []cpu.InfoStat) []*sonm.CPUDevice {
	result := make([]*sonm.CPUDevice, 0)

	for _, i := range c {
		ext := make(map[string]string)
		ext["family"] = i.Family
		ext["cache_size"] = strconv.Itoa(int(i.CacheSize))
		ext["flags"] = strings.Join(i.Flags, " ")

		device := &sonm.CPUDevice{
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

func CPUFromProto(c []*sonm.CPUDevice) ([]cpu.InfoStat, error) {
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

func GPUIntoProto(devices []gpu.Device) []*sonm.GPUDevice {
	result := make([]*sonm.GPUDevice, 0)

	for _, device := range devices {
		dump, err := json.Marshal(device)
		if err != nil {
			continue
		}
		proto := &sonm.GPUDevice{}
		if err := jsonpb.Unmarshal(bytes.NewReader(dump), proto); err != nil {
			continue
		}

		result = append(result, proto)
	}

	return result
}

func GPUFromProto(g []*sonm.GPUDevice) ([]gpu.Device, error) {
	result := []gpu.Device{}

	for _, i := range g {
		device, err := gpu.NewDevice(
			i.GetName(),
			i.GetVendorName(),
			i.GetMaxClockFrequency(),
			i.GetMaxMemorySize(),
			gpu.WithVendorId(uint(i.GetVendorId())),
			gpu.WithOpenClDeviceVersionSpec(i.GetOpenCLDeviceVersionMajor(), i.GetOpenCLDeviceVersionMinor()),
		)
		if err != nil {
			return nil, err
		}
		result = append(result, device)
	}

	return result, nil
}

func HardwareFromProto(cap *sonm.Capabilities) (*Hardware, error) {
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
