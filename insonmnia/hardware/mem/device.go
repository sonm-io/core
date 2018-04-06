package mem

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/proto"
)

func NewMemoryDevice() (*sonm.RAMDevice, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &sonm.RAMDevice{Total: m.Total}, err
}
