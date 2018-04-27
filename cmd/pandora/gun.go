package main

import (
	"context"

	"github.com/yandex/pandora/core"
	"github.com/yandex/pandora/core/aggregate/netsample"
	"go.uber.org/zap"
)

type Gun interface {
	core.Gun
}

type gun struct {
	aggregator core.Aggregator
	ext        interface{}
	log        *zap.Logger
}

func newGun(ext interface{}, log *zap.Logger) Gun {
	return &gun{
		ext: ext,
		log: log,
	}
}

func (m *gun) Bind(aggregator core.Aggregator) {
	m.aggregator = aggregator
}

func (m *gun) Shoot(ctx context.Context, ammo core.Ammo) {
	sample := netsample.Acquire("REQUEST")

	err := ammo.(Ammo).Execute(ctx, m.ext)

	if err == nil {
		sample.SetProtoCode(200)
	} else {
		m.log.Warn("failed to process ammo", zap.Error(err))
		sample.SetErr(err)
		sample.SetProtoCode(500)
	}

	m.aggregator.Report(sample)
}
