package optimus

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortDescending(t *testing.T) {
	orders := []WeightedOrder{
		{
			Weight: 0.1,
		},
		{
			Weight: 0.3,
		},
		{
			Weight: 0.2,
		},
	}

	SortOrders(orders)

	assert.True(t, math.Abs(orders[0].Weight-0.3) < 1e-3)
	assert.True(t, math.Abs(orders[1].Weight-0.2) < 1e-3)
	assert.True(t, math.Abs(orders[2].Weight-0.1) < 1e-3)
}

func TestLearning(t *testing.T) {
	model := newLLSModel(ModelConfig{
		Alpha:          1e-6,
		Regularization: 6.0,
		MaxIterations:  1000,
	})

	sigmoid := newSigmoid(sigmoidConfig{
		Alpha: 10.0,
		Delta: 43200.0,
	})

	n := 1000

	orders := make([]*MarketOrder, 0, 100)
	for i := 0; i < int(0.3*float64(n)); i++ {
		orders = append(orders, &MarketOrder{
			Order: &sonm.Order{
				Price: sonm.NewBigIntFromInt(77160493827160 + rand.Int63n(1000)),
				Benchmarks: &sonm.Benchmarks{
					Values: []uint64{
						40,
						21,
						2,
						256,
						160,
						1000,
						1000,
						6,
						3,
						1200,
						1860000,
						3000,
					},
				},
			},
		})
	}

	for i := 0; i < int(0.3*float64(n)); i++ {
		orders = append(orders, &MarketOrder{
			Order: &sonm.Order{
				Price: sonm.NewBigIntFromInt(115740740740741 + rand.Int63n(1000)),
				Benchmarks: &sonm.Benchmarks{
					Values: []uint64{
						40,
						21,
						2,
						256,
						160,
						1000,
						1000,
						6,
						3,
						1620,
						2700000,
						3000,
					},
				},
			},
		})
	}

	for i := 0; i < int(0.4*float64(n)); i++ {
		orders = append(orders, &MarketOrder{
			Order: &sonm.Order{
				Price: sonm.NewBigIntFromInt(77160493827160 + rand.Int63n(1000)),
				Benchmarks: &sonm.Benchmarks{
					Values: []uint64{
						40,
						21,
						2,
						256,
						160,
						1000,
						1000,
						6,
						3,
						1260,
						1560000,
						3000,
					},
				},
			},
		})
	}

	classifier := newRegressionClassifier(model, sigmoid, time.Now)
	weightedOrders, err := classifier.Classify(orders)

	require.NoError(t, err)
	assert.Equal(t, n, len(weightedOrders))
}
