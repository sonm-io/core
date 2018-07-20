package connor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/connor/price"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
)

type Connor struct {
	cfg *Config
	log *zap.Logger
	key *ecdsa.PrivateKey

	engine             *engine
	tokenPriceProvider price.Provider
	benchmarkList      benchmarks.BenchList
	marketClient       sonm.MarketClient
	dealsClient        sonm.DealManagementClient
}

func New(ctx context.Context, cfg *Config, log *zap.Logger) (*Connor, error) {
	log.Debug("building new instance", zap.Any("config", *cfg))

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("cannot load eth keys: %v", err)
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), key)
	if err != nil {
		return nil, fmt.Errorf("cannot create cert TLS config: %v", err)
	}

	creds := auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(key.PublicKey))
	cc, err := xgrpc.NewClient(ctx, cfg.Node.Endpoint.String(), creds)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection to node: %v", err)
	}

	tokenPrice := price.NewProvider(&price.ProviderConfig{
		Token:  cfg.Mining.Token,
		Margin: cfg.Market.PriceMarginality,
		URL:    cfg.Mining.TokenPrice.PriceURL,
	})

	return &Connor{
		key:                key,
		cfg:                cfg,
		log:                log,
		tokenPriceProvider: tokenPrice,
		marketClient:       sonm.NewMarketClient(cc),
		dealsClient:        sonm.NewDealManagementClient(cc),
		engine:             NewEngine(ctx, cfg, tokenPrice, log, cc),
	}, nil
}

func (c *Connor) Serve(ctx context.Context) error {
	c.log.Debug("starting Connor instance")

	if err := c.loadInitialData(ctx); err != nil {
		return fmt.Errorf("initializind failed: %v", err)
	}

	// perform extra config validation using external list of required benchmarks
	if err := c.cfg.validateBenchmarks(c.benchmarkList); err != nil {
		return fmt.Errorf("benchmarks validation failed: %v", err)
	}

	c.log.Debug("price",
		zap.String(c.cfg.Mining.Token, c.tokenPriceProvider.GetPrice().String()),
		zap.Float64("margin", c.cfg.Market.PriceMarginality))

	c.engine.start(ctx)
	c.startPriceTracking(ctx)

	// restore two subsets of orders, then separate on non-existing orders that
	// should be placed on market and active orders that should be watched
	// for deal opening.
	existingOrders, err := c.marketClient.GetOrders(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load orders from market: %v", err)
	}

	existingDeals, err := c.dealsClient.List(ctx, &sonm.Count{Count: 1000})
	if err != nil {
		return fmt.Errorf("cannot load deals from market: %v", err)
	}

	existingCorders := NewCordersSlice(existingOrders.GetOrders(), c.cfg.Mining.Token)
	targetCorders := c.getTargetCorders()

	set := c.divideOrdersSets(existingCorders, targetCorders)
	c.log.Debug("restoring existing entities",
		zap.Int("orders_restore", len(set.toRestore)),
		zap.Int("orders_create", len(set.toCreate)),
		zap.Int("deals_restore", len(existingDeals.GetDeal())))

	for _, deal := range existingDeals.GetDeal() {
		c.engine.RestoreDeal(deal)
	}

	for _, ord := range set.toCreate {
		c.engine.CreateOrder(ord)
	}

	for _, ord := range set.toRestore {
		c.engine.RestoreOrder(ord)
	}

	<-ctx.Done()
	c.close()
	return nil
}

func (c *Connor) loadInitialData(ctx context.Context) error {
	if err := c.tokenPriceProvider.Update(ctx); err != nil {
		return fmt.Errorf("cannot update %s price: %v", c.cfg.Mining.Token, err)
	}

	benchList, err := benchmarks.NewBenchmarksList(ctx, c.cfg.BenchmarkList)
	if err != nil {
		return fmt.Errorf("cannot load benchmark list: %v", err)
	}

	c.benchmarkList = benchList

	return nil
}

func (c *Connor) startPriceTracking(ctx context.Context) {
	go func() {
		log := c.log.Named("token-price")
		t := time.NewTicker(c.cfg.Mining.TokenPrice.UpdateInterval)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Debug("stop price tracking")
				return

			case <-t.C:
				if err := c.tokenPriceProvider.Update(ctx); err != nil {
					log.Warn("cannot update token price", zap.Error(err))
				} else {
					log.Debug("received new token price",
						zap.String("new_price", c.tokenPriceProvider.GetPrice().String()))
				}
			}
		}
	}()
}

type ordersSets struct {
	toCreate  []*Corder
	toRestore []*Corder
}

func (c *Connor) divideOrdersSets(existingCorders, targetCorders []*Corder) *ordersSets {
	byHashrate := map[uint64]*Corder{}
	for _, ord := range existingCorders {
		byHashrate[ord.GetHashrate()] = ord
	}

	set := &ordersSets{
		toCreate:  make([]*Corder, 0),
		toRestore: make([]*Corder, 0),
	}

	for _, ord := range targetCorders {
		if ex, ok := byHashrate[ord.GetHashrate()]; ok {
			set.toRestore = append(set.toRestore, ex)
		} else {
			set.toCreate = append(set.toCreate, ord)
		}
	}

	return set
}

func (c *Connor) getTargetCorders() []*Corder {
	v := make([]*Corder, 0)

	for hashrate := c.cfg.Market.FromHashRate; hashrate <= c.cfg.Market.ToHashRate; hashrate += c.cfg.Market.Step {
		bigHashrate := big.NewInt(int64(hashrate))
		p := big.NewInt(0).Mul(bigHashrate, c.tokenPriceProvider.GetPrice())

		bench := newBenchmarkFromMap(c.cfg.Market.Benchmarks)
		order, _ := NewCorderFromParams(c.cfg.Mining.Token, p, hashrate, bench)
		v = append(v, order)
	}

	return v
}

func (c *Connor) close() {
	c.log.Debug("closing Connor")
}
