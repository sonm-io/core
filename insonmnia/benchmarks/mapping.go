package benchmarks

import (
	"context"

	sonm "github.com/sonm-io/core/proto"
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

	deviceTypes := map[uint64]sonm.DeviceType{}
	splittingAlgorithms := map[uint64]sonm.SplittingAlgorithm{}
	for _, benchmarks := range benchmarkList.MapByDeviceType() {
		for _, benchmark := range benchmarks {
			deviceTypes[benchmark.ID] = benchmark.Type
			splittingAlgorithms[benchmark.ID] = benchmark.SplittingAlgorithm
		}
	}

	return &mapping{
		deviceTypes:         deviceTypes,
		splittingAlgorithms: splittingAlgorithms,
	}, nil
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
