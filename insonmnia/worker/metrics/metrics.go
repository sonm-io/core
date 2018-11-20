package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/disk"
	"github.com/sonm-io/core/insonmnia/hardware/ram"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.Logger

	GPUs map[sonm.GPUVendorType]gpu.MetricsHandler
	// in future, special handlers for other hardware units can be controlled by this plugin

	mu        sync.Mutex
	lastState *sonm.WorkerMetricsResponse
}

func NewHandler(log *zap.Logger, GPUConfig map[string]map[string]string) (*Handler, error) {
	handler := &Handler{
		lastState: &sonm.WorkerMetricsResponse{},
		GPUs:      make(map[sonm.GPUVendorType]gpu.MetricsHandler),
		logger:    log.Named("metrix"),
	}

	handler.logger.Debug("initializing metrics handler")

	for vendor := range GPUConfig {
		typeID, err := gpu.GetVendorByName(vendor)
		if err != nil {
			return nil, err
		}

		h, err := gpu.NewMetricsHandler(typeID)
		if err != nil {
			handler.logger.Error("failed to initialize GPU metrics plugin", zap.String("vendor", vendor), zap.Error(err))
			return nil, err
		}

		handler.logger.Debug("successfully created GPU metrics handler", zap.String("vendor", vendor))
		handler.GPUs[typeID] = h
	}

	return handler, nil
}

func (m *Handler) Run(ctx context.Context) {
	go func() {
		m.logger.Debug("starting metrics collection")

		tk := util.NewImmediateTicker(time.Minute)
		defer tk.Stop()

		for {
			select {
			case <-ctx.Done():
				m.logger.Warn("context cancelled", zap.Error(ctx.Err()))
				return
			case <-tk.C:
				if err := m.update(ctx); err != nil {
					m.logger.Warn("failed to update metrics", zap.Error(err))
				}
			}
		}
	}()
}

func (m *Handler) update(ctx context.Context) error {
	m.logger.Debug("updating hardware metrics")
	merr := multierror.NewMultiError()

	gpuMetrics, err := m.updateGPUMetrics()
	if err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to update GPU metrics: %v", err))
	}

	cpuMetrics, err := m.updateCPUMetrics()
	if err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to update CPU metrics: %v", err))
	}

	diskMetrics, err := m.updateDiskMetrics(ctx)
	if err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to update disk metrics: %v", err))
	}

	ramMetrics, err := m.updateRAMMetrics()
	if err != nil {
		merr = multierror.Append(merr, fmt.Errorf("failed to update RAM metrics: %v", err))
	}

	if merr.ErrorOrNil() != nil {
		return merr.ErrorOrNil()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	newState := &sonm.WorkerMetricsResponse{}
	newState.
		Append(gpuMetrics).
		Append(cpuMetrics).
		Append(diskMetrics).
		Append(ramMetrics)
	m.lastState = newState

	return nil
}

func (m *Handler) updateGPUMetrics() (map[string]float64, error) {
	result := make(map[string]float64)
	for _, h := range m.GPUs {
		metrics, err := h.GetMetrics()
		if err != nil {
			return nil, fmt.Errorf("failed to update GPU metrics: %v", err)
		}

		for k, v := range metrics {
			result[k] = v
		}
	}

	return result, nil
}

func (m *Handler) updateCPUMetrics() (map[string]float64, error) {
	loads, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	if len(loads) == 0 {
		return nil, fmt.Errorf("CPU meter returns empty set")
	}

	return map[string]float64{
		sonm.MetricsKeyCPUUtilization: loads[0],
	}, nil
}

func (m *Handler) updateDiskMetrics(ctx context.Context) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	info, err := disk.FreeDiskSpace(ctx)
	if err != nil {
		return nil, err
	}

	metrics := map[string]float64{
		sonm.MetricsKeyDiskTotal: float64(info.AvailableBytes),
		sonm.MetricsKeyDiskFree:  float64(info.FreeBytes),
	}

	return metrics, nil
}

func (m *Handler) updateRAMMetrics() (map[string]float64, error) {
	meme, err := ram.NewRAMDevice()
	if err != nil {
		return nil, err
	}

	metrics := map[string]float64{
		sonm.MetricsKeyRAMTotal: float64(meme.GetTotal()),
		sonm.MetricsKeyRAMFree:  float64(meme.GetAvailable()),
	}

	return metrics, nil
}

func (m *Handler) Get() *sonm.WorkerMetricsResponse {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.lastState
}

func (m *Handler) Close() error {
	merr := multierror.NewMultiError()

	for _, h := range m.GPUs {
		if err := h.Close(); err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr.ErrorOrNil()
}
