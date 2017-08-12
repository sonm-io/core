package hardware

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Hardware accumulates the finest hardware information about system the miner
// is running on.
type Hardware struct {
	CPU    []cpu.InfoStat
	Memory *mem.VirtualMemoryStat
}

// LogicalCPUCount returns the number of logical CPUs in the system.
func (h *Hardware) LogicalCPUCount() int {
	res := 0
	for _, c := range h.CPU {
		res += int(c.Cores)
	}

	return res
}

type HardwareInfo interface {
	// CPU returns statistics about system CPU.
	//
	// This includes vendor name, model name, number of cores, cache info,
	// instruction flags and many others to be able to identify and to properly
	// account the CPU.
	CPU() ([]cpu.InfoStat, error)

	// Memory returns statistics about system memory.
	//
	// This includes total physical  memory, available memory and many others,
	// expressed in bytes.
	Memory() (*mem.VirtualMemoryStat, error)

	//GPU()

	// Info returns all described above hardware statistics.
	Info() (*Hardware, error)
}

type hardwareInfo struct {
}

func (*hardwareInfo) CPU() ([]cpu.InfoStat, error) {
	return cpu.Info()
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
	}

	return hardware, nil
}

// New constructs a new hardware info collector.
func New() HardwareInfo {
	return &hardwareInfo{}
}
