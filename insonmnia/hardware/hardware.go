package hardware

import (
	"fmt"

	"github.com/cnf/structhash"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/mem"
	"github.com/sonm-io/core/proto"
)

// CPU and GPU want an ID based on properties hash.

type CPUProperties struct {
	Device    *cpu.Device                `json:"device"`
	Benchmark map[uint64]*sonm.Benchmark `json:"benchmark"`
}

type MemoryProperties struct {
	ID        string                     `json:"id"`
	Device    *mem.Device                `json:"device"`
	Benchmark map[uint64]*sonm.Benchmark `json:"benchmark"`
}

type NetworkProperties struct {
	Device    interface{}                `json:"device"`
	Benchmark map[uint64]*sonm.Benchmark `json:"benchmark"`
}

type StorageProperties struct {
	Device    interface{}                `json:"device"`
	Benchmark map[uint64]*sonm.Benchmark `json:"benchmark"`
}

// Hardware accumulates the finest hardware information about system the worker
// is running on.
type Hardware struct {
	CPU     *CPUProperties     `json:"cpu"`
	GPU     []*sonm.GPUDevice  `json:"gpu"`
	Memory  *MemoryProperties  `json:"memory"`
	Network *NetworkProperties `json:"network"`
	Storage *StorageProperties `json:"storage"`
}

// NewHardware returns initial hardware capabilities for Worker's host.
// Parts of the struct may be filled later by HW-plugins.
func NewHardware() (*Hardware, error) {
	var err error
	hw := &Hardware{
		CPU:     &CPUProperties{Benchmark: make(map[uint64]*sonm.Benchmark)},
		Memory:  &MemoryProperties{Benchmark: make(map[uint64]*sonm.Benchmark)},
		Network: &NetworkProperties{Benchmark: make(map[uint64]*sonm.Benchmark)},
		Storage: &StorageProperties{Benchmark: make(map[uint64]*sonm.Benchmark)},
	}

	hw.CPU.Device, err = cpu.GetCPUDevice()
	if err != nil {
		return nil, err
	}

	vm, err := mem.NewMemoryDevice()
	if err != nil {
		return nil, err
	}

	hw.Memory = &MemoryProperties{
		Device:    vm,
		Benchmark: make(map[uint64]*sonm.Benchmark),
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

type HashableMemory struct {
	Total uint64 `json:"total"`
}

// DeviceMapping maps hardware capabilities to device description, hashing-friendly
type DeviceMapping struct {
	CPU     *cpu.Device       `json:"cpu"`
	GPU     []*sonm.GPUDevice `json:"gpu"`
	Memory  HashableMemory    `json:"memory"`
	Network interface{}       `json:"network"`
	Storage interface{}       `json:"storage"`
}

func (dm *DeviceMapping) Hash() string {
	return fmt.Sprintf("%x", structhash.Md5(dm, 1))
}

func (h *Hardware) devicesMap() *DeviceMapping {
	return &DeviceMapping{
		CPU:     h.CPU.Device,
		GPU:     h.GPU,
		Memory:  HashableMemory{Total: h.Memory.Device.Total},
		Network: h.Network.Device,
		Storage: h.Storage.Device,
	}
}
