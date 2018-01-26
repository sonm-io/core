package hardware

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	pb "github.com/sonm-io/core/proto"
)

// Hardware accumulates the finest hardware information about system the miner
// is running on.
type Hardware struct {
	CPU    []cpu.Device
	Memory *mem.VirtualMemoryStat
	GPU    []gpu.Device
}

// LogicalCPUCount returns the number of logical CPUs in the system.
func (h *Hardware) LogicalCPUCount() int {
	count := 0
	for _, c := range h.CPU {
		count += int(c.Cores)
	}

	return count
}

// TotalMemory returns the total number of bytes.
func (h *Hardware) TotalMemory() uint64 {
	return h.Memory.Total
}

// HasGPU returns true if a system has GPU on the board.
func (h *Hardware) HasGPU() bool {
	return len(h.GPU) > 0
}

type HardwareInfo interface {
	// CPU returns information about system CPU.
	//
	// This includes vendor name, model name, number of cores, cache info,
	// instruction flags and many others to be able to identify and to properly
	// account the CPU.
	CPU() ([]cpu.Device, error)

	// Memory returns information about system memory.
	//
	// This includes total physical  memory, available memory and many others,
	// expressed in bytes.
	Memory() (*mem.VirtualMemoryStat, error)

	// GPU returns information about GPU devices on the machine.
	GPU() ([]gpu.Device, error)

	// Info returns all described above hardware statistics.
	Info(filters ...FilterFunc) (*Hardware, error)
}

type hardwareInfo struct{}

func (*hardwareInfo) CPU() ([]cpu.Device, error) {
	return cpu.GetCPUDevices()
}

func (h *hardwareInfo) Memory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (*hardwareInfo) GPU() ([]gpu.Device, error) {
	return gpu.GetGPUDevices()
}

type FilterFunc func(hw *Hardware) *Hardware

func FilterGPUByVendor(t pb.GPUVendorType) FilterFunc {
	return func(hw *Hardware) *Hardware {
		var filtered []gpu.Device
		for _, d := range hw.GPU {
			if t != pb.GPUVendorType_GPU_UNKNOWN && t == d.VendorType() {
				filtered = append(filtered, d)
			}
		}

		hw.GPU = filtered

		return hw
	}
}

func (h *hardwareInfo) Info(filters ...FilterFunc) (*Hardware, error) {
	cpuInfo, err := h.CPU()
	if err != nil {
		return nil, err
	}

	memory, err := h.Memory()
	if err != nil {
		return nil, err
	}

	gpuInfo, err := h.GPU()
	if err != nil {
		if err != gpu.ErrUnsupportedPlatform {
			return nil, err
		}

		gpuInfo = make([]gpu.Device, 0)
	}

	hardware := &Hardware{
		CPU:    cpuInfo,
		Memory: memory,
		GPU:    gpuInfo,
	}

	// apply filters
	for _, ff := range filters {
		hardware = ff(hardware)
	}

	return hardware, nil
}

// New constructs a new hardware info collector.
func New() HardwareInfo {
	return &hardwareInfo{}
}
