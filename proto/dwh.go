package sonm

import (
	"github.com/pkg/errors"
)

const (
	NumBenchmarks = 12
)

func NewBenchmarks(benchmarks []uint64) (*DWHBenchmarks, error) {
	if len(benchmarks) < NumBenchmarks {
		return nil, errors.Errorf("expected %d benchmarks, got %d", NumBenchmarks, len(benchmarks))
	}

	return &DWHBenchmarks{
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

func (m *DWHBenchmarks) ToArray() [NumBenchmarks]uint64 {
	return [NumBenchmarks]uint64{
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
