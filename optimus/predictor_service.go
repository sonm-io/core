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
	request.Normalize()

	worker := newMockWorker(request.GetDevices())
	engine, err := m.engineFactory(worker)
	if err != nil {
		return nil, err
	}
	if err := engine.execute(ctx); err != nil {
		return nil, err
	}

	orderIDs := make([]*sonm.BigInt, 0, len(worker.Result))
	for _, plan := range worker.Result {
		orderIDs = append(orderIDs, plan.GetOrderID())
	}

	return &sonm.PredictSupplierReply{
		Price:    sonm.SumPrice(worker.Result),
		OrderIDs: orderIDs,
	}, nil
}
