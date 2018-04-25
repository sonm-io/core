package hardware

import (
	"fmt"
	"net"

	"github.com/cnf/structhash"
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

func (h *Hardware) SetNetworkIncoming(IPs []string) {
	for _, ip := range IPs {
		if !util.IsPrivateIP(net.ParseIP(ip)) {
			h.Network.Incoming = true
			break
		}
	}
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
