package benchmarks

import (
	"context"

	sonm "github.com/sonm-io/core/proto"
)

const (
	arrayMappingThreshold = 1024
)

type Loader interface {
	Load(ctx context.Context) (Mapping, error)
}

type loader struct {
	cfg Config
}

func NewLoader(uri string) Loader {
	return &loader{
		cfg: Config{URL: uri},
	}
}

func (m *loader) Load(ctx context.Context) (Mapping, error) {
	benchmarkList, err := NewBenchmarksList(ctx, m.cfg)
	if err != nil {
		return nil, err
	}

	maxID := benchmarkList.Max()
	if maxID <= arrayMappingThreshold {
		return NewArrayMapping(benchmarkList, maxID), nil
	}

	return NewMapMapping(benchmarkList), nil
}

func NewArrayMapping(benchmarks BenchList, maxID uint64) Mapping {
	deviceTypes := make([]sonm.DeviceType, maxID+1)
	splittingAlgorithms := make([]sonm.SplittingAlgorithm, maxID+1)

	for _, benchmark := range benchmarks.ByID() {
		deviceTypes[benchmark.ID] = benchmark.Type
		splittingAlgorithms[benchmark.ID] = benchmark.SplittingAlgorithm
	}
	return &arrayMapping{
		deviceTypes:         deviceTypes,
		splittingAlgorithms: splittingAlgorithms,
	}
}

func NewMapMapping(benchmarks BenchList) Mapping {
	deviceTypes := map[uint64]sonm.DeviceType{}
	splittingAlgorithms := map[uint64]sonm.SplittingAlgorithm{}
	for _, benchmarks := range benchmarks.MapByDeviceType() {
		for _, benchmark := range benchmarks {
			deviceTypes[benchmark.ID] = benchmark.Type
			splittingAlgorithms[benchmark.ID] = benchmark.SplittingAlgorithm
		}
	}

	return &mapping{
		deviceTypes:         deviceTypes,
		splittingAlgorithms: splittingAlgorithms,
	}
}

type Mapping interface {
	DeviceType(id int) sonm.DeviceType
	SplittingAlgorithm(id int) sonm.SplittingAlgorithm
}

type mapping struct {
	deviceTypes         map[uint64]sonm.DeviceType
	splittingAlgorithms map[uint64]sonm.SplittingAlgorithm
}

func (m *mapping) DeviceType(id int) sonm.DeviceType {
	ty, ok := m.deviceTypes[uint64(id)]
	if !ok {
		return sonm.DeviceType_DEV_UNKNOWN
	}
	return ty
}

func (m *mapping) SplittingAlgorithm(id int) sonm.SplittingAlgorithm {
	ty, ok := m.splittingAlgorithms[uint64(id)]
	if !ok {
		return sonm.SplittingAlgorithm_NONE
	}
	return ty
}

type arrayMapping struct {
	deviceTypes         []sonm.DeviceType
	splittingAlgorithms []sonm.SplittingAlgorithm
}

func (m *arrayMapping) DeviceType(id int) sonm.DeviceType {
	if id < 0 || id >= len(m.deviceTypes) {
		return sonm.DeviceType_DEV_UNKNOWN
	}

	return m.deviceTypes[id]
}

func (m *arrayMapping) SplittingAlgorithm(id int) sonm.SplittingAlgorithm {
	if id < 0 || id >= len(m.splittingAlgorithms) {
		return sonm.SplittingAlgorithm_NONE
	}

	return m.splittingAlgorithms[id]
}
