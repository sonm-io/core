package sonm

func (m *PredictSupplierRequest) Normalize() {
	for _, dev := range m.GetDevices().GetGPUs() {
		dev.GetDevice().FillHashID()
	}
}
