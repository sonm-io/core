package mem

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/cgroups"
)

type Device struct {
	Used  uint64 `json:"used"`
	Total uint64 `json:"total"`
}

func getTotal(m *mem.VirtualMemoryStat, cg cgroups.CGroup) uint64 {
	hostMem := m.Total

	cgStat, err := cg.Stats()
	if err != nil {
		// cannot read mem limit from cGroup,
		// return amount of memory available for host system
		return hostMem
	}

	if hostMem > cgStat.MemoryLimit {
		return cgStat.MemoryLimit
	} else {
		return hostMem
	}
}

func NewMemoryDevice(cg cgroups.CGroup) (*Device, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	total := getTotal(m, cg)
	return &Device{Used: m.Used, Total: total}, err
}
