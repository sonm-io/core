package hardware

import (
	"testing"

	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestHardware(t *testing.T) *Hardware {
	x, err := NewHardware()
	require.NoError(t, err)

	x.RAM.Device.Available = 1024
	x.CPU.Device = &sonm.CPUDevice{
		ModelName: "Intel", Cores: 2, Sockets: 1,
	}
	x.Network.In = 100
	x.Network.Out = 200

	x.Storage.Device = &sonm.StorageDevice{
		BytesAvailable: 100500,
	}
	x.GPU = append(x.GPU, &sonm.GPU{
		Device:     &sonm.GPUDevice{ID: "1234", Memory: 123546},
		Benchmarks: make(map[uint64]*sonm.Benchmark),
	})
	return x
}

func TestHardwareHash(t *testing.T) {
	x := getTestHardware(t)
	hash1 := x.Hash()
	assert.NotEmpty(t, hash1)

	bench := &sonm.Benchmark{ID: 666, Result: 1337}
	x.CPU.Benchmarks[111] = bench
	x.Network.BenchmarksIn[222] = bench
	x.Network.BenchmarksOut[234] = bench
	x.Storage.Benchmarks[333] = bench
	x.RAM.Benchmarks[444] = bench
	x.GPU[0].Benchmarks[123] = bench

	hash2 := x.Hash()
	assert.NotEmpty(t, hash2)
	assert.Equal(t, hash1, hash2)
}

func TestHardwareLimitTo(t *testing.T) {
	hardware := getTestHardware(t)
	hardware.CPU.Benchmarks[0] = &sonm.Benchmark{
		ID:                 0,
		SplittingAlgorithm: sonm.SplittingAlgorithm_MAX,
		Result:             100,
	}
	hardware.CPU.Benchmarks[1] = &sonm.Benchmark{
		ID:                 1,
		SplittingAlgorithm: sonm.SplittingAlgorithm_MIN,
		Result:             100,
	}
	hardware.CPU.Benchmarks[2] = &sonm.Benchmark{
		ID:                 2,
		SplittingAlgorithm: sonm.SplittingAlgorithm_NONE,
		Result:             100,
	}
	hardware.CPU.Benchmarks[3] = &sonm.Benchmark{
		ID:                 3,
		SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL,
		Result:             100,
	}

	resources := &sonm.AskPlanResources{
		CPU: &sonm.AskPlanCPU{CorePercents: 150},
	}
	limitedHardware, err := hardware.LimitTo(resources)
	require.NoError(t, err)
	require.Equal(t, limitedHardware.CPU.Benchmarks[0].Result, uint64(100))
	require.Equal(t, limitedHardware.CPU.Benchmarks[1].Result, uint64(100))
	require.Equal(t, limitedHardware.CPU.Benchmarks[2].Result, uint64(100))
	//150 core percents out of total 200
	require.Equal(t, limitedHardware.CPU.Benchmarks[3].Result, uint64(75))
}

// Dev-770
func TestHardware_ResourcesToBenchmarks(t *testing.T) {
	hardware := getTestHardware(t)
	hardware.RAM.Device.Available = 8325287936
	hardware.RAM.Benchmarks[benchmarks.RamSize] = &sonm.Benchmark{
		ID:                 benchmarks.RamSize,
		Result:             hardware.RAM.Device.Available,
		SplittingAlgorithm: sonm.SplittingAlgorithm_PROPORTIONAL,
		Type:               sonm.DeviceType_DEV_RAM,
	}
	resources := &sonm.AskPlanResources{
		RAM: &sonm.AskPlanRAM{
			Size: &sonm.DataSize{
				Bytes: 4194304,
			},
		},
	}
	benches, err := hardware.ResourcesToBenchmarks(resources)
	require.NoError(t, err)
	require.Equal(t, uint64(4194304), benches.GetValues()[benchmarks.RamSize])
}
