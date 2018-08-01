package optimus

import (
	"fmt"
	"math/big"

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

	benchmarks, err := sonm.NewBenchmarks(benchmarksValues)
	if err != nil {
		return nil, fmt.Errorf("could not parse benchmark values: %s", err)
	}

	order := &sonm.Order{
		Benchmarks: benchmarks,
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
