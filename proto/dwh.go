package sonm

import (
	"github.com/pkg/errors"
)

const (
	NumBenchmarks = 12
	NumNetflags   = 3
)

func NewBenchmarks(benchmarks []uint64) (*Benchmarks, error) {
	if len(benchmarks) < NumBenchmarks {
		return nil, errors.Errorf("expected %d benchmarks, got %d", NumBenchmarks, len(benchmarks))
	}

	return &Benchmarks{
		CPUSysbenchMulti: benchmarks[0],
		CPUSysbenchOne:   benchmarks[1],
		CPUCores:         benchmarks[2],
		RAMSize:          benchmarks[3],
		StorageSize:      benchmarks[4],
		NetTrafficIn:     benchmarks[5],
		NetTrafficOut:    benchmarks[6],
		GPUCount:         benchmarks[7],
		GPUMem:           benchmarks[8],
		GPUEthHashrate:   benchmarks[9],
		GPUCashHashrate:  benchmarks[10],
		GPURedshift:      benchmarks[11],
	}, nil
}

func (m *Benchmarks) ToArray() []uint64 {
	return []uint64{
		m.CPUSysbenchMulti,
		m.CPUSysbenchOne,
		m.CPUCores,
		m.RAMSize,
		m.StorageSize,
		m.NetTrafficIn,
		m.NetTrafficOut,
		m.GPUCount,
		m.GPUMem,
		m.GPUEthHashrate,
		m.GPUCashHashrate,
		m.GPURedshift,
	}
}

func UintToNetflags(flags uint64) [NumNetflags]bool {
	var fixedNetflags [3]bool
	for idx := 0; idx < NumNetflags; idx++ {
		fixedNetflags[NumNetflags-1-idx] = flags&(1<<uint64(idx)) != 0
	}

	return fixedNetflags
}

func NetflagsToUint(flags [NumNetflags]bool) uint64 {
	var netflags uint64
	for idx, flag := range flags {
		if flag {
			netflags |= 1 << uint64(NumNetflags-1-idx)
		}
	}

	return netflags
}
