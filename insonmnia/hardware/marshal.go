package hardware

import (
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.DevicesReply {
	var cpus []*sonm.CPUDevice
	for _, c := range h.CPU {
		dev := c.Device.Marshal()
		dev.Benchmarks = c.Benchmark
		cpus = append(cpus, c.Device.Marshal())
	}

	var gpus []*sonm.GPUDevice
	for _, g := range h.GPU {
		gpus = append(gpus, g.Device)
	}

	ram := &sonm.RAMDevice{
		Total:      h.Memory.Device.Available,
		Benchmarks: h.Memory.Benchmark,
	}

	net := &sonm.NetworkDevice{
		Benchmarks: h.Network.Benchmark,
	}

	stor := &sonm.StorageDevice{
		Benchmarks: h.Storage.Benchmark,
	}

	return &sonm.DevicesReply{
		CPUs:    cpus,
		GPUs:    gpus,
		Memory:  ram,
		Network: net,
		Storage: stor,
	}
}
