// +build !darwin,cl

package gpu

import (
	"context"
	"fmt"
	"sync"

	"github.com/noxiouz/zapctx/ctxlog"
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

		prefix := fmt.Sprintf("gpu%d_", i)
		if status.Temperature != nil {
			metrics[prefix+"temp"] = float64(*status.Temperature)
		} else {
			metrics[prefix+"temp"] = 0
		}

		if status.FanSpeed != nil {
			metrics[prefix+"fan"] = float64(*status.FanSpeed)
		} else {
			metrics[prefix+"fan"] = 0
		}

		if status.Power != nil {
			metrics[prefix+"power"] = float64(*status.Power)
		} else {
			metrics[prefix+"power"] = 0
		}
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

		prefix := fmt.Sprintf("gpu%d_", i)
		metrics[prefix+"temp"] = status.Temperature
		metrics[prefix+"fan"] = status.Fan
		metrics[prefix+"power"] = status.Power
	}

	return metrics, nil
}

func (m *radeonMetrics) Close() error {
	return nil
}
