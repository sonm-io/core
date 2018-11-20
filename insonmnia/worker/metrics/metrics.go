package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/noxiouz/zapctx/ctxlog"
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

// TODO(sshaman1101):
// TODO(sshaman1101):
// TODO(sshaman1101): COLLECT THE FUCKING PREFIXES
// TODO(sshaman1101):
// TODO(sshaman1101):

func NewHandler(ctx context.Context, GPUConfig map[string]map[string]string) (*Handler, error) {
	handler := &Handler{
		GPUs:   make(map[sonm.GPUVendorType]gpu.MetricsHandler),
		logger: ctxlog.GetLogger(ctx).Named("metrix"),
	}

	handler.logger.Debug("initializing metrics handler")

	for vendor := range GPUConfig {
		typeID, err := gpu.GetVendorByName(vendor)
		if err != nil {
			return nil, err
		}

		h, err := gpu.NewMetricsHandler(ctx, typeID)
		if err != nil {
			handler.logger.Error("failed to initialize GPU metrics plugin", zap.String("vendor", vendor), zap.Error(err))
			return nil, err
		}

		handler.logger.Debug("successfully created GPU handler", zap.String("vendor", vendor))
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
				m.update(ctx)
			}
		}
	}()
}

func (m *Handler) update(ctx context.Context) {
	m.logger.Debug("updating hardware metrics")
	isError := false

	gpuMetrics, err := m.updateGPUMetrics()
	if err != nil {
		m.logger.Warn("failed to update GPU metrics", zap.Error(err))
		isError = true
	}

	cpuMetrics, err := m.updateCPUMetrics()
	if err != nil {
		m.logger.Warn("failed to update CPU metrics", zap.Error(err))
		isError = true
	}

	diskMetrics, err := m.updateDiskMetrics(ctx)
	if err != nil {
		m.logger.Warn("failed to update disk metrics", zap.Error(err))
		isError = true
	}

	ramMetrics, err := m.updateRAMMetrics()
	if err != nil {
		m.logger.Warn("failed to update RAM metrics", zap.Error(err))
		isError = true
	}

	// do not update metrics if some part is failed
	if !isError {
		m.mu.Lock()

		newState := &sonm.WorkerMetricsResponse{}
		newState.
			Append(gpuMetrics).
			Append(cpuMetrics).
			Append(diskMetrics).
			Append(ramMetrics)
		m.lastState = newState

		m.mu.Unlock()
	}
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

	return map[string]float64{"cpu_utilization": loads[0]}, nil
}

func (m *Handler) updateDiskMetrics(ctx context.Context) (map[string]float64, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	info, err := disk.FreeDiskSpace(ctx)
	if err != nil {
		return nil, err
	}

	metrics := map[string]float64{
		"disk_total": float64(info.AvailableBytes),
		"disk_free":  float64(info.FreeBytes),
	}

	return metrics, nil
}

func (m *Handler) updateRAMMetrics() (map[string]float64, error) {
	meme, err := ram.NewRAMDevice()
	if err != nil {
		return nil, err
	}

	metrics := map[string]float64{
		"ram_total": float64(meme.GetTotal()),
		"ram_free":  float64(meme.GetAvailable()),
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
