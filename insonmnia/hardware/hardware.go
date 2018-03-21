package hardware

import (
	"fmt"

	"github.com/cnf/structhash"
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/proto"
)

type CPUProperties struct {
	Device    cpu.Device                 `json:"device"`
	Benchmark map[string]*sonm.Benchmark `json:"benchmark"`
}

type MemoryProperties struct {
	Device    *mem.VirtualMemoryStat     `json:"device"`
	Benchmark map[string]*sonm.Benchmark `json:"benchmark"`
}

type GPUProperties struct {
	Device    *sonm.GPUDevice            `json:"device"`
	Benchmark map[string]*sonm.Benchmark `json:"benchmark"`
}

type NetworkProperties struct {
	Device    interface{}                `json:"device"`
	Benchmark map[string]*sonm.Benchmark `json:"benchmark"`
}

type StorageProperties struct {
	Device    interface{}                `json:"device"`
	Benchmark map[string]*sonm.Benchmark `json:"benchmark"`
}

// Hardware accumulates the finest hardware information about system the worker
// is running on.
type Hardware struct {
	CPU     []*CPUProperties   `json:"cpu"`
	GPU     []*GPUProperties   `json:"gpu"`
	Memory  *MemoryProperties  `json:"memory"`
	Network *NetworkProperties `json:"network"`
	Storage *StorageProperties `json:"storage"`
}

// NewHardware returns initial hardware capabilities for Worker's host.
// Parts of the struct may be filled later by HW-plugins.
func NewHardware() (*Hardware, error) {
	hw := &Hardware{
		CPU:     []*CPUProperties{},
		GPU:     []*GPUProperties{},
		Memory:  &MemoryProperties{Benchmark: make(map[string]*sonm.Benchmark)},
		Network: &NetworkProperties{Benchmark: make(map[string]*sonm.Benchmark)},
		Storage: &StorageProperties{Benchmark: make(map[string]*sonm.Benchmark)},
	}

	CPUs, err := cpu.GetCPUDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range CPUs {
		hw.CPU = append(hw.CPU, &CPUProperties{
			Device:    dev,
			Benchmark: make(map[string]*sonm.Benchmark),
		})
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	hw.Memory = &MemoryProperties{
		Device:    vm,
		Benchmark: make(map[string]*sonm.Benchmark),
	}

	return hw, nil
}

// LogicalCPUCount returns the number of logical CPUs in the system.
func (h *Hardware) LogicalCPUCount() int {
	count := 0
	for _, c := range h.CPU {
		count += int(c.Device.Cores)
	}

	return count
}

// AvailableMemory returns amount of available RAM which can be used to allocate for containers
// (in bytes).
//
// TODO(sshaman1101): check for parent cGroup settings
func (h *Hardware) AvailableMemory() uint64 {
	return h.Memory.Device.Total
}

// GPUCount returns amount of available GPU devices.
func (h *Hardware) GPUCount() uint64 {
	return uint64(len(h.GPU))
}

func (h *Hardware) Hash() string {
	return h.devicesMap().Hash()
}

type HashableMemory struct {
	Total uint64 `json:"total"`
}

// DeviceMapping maps hardware capabilities to device description, hashing-friendly
type DeviceMapping struct {
	CPU     []cpu.Device      `json:"cpu"`
	GPU     []*sonm.GPUDevice `json:"gpu"`
	Memory  HashableMemory    `json:"memory"`
	Network interface{}       `json:"network"`
	Storage interface{}       `json:"storage"`
}

func (dm *DeviceMapping) Hash() string {
	return fmt.Sprintf("%x", structhash.Md5(dm, 1))
}

func (h *Hardware) devicesMap() *DeviceMapping {
	m := &DeviceMapping{
		CPU:     []cpu.Device{},
		GPU:     []*sonm.GPUDevice{},
		Memory:  HashableMemory{Total: h.Memory.Device.Total},
		Network: h.Network.Device,
		Storage: h.Storage.Device,
	}

	for _, c := range h.CPU {
		m.CPU = append(m.CPU, c.Device)
	}

	for _, g := range h.GPU {
		m.GPU = append(m.GPU, g.Device)
	}

	return m
}
