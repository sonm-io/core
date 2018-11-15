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

	g, err := m.updateGPUMetrics()
	if err != nil {
		m.logger.Warn("failed to update GPU metrics", zap.Error(err))
		isError = true
	}

	c, err := m.updateCPUMetrics()
	if err != nil {
		m.logger.Warn("failed to update CPU metrics", zap.Error(err))
		isError = true
	}

	d, err := m.updateDiskMetrics(ctx)
	if err != nil {
		m.logger.Warn("failed to update disk metrics", zap.Error(err))
		isError = true
	}

	r, err := m.updateRAMMetrics()
	if err != nil {
		m.logger.Warn("failed to update RAM metrics", zap.Error(err))
		isError = true
	}

	// do not update metrics if some part is failed
	if !isError {
		m.mu.Lock()
		m.lastState = &sonm.WorkerMetricsResponse{
			GPUMetrics:  g,
			CpuMetrics:  c,
			DiskMetrics: d,
			RamMetrics:  r,
		}
		m.mu.Unlock()

		m.logger.Debug("new metrics received") // todo: remove
	}
}

func (m *Handler) updateGPUMetrics() ([]*sonm.GPUMetrics, error) {
	var result []*sonm.GPUMetrics

	for _, h := range m.GPUs {
		metrics, err := h.GetMetrics()
		if err != nil {
			return nil, fmt.Errorf("failed to update GPU metrics: %v", err)
		}

		result = append(result, metrics...)
	}

	return result, nil
}

func (m *Handler) updateCPUMetrics() (*sonm.CPUMetrics, error) {
	loads, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	u := float32(loads[0])
	return &sonm.CPUMetrics{Utilization: u}, nil
}

func (m *Handler) updateDiskMetrics(ctx context.Context) (*sonm.DiskMetrics, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	info, err := disk.FreeDiskSpace(ctx)
	if err != nil {
		return nil, err
	}

	dm := &sonm.DiskMetrics{
		Total: info.AvailableBytes,
		Free:  info.FreeBytes,
	}

	return dm, nil
}

func (m *Handler) updateRAMMetrics() (*sonm.RAMMetrics, error) {
	meme, err := ram.NewRAMDevice()
	if err != nil {
		return nil, err
	}

	// FIXME: should we collect all available RAM
	// or only limited by cgroup for the worker?
	r := &sonm.RAMMetrics{
		Total: meme.GetTotal(),
		Free:  meme.GetAvailable(),
	}

	return r, nil
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
