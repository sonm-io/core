package hardware

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Hardware struct {
	CPU    []cpu.InfoStat
	Memory *mem.VirtualMemoryStat
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

func New() HardwareInfo {
	return &hardwareInfo{}
}
