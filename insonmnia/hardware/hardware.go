package hardware

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/cnf/structhash"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/ram"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
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
		},
		Storage: &sonm.Storage{Benchmarks: make(map[uint64]*sonm.Benchmark)},
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
	hashes := []string{}
	for _, idx := range indexes {
		if idx >= uint64(len(m.GPU)) {
			return nil, fmt.Errorf("invalid GPU index %d", idx)
		}
		hashes = append(hashes, m.GPU[idx].Device.Hash)
	}
	return hashes, nil
}

func (h *Hardware) SetNetworkIncoming(IPs []string) {
	for _, ip := range IPs {
		if !util.IsPrivateIP(net.ParseIP(ip)) {
			h.Network.Incoming = true
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
	//TODO: Looks like this should be fixed
	result.Network.Outbound = true
	result.Network.Overlay = h.Network.Overlay
	result.Network.Incoming = h.Network.Incoming
	//TODO: Make network device use DataSizeRate
	result.Network.ThroughputIn.BitsPerSecond = h.Network.GetIn()
	result.Network.ThroughputOut.BitsPerSecond = h.Network.GetOut()
	return result
}

func insertBench(to []uint64, bench *sonm.Benchmark, proportion float64) ([]uint64, error) {
	for len(to) <= int(bench.ID) {
		to = append(to, uint64(0))
	}
	if bench.SplittingAlgorithm == sonm.SplittingAlgorithm_NONE {
		if to[bench.ID] != 0 {
			return nil, fmt.Errorf("duplicate benchmark with id %d and type none", bench.ID)
		}
		to[bench.ID] = bench.GetResult()
	} else if bench.SplittingAlgorithm == sonm.SplittingAlgorithm_PROPORTIONAL {
		to[bench.ID] += uint64(float64(bench.Result) * proportion)
	} else if bench.SplittingAlgorithm == sonm.SplittingAlgorithm_MAX {
		if bench.Result > to[bench.ID] {
			to[bench.ID] = bench.Result
		}
	} else if bench.SplittingAlgorithm == sonm.SplittingAlgorithm_MIN {
		if bench.Result < to[bench.ID] {
			to[bench.ID] = bench.Result
		}
	}
	return to, nil
}

func (h *Hardware) ResourcesToBenchmarks(resources *sonm.AskPlanResources) (*sonm.Benchmarks, error) {
	if !resources.GPU.Normalized() {
		return nil, errors.New("passed resources are not normalized, call resources.GPU.Normalize(hardware) first")
	}
	var err error
	benchmarks := make([]uint64, sonm.MinNumBenchmarks)

	proportions := []float64{}
	hwBenches := []map[uint64]*sonm.Benchmark{}
	for _, hash := range resources.GetGPU().GetHashes() {
		found := false
		for _, gpu := range h.GPU {
			if gpu.GetDevice().GetHash() == hash {
				hwBenches = append(hwBenches, gpu.Benchmarks)
				proportions = append(proportions, 1.0)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown hash in passed resources - %s", hash)
		}
	}

	hwBenches = append(hwBenches,
		h.CPU.GetBenchmarks(),
		h.Storage.GetBenchmarks(),
		h.RAM.GetBenchmarks(),
		h.Network.GetBenchmarksIn(),
		h.Network.GetBenchmarksOut())
	proportions = append(proportions,
		float64(resources.GetCPU().GetCorePercents())/float64(h.CPU.GetDevice().GetCores())/100,
		float64(resources.GetStorage().GetSize().GetBytes())/float64(h.Storage.GetDevice().GetBytesAvailable()),
		float64(resources.GetRAM().GetSize().GetBytes())/float64(h.RAM.GetDevice().GetAvailable()),
		float64(resources.GetNetwork().GetThroughputIn().GetBitsPerSecond())/float64(h.Network.GetIn()),
		float64(resources.GetNetwork().GetThroughputOut().GetBitsPerSecond())/float64(h.Network.GetOut()))

	for idx, benchMap := range hwBenches {
		for _, bench := range benchMap {
			if benchmarks, err = insertBench(benchmarks, bench, proportions[idx]); err != nil {
				return nil, err
			}
		}
	}

	ctxlog.S(context.Background()).Infof("%v", benchmarks)

	return sonm.NewBenchmarks(benchmarks)
}

type hashableRAM struct {
	Available uint64 `json:"available"`
}

type hashableNetworkCapabilities struct {
	Overlay  bool `json:"overlay"`
	Incoming bool `json:"incoming"`
}

// DeviceMapping maps hardware capabilities to device description, hashing-friendly
type DeviceMapping struct {
	CPU         *sonm.CPUDevice             `json:"cpu"`
	GPU         []*sonm.GPUDevice           `json:"gpu"`
	RAM         hashableRAM                 `json:"ram"`
	NetworkIn   uint64                      `json:"network_in"`
	NetworkOut  uint64                      `json:"network_out"`
	Storage     *sonm.StorageDevice         `json:"storage"`
	NetworkCaps hashableNetworkCapabilities `json:"network_caps"`
}

func (dm *DeviceMapping) Hash() string {
	return fmt.Sprintf("%x", structhash.Md5(dm, 1))
}

func (h *Hardware) devicesMap() *DeviceMapping {
	var GPUs []*sonm.GPUDevice
	for _, dev := range h.GPU {
		GPUs = append(GPUs, dev.Device)
	}

	return &DeviceMapping{
		CPU:        h.CPU.Device,
		GPU:        GPUs,
		RAM:        hashableRAM{Available: h.RAM.Device.Available},
		NetworkIn:  h.Network.In,
		NetworkOut: h.Network.Out,
		Storage:    h.Storage.Device,
		NetworkCaps: hashableNetworkCapabilities{
			Incoming: h.Network.Incoming,
			Overlay:  h.Network.Overlay,
		},
	}
}
