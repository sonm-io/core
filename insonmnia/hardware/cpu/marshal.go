package cpu

import (
	"github.com/sonm-io/core/proto"
)

func (d *Device) Marshal(benchmarks map[uint64]*sonm.Benchmark) *sonm.CPUDevice {
	return &sonm.CPUDevice{
		ModelName:  d.Name,
		Cores:      uint32(d.Cores),
		Sockets:    uint32(d.Sockets),
		Benchmarks: benchmarks,
	}
}
