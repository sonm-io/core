package sonm

import (
	"fmt"
)

func (m *PredictSupplierRequest) Validate() error {
	if len(m.GetDevices().GetGPUs()) > 16 {
		return fmt.Errorf("number of GPUs must be <= 16")
	}

	return nil
}

func (m *PredictSupplierRequest) Normalize() {
	for id, dev := range m.GetDevices().GetGPUs() {
		dev.GetDevice().ID = fmt.Sprintf("%x", id)
		dev.GetDevice().FillHashID()
	}
}
