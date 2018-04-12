package sonm

import "errors"

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

	return nil
}
