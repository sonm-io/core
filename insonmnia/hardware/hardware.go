package hardware

import "github.com/shirou/gopsutil/cpu"

type Hardware struct {
	CPU []cpu.InfoStat
}

type HardwareInfo interface {
	CPU() ([]cpu.InfoStat, error)
	//GPU()
	//RAM()

	Info() (*Hardware, error)
}

type hardwareInfo struct {
}

func (hardwareInfo) CPU() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

func (h *hardwareInfo) Info() (*Hardware, error) {
	cpuInfo, err := h.CPU()
	if err != nil {
		return nil, err
	}

	hardware := &Hardware{
		CPU: cpuInfo,
	}

	return hardware, nil
}

func New() HardwareInfo {
	return &hardwareInfo{}
}
