package sonm

import "errors"

const (
	minRamSize    = 4 * 1024 * 1024
	minCPUPercent = 1
)

func (m *AskPlan) Validate() error {
	if m.GetResources().GetCPU().GetCores() < minCPUPercent {
		return errors.New("CPU count is too low")
	}

	if m.GetResources().GetRAM().GetSize().GetSize() < minRamSize {
		return errors.New("RAM size is too low")
	}

	return nil
}
