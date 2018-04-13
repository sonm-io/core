package optimus

import "github.com/sonm-io/core/proto"

// TODO: This is shit!
func newBenchmarksFromDevices(devices *sonm.DevicesReply) [sonm.MinNumBenchmarks]uint64 {
	benchmarks := [sonm.MinNumBenchmarks]uint64{}
	for id := uint64(0); id < sonm.MinNumBenchmarks; id++ {
		if v, ok := devices.CPU.GetBenchmarks()[id]; ok {
			benchmarks[id] = v.Result
		}
		for _, gpu := range devices.GPUs {
			if v, ok := gpu.Benchmarks[id]; ok {
				benchmarks[id] += v.Result
			}
		}
		if v, ok := devices.RAM.Benchmarks[id]; ok {
			benchmarks[id] = v.Result
		}
		if v, ok := devices.Network.BenchmarksIn[id]; ok {
			benchmarks[id] = v.Result
		}
		if v, ok := devices.Network.BenchmarksOut[id]; ok {
			benchmarks[id] = v.Result
		}
		if v, ok := devices.Storage.Benchmarks[id]; ok {
			benchmarks[id] = v.Result
		}
	}
	return benchmarks
}
