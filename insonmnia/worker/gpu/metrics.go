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

func (m *nvidiaMetrics) GetMetrics() ([]*sonm.GPUMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var metrics []*sonm.GPUMetrics
	for _, dev := range m.devices {
		status, err := dev.Status()
		if err != nil {
			return nil, fmt.Errorf("failed to get device status for GPU `%s`: %v", *dev.Model, err)
		}

		m := &sonm.GPUMetrics{
			ID:   dev.PCI.BusID,
			Name: *dev.Model,
		}

		if status.Temperature != nil {
			m.Temperature = float32(*status.Temperature)
		}

		if status.FanSpeed != nil {
			m.Fan = float32(*status.FanSpeed)
		}

		if status.Power != nil {
			m.Power = float32(*status.Power)
		}

		metrics = append(metrics, m)
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

func (m *radeonMetrics) GetMetrics() ([]*sonm.GPUMetrics, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var metrics []*sonm.GPUMetrics
	for _, dev := range m.devices {
		status, err := dev.Metrics()
		if err != nil {
			return nil, fmt.Errorf("failed to get device status for GPU `%s`: %v", dev.Name, err)
		}

		metrics = append(metrics, &sonm.GPUMetrics{
			ID:          dev.PCIBusID,
			Name:        dev.Name,
			Temperature: float32(status.Temperature),
			Fan:         float32(status.Fan),
			Power:       float32(status.Power),
		})
	}

	return metrics, nil
}

func (m *radeonMetrics) Close() error {
	return nil
}
