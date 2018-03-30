package mem

import (
	"github.com/shirou/gopsutil/mem"
)

type Device struct {
	// Total is total mem present on the host system
	Total uint64 `json:"total"`
	// Available is available mem for tasks scheduling
	Available uint64 `json:"available"`
}

func NewMemoryDevice() (*Device, error) {
	m, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &Device{Total: m.Total}, err
}
