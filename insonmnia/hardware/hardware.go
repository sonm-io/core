package hardware

import (
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/cnf/structhash"
	"github.com/mohae/deepcopy"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/ram"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/netutil"
)

// Hardware accumulates the finest hardware information about system the worker
// is running on.
type Hardware struct {
	CPU     *sonm.CPU     `json:"cpu"`
	GPU     []*sonm.GPU   `json:"gpu"`
	RAM     *sonm.RAM     `json:"ram"`
	Network *sonm.Network `json:"network_in"`
	Storage *sonm.Storage `json:"storage"`
}

// NewHardware returns initial hardware capabilities for Worker's host.
// Parts of the struct may be filled later by HW-plugins.
func NewHardware() (*Hardware, error) {
	var err error
	hw := &Hardware{
		CPU: &sonm.CPU{Benchmarks: make(map[uint64]*sonm.Benchmark)},
		RAM: &sonm.RAM{Benchmarks: make(map[uint64]*sonm.Benchmark, 5)},
		Network: &sonm.Network{
			BenchmarksIn:  make(map[uint64]*sonm.Benchmark),
			BenchmarksOut: make(map[uint64]*sonm.Benchmark),
			NetFlags:      &sonm.NetFlags{},
		},
		Storage: &sonm.Storage{Device: &sonm.StorageDevice{}, Benchmarks: make(map[uint64]*sonm.Benchmark)},
	}

	hw.CPU.Device, err = cpu.GetCPUDevice()
	if err != nil {
		return nil, err
	}

	hw.RAM.Device, err = ram.NewRAMDevice()
	if err != nil {
		return nil, err
	}

	return hw, nil
}

// LogicalCPUCount returns the number of logical CPUs in the system.
//
// Method is deprecated.
func (h *Hardware) LogicalCPUCount() int {
	return int(h.CPU.Device.Cores)
}

func (h *Hardware) Hash() string {
	return h.devicesMap().Hash()
}

func (m *Hardware) HashGPU(indexes []uint64) ([]string, error) {
	hashes := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		if idx >= uint64(len(m.GPU)) {
			return nil, fmt.Errorf("invalid GPU index %d", idx)
		}
		hashes = append(hashes, m.GPU[idx].Device.Hash)
	}
	return hashes, nil
}

func (m *Hardware) GPUIDs(gpuResources *sonm.AskPlanGPU) ([]gpu.GPUID, error) {
	if gpuResources == nil {
		return nil, nil
	}
	if !gpuResources.Normalized() {
		return nil, fmt.Errorf("GPU devices are not normalized")
	}
	result := make([]gpu.GPUID, 0, len(gpuResources.Hashes))
	for _, hash := range gpuResources.Hashes {
		set := false
		for _, gpuDevice := range m.GPU {
			if gpuDevice.GetDevice().GetHash() == hash {
				set = true
				result = append(result, gpu.GPUID(gpuDevice.GetDevice().GetID()))
				break
			}
		}
		if set == false {
			return nil, fmt.Errorf("could not find id for gpu hash %s", hash)
		}
	}
	return result, nil
}

func (h *Hardware) SetNetworkIncoming(IPs []string) {
	for _, ip := range IPs {
		if !netutil.IsPrivateIP(net.ParseIP(ip)) {
			h.Network.NetFlags.SetIncoming(true)
			break
		}
	}
}

func (h *Hardware) AskPlanResources() *sonm.AskPlanResources {
	result := sonm.NewEmptyAskPlanResources()
	result.CPU.CorePercents = uint64(h.CPU.GetDevice().GetCores()) * 100
	result.RAM.Size.Bytes = h.RAM.Device.Available
	result.Storage.Size.Bytes = h.Storage.GetDevice().GetBytesAvailable()
	for _, gpu := range h.GPU {
		result.GPU.Hashes = append(result.GPU.Hashes, gpu.Device.Hash)
	}
	result.Network.NetFlags = h.Network.NetFlags
	//TODO: Make network device use DataSizeRate
	result.Network.ThroughputIn.BitsPerSecond = h.Network.GetIn()
	result.Network.ThroughputOut.BitsPerSecond = h.Network.GetOut()
	return result
}

func (h *Hardware) SetDevicesFromBenches() {
	netIn, ok := h.Network.BenchmarksIn[benchmarks.NetworkIn]
	if ok {
		h.Network.In = netIn.GetResult()
	}

	netOut, ok := h.Network.BenchmarksOut[benchmarks.NetworkOut]
	if ok {
		h.Network.Out = netOut.GetResult()
	}

	storageSize, ok := h.Storage.Benchmarks[benchmarks.StorageSize]
	if ok {
		h.Storage.Device.BytesAvailable = storageSize.GetResult()
	}
}

type benchValue struct {
	isSet bool
	value uint64
}

func (m *benchValue) Set(value uint64) {
	m.value = value
	m.isSet = true
}

func (m *benchValue) Add(value uint64) {
	m.value += value
	m.isSet = true
}

func (m *benchValue) IsSet() bool {
	return m.isSet
}

func (m *benchValue) Value() uint64 {
	return m.value
}

func insertBenches(to map[uint64]*sonm.Benchmark, from map[uint64]*sonm.Benchmark, proportion float64) error {
	for _, bench := range from {
		if err := insertBench(to, bench, proportion); err != nil {
			return err
		}
	}
	return nil
}

func insertBench(to map[uint64]*sonm.Benchmark, bench *sonm.Benchmark, proportion float64) error {
	if math.IsNaN(proportion) || math.IsInf(proportion, 0) {
		proportion = 0.0
	}

	id := bench.GetID()
	_, wasSet := to[id]
	if !wasSet {
		to[id] = deepcopy.Copy(bench).(*sonm.Benchmark)
		to[id].Result = 0
	}
	target := to[id]
	switch bench.SplittingAlgorithm {
	case sonm.SplittingAlgorithm_NONE:
		to[id].Result = bench.GetResult()
		if wasSet {
			return fmt.Errorf("duplicate benchmark with id %d and splitting algorithm none", bench.ID)
		}
	case sonm.SplittingAlgorithm_PROPORTIONAL:
		target.Result += uint64(float64(bench.Result) * proportion)
	case sonm.SplittingAlgorithm_MAX:
		if bench.Result >= target.Result {
			target.Result = bench.Result
		}
	case sonm.SplittingAlgorithm_MIN:
		if !wasSet {
			target.Result = bench.GetResult()
		} else {
			if bench.Result < target.GetResult() {
				target.Result = bench.Result
			}
		}
	}
	return nil
}

func (h *Hardware) ResourcesToBenchmarks(resources *sonm.AskPlanResources) (*sonm.Benchmarks, error) {
	benchMap, err := h.ResourcesToBenchmarkMap(resources)
	if err != nil {
		return nil, err
	}
	var benchCount uint64 = sonm.MinNumBenchmarks
	for id := range benchMap {
		if benchCount < id+1 {
			benchCount = id + 1
		}
	}
	resultBenchmarks := make([]uint64, benchCount)
	for k, v := range benchMap {
		resultBenchmarks[k] = v.GetResult()
	}
	return sonm.NewBenchmarks(resultBenchmarks)
}

// This one is very similar to ResourcesToBenchmarkMap
// except it stores corresponding benchmarks in corresponding devices.
// This can not be simply reused due to MIN splitting algorithm (e.g. GPU mem - in this case it is stored separately)
// TODO: find a way to refactor all this shit.
func (h *Hardware) LimitTo(resources *sonm.AskPlanResources) (*Hardware, error) {
	if !h.Network.NetFlags.ConverseImplication(resources.GetNetwork().GetNetFlags()) {
		return nil, fmt.Errorf("provided resources netfalgs do not match")
	}
	hardware := &Hardware{
		CPU: &sonm.CPU{Device: h.CPU.Device, Benchmarks: map[uint64]*sonm.Benchmark{}},
		GPU: []*sonm.GPU{},
		RAM: &sonm.RAM{Device: h.RAM.Device, Benchmarks: map[uint64]*sonm.Benchmark{}},
		Network: &sonm.Network{
			In:            h.Network.In,
			Out:           h.Network.Out,
			NetFlags:      resources.GetNetwork().GetNetFlags(),
			BenchmarksIn:  map[uint64]*sonm.Benchmark{},
			BenchmarksOut: map[uint64]*sonm.Benchmark{},
		},
		Storage: &sonm.Storage{Device: h.Storage.Device, Benchmarks: map[uint64]*sonm.Benchmark{}},
	}

	for _, gpu := range h.GPU {
		hardware.GPU = append(hardware.GPU, &sonm.GPU{Device: gpu.Device, Benchmarks: map[uint64]*sonm.Benchmark{}})
	}

	for _, hash := range resources.GetGPU().GetHashes() {
		found := false
		for idx, gpu := range h.GPU {
			if gpu.GetDevice().GetHash() == hash {
				if err := insertBenches(hardware.GPU[idx].Benchmarks, gpu.GetBenchmarks(), 1.0); err != nil {
					return nil, err
				}

				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown hash in passed resources: %s", hash)
		}
	}

	proportion := float64(resources.GetCPU().GetCorePercents()) / float64(h.CPU.GetDevice().GetCores()) / 100
	if err := insertBenches(hardware.CPU.Benchmarks, h.CPU.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetStorage().GetSize().GetBytes()) / float64(h.Storage.GetDevice().GetBytesAvailable())
	if err := insertBenches(hardware.Storage.Benchmarks, h.Storage.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetRAM().GetSize().GetBytes()) / float64(h.RAM.GetDevice().GetAvailable())
	if err := insertBenches(hardware.RAM.Benchmarks, h.RAM.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetNetwork().GetThroughputIn().GetBitsPerSecond()) / float64(h.Network.GetIn())
	if err := insertBenches(hardware.Network.BenchmarksIn, h.Network.GetBenchmarksIn(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetNetwork().GetThroughputOut().GetBitsPerSecond()) / float64(h.Network.GetOut())
	if err := insertBenches(hardware.Network.BenchmarksOut, h.Network.GetBenchmarksOut(), proportion); err != nil {
		return nil, err
	}

	return hardware, nil
}

func (h *Hardware) ResourcesToBenchmarkMap(resources *sonm.AskPlanResources) (benchmarks map[uint64]*sonm.Benchmark, err error) {
	if !resources.GPU.Normalized() {
		return nil, errors.New("passed resources are not normalized, call resources.GPU.Normalize(hardware) first")
	}

	benchmarks = map[uint64]*sonm.Benchmark{}

	for _, hash := range resources.GetGPU().GetHashes() {
		found := false
		for _, gpu := range h.GPU {
			if gpu.GetDevice().GetHash() == hash {
				if err := insertBenches(benchmarks, gpu.GetBenchmarks(), 1.0); err != nil {
					return nil, err
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown hash in passed resources: %s", hash)
		}
	}

	proportion := float64(resources.GetCPU().GetCorePercents()) / float64(h.CPU.GetDevice().GetCores()) / 100
	if err := insertBenches(benchmarks, h.CPU.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetStorage().GetSize().GetBytes()) / float64(h.Storage.GetDevice().GetBytesAvailable())
	if err := insertBenches(benchmarks, h.Storage.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetRAM().GetSize().GetBytes()) / float64(h.RAM.GetDevice().GetAvailable())
	if err := insertBenches(benchmarks, h.RAM.GetBenchmarks(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetNetwork().GetThroughputIn().GetBitsPerSecond()) / float64(h.Network.GetIn())
	if err := insertBenches(benchmarks, h.Network.GetBenchmarksIn(), proportion); err != nil {
		return nil, err
	}

	proportion = float64(resources.GetNetwork().GetThroughputOut().GetBitsPerSecond()) / float64(h.Network.GetOut())
	if err := insertBenches(benchmarks, h.Network.GetBenchmarksOut(), proportion); err != nil {
		return nil, err
	}

	return benchmarks, nil
}

type hashableRAM struct {
	Available uint64 `json:"available"`
}

// DeviceMapping maps hardware capabilities to device description, hashing-friendly
type DeviceMapping struct {
	CPU        *sonm.CPUDevice     `json:"cpu"`
	GPU        []*sonm.GPUDevice   `json:"gpu"`
	RAM        hashableRAM         `json:"ram"`
	NetworkIn  uint64              `json:"network_in"`
	NetworkOut uint64              `json:"network_out"`
	Storage    *sonm.StorageDevice `json:"storage"`
	NetFlags   *sonm.NetFlags      `json:"netflags"`
}

func (dm *DeviceMapping) Hash() string {
	return fmt.Sprintf("%x", structhash.Md5(dm, 1))
}

func (h *Hardware) devicesMap() *DeviceMapping {
	var GPUs []*sonm.GPUDevice
	for _, dev := range h.GPU {
		GPUs = append(GPUs, dev.Device)
	}

	// We intentionally use zero values for net and storage as it is not really hardware
	return &DeviceMapping{
		CPU:        h.CPU.Device,
		GPU:        GPUs,
		RAM:        hashableRAM{Available: h.RAM.Device.Available},
		NetworkIn:  0,
		NetworkOut: 0,
		Storage:    &sonm.StorageDevice{BytesAvailable: 0},
		NetFlags:   &sonm.NetFlags{Flags: h.Network.NetFlags.GetFlags()},
	}
}
