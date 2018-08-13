package antifraud

import (
	"context"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	LogFormatClaymore = "claymore"
	PoolFormatDwarf   = "dwarf"
)

// Processor is part of AntiFraud module that continuously
// processing results of container execution (like logs, pool reports, etc)
// and calculate task quality value.
type Processor interface {
	Run(ctx context.Context) error
	TaskID() string
	TaskQuality() (accurate bool, quality float64)
}

// ProcessorFactory builds particular processors for the anti-fraud
type ProcessorFactory interface {
	LogProcessor(deal *sonm.Deal, taskID string, opts ...Option) Processor
	PoolProcessor(deal *sonm.Deal, taskID string, opts ...Option) Processor
}

func NewProcessorFactory(cfg *Config) ProcessorFactory {
	var pool, log builderFunc

	switch cfg.PoolProcessorConfig.Format {
	case PoolFormatDwarf:
		pool = func(deal *sonm.Deal, taskID string, opts ...Option) Processor {
			o := &processorOpts{}
			for _, opt := range opts {
				opt(o)
			}
			return newDwarfPoolProcessor(&cfg.PoolProcessorConfig, o.logger, deal, taskID)
		}
	}

	switch cfg.LogProcessorConfig.Format {
	case LogFormatClaymore:
		log = func(deal *sonm.Deal, taskID string, opts ...Option) Processor {
			o := &processorOpts{}
			for _, opt := range opts {
				opt(o)
			}
			return newClaymoreLogProcessor(&cfg.LogProcessorConfig, o.logger, o.cc, deal, taskID)
		}
	}

	return &processorFactory{
		pool: pool,
		log:  log,
	}

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

type builderFunc func(deal *sonm.Deal, taskID string, opts ...Option) Processor

type processorFactory struct {
	pool builderFunc
	log  builderFunc
}

func (p *processorFactory) LogProcessor(deal *sonm.Deal, taskID string, opts ...Option) Processor {
	return p.log(deal, taskID, opts...)
}

func (p *processorFactory) PoolProcessor(deal *sonm.Deal, taskID string, opts ...Option) Processor {
	return p.pool(deal, taskID, opts...)
}
