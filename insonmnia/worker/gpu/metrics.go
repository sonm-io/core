// +build !darwin,cl

package gpu

import (
	"context"
	"fmt"
	"sync"

	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/proto"
	"github.com/sshaman1101/nvidia-docker/nvidia"
	"go.uber.org/zap"
)

type nvidiaMetrics struct {
	mu      sync.Mutex
	devices []nvidia.Device
}

func newNvidiaMetricsHandler(ctx context.Context) (MetricsHandler, error) {
	if err := nvidia.Init(); err != nil {
		ctxlog.G(ctx).Error("failed to load NVML", zap.Error(err))
		return nil, err
	}

	devices, err := nvidia.LookupDevices()
	if err != nil {
		ctxlog.G(ctx).Error("failed to collect devices via NVML", zap.Error(err))
		return nil, err
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

func newRadeonMetricsHandler(ctx context.Context) (MetricsHandler, error) {
	devices, err := CollectDRICardDevices()
	if err != nil {
		ctxlog.G(ctx).Error("failed to collect DRI devices", zap.Error(err))
		return nil, err
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
