package connor

import (
	"context"

	"go.uber.org/zap"
)

type Connor struct {
	cfg *Config
	log *zap.Logger
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*Connor, error) {
	log.Debug("building new instance", zap.Any("config", *cfg))
	return &Connor{
		cfg: cfg,
		log: log,
	}, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.log.Debug("starting Connor instance")
	<-ctx.Done()
	return nil
}

func (c *Connor) close() {
	c.log.Debug("closing Connor")
}
