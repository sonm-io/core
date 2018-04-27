package sonm

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	minRamSize    = 4 * 1024 * 1024
	minCPUPercent = 1
)

func (c *AskPlanCPU) MarshalYAML() (interface{}, error) {
	return map[string]float64{"cores": float64(c.CorePercents) / 100.}, nil
}

func (c *AskPlanCPU) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// NOTE: this works till AskPlanCPU has only one field.
	// When another fields are added we may use yaml.MapSlice (or better representation announced in yaml.v3)
	// or unmarshaller for each field.
	var cpuData map[string]float64
	err := unmarshal(&cpuData)
	if err != nil {
		return err
	}
	percents, ok := cpuData["cores"]
	if !ok {
		return errors.New("missing cores section in CPU description")
	}
	c.CorePercents = uint64(percents * 100)
	return nil
}

func (m *AskPlan) Validate() error {
	if m.GetResources().GetCPU().GetCorePercents() < minCPUPercent {
		return errors.New("CPU count is too low")
	}

	if m.GetResources().GetRAM().GetSize().GetBytes() < minRamSize {
		return errors.New("RAM size is too low")
	}

	return m.GetResources().GetGPU().Validate()
}

func (m *AskPlanResources) Add(resources *AskPlanResources) error {
	if err := m.GPU.Add(resources.GPU); err != nil {
		return err
	}
	m.CPU.CorePercents += resources.CPU.CorePercents
	m.RAM.Size.Bytes += resources.RAM.Size.Bytes
	m.Storage.Size.Bytes += resources.Storage.Size.Bytes
	m.Network.Incoming = m.Network.Incoming && resources.Network.Incoming
	m.Network.Outbound = m.Network.Outbound && resources.Network.Outbound
	m.Network.Overlay = m.Network.Overlay && resources.Network.Overlay
	m.Network.ThroughputIn.BitsPerSecond += resources.Network.ThroughputIn.BitsPerSecond
	m.Network.ThroughputOut.BitsPerSecond += resources.Network.ThroughputOut.BitsPerSecond
	return nil
}

func (m *AskPlanResources) Sub(resources *AskPlanResources) error {
	if ok, desc := m.Contains(resources); !ok {
		return errors.New(desc)
	}
	m.CPU.CorePercents -= resources.CPU.CorePercents
	m.RAM.Size.Bytes -= resources.RAM.Size.Bytes
	m.Storage.Size.Bytes -= resources.Storage.Size.Bytes
	m.GPU.Sub(resources.GPU)
	m.Network.ThroughputIn.BitsPerSecond -= resources.Network.ThroughputIn.BitsPerSecond
	m.Network.ThroughputOut.BitsPerSecond -= resources.Network.ThroughputOut.BitsPerSecond
	return nil
}

func (m *AskPlanResources) ToHostConfigResources(cgroupParent string) container.Resources {
	panic("implement me")
}

func (m *AskPlanResources) ToCgroupResources() *specs.LinuxResources {
	panic("implement me")
}

func converseImplication(lhs, rhs bool) bool {
	return lhs && !rhs
}

func (m *AskPlanResources) Contains(resources *AskPlanResources) (result bool, detailedDescription string) {
	if m.CPU.CorePercents >= resources.CPU.CorePercents {
		return false, "not enough CPU"
	}
	if m.RAM.Size.Bytes >= resources.RAM.Size.Bytes {
		return false, "not enough RAM"
	}
	if m.Storage.Size.Bytes >= resources.Storage.Size.Bytes {
		return false, "not enough Storage"
	}
	if m.GPU.Contains(resources.GPU) {
		return false, "specified GPU is occupied"
	}
	if converseImplication(m.Network.Incoming, resources.Network.Incoming) {
		return false, "incoming traffic is prohibited"
	}
	if converseImplication(m.Network.Outbound, resources.Network.Outbound) {
		return false, "outbound traffic is prohibited"
	}
	if converseImplication(m.Network.Overlay, resources.Network.Overlay) {
		return false, "overlay traffic is prohibited"
	}
	if m.Network.ThroughputIn.BitsPerSecond >= resources.Network.ThroughputIn.BitsPerSecond {
		return false, "incoming traffic limit exceeded"
	}
	if m.Network.ThroughputOut.BitsPerSecond >= resources.Network.ThroughputOut.BitsPerSecond {
		return false, "incoming traffic limit exceeded"
	}
	return true, ""
}

func (m *AskPlanGPU) Validate() error {
	if len(m.GetHashes()) > 0 && len(m.GetIndexes()) > 0 {
		return errors.New("cannot set GPUs using both hashes and IDs")
	}
	return nil
}

type GPUHasher interface {
	HashGPU(indexes []uint64) (hashes []string, err error)
}

func (m *AskPlanGPU) Normalize(hasher GPUHasher) error {
	hashes, err := hasher.HashGPU(m.Indexes)
	if err != nil {
		return err
	}
	m.Indexes = []uint64{}
	m.Hashes = hashes
	return nil
}

func (m *AskPlanGPU) Normalized() bool {
	return len(m.Indexes) == 0
}

func (m *AskPlanGPU) Add(other *AskPlanGPU) error {
	// Fuck go
	result := m.deviceSet()
	for _, dev := range other.GetHashes() {
		if _, ok := result[dev]; ok {
			return fmt.Errorf("could not add up overlapping AskPlanGPU, %s is present in both", dev)
		}
		result[dev] = struct{}{}
	}
	m.restoreFromSet(result)
	return nil
}

func (m *AskPlanGPU) Sub(other *AskPlanGPU) error {
	if !m.Contains(other) {
		return errors.New("can not subtract gpu resources - minuend is less than subtrahend")
	}
	result := m.deviceSet()
	for _, dev := range other.GetHashes() {
		delete(result, dev)
	}
	m.restoreFromSet(result)
	return nil
}

func (m *AskPlanGPU) Contains(other *AskPlanGPU) bool {
	result := m.deviceSet()
	for _, dev := range other.GetHashes() {
		if _, ok := result[dev]; !ok {
			return false
		}
	}
	return true
}

func (m *AskPlanGPU) deviceSet() map[string]struct{} {
	result := map[string]struct{}{}
	for _, dev := range m.GetHashes() {
		result[dev] = struct{}{}
	}
	return result
}

func (m *AskPlanGPU) restoreFromSet(from map[string]struct{}) {
	m.Hashes = make([]string, 0, len(from))
	for dev := range from {
		m.Hashes = append(m.GetHashes(), dev)
	}
}
