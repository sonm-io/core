package sonm

import (
	"unicode/utf8"
)

const (
	MetricsKeyDiskFree        = "disk_free"
	MetricsKeyDiskFreePercent = "disk_free_percent"
	MetricsKeyDiskTotal       = "disk_total"
	MetricsKeyRAMFree         = "ram_free"
	MetricsKeyRAMFreePercent  = "ram_free_percent"
	MetricsKeyRAMTotal        = "ram_total"
	MetricsKeyCPUUtilization  = "cpu_utilization"
	MetricsKeyGPUPrefix       = "gpu"
	MetricsKeyGPUTemperature  = "temp"
	MetricsKeyGPUFan          = "fan"
	MetricsKeyGPUPower        = "power"
)

func (m *TaskTag) MarshalYAML() (interface{}, error) {
	if m.GetData() == nil {
		return nil, nil
	}
	if utf8.Valid(m.GetData()) {
		return string(m.GetData()), nil
	}
	return m.GetData(), nil
}

func (m *TaskTag) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&m.Data); err == nil {
		return nil
	}
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}
	m.Data = []byte(str)
	return nil
}

func (m *WorkerMetricsResponse) Append(x ...map[string]float64) *WorkerMetricsResponse {
	if m.Metrics == nil {
		m.Metrics = make(map[string]float64)
	}

	for _, item := range x {
		for k, v := range item {
			m.Metrics[k] = v
		}
	}

	return m
}
