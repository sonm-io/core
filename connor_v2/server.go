package connor

import (
	"context"
	"fmt"
	"math/big"

	"github.com/sonm-io/core/connor_v2/price"
	"go.uber.org/zap"
)

type Connor struct {
	cfg *Config
	log *zap.Logger

	snmPriceProvider   price.Provider
	tokenPriceProvider price.Provider
	orderManager       *orderManager
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*Connor, error) {
	log.Debug("building new instance", zap.Any("config", *cfg))

	return &Connor{
		cfg:                cfg,
		log:                log,
		snmPriceProvider:   price.NewSonmPriceProvider(),
		tokenPriceProvider: price.NewProvider(cfg.Mining.Token),
		orderManager:       NewOrderManager(ctx, log, nil),
	}, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.log.Debug("starting Connor instance")

	if err := c.loadInitialData(ctx); err != nil {
		return fmt.Errorf("initializind failed: %v", err)
	}

	c.log.Debug("price", zap.String("SNM", c.snmPriceProvider.GetPrice().String()),
		zap.String(c.cfg.Mining.Token, c.tokenPriceProvider.GetPrice().String()))

	// todo: detach in background
	go c.orderManager.start(ctx)

	newOrder := newBidTemplate(c.cfg.Mining.Token)
	c.log.Debug("steps", zap.Uint64("from", c.cfg.Market.FromHashRate),
		zap.Uint64("to", c.cfg.Market.ToHashRate),
		zap.Uint64("step", c.cfg.Market.Step))

	for hr := c.cfg.Market.FromHashRate; hr <= c.cfg.Market.ToHashRate; hr += c.cfg.Market.Step {
		p := big.NewInt(0).Mul(c.tokenPriceProvider.GetPrice(), big.NewInt(int64(hr)))
		c.log.Debug("requesting initial order placement", zap.String("price", p.String()), zap.Uint64("hashrate", hr))
		c.orderManager.Create(newOrder(p, hr))
	}

	<-ctx.Done()
	return nil
}

func (c *Connor) loadInitialData(ctx context.Context) error {
	if err := c.snmPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update SNM price: %v", err)
	}

	if err := c.tokenPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update %s price: %v", c.cfg.Mining.Token, err)
	}

	// todo: load orders storage

	return nil
}

func (c *Connor) close() {
	c.log.Debug("closing Connor")
}
