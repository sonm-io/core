package sonm

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

const (
	// The CPU CFS scheduler period in nanoseconds. Used alongside CPU quota.
	defaultCPUPeriod = uint64(100000)

	MinCPUPercent  = 1
	MinRamSize     = 4 * 1 << 20
	MinStorageSize = 64 * 1 << 20
)

func (c *AskPlanCPU) MarshalYAML() (interface{}, error) {
	if c == nil {
		return nil, nil
	}
	return map[string]float64{"cores": float64(c.GetCorePercents()) / 100.}, nil
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
	if m.GetIdentity() == IdentityLevel_UNKNOWN {
		return errors.New("identity level is required and should not be 0")
	}
	if m.GetResources().GetCPU().GetCorePercents() < MinCPUPercent {
		return errors.New("CPU count is too low")
	}

	if m.GetResources().GetRAM().GetSize().GetBytes() < MinRamSize {
		return errors.New("RAM size is too low")
	}

	if m.GetResources().GetStorage().GetSize().GetBytes() < MinStorageSize {
		return errors.New("storage size is too low")
	}

	return m.GetResources().GetGPU().Validate()
}

func NewEmptyAskPlanResources() *AskPlanResources {
	res := &AskPlanResources{}
	res.initNilWithZero()
	return res
}

func (m *AskPlanResources) initNilWithZero() {
	if m.CPU == nil {
		m.CPU = &AskPlanCPU{}
	}
	if m.RAM == nil {
		m.RAM = &AskPlanRAM{}
	}
	if m.RAM.Size == nil {
		m.RAM.Size = &DataSize{}
	}
	if m.Storage == nil {
		m.Storage = &AskPlanStorage{}
	}
	if m.Storage.Size == nil {
		m.Storage.Size = &DataSize{}
	}
	if m.GPU == nil {
		m.GPU = &AskPlanGPU{}
	}
	if m.GPU.Hashes == nil {
		m.GPU.Hashes = []string{}
	}
	if m.GPU.Indexes == nil {
		m.GPU.Indexes = []uint64{}
	}
	if m.Network == nil {
		m.Network = &AskPlanNetwork{
			NetFlags: &NetFlags{},
		}
	}
	if m.Network.ThroughputOut == nil {
		m.Network.ThroughputOut = &DataSizeRate{}
	}
	if m.Network.ThroughputIn == nil {
		m.Network.ThroughputIn = &DataSizeRate{}
	}
}

func (m *AskPlanResources) Add(resources *AskPlanResources) error {
	m.initNilWithZero()
	if err := m.GetGPU().Add(resources.GetGPU()); err != nil {
		return err
	}
	m.CPU.CorePercents += resources.GetCPU().GetCorePercents()
	m.RAM.Size.Bytes += resources.GetRAM().GetSize().GetBytes()
	m.Storage.Size.Bytes += resources.GetStorage().GetSize().GetBytes()
	m.Network.NetFlags.Flags |= resources.GetNetwork().GetNetFlags().GetFlags()
	m.Network.ThroughputIn.BitsPerSecond += resources.GetNetwork().GetThroughputIn().GetBitsPerSecond()
	m.Network.ThroughputOut.BitsPerSecond += resources.GetNetwork().GetThroughputOut().GetBitsPerSecond()
	return nil
}

func (m *AskPlanResources) Sub(resources *AskPlanResources) error {
	m.initNilWithZero()
	if ok, desc := m.Contains(resources); !ok {
		return errors.New(desc)
	}
	m.CPU.CorePercents -= resources.GetCPU().GetCorePercents()
	m.RAM.Size.Bytes -= resources.GetRAM().GetSize().GetBytes()
	m.Storage.Size.Bytes -= resources.GetStorage().GetSize().GetBytes()
	m.GPU.Sub(resources.GetGPU())
	m.Network.ThroughputIn.BitsPerSecond -= resources.GetNetwork().GetThroughputIn().GetBitsPerSecond()
	m.Network.ThroughputOut.BitsPerSecond -= resources.GetNetwork().GetThroughputOut().GetBitsPerSecond()
	if m.Network.NetFlags.GetIncoming() && resources.GetNetwork().GetNetFlags().GetIncoming() {
		m.Network.NetFlags.SetIncoming(false)
	}
	return nil
}

func (m *AskPlanResources) ToHostConfigResources(cgroupParent string) container.Resources {
	//TODO: check and discuss
	return container.Resources{
		Memory:       int64(m.GetRAM().GetSize().GetBytes()),
		NanoCPUs:     int64(m.GetCPU().GetCorePercents() * 1e9 / 100),
		CgroupParent: cgroupParent,
	}

}

func (m *AskPlanResources) ToCgroupResources() *specs.LinuxResources {
	//TODO: Is it enough?
	maxMemory := int64(m.GetRAM().GetSize().GetBytes())
	quota := m.CPUQuota()
	period := defaultCPUPeriod
	return &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Quota:  &quota,
			Period: &period,
		},
		Memory: &specs.LinuxMemory{
			Limit: &maxMemory,
		},
	}
}

func (m *AskPlanResources) CPUQuota() int64 {
	if m == nil {
		return 0
	}
	return int64(defaultCPUPeriod) * int64(m.GetCPU().GetCorePercents()) / 100
}

func (m *AskPlanResources) Contains(resources *AskPlanResources) (result bool, detailedDescription string) {
	if m.GetCPU().GetCorePercents() < resources.GetCPU().GetCorePercents() {
		return false, fmt.Sprintf("not enough CPU, required %d core percents, available %d core percents",
			resources.GetCPU().GetCorePercents(), m.GetCPU().GetCorePercents())
	}
	if m.GetRAM().GetSize().GetBytes() < resources.GetRAM().GetSize().GetBytes() {
		return false, fmt.Sprintf("not enough RAM, required %s, available %s",
			resources.GetRAM().GetSize().Unwrap().HumanReadable(), m.GetRAM().GetSize().Unwrap().HumanReadable())
	}
	if m.GetStorage().GetSize().GetBytes() < resources.GetStorage().GetSize().GetBytes() {
		return false, "not enough Storage"
	}
	if !m.GetGPU().Contains(resources.GetGPU()) {
		return false, "specified GPU is occupied"
	}
	if !m.GetNetwork().GetNetFlags().ConverseImplication(resources.GetNetwork().GetNetFlags()) {
		return false, "net flags are not satisfied"
	}
	if m.GetNetwork().GetThroughputIn().GetBitsPerSecond() < resources.GetNetwork().GetThroughputIn().GetBitsPerSecond() {
		return false, "incoming traffic limit exceeded"
	}
	if m.GetNetwork().GetThroughputOut().GetBitsPerSecond() < resources.GetNetwork().GetThroughputOut().GetBitsPerSecond() {
		return false, "outbound traffic limit exceeded"
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

type AskPlanHasher struct {
	*AskPlanResources
}

func (m *AskPlanHasher) HashGPU(indexes []uint64) ([]string, error) {
	askPlanHashes := m.GetGPU().GetHashes()
	resultHashes := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		if idx >= uint64(len(askPlanHashes)) {
			return nil, fmt.Errorf("invalid GPU index %d", idx)
		}
		resultHashes = append(resultHashes, askPlanHashes[idx])
	}
	return resultHashes, nil
}

func (m *AskPlanGPU) Normalize(hasher GPUHasher) error {
	if m == nil || m.Normalized() {
		return nil
	}
	if err := m.Validate(); err != nil {
		return err
	}
	hashes, err := hasher.HashGPU(m.Indexes)
	if err != nil {
		return err
	}
	m.Indexes = []uint64{}
	m.Hashes = hashes
	return nil
}

func (m *AskPlanGPU) Normalized() bool {
	if m == nil {
		return true
	}
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
	if !m.Normalized() || !other.Normalized() {
		return errors.New("can not subtract gpu resources - not normalized")
	}
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
	if other == nil {
		return true
	}
	if m == nil {
		return false
	}
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

func (m *AskPlanNetwork) UnmarshalYAML(unmarshal func(interface{}) error) error {

	type Impl struct {
		ThroughputIn  *DataSizeRate
		ThroughputOut *DataSizeRate
		Overlay       bool
		Outbound      bool
		Incoming      bool
	}
	impl := &Impl{}

	if err := unmarshal(impl); err != nil {
		return err
	}

	m.ThroughputIn = impl.ThroughputIn
	m.ThroughputOut = impl.ThroughputOut
	m.NetFlags = &NetFlags{}
	m.NetFlags.SetOverlay(impl.Overlay)
	m.NetFlags.SetOutbound(impl.Outbound)
	m.NetFlags.SetIncoming(impl.Incoming)

	return nil
}
