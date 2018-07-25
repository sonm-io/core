package optimus

import (
	"fmt"
	"math"

	"go.uber.org/zap"
)

// GreedyLinearRegressionModel implements greedy knapsack optimization
// algorithm.
// The basic idea is to train the model using BID orders from the marketplace
// by optimizing multidimensional linear regression over order benchmarks to
// reduce the number of parameters to a single one - predicted price. This
// price can be used to assign weights to orders to be able to determine which
// orders are better to buy than others.
type GreedyLinearRegressionModel struct {
	orders          []*MarketOrder
	classifier      OrderClassifier
	exhaustionLimit int
	log             *zap.SugaredLogger
}

func (m *GreedyLinearRegressionModel) Optimize(knapsack *Knapsack, orders []*MarketOrder) error {
	if len(m.orders) <= minNumOrders {
		return fmt.Errorf("not enough orders to perform optimization")
	}

	weightedOrders, err := m.classifier.Classify(m.orders)
	if err != nil {
		return fmt.Errorf("failed to classify orders: %v", err)
	}

	// Here we create an index of matching orders to be able to filter
	// the entire training set for only interesting features.
	filter := map[string]struct{}{}
	for _, order := range orders {
		filter[order.GetOrder().GetId().Unwrap().String()] = struct{}{}
	}

	exhaustedCounter := 0
	for _, weightedOrder := range weightedOrders {
		// Ignore orders with too low relative weight, i.e. orders that have
		// quotient of its price to predicted price less than 1%.
		// It may be, for example, when an order has 0 price.
		// TODO: For now not sure where to perform this filtering. Let it be here.
		if math.Abs(weightedOrder.Weight) < 0.01 {
			m.log.Debugf("ignore `%s` order - weight too low: %.6f", weightedOrder.ID().String(), weightedOrder.Weight)
			continue
		}

		if _, ok := filter[weightedOrder.ID().String()]; !ok {
			continue
		}

		if exhaustedCounter >= m.exhaustionLimit {
			break
		}

		order := weightedOrder.Order.Order

		m.log.Debugw("trying to put an order into resources pool",
			zap.Any("order", *weightedOrder.Order),
			zap.Float64("weight", weightedOrder.Weight),
			zap.String("price", order.Price.ToPriceString()),
			zap.Float64("predictedPrice", weightedOrder.PredictedPrice),
		)

		switch err := knapsack.Put(order); err {
		case nil:
		case errExhausted:
			exhaustedCounter += 1
			continue
		default:
			return fmt.Errorf("failed to consume order: %v", err)
		}
	}

	return nil
}
