// The whole module is a piece of shit. Do we really need such kind of
// benchmark madness?

package optimus

import (
	"errors"
	"math"

	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
)

var (
	errExhausted = errors.New("no resources left")
)

type Consumer interface {
	LowerBound() []Rational
	DeviceType() sonm.DeviceType
	DeviceBenchmark(id int) (*sonm.Benchmark, bool)
	SplittingAlgorithm() sonm.SplittingAlgorithm
	Result(criteria Rational) interface{}
}

type consumer struct{}

func (m *consumer) SplittingAlgorithm() sonm.SplittingAlgorithm {
	return sonm.SplittingAlgorithm_PROPORTIONAL
}

type cpuConsumer struct {
	consumer
	cpu *sonm.CPU
}

func (m *cpuConsumer) LowerBound() []Rational {
	return []Rational{
		NewRational(sonm.MinCPUPercent, uint64(m.cpu.Device.Cores)).Div(100),
	}
}

func (m *cpuConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_CPU
}

func (m *cpuConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.cpu.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *cpuConsumer) Result(criteria Rational) interface{} {
	return &sonm.AskPlanCPU{CorePercents: criteria.Mul(100).Mul(uint64(m.cpu.Device.Cores)).Uint64Ceil()}
}

type ramConsumer struct {
	consumer
	ram *sonm.RAM
}

func (m *ramConsumer) LowerBound() []Rational {
	return []Rational{
		NewRational(sonm.MinRamSize, m.ram.Device.Total),
	}
}

func (m *ramConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_RAM
}

func (m *ramConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.ram.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *ramConsumer) Result(criteria Rational) interface{} {
	return &sonm.AskPlanRAM{Size: &sonm.DataSize{Bytes: criteria.Mul(m.ram.Device.Total).Uint64Ceil()}}
}

type storageConsumer struct {
	consumer
	dev *sonm.Storage
}

func (m *storageConsumer) LowerBound() []Rational {
	return []Rational{
		NewRational(sonm.MinStorageSize, m.dev.Device.BytesAvailable),
	}
}

func (m *storageConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_STORAGE
}

func (m *storageConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *storageConsumer) Result(criteria Rational) interface{} {
	return &sonm.AskPlanStorage{Size: &sonm.DataSize{Bytes: criteria.Mul(m.dev.Device.BytesAvailable).Uint64Ceil()}}
}

type networkInConsumer struct {
	consumer
	dev *sonm.Network
}

func (m *networkInConsumer) LowerBound() []Rational {
	return []Rational{}
}

func (m *networkInConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_NETWORK_IN
}

func (m *networkInConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.BenchmarksIn[uint64(id)]
	return benchmark, ok
}

func (m *networkInConsumer) Result(criteria Rational) interface{} {
	return &sonm.DataSizeRate{BitsPerSecond: criteria.Mul(m.dev.In).Uint64Ceil()}
}

type networkOutConsumer struct {
	consumer
	dev *sonm.Network
}

func (m *networkOutConsumer) LowerBound() []Rational {
	return []Rational{}
}

func (m *networkOutConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_NETWORK_OUT
}

func (m *networkOutConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.BenchmarksOut[uint64(id)]
	return benchmark, ok
}

func (m *networkOutConsumer) Result(criteria Rational) interface{} {
	return &sonm.DataSizeRate{BitsPerSecond: criteria.Mul(m.dev.Out).Uint64Ceil()}
}

type DeviceManager struct {
	devices        *sonm.DevicesReply
	mapping        benchmarks.Mapping
	freeGPUs       []*sonm.GPU
	freeBenchmarks [sonm.MinNumBenchmarks]uint64
}

func newDeviceManager(devices *sonm.DevicesReply, freeDevices *sonm.DevicesReply, mapping benchmarks.Mapping) (*DeviceManager, error) {
	m := &DeviceManager{
		devices:        devices,
		mapping:        mapping,
		freeGPUs:       append([]*sonm.GPU{}, freeDevices.GPUs...),
		freeBenchmarks: newBenchmarksFromDevices(freeDevices),
	}

	return m, nil
}

func (m *DeviceManager) Consume(benchmarks sonm.Benchmarks) (*sonm.AskPlanResources, error) {
	cpu, err := m.consumeCPU(benchmarks.ToArray())
	if err != nil {
		return nil, err
	}

	ram, err := m.consumeRAM(benchmarks.ToArray())
	if err != nil {
		return nil, err
	}

	gpu, err := m.consumeGPU(benchmarks.GPUCount(), benchmarks.ToArray())
	if err != nil {
		return nil, err
	}

	storage, err := m.consumeStorage(benchmarks.ToArray())
	if err != nil {
		return nil, err
	}

	network, err := m.consumeNetwork(benchmarks.ToArray())
	if err != nil {
		return nil, err
	}

	plan := &sonm.AskPlanResources{
		CPU:     cpu,
		RAM:     ram,
		GPU:     gpu,
		Storage: storage,
		Network: network,
	}

	return plan, nil
}

func (m *DeviceManager) consumeCPU(benchmarks []uint64) (*sonm.AskPlanCPU, error) {
	consumer := &cpuConsumer{cpu: m.devices.CPU}
	value, err := m.consume(benchmarks, consumer)
	if err != nil {
		return nil, err
	}

	return value.(*sonm.AskPlanCPU), nil
}

func (m *DeviceManager) consumeRAM(benchmarks []uint64) (*sonm.AskPlanRAM, error) {
	consumer := &ramConsumer{ram: m.devices.RAM}
	value, err := m.consume(benchmarks, consumer)
	if err != nil {
		return nil, err
	}

	return value.(*sonm.AskPlanRAM), nil
}

func (m *DeviceManager) consumeStorage(benchmarks []uint64) (*sonm.AskPlanStorage, error) {
	consumer := &storageConsumer{dev: m.devices.Storage}
	value, err := m.consume(benchmarks, consumer)
	if err != nil {
		return nil, err
	}

	return value.(*sonm.AskPlanStorage), nil
}

func (m *DeviceManager) consumeNetwork(benchmarks []uint64) (*sonm.AskPlanNetwork, error) {
	throughputIn, err := m.consumeNetworkIn(benchmarks)
	if err != nil {
		return nil, err
	}

	throughputOut, err := m.consumeNetworkOut(benchmarks)
	if err != nil {
		return nil, err
	}

	return &sonm.AskPlanNetwork{
		ThroughputIn:  throughputIn,
		ThroughputOut: throughputOut,
	}, nil
}

func (m *DeviceManager) consumeNetworkIn(benchmarks []uint64) (*sonm.DataSizeRate, error) {
	consumer := &networkInConsumer{dev: m.devices.Network}
	value, err := m.consume(benchmarks, consumer)
	if err != nil {
		return nil, err
	}

	return value.(*sonm.DataSizeRate), nil
}

func (m *DeviceManager) consumeNetworkOut(benchmarks []uint64) (*sonm.DataSizeRate, error) {
	consumer := &networkOutConsumer{dev: m.devices.Network}
	value, err := m.consume(benchmarks, consumer)
	if err != nil {
		return nil, err
	}

	return value.(*sonm.DataSizeRate), nil
}

func (m *DeviceManager) consume(benchmarks []uint64, consumer Consumer) (interface{}, error) {
	values := consumer.LowerBound()

	filter := func(id int) (uint64, bool) {
		if m.mapping.DeviceType(id) == consumer.DeviceType() {
			if m.mapping.SplittingAlgorithm(id) == consumer.SplittingAlgorithm() {
				if deviceBenchmark, ok := consumer.DeviceBenchmark(id); ok {
					return deviceBenchmark.Result, true
				}
			}
		}

		return 0, false
	}

	for id, value := range benchmarks {
		if benchmarkResult, ok := filter(id); ok {
			values = append(values, NewRational(value, benchmarkResult))
		}
	}

	value, err := Max(values)
	if err != nil {
		return 0, err
	}

	for id := range m.freeBenchmarks {
		if benchmarkResult, ok := filter(id); ok {
			if m.freeBenchmarks[id] < value.Mul(benchmarkResult).Uint64Ceil() {
				return 0, errExhausted
			}
		}
	}

	for id := range m.freeBenchmarks {
		if benchmarkResult, ok := filter(id); ok {
			m.freeBenchmarks[id] -= value.Mul(benchmarkResult).Uint64Ceil()
		}
	}

	return consumer.Result(value), nil
}

func (m *DeviceManager) consumeGPU(count uint64, benchmarks []uint64) (*sonm.AskPlanGPU, error) {
	if count == 0 {
		return &sonm.AskPlanGPU{}, nil
	}

	score := float64(math.MaxFloat64)
	var candidates []*sonm.GPU
	for _, subset := range m.combinationsGPU(int(count)) {
		currentScore := 0.0
		currentBenchmarks := append([]uint64{}, benchmarks...)

		for _, gpu := range subset {
			for id := range currentBenchmarks {
				if m.mapping.DeviceType(id) == sonm.DeviceType_DEV_GPU {
					if m.mapping.SplittingAlgorithm(id) == sonm.SplittingAlgorithm_PROPORTIONAL {
						if benchmark, ok := gpu.Benchmarks[uint64(id)]; ok {
							if currentBenchmarks[id] > benchmark.Result {
								currentBenchmarks[id] -= benchmark.Result
							} else {
								currentBenchmarks[id] = 0
							}

							currentScore += math.Pow(math.Max(0, float64(benchmark.Result)-float64(benchmarks[id]))/float64(benchmark.Result), 2)
						}
					}
				}
			}
		}

		mismatch := false
		for id := range currentBenchmarks {
			if m.mapping.DeviceType(id) == sonm.DeviceType_DEV_GPU {
				if m.mapping.SplittingAlgorithm(id) == sonm.SplittingAlgorithm_PROPORTIONAL {
					if currentBenchmarks[id] != 0 {
						mismatch = true
						break
					}
				}
			}
		}

		if mismatch {
			continue
		}

		currentScore = math.Sqrt(currentScore)

		if currentScore < score {
			score = currentScore
			candidates = append([]*sonm.GPU{}, subset...)
		}
	}

	if len(candidates) == 0 {
		return nil, errExhausted
	}

	var idx []string
	for _, dev := range candidates {
		idx = append(idx, dev.Device.Hash)
	}

	var freeGPUs []*sonm.GPU
	for _, gpu := range m.freeGPUs {
		// Exclude device.
		exclude := false
		for _, dev := range candidates {
			if gpu.Device.Hash == dev.Device.Hash {
				exclude = true
				break
			}
		}

		if !exclude {
			freeGPUs = append(freeGPUs, gpu)
		}
	}

	m.freeGPUs = freeGPUs

	return &sonm.AskPlanGPU{Hashes: idx}, nil
}

func (m *DeviceManager) combinationsGPU(k int) [][]*sonm.GPU {
	return combinationsGPU(m.freeGPUs, k)
}

func combinationsGPU(gpu []*sonm.GPU, k int) [][]*sonm.GPU {
	var GPUs [][]*sonm.GPU
	yieldCombinationsGPU(gpu, k, func(gpu []*sonm.GPU) {
		GPUs = append(GPUs, append([]*sonm.GPU{}, gpu...))
	})

	return GPUs
}

func yieldCombinationsGPU(gpu []*sonm.GPU, k int, fn func([]*sonm.GPU)) {
	subset := make([]*sonm.GPU, k)
	last := k - 1

	var recurse func(int, int)
	recurse = func(id, nextID int) {
		for j := nextID; j < len(gpu); j++ {
			subset[id] = gpu[j]
			if id == last {
				fn(subset)
			} else {
				recurse(id+1, j+1)
			}
		}
		return
	}

	recurse(0, 0)
}
