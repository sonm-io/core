package optimus

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m *PredictorService) Predict(ctx context.Context, request *sonm.BidResources) (*sonm.Price, error) {
	classification := m.Classification()
	if classification == nil {
		return nil, status.Errorf(codes.Unavailable, "regression is not finished yet")
	}

	knownBenchmarks := m.benchmarks.MapByCode()
	givenBenchmarks := request.GetBenchmarks()

	if len(givenBenchmarks) > len(knownBenchmarks) {
		return nil, fmt.Errorf("benchmark list too large")
	}

	benchmarksValues := make([]uint64, len(knownBenchmarks))
	for code, value := range givenBenchmarks {
		bench, ok := knownBenchmarks[code]
		if !ok {
			return nil, fmt.Errorf("unknown benchmark code: %s", code)
		}

		benchmarksValues[bench.GetID()] = value
	}

	orderBenchmarks, err := sonm.NewBenchmarks(benchmarksValues)
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark values: %s", err)
	}

	order := &sonm.Order{
		Benchmarks: orderBenchmarks,
	}

	price, err := classification.Predictor.PredictPrice(&MarketOrder{
		Order: order,
	})

	if err != nil {
		return nil, err
	}

	price /= priceMultiplier
	priceBigF := big.NewFloat(price)
	priceBigI, _ := priceBigF.Int(nil)

	return &sonm.Price{
		PerSecond: sonm.NewBigInt(priceBigI),
	}, nil
}

func (m *PredictorService) PredictSupplier(ctx context.Context, request *sonm.PredictSupplierRequest) (*sonm.PredictSupplierReply, error) {
	priceThresholdValue, err := NewRelativePriceThreshold(5.0)
	if err != nil {
		return nil, err
	}

	cfg := &workerConfig{
		PrivateKey:  m.cfg.Marketplace.PrivateKey,
		Epoch:       60 * time.Second,
		OrderPolicy: 0,
		DryRun:      false,
		Identity:    sonm.IdentityLevel_ANONYMOUS,
		PriceThreshold: priceThreshold{
			PriceThreshold: priceThresholdValue,
		},
		StaleThreshold: 5 * time.Minute,
		PreludeTimeout: 30 * time.Second,
		Optimization: OptimizationConfig{
			Model: optimizationMethodFactory{
				OptimizationMethodFactory: &defaultOptimizationMethodFactory{},
			},
		},
	}

	blacklist := newEmptyBlacklist()
	worker := newMockWorker(request.GetDevices())
	benchmarkMapping := benchmarks.NewArrayMapping(m.benchmarks, m.benchmarks.Max())
	tagger := newTagger("predictor")

	engine, err := newWorkerEngine(cfg, common.Address{}, common.Address{}, blacklist, worker, m.market, m.marketCache, benchmarkMapping, tagger, m.log)
	if err != nil {
		return nil, err
	}

	if err := engine.execute(ctx); err != nil {
		return nil, err
	}

	return &sonm.PredictSupplierReply{Price: sonm.SumPrice(worker.Result)}, nil
}
