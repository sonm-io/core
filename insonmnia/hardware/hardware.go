package hardware

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	pb "github.com/sonm-io/core/proto"
)

// Hardware accumulates the finest hardware information about system the miner
// is running on.
type Hardware struct {
	CPU    []cpu.Device
	Memory *mem.VirtualMemoryStat
	GPU    []*pb.GPUDevice
}

// LogicalCPUCount returns the number of logical CPUs in the system.
func (h *Hardware) LogicalCPUCount() int {
	count := 0
	for _, c := range h.CPU {
		count += int(c.Cores)
	}

	return count
}

type Info interface {
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
	// GPU() ([]*pb.GPUDevice, error)

	// Info returns all described above hardware statistics.
	Info() (*Hardware, error)
}

type hardwareInfo struct{}

func (*hardwareInfo) CPU() ([]cpu.Device, error) {
	return cpu.GetCPUDevices()
}

func (h *hardwareInfo) Memory() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (h *hardwareInfo) Info() (*Hardware, error) {
	cpuInfo, err := h.CPU()
	if err != nil {
		return nil, err
	}

	memory, err := h.Memory()
	if err != nil {
		return nil, err
	}

	hardware := &Hardware{
		CPU:    cpuInfo,
		Memory: memory,
		GPU:    nil,
	}

	return hardware, nil
}

// New constructs a new hardware info collector.
func New() Info {
	return &hardwareInfo{}
}
