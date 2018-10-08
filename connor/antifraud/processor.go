package antifraud

import (
	"context"

	"github.com/sonm-io/core/connor/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	LogFormatCommon         = "common"
	PoolFormatDwarf         = "dwarf"
	ProcessorFormatDisabled = "disabled"
)

// Processor is part of AntiFraud module that continuously
// processing results of container execution (like logs, pool reports, etc)
// and calculate task quality value.
type Processor interface {
	Run(ctx context.Context) error
	TaskID() string
	TaskQuality() (accurate bool, quality float64)
}

type disabledProcessor struct {
	taskID string
}

func (p *disabledProcessor) TaskID() string                                { return p.taskID }
func (p *disabledProcessor) TaskQuality() (accurate bool, quality float64) { return true, 1.0 }
func (p *disabledProcessor) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// ProcessorFactory builds particular processors for the anti-fraud
type ProcessorFactory interface {
	LogProcessor(deal *types.Deal, taskID string, opts ...Option) Processor
	PoolProcessor(deal *types.Deal, taskID string, opts ...Option) Processor
}

func NewProcessorFactory(cfg *Config) ProcessorFactory {
	var pool, log builderFunc

	switch cfg.PoolProcessorConfig.Format {
	case PoolFormatDwarf:
		pool = func(deal *types.Deal, taskID string, opts ...Option) Processor {
			o := makeOpts(opts...)
			return newDwarfPoolProcessor(&cfg.PoolProcessorConfig, o.logger, deal, taskID)
		}
	case ProcessorFormatDisabled:
		pool = func(deal *types.Deal, taskID string, opts ...Option) Processor {
			return &disabledProcessor{}
		}
	}

	switch cfg.LogProcessorConfig.Format {
	case LogFormatCommon:
		log = func(deal *types.Deal, taskID string, opts ...Option) Processor {
			o := makeOpts(opts...)
			return newLogProcessor(&cfg.LogProcessorConfig, o.logger, o.cc, deal, taskID)
		}
	case ProcessorFormatDisabled:
		log = func(deal *types.Deal, taskID string, opts ...Option) Processor {
			return &disabledProcessor{taskID: taskID}
		}
	}

	return &processorFactory{
		pool: pool,
		log:  log,
	}
}

func makeOpts(opts ...Option) *processorOpts {
	o := &processorOpts{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type processorOpts struct {
	logger *zap.Logger
	cc     *grpc.ClientConn
}

type Option func(options *processorOpts)

func WithLogger(log *zap.Logger) Option {
	return func(o *processorOpts) {
		o.logger = log
	}
}

func WithClientConn(cc *grpc.ClientConn) Option {
	return func(o *processorOpts) {
		o.cc = cc
	}
}

type builderFunc func(deal *types.Deal, taskID string, opts ...Option) Processor

type processorFactory struct {
	pool builderFunc
	log  builderFunc
}

func (p *processorFactory) LogProcessor(deal *types.Deal, taskID string, opts ...Option) Processor {
	return p.log(deal, taskID, opts...)
}

func (p *processorFactory) PoolProcessor(deal *types.Deal, taskID string, opts ...Option) Processor {
	return p.pool(deal, taskID, opts...)
}
