package hardware

import (
	"github.com/sonm-io/core/proto"
)

func (h *Hardware) IntoProto() *sonm.DevicesReply {
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

	storage := &sonm.StorageDevice{
		Benchmarks: h.Storage.Benchmark,
	}

	return &sonm.DevicesReply{
		CPUs:    h.CPU.Device.Marshal(h.CPU.Benchmark),
		GPUs:    gpus,
		Memory:  ram,
		Network: net,
		Storage: storage,
	}
}
