package hardware

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHardwareHash(t *testing.T) {
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
