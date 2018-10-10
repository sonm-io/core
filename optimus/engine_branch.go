package optimus

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type ordersPool map[string]*MarketOrder

func (m ordersPool) Clone() ordersPool {
	pool := map[string]*MarketOrder{}
	for _, order := range m {
		id := order.GetOrder().GetId().Unwrap().String()
		pool[id] = order
	}

	return pool
}

type node struct {
	Knapsack   *Knapsack
	OrdersPool map[string]*MarketOrder

	depth    int
	children []*node
	log      *zap.SugaredLogger
}

func newNode(ctx context.Context, knapsack *Knapsack, ordersPool ordersPool, depth int, log *zap.SugaredLogger) (*node, error) {
	if err := contextDone(ctx); err != nil {
		return nil, err
	}

	children := make([]*node, 0, len(ordersPool))

	for id, order := range ordersPool {
		knapsack := knapsack.Clone()
		if err := knapsack.Put(order.GetOrder()); err != nil {
			continue
		}

		pool := ordersPool.Clone()
		delete(pool, id)

		node, err := newNode(ctx, knapsack, pool, depth+1, log)
		if err != nil {
			return nil, err
		}

		if node != nil {
			children = append(children, node)
		}
	}

	if len(children) == 0 {
		log.Debugf("found leaf node %d", depth)
	}

	m := &node{
		Knapsack:   knapsack,
		OrdersPool: ordersPool,

		depth:    depth,
		children: children,
		log:      log,
	}

	return m, nil
}

func (m *node) FindOptimum() *Knapsack {
	leafNodes := m.appendLeaf(make([]*node, 0, len(m.OrdersPool)*100))

	var winnerNode *node
	var winnerPrice = 0.0

	for _, node := range leafNodes {
		price := node.Knapsack.PPSf64()
		if price >= winnerPrice {
			winnerNode = node
			winnerPrice = price
		}
	}

	if winnerNode == nil {
		return nil
	}

	return winnerNode.Knapsack
}

func (m *node) appendLeaf(nodes []*node) []*node {
	if len(m.children) == 0 {
		return append(nodes, m)
	}

	for _, child := range m.children {
		nodes = child.appendLeaf(nodes)
	}

	return nodes
}

type BranchBoundModelConfig struct {
	HeightLimit int `yaml:"height_limit" default:"6"`
}

type BranchBoundModelFactory struct {
	BranchBoundModelConfig
}

func (m *BranchBoundModelFactory) Config() interface{} {
	return &m.BranchBoundModelConfig
}

func (*BranchBoundModelFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	return &BranchBoundModel{
		Log: log.With("model", "BBM"),
	}
}

type BranchBoundModel struct {
	Log *zap.SugaredLogger
}

func (m *BranchBoundModel) Optimize(ctx context.Context, knapsack *Knapsack, orders []*MarketOrder) error {
	ordersPool := map[string]*MarketOrder{}
	for _, order := range orders {
		id := order.GetOrder().GetId().Unwrap().String()
		ordersPool[id] = order
	}

	root, err := newNode(ctx, knapsack, ordersPool, 0, m.Log)
	if err != nil {
		return err
	}

	if root == nil {
		return fmt.Errorf("failed to construct decision tree")
	}

	m.Log.Infof("successfully build branch bound model")

	winner := root.FindOptimum()
	if winner == nil {
		return fmt.Errorf("failed to found optimum branch")
	}

	m.Log.Infof("successfully found optimum branch")

	*knapsack = *winner

	return nil
}
