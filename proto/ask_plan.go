package sonm

import "errors"

const (
	MinCPUPercent = 1
	MinRamSize    = 4 * 1 << 20
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
	if m.GetResources().GetCPU().GetCorePercents() < MinCPUPercent {
		return errors.New("CPU count is too low")
	}

	if m.GetResources().GetRAM().GetSize().GetBytes() < MinRamSize {
		return errors.New("RAM size is too low")
	}

	if len(m.GetResources().GetGPU().GetHashes()) > 0 && len(m.GetResources().GetGPU().GetIndexes()) > 0 {
		return errors.New("cannot set GPUs using both hashes and IDs")
	}

	return nil
}
