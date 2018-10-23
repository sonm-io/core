// The whole module is a piece of shit. Do we really need such kind of
// benchmark madness?

package optimus

import (
	"errors"
	"math"

	"github.com/montanaflynn/stats"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/proto"
)

const (
	minCores = sonm.MinCPUPercent * 0.01
	minRAM   = sonm.MinRamSize
)

var (
	errExhausted = errors.New("no resources left")
)

type Consumer interface {
	LowerBound() []float64
	DeviceType() sonm.DeviceType
	DeviceBenchmark(id int) (*sonm.Benchmark, bool)
	SplittingAlgorithm() sonm.SplittingAlgorithm
	Result(criteria float64) interface{}
}

type consumer struct{}

func (m *consumer) SplittingAlgorithm() sonm.SplittingAlgorithm {
	return sonm.SplittingAlgorithm_PROPORTIONAL
}

type cpuConsumer struct {
	consumer
	cpu *sonm.CPU
}

func (m *cpuConsumer) LowerBound() []float64 {
	return []float64{
		minCores / float64(m.cpu.Device.Cores),
	}
}

func (m *cpuConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_CPU
}

func (m *cpuConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.cpu.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *cpuConsumer) Result(criteria float64) interface{} {
	return &sonm.AskPlanCPU{CorePercents: uint64(math.Ceil(100.0 * criteria * float64(m.cpu.Device.Cores)))}
}

type ramConsumer struct {
	consumer
	ram *sonm.RAM
}

func (m *ramConsumer) LowerBound() []float64 {
	return []float64{
		float64(minRAM) / float64(m.ram.Device.Total),
	}
}

func (m *ramConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_RAM
}

func (m *ramConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.ram.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *ramConsumer) Result(criteria float64) interface{} {
	return &sonm.AskPlanRAM{Size: &sonm.DataSize{Bytes: uint64(math.Ceil(criteria * float64(m.ram.Device.Total)))}}
}

type storageConsumer struct {
	consumer
	dev *sonm.Storage
}

func (m *storageConsumer) LowerBound() []float64 {
	return []float64{
		float64(sonm.MinStorageSize) / float64(m.dev.Device.BytesAvailable),
	}
}

func (m *storageConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_STORAGE
}

func (m *storageConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.Benchmarks[uint64(id)]
	return benchmark, ok
}

func (m *storageConsumer) Result(criteria float64) interface{} {
	return &sonm.AskPlanStorage{Size: &sonm.DataSize{Bytes: uint64(math.Ceil(criteria * float64(m.dev.Device.BytesAvailable)))}}
}

type networkInConsumer struct {
	consumer
	dev *sonm.Network
}

func (m *networkInConsumer) LowerBound() []float64 {
	return []float64{}
}

func (m *networkInConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_NETWORK_IN
}

func (m *networkInConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.BenchmarksIn[uint64(id)]
	return benchmark, ok
}

func (m *networkInConsumer) Result(criteria float64) interface{} {
	return &sonm.DataSizeRate{BitsPerSecond: uint64(math.Ceil(criteria * float64(m.dev.In)))}
}

type networkOutConsumer struct {
	consumer
	dev *sonm.Network
}

func (m *networkOutConsumer) LowerBound() []float64 {
	return []float64{}
}

func (m *networkOutConsumer) DeviceType() sonm.DeviceType {
	return sonm.DeviceType_DEV_NETWORK_OUT
}

func (m *networkOutConsumer) DeviceBenchmark(id int) (*sonm.Benchmark, bool) {
	benchmark, ok := m.dev.BenchmarksOut[uint64(id)]
	return benchmark, ok
}

func (m *networkOutConsumer) Result(criteria float64) interface{} {
	return &sonm.DataSizeRate{BitsPerSecond: uint64(math.Ceil(criteria * float64(m.dev.Out)))}
}

type DeviceManager struct {
	devices             *sonm.DevicesReply
	mapping             benchmarks.Mapping
	freeGPUs            []*sonm.GPU
	freeBenchmarks      []uint64
	freeIncomingNetwork bool
}

func newDeviceManager(devices *sonm.DevicesReply, freeDevices *sonm.DevicesReply, mapping benchmarks.Mapping) (*DeviceManager, error) {
	freeHardware := hardware.Hardware{
		CPU:     freeDevices.CPU,
		GPU:     freeDevices.GPUs,
		RAM:     freeDevices.RAM,
		Network: freeDevices.Network,
		Storage: freeDevices.Storage,
	}

	freeBenchmarks, err := freeHardware.FullBenchmarks()
	if err != nil {
		return nil, err
	}

	m := &DeviceManager{
		devices:             devices,
		mapping:             mapping,
		freeGPUs:            append([]*sonm.GPU{}, freeDevices.GPUs...),
		freeBenchmarks:      freeBenchmarks.ToArray(), // TODO: <
		freeIncomingNetwork: freeDevices.GetNetwork().GetNetFlags().GetIncoming(),
	}

	return m, nil
}

func (m *DeviceManager) Clone() *DeviceManager {
	freeGPUs := make([]*sonm.GPU, len(m.freeGPUs))
	freeBenchmarks := make([]uint64, len(m.freeBenchmarks))

	copy(freeGPUs, m.freeGPUs)
	copy(freeBenchmarks, m.freeBenchmarks)

	return &DeviceManager{
		devices:             m.devices,
		mapping:             m.mapping,
		freeGPUs:            freeGPUs,
		freeBenchmarks:      freeBenchmarks,
		freeIncomingNetwork: m.freeIncomingNetwork,
	}
}

func (m *DeviceManager) Contains(benchmarks sonm.Benchmarks, netflags sonm.NetFlags) bool {
	copyFreeBenchmarks := append([]uint64{}, m.freeBenchmarks...)
	copyFreeGPUs := append([]*sonm.GPU{}, m.freeGPUs...)
	copyFreeIncomingNetwork := m.freeIncomingNetwork

	defer func() {
		m.freeBenchmarks = copyFreeBenchmarks
	}()
	defer func() {
		m.freeGPUs = copyFreeGPUs
	}()
	defer func() {
		m.freeIncomingNetwork = copyFreeIncomingNetwork
	}()

	_, err := m.consumeBenchmarks(benchmarks, netflags)
	return err == nil
}

func (m *DeviceManager) Consume(benchmarks sonm.Benchmarks, netflags sonm.NetFlags) (*sonm.AskPlanResources, error) {
	// Transaction-like resource restoring while consuming in case of errors.
	copyFreeBenchmarks := append([]uint64{}, m.freeBenchmarks...)
	copyFreeGPUs := append([]*sonm.GPU{}, m.freeGPUs...)
	copyFreeIncomingNetwork := m.freeIncomingNetwork

	plan, err := m.consumeBenchmarks(benchmarks, netflags)
	if err != nil {
		m.freeIncomingNetwork = copyFreeIncomingNetwork
		m.freeGPUs = copyFreeGPUs
		m.freeBenchmarks = copyFreeBenchmarks
		return nil, err
	}

	return plan, nil
}

func (m *DeviceManager) consumeBenchmarks(benchmarks sonm.Benchmarks, netflags sonm.NetFlags) (*sonm.AskPlanResources, error) {
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

	network, err := m.consumeNetwork(benchmarks.ToArray(), netflags)
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

func (m *DeviceManager) consumeNetwork(benchmarks []uint64, netflags sonm.NetFlags) (*sonm.AskPlanNetwork, error) {
	throughputIn, err := m.consumeNetworkIn(benchmarks)
	if err != nil {
		return nil, err
	}

	throughputOut, err := m.consumeNetworkOut(benchmarks)
	if err != nil {
		return nil, err
	}

	if netflags.GetIncoming() {
		if m.freeIncomingNetwork {
			m.freeIncomingNetwork = false
		} else {
			return nil, errExhausted
		}
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
			switch m.mapping.SplittingAlgorithm(id) {
			case sonm.SplittingAlgorithm_NONE, sonm.SplittingAlgorithm_MIN:
				if deviceBenchmark, ok := consumer.DeviceBenchmark(id); ok {
					if deviceBenchmark.Result < benchmarks[id] {
						return deviceBenchmark.Result, true
					}
				}
			case consumer.SplittingAlgorithm():
				if deviceBenchmark, ok := consumer.DeviceBenchmark(id); ok {
					return deviceBenchmark.Result, true
				} else {
					return 0, true
				}
			}
		}

		return 0, false
	}

	for id, value := range benchmarks {
		if benchmarkResult, ok := filter(id); ok {
			values = append(values, float64(value)/float64(benchmarkResult))
		}
	}

	value, err := stats.Max(values)
	if err != nil {
		return nil, err
	}

	for id := range m.freeBenchmarks {
		if benchmarkResult, ok := filter(id); ok {
			if m.freeBenchmarks[id] < uint64(math.Ceil(value*float64(benchmarkResult))) {
				return nil, errExhausted
			}
		}
	}

	for id := range m.freeBenchmarks {
		if benchmarkResult, ok := filter(id); ok {
			m.freeBenchmarks[id] -= uint64(math.Ceil(value * float64(benchmarkResult)))
		}
	}

	return consumer.Result(value), nil
}

type wrappedGPU struct {
	*sonm.GPU
	BenchmarksCache []uint64
}

func (m *DeviceManager) consumeGPU(minCount uint64, benchmarks []uint64) (*sonm.AskPlanGPU, error) {
	if minCount == 0 {
		if m.isGPURequired(benchmarks) {
			minCount = 1
		} else {
			return &sonm.AskPlanGPU{}, nil
		}
	}

	score := float64(math.MaxFloat64)
	var candidates []*wrappedGPU

	// Index of benchmark mapping.
	deviceTypes := make([]sonm.DeviceType, len(benchmarks))
	splittingAlgorithms := make([]sonm.SplittingAlgorithm, len(benchmarks))
	for id := range benchmarks {
		deviceTypes[id] = m.mapping.DeviceType(id)
		splittingAlgorithms[id] = m.mapping.SplittingAlgorithm(id)
	}

	// Fast filter by GPU memory benchmark.
	// All GPUs in the subset must have at least(!) the required memory
	// number.
	filteredFreeGPUs := make([]*wrappedGPU, 0, len(m.freeGPUs))
subsetLoop:
	for _, gpu := range m.freeGPUs {
		benchmarksCache := make([]uint64, len(benchmarks))
		for id, benchmarkValue := range benchmarks {
			if benchmark, ok := gpu.Benchmarks[uint64(id)]; ok {
				benchmarksCache[id] = benchmark.Result
			}

			if splittingAlgorithms[id] == sonm.SplittingAlgorithm_MIN {
				if benchmark, ok := gpu.Benchmarks[uint64(id)]; ok {
					if benchmarkValue > benchmark.Result {
						continue subsetLoop
					}
				}
			}
		}

		filteredFreeGPUs = append(filteredFreeGPUs, &wrappedGPU{
			GPU:             gpu,
			BenchmarksCache: benchmarksCache,
		})
	}

	// Move this allocation out of the cycle to prevent continuous reallocation.
	currentBenchmarks := make([]uint64, len(benchmarks))

	for k := int(minCount); k <= len(filteredFreeGPUs); k++ {
		for _, subset := range combinationsGPU(filteredFreeGPUs, k) {
			currentScore := 0.0
			copy(currentBenchmarks, benchmarks)

			for _, gpu := range subset {
				for id := range currentBenchmarks {
					if deviceTypes[id] == sonm.DeviceType_DEV_GPU {
						if splittingAlgorithms[id] == sonm.SplittingAlgorithm_PROPORTIONAL {
							//if benchmark, ok := gpu.Benchmarks[uint64(id)]; ok {
							benchmark := gpu.BenchmarksCache[id]
							if benchmark == 0 {
								if benchmarks[id] == 0 {
									// Nothing to subtract using this benchmark. Nothing to add to the score.
									// Still try the rest of benchmarks.
									continue
								} else {
									// The GPU set can't fit the benchmark. Well, possibly can,
									// but without this GPU.
									// Anyway the score will be +Inf, so definitely it's not the minimum one.
									break
								}
							}

							if currentBenchmarks[id] > benchmark {
								currentBenchmarks[id] -= benchmark
							} else {
								currentBenchmarks[id] = 0
							}

							currentScore += math.Pow(math.Max(0, float64(benchmark)-float64(benchmarks[id]))/float64(benchmark), 2)
						}
					}
				}
			}

			mismatch := false
			for id := range currentBenchmarks {
				if deviceTypes[id] == sonm.DeviceType_DEV_GPU {
					if splittingAlgorithms[id] == sonm.SplittingAlgorithm_PROPORTIONAL {
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
				candidates = append([]*wrappedGPU{}, subset...)
			}
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

func (m *DeviceManager) isGPURequired(benchmarks []uint64) bool {
	for id, value := range benchmarks {
		if m.mapping.DeviceType(id) == sonm.DeviceType_DEV_GPU {
			if value != 0 {
				return true
			}
		}
	}

	return false
}

//func (m *DeviceManager) combinationsGPU(k int) [][]*sonm.GPU {
//	return combinationsGPU(m.freeGPUs, k)
//}

func combinationsGPU(gpu []*wrappedGPU, k int) [][]*wrappedGPU {
	var GPUs [][]*wrappedGPU
	yieldCombinationsGPU(gpu, k, func(gpu []*wrappedGPU) {
		GPUs = append(GPUs, append([]*wrappedGPU{}, gpu...))
	})

	return GPUs
}

func yieldCombinationsGPU(gpu []*wrappedGPU, k int, fn func([]*wrappedGPU)) {
	subset := make([]*wrappedGPU, k)
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
