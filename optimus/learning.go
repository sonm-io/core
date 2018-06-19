package optimus

import (
	"errors"
	"math/big"
	"sort"

	"github.com/sonm-io/core/proto"
)

type WeightedOrder struct {
	Order    *MarketOrder
	Distance float64
	Weight   float64
}

// TODO: Docs.
type OrderClassifier interface {
	Classify(orders []*MarketOrder) ([]WeightedOrder, error)
}

type regressionClassifier struct {
	newModel newModel
	sigmoid  sigmoid
	clock    Clock
}

func newRegressionClassifier(newModel newModel, sigmoid sigmoid, clock Clock) OrderClassifier {
	return &regressionClassifier{
		newModel: newModel,
		sigmoid:  sigmoid,
		clock:    clock,
	}
}

func (m *regressionClassifier) Classify(orders []*MarketOrder) ([]WeightedOrder, error) {
	trainingSet := m.TrainingSet(orders)
	expectation := m.Expectation(orders)
	expectationN := append([]float64{}, expectation...)

	normalizer, err := m.Normalize(&trainingSet, expectationN)
	if err != nil {
		return nil, err
	}

	predictor, err := m.newModel().Train(trainingSet, expectationN)
	if err != nil {
		return nil, err
	}

	weightedOrders := make([]WeightedOrder, 0, len(trainingSet))
	for i, values := range trainingSet {
		normalizedPrice, err := predictor.Predict(values)
		if err != nil {
			return nil, err
		}

		distance := normalizer.Denormalize(normalizedPrice) - expectation[i]

		if orders[i].CreatedTS == nil {
			orders[i].CreatedTS = &sonm.Timestamp{}
		}
		weightedOrders = append(weightedOrders, WeightedOrder{
			Order:    orders[i],
			Distance: distance,
		})
	}

	m.RecalculateWeights(weightedOrders)
	SortOrders(weightedOrders)

	// TODO: Also we can predict worker's price.

	return weightedOrders, nil
}

func (m *regressionClassifier) TrainingSet(orders []*MarketOrder) [][]float64 {
	benchmarksCount := m.benchmarksCount(orders)
	trainingSet := make([][]float64, len(orders))
	for i, order := range orders {
		trainingSet[i] = make([]float64, benchmarksCount)
		for j, value := range order.Order.Benchmarks.ToArray() {
			trainingSet[i][j] = float64(value)
		}
	}

	return trainingSet
}

func (m *regressionClassifier) Expectation(orders []*MarketOrder) []float64 {
	expectation := make([]float64, len(orders))
	for i, order := range orders {
		price, _ := new(big.Float).SetInt(order.Order.Price.Unwrap()).Float64()
		expectation[i] = price
	}

	return expectation
}

func (m *regressionClassifier) Normalize(trainingSet *[][]float64, expectation []float64) (Normalizer, error) {
	if trainingSet == nil || len(*trainingSet) == 0 {
		return nil, errors.New("empty training set")
	}

	transposed := transpose(*trainingSet)
	filtered := transposed[:0]
	for _, values := range transposed {
		normalizer, err := newNormalizer(values...)
		switch err {
		case nil:
			if normalizer.IsDegenerated() {
				continue
			}
			filtered = append(filtered, values)
		case ErrDegenerateVector:
			continue
		default:
			return nil, err
		}

		normalizer.NormalizeBatch(values)
	}
	*trainingSet = transpose(filtered)

	normalizer, err := newNormalizer(expectation...)
	if err != nil {
		return nil, err
	}

	normalizer.NormalizeBatch(expectation)

	return normalizer, nil
}

func (m *regressionClassifier) RecalculateWeights(orders []WeightedOrder) {
	sumDistance := 0.0
	for _, order := range orders {
		sumDistance += order.Distance
	}
	meanDistance := sumDistance / float64(len(orders))

	for _, order := range orders {
		order.Weight = order.Distance + meanDistance
	}

	now := float64(m.clock().Unix())
	for _, order := range orders {
		order.Weight = order.Weight * m.sigmoid(now-float64(order.Order.CreatedTS.Seconds))
	}
}

func SortOrders(orders []WeightedOrder) {
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Weight > orders[j].Weight
	})
}

func (m *regressionClassifier) benchmarksCount(orders []*MarketOrder) int {
	max := 0
	for _, order := range orders {
		if length := len(order.Order.Benchmarks.ToArray()); length > max {
			max = length
		}
	}

	return max
}
