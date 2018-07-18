package connor

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/sonm-io/core/connor/database"
	"github.com/sonm-io/core/connor/watchers"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	daysPerMonth = 30
	secsPerDay   = 86400
)

func (t *TraderModule) ChargeOrdersOnce(ctx context.Context, token watchers.TokenWatcher, snm watchers.PriceWatcher, balanceReply *sonm.BalanceReply) error {
	if err := t.c.db.CreateOrdersTable(); err != nil {
		return err
	}

	count, err := t.c.db.GetCountFromDB()
	if err != nil {
		return fmt.Errorf("cannot get count from database: %v", err)
	}

	var mhashForToken float64

	if count == 0 {
		mhashForToken = t.c.cfg.ChargeOrders.Start
	} else {
		mhashForToken, err = t.c.db.GetLastActualStepFromDb()
		if err != nil {
			t.c.logger.Error("cannot get last actual step from DB", zap.Error(err))
			return err
		}
		mhashForToken = mhashForToken + t.c.cfg.ChargeOrders.Step
	}

	pricePerMonthUSD, pricePerSecMh, err := t.GetPriceForTokenPerSec(token)
	if err != nil {
		t.c.logger.Error("cannot get profit for tokens", zap.Error(err))
		return err
	}

	limitChargeInSNM := t.profit.LimitChargeSNM(balanceReply.GetSideBalance().Unwrap(), t.c.cfg.Trade.PartCharge)
	limitChargeInSNMClone := big.NewInt(0).Set(limitChargeInSNM)
	// snm price in USD per Ether, like 0.12
	limitChargeInUSD := t.profit.ConvertSNMBalanceToUSD(limitChargeInSNMClone, snm.GetPrice())

	pricePackMhInUSDPerMonth := mhashForToken * (pricePerMonthUSD * t.c.cfg.Trade.MarginAccounting)
	if pricePackMhInUSDPerMonth == 0 {
		return fmt.Errorf("price for pack Mhash in USD per month = 0")
	}
	sumOrdersPerMonth := limitChargeInUSD / pricePackMhInUSDPerMonth

	if limitChargeInSNM.Cmp(big.NewInt(0)) <= -1 {
		return fmt.Errorf("balance SNM is not enough for create orders")
	}

	t.c.logger.Info("start charging orders", zap.String("symbol ", t.c.cfg.UsingToken),
		zap.Float64("limit balance for charge USD", limitChargeInUSD),
		zap.Int64("sum orders per month", int64(sumOrdersPerMonth)),
		zap.Float64("step", t.c.cfg.ChargeOrders.Step),
		zap.Int64("start", int64(t.c.cfg.ChargeOrders.Start)),
		zap.Int64("destination", int64(t.c.cfg.ChargeOrders.Destination)))

	for i := 0; i < int(sumOrdersPerMonth); i++ {
		if mhashForToken >= t.c.cfg.ChargeOrders.Destination {
			t.c.logger.Info("charge is finished cause reached the limit", zap.Float64("charge_destination", t.c.cfg.ChargeOrders.Destination))
			break
		}

		pricePerSecPack := t.FloatToBigInt(mhashForToken * (pricePerSecMh * t.c.cfg.Trade.MarginAccounting))
		pricePerSecPackWithoutMargin := t.FloatToBigInt(mhashForToken * (pricePerSecMh))

		t.c.logger.Info("price", zap.Float64("hashes", mhashForToken), zap.Float64("price_per_sec_for_Mh", pricePerSecMh),
			zap.String("ending_price_with_margin_for_pack", sonm.NewBigInt(pricePerSecPack).ToPriceString()),
			zap.String("price_without_margin_for_pack", sonm.NewBigInt(pricePerSecPackWithoutMargin).ToPriceString()),
		)

		mhashForToken, err = t.ChargeOrders(ctx, pricePerSecPack, t.c.cfg.ChargeOrders.Step, mhashForToken)
		if err != nil {
			return fmt.Errorf("cannot charging market: %v", err)
		}
	}
	return nil
}

// Prepare price and Map depends on token symbol. Create orders to the market, until the budget is over.
func (t *TraderModule) ChargeOrders(ctx context.Context, priceForHashPerSec *big.Int, step float64, buyMghash float64) (float64, error) {
	requiredHashRate := uint64(buyMghash * hashes)

	if t.c.cfg.UsingToken != "ETH" {
		requiredHashRate = uint64(buyMghash)
	}

	t.c.logger.Sugar().Infof("required hashrate %v H/s", requiredHashRate)
	benchmarks, err := t.getBenchmarksForSymbol(uint64(requiredHashRate))
	if err != nil {
		return 0, err
	}

	buyMghash, err = t.CreateOrderOnMarketStep(ctx, step, benchmarks, buyMghash, priceForHashPerSec)
	if err != nil {
		return 0, err
	}

	return buyMghash, nil
}

// Create order on market depends on token.
func (t *TraderModule) CreateOrderOnMarketStep(ctx context.Context, step float64, benchmarks map[string]uint64, buyMgHash float64, price *big.Int) (float64, error) {
	actOrder, err := t.c.Market.CreateOrder(ctx, &sonm.BidOrder{
		Tag:      t.c.cfg.Trade.Tag,
		Duration: &sonm.Duration{},
		Price: &sonm.Price{PerSecond: sonm.NewBigInt(price)},
		Identity: t.c.cfg.Trade.IdentityForBid,
		Resources: &sonm.BidResources{
			Benchmarks: benchmarks,
			Network: &sonm.BidNetwork{
				Overlay:  false,
				Outbound: true,
				Incoming: false,
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("cannot create bid order: %v", err)
	}

	if actOrder.GetId() != nil && actOrder.GetPrice() != nil {
		if err := t.c.db.SaveOrderIntoDB(&database.OrderDb{
			OrderID:    actOrder.GetId().Unwrap().Int64(),
			Price:      actOrder.GetPrice().Unwrap().Int64(),
			Hashrate:   actOrder.GetBenchmarks().GPUEthHashrate(),
			StartTime:  time.Now(),
			Status:     int64(actOrder.GetOrderStatus()),
			ActualStep: buyMgHash,
		}); err != nil {
			return 0, fmt.Errorf("cannot save order to database: %v", err)
		}

		t.c.logger.Info("order created", zap.Any("order", actOrder))
		reBuyHash := buyMgHash + step
		return reBuyHash, nil
	}
	return buyMgHash, nil
}
func (t *TraderModule) GetProfitForTokenBySymbol(tokens []*TokenMainData, symbol string) (float64, error) {
	for _, t := range tokens {
		if t.Symbol == symbol {
			return t.ProfitPerMonthUsd, nil
		}
	}
	return 0, fmt.Errorf("cannot get price from token")
}

//this function determines price for 1 Mhash per second

func (t *TraderModule) GetPriceForTokenPerSec(token watchers.TokenWatcher) (float64, float64, error) {
	tokens, err := t.profit.CollectTokensMiningProfit(token)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot calculate token prices %v", err)
	}

	pricePerMonthUSD, err := t.GetProfitForTokenBySymbol(tokens, t.c.cfg.UsingToken)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get profit for tokens: %v", err)
	}

	if pricePerMonthUSD == 0 {
		return 0, 0, fmt.Errorf("price per month in USD = 0")
	}

	pricePerSec := pricePerMonthUSD / (secsPerDay * daysPerMonth)

	return pricePerMonthUSD, pricePerSec, nil
}

func (t *TraderModule) newBaseBenchmarks() map[string]uint64 {
	return map[string]uint64{
		"ram-size":            t.c.cfg.Benchmark.RamSize,
		"cpu-cores":           t.c.cfg.Benchmark.CpuCores,
		"cpu-sysbench-single": t.c.cfg.Benchmark.CpuSysbenchSingle,
		"cpu-sysbench-multi":  t.c.cfg.Benchmark.CpuSysbenchMulti,
		"net-download":        t.c.cfg.Benchmark.NetDownload,
		"net-upload":          t.c.cfg.Benchmark.NetUpload,
		"gpu-count":           t.c.cfg.Benchmark.GpuCount,
		"gpu-mem":             t.c.cfg.Benchmark.GpuMem,
	}
}

func (t *TraderModule) newBenchmarksWithGPU(ethHashRate uint64) map[string]uint64 {
	b := t.newBaseBenchmarks()
	b["gpu-eth-hashrate"] = ethHashRate
	return b
}

func (t *TraderModule) newBenchmarksWithoutGPU(zecHashrate uint64) map[string]uint64 {
	b := t.newBaseBenchmarks()
	b["gpu-cash-hashrate"] = zecHashrate
	return b
}

func (t *TraderModule) getBenchmarksForSymbol(ethHashRate uint64) (map[string]uint64, error) {
	switch t.c.cfg.UsingToken {
	case "ETH":
		return t.newBenchmarksWithGPU(ethHashRate), nil
	case "ZEC":
		return t.newBenchmarksWithoutGPU(ethHashRate), nil
	case "XMR":
		return t.newBenchmarksWithGPU(ethHashRate), nil
	default:
		return nil, fmt.Errorf("cannot create benchmakes for symbol \"%s\"", t.c.cfg.UsingToken)
	}
}

func (t *TraderModule) FloatToBigInt(val float64) *big.Int {
	realV := big.NewFloat(val)
	price := realV.Mul(realV, big.NewFloat(params.Ether))
	v, _ := price.Int(nil)
	return v
}
