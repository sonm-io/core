package hardware

import (
	"fmt"

	"github.com/cnf/structhash"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/mem"
	"github.com/sonm-io/core/proto"
)

// Hardware accumulates the finest hardware information about system the worker
// is running on.
type Hardware struct {
	CPU     *sonm.CPU     `json:"cpu"`
	GPU     []*sonm.GPU   `json:"gpu"`
	RAM     *sonm.RAM     `json:"ram"`
	Network *sonm.Network `json:"network"`
	Storage *sonm.Storage `json:"storage"`
}

// NewHardware returns initial hardware capabilities for Worker's host.
// Parts of the struct may be filled later by HW-plugins.
func NewHardware() (*Hardware, error) {
	var err error
	hw := &Hardware{
		CPU:     &sonm.CPU{Benchmarks: make(map[uint64]*sonm.Benchmark)},
		RAM:     &sonm.RAM{Benchmarks: make(map[uint64]*sonm.Benchmark)},
		Network: &sonm.Network{Benchmarks: make(map[uint64]*sonm.Benchmark)},
		Storage: &sonm.Storage{Benchmarks: make(map[uint64]*sonm.Benchmark)},
	}

	hw.CPU.Device, err = cpu.GetCPUDevice()
	if err != nil {
		return nil, err
	}

	hw.RAM.Device, err = mem.NewMemoryDevice()
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

type hashableMemory struct {
	Available uint64 `json:"total"`
}

// DeviceMapping maps hardware capabilities to device description, hashing-friendly
type DeviceMapping struct {
	CPU     *sonm.CPUDevice     `json:"cpu"`
	GPU     []*sonm.GPUDevice   `json:"gpu"`
	Memory  hashableMemory      `json:"memory"`
	Network *sonm.NetworkDevice `json:"network"`
	Storage *sonm.StorageDevice `json:"storage"`
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
		CPU:     h.CPU.Device,
		GPU:     GPUs,
		Memory:  hashableMemory{Available: h.RAM.Device.Available},
		Network: h.Network.Device,
		Storage: h.Storage.Device,
	}
}
