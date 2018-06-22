package optimus

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"

	"go.uber.org/zap"
)

const (
	priceMultiplier = 1e-18
)

type WeightedOrder struct {
	// Order is an initial market order.
	Order *MarketOrder
	// PredictedPrice is a order price, that is calculated during scanning and
	// analysing the market.
	PredictedPrice float64
	// Distance represents the difference between denormalized order price and
	// the market predicted price.
	Distance float64
	// Weight represents some specific order weight.
	//
	// It fits in [0; 1] range and is used to reduce an order attractiveness
	// if it has been laying on the market for a long time without being sold.
	Weight float64
}

type OrderPredictor struct {
	predictor   TrainedModel
	normalizer  Normalizer
	normalizers []Normalizer
}

func (m *OrderPredictor) PredictPrice(order *MarketOrder) (float64, error) {
	benchmarks := order.GetOrder().GetBenchmarks().ToArray()
	if len(benchmarks) != len(m.normalizers) {
		return math.NaN(), fmt.Errorf("number of benchmarks have changed")
	}

	vec := make([]float64, 0, len(benchmarks))
	for id, benchmark := range benchmarks {
		// Skip this benchmark for degenerated normalizers.
		if m.normalizers[id] == nil {
			continue
		}
		vec = append(vec, m.normalizers[id].Normalize(float64(benchmark)))
	}

	price, err := m.predictor.Predict(vec)
	if err != nil {
		return math.NaN(), err
	}

	return m.normalizer.Denormalize(price) * priceMultiplier, nil
}

// OrderClassification is a struct that is returned after market orders
// classification. Contains weighted orders for some epoch and is able to
// predict some order's market price.
type OrderClassification struct {
	WeightedOrders []WeightedOrder
	Predictor      *OrderPredictor
}

// TODO: Docs.
type OrderClassifier interface {
	Classify(orders []*MarketOrder) ([]WeightedOrder, error)
	ClassifyExt(orders []*MarketOrder) (*OrderClassification, error)
}

type regressionClassifier struct {
	modelFactory modelFactory
	sigmoid      sigmoid
	clock        Clock
	log          *zap.Logger
}

func newRegressionClassifier(modelFactory modelFactory, sigmoid sigmoid, clock Clock, log *zap.Logger) OrderClassifier {
	return &regressionClassifier{
		modelFactory: modelFactory,
		sigmoid:      sigmoid,
		clock:        clock,
		log:          log,
	}
}

func (m *regressionClassifier) Classify(orders []*MarketOrder) ([]WeightedOrder, error) {
	classification, err := m.ClassifyExt(orders)
	if err != nil {
		return nil, err
	}
	return classification.WeightedOrders, nil
}

func (m *regressionClassifier) ClassifyExt(orders []*MarketOrder) (*OrderClassification, error) {
	trainingSet := m.TrainingSet(orders)
	expectation := m.Expectation(orders)
	expectationN := append([]float64{}, expectation...)

	trainingSetNormalizers, expectationNormalizer, err := m.Normalize(&trainingSet, expectationN)
	if err != nil {
		return nil, err
	}

	predictor, err := m.modelFactory(m.log).Train(trainingSet, expectationN)
	if err != nil {
		return nil, err
	}

	weightedOrders := make([]WeightedOrder, 0, len(trainingSet))
	for i, values := range trainingSet {
		normalizedPrice, err := predictor.Predict(values)
		if err != nil {
			return nil, err
		}

		price := expectationNormalizer.Denormalize(normalizedPrice)
		distance := price - expectation[i]

		weightedOrders = append(weightedOrders, WeightedOrder{
			Order:          orders[i],
			PredictedPrice: math.Max(0.0, price*priceMultiplier),
			Distance:       distance,
			Weight:         1.0,
		})
	}

	if err := m.RecalculateWeights(weightedOrders); err != nil {
		return nil, err
	}

	SortOrders(weightedOrders)

	orderClassification := &OrderClassification{
		WeightedOrders: weightedOrders,
		Predictor: &OrderPredictor{
			predictor:   predictor,
			normalizer:  expectationNormalizer,
			normalizers: trainingSetNormalizers,
		},
	}

	return orderClassification, nil
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

func (m *regressionClassifier) Normalize(trainingSet *[][]float64, expectation []float64) ([]Normalizer, Normalizer, error) {
	if trainingSet == nil || len(*trainingSet) == 0 {
		return nil, nil, errors.New("empty training set")
	}

	transposed := transpose(*trainingSet)
	filtered := transposed[:0]
	// Some normalizers can be nil, because of degenerated input. However we
	// must leave them, because in the future it will be used to normalize
	// order's benchmarks we want to predict the price for.
	normalizers := make([]Normalizer, len(transposed))
	for id, values := range transposed {
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
			return nil, nil, err
		}

		normalizer.NormalizeBatch(values)
		normalizers[id] = normalizer
	}
	*trainingSet = transpose(filtered)

	normalizer, err := newNormalizer(expectation...)
	if err != nil {
		return nil, nil, err
	}

	normalizer.NormalizeBatch(expectation)

	return normalizers, normalizer, nil
}

func (m *regressionClassifier) RecalculateWeights(orders []WeightedOrder) error {
	if len(orders) == 0 {
		return errors.New("empty input")
	}

	sumDistance := 0.0
	for _, order := range orders {
		sumDistance += order.Distance
	}
	meanDistance := sumDistance / float64(len(orders))

	for id, order := range orders {
		orders[id].Weight = order.Distance + meanDistance
	}

	weights := make([]float64, len(orders))
	for id := range orders {
		weights[id] = orders[id].Weight
	}

	normalizer, err := newNormalizer(weights...)
	if err != nil {
		return err
	}

	for id := range orders {
		orders[id].Weight = normalizer.Normalize(orders[id].Weight)
	}

	now := m.clock()
	for id, order := range orders {
		scale := m.sigmoid(float64(now.Unix() - order.Order.GetCreatedTS().GetSeconds()))
		if math.IsNaN(scale) {
			orders[id].Weight = 0.0
		} else {
			orders[id].Weight = order.Weight * scale
		}
	}

	return nil
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
