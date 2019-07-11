package optimus

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type bruteConfig struct {
	Match int                       `yaml:"match" default:"128"`
	Model optimizationMethodFactory `yaml:"model"`
}

type BatchModelConfig struct {
	Brute   bruteConfig                 `yaml:"brute"`
	Methods []optimizationMethodFactory `yaml:"models"`
}

type BatchModelFactory struct {
	BatchModelConfig
}

func (m *BatchModelFactory) Config() interface{} {
	return &m.BatchModelConfig
}

func (m *BatchModelFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	if len(matchedOrders) <= m.Brute.Match {
		return m.Brute.Model.Create(orders, matchedOrders, log)
	}

	methods := make([]OptimizationMethod, len(m.Methods))
	for id := range m.Methods {
		methods[id] = m.Methods[id].Create(orders, matchedOrders, log)
	}

	return &BatchModel{
		Methods: methods,
		Log:     log,
	}
}

type BatchModel struct {
	Methods []OptimizationMethod
	Log     *zap.SugaredLogger
}

func (m *BatchModel) Optimize(ctx context.Context, knapsack *Knapsack, orders []*MarketOrder) error {
	if len(m.Methods) == 0 {
		return fmt.Errorf("no optimization methods found")
	}

	wg, ctx := errgroup.WithContext(ctx)

	knapsacks := make([]*Knapsack, 0, len(m.Methods))
	for range m.Methods {
		knapsacks = append(knapsacks, knapsack.Clone())
	}

	for id := range m.Methods {
		method := m.Methods[id]
		knapsack := knapsacks[id]

		wg.Go(func() error {
			return method.Optimize(ctx, knapsack, orders)
		})
	}

	if err := wg.Wait(); err != nil {
		return err
	}

	winnerId := 0
	winnerPrice := 0.0
	for id := range knapsacks {
		price := knapsacks[id].PPSf64()
		m.Log.Debugf("[%d] %T optimization resulted in %.12f price", id, m.Methods[id], price)

		if price > winnerPrice {
			winnerId = id
			winnerPrice = price
		}
	}

	*knapsack = *knapsacks[winnerId].Clone()

	return nil
}
