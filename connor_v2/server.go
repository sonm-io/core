package connor

import (
	"context"
	"fmt"

	"github.com/sonm-io/core/connor_v2/price"
	"go.uber.org/zap"
)

type Connor struct {
	cfg *Config
	log *zap.Logger

	snmPriceProvider   price.Provider
	tokenPriceProvider price.Provider
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*Connor, error) {
	log.Debug("building new instance", zap.Any("config", *cfg))

	return &Connor{
		cfg:                cfg,
		log:                log,
		snmPriceProvider:   price.NewSonmPriceProvider(),
		tokenPriceProvider: price.NewProvider(cfg.Mining.Token),
	}, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.log.Debug("starting Connor instance")

	if err := c.snmPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update SNM price: %v", err)
	}

	if err := c.tokenPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update %s price: %v", c.cfg.Mining.Token, err)
	}

	c.log.Debug("price", zap.String("SNM", c.snmPriceProvider.GetPrice().String()),
		zap.String(c.cfg.Mining.Token, c.tokenPriceProvider.GetPrice().String()))

	<-ctx.Done()
	return nil
}

func (c *Connor) close() {
	c.log.Debug("closing Connor")
}
