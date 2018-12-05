// +build !darwin,cl

package gpu

import (
	"fmt"
	"sync"

	"github.com/sonm-io/core/proto"
	"github.com/sshaman1101/nvidia-docker/nvidia"
)

type nvidiaMetrics struct {
	mu      sync.Mutex
	devices []nvidia.Device
}

func newNvidiaMetricsHandler() (MetricsHandler, error) {
	if err := nvidia.Init(); err != nil {
		return nil, fmt.Errorf("failed to load NVML: %v", err)
	}

	devices, err := nvidia.LookupDevices()
	if err != nil {
		return nil, fmt.Errorf("failed to collect devices via NVML: %v", err)
	}

	return &nvidiaMetrics{devices: devices}, nil
}

func (m *nvidiaMetrics) GetMetrics() (map[string]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make(map[string]float64)
	for i, dev := range m.devices {
		status, err := dev.Status()
		if err != nil {
			return nil, fmt.Errorf("failed to get device status for GPU `%s`: %v", *dev.Model, err)
		}

		var temp float64 = 0
		var fan float64 = 0
		var power float64 = 0

		if status.Temperature != nil {
			temp = float64(*status.Temperature)
		}

		if status.FanSpeed != nil {
			fan = float64(*status.FanSpeed)
		}

		if status.Power != nil {
			power = float64(*status.Power)
		}

		metrics[tempKey(i)] = temp
		metrics[fanKey(i)] = fan
		metrics[powerKey(i)] = power
	}

	return metrics, nil
}

func (m *nvidiaMetrics) Close() error {
	return nvidia.Shutdown()
}

type radeonMetrics struct {
	mu      sync.Mutex
	devices []DRICard
}

func newRadeonMetricsHandler() (MetricsHandler, error) {
	devices, err := collectDRICardsWithOpenCL()
	if err != nil {
		return nil, fmt.Errorf("failed to collect DRI devices: %v", err)
	}

	return &radeonMetrics{devices: devices}, nil
}

func (m *radeonMetrics) GetMetrics() (map[string]float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := make(map[string]float64)
	for i, dev := range m.devices {
		status, err := dev.Metrics()
		if err != nil {
			return nil, fmt.Errorf("failed to get device status for GPU `%s`: %v", dev.Name, err)
		}

		metrics[tempKey(i)] = status.Temperature
		metrics[fanKey(i)] = status.Fan
		metrics[powerKey(i)] = status.Power
	}

	return metrics, nil
}

func (m *radeonMetrics) Close() error {
	return nil
}

func tempKey(i int) string {
	return fmt.Sprintf("%s%d_%s", sonm.MetricsKeyGPUPrefix, i, sonm.MetricsKeyGPUTemperature)
}

func fanKey(i int) string {
	return fmt.Sprintf("%s%d_%s", sonm.MetricsKeyGPUPrefix, i, sonm.MetricsKeyGPUFan)
}

func powerKey(i int) string {
	return fmt.Sprintf("%s%d_%s", sonm.MetricsKeyGPUPrefix, i, sonm.MetricsKeyGPUPower)
}
