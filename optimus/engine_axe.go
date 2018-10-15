package optimus

import (
	"context"
	"math/big"
	"sync"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type AxeModelConfig struct{}

type AxeModelFactory struct {
	AxeModelConfig
}

func (a *AxeModelFactory) Config() interface{} {
	return &a.AxeModelConfig
}

func (a *AxeModelFactory) Create(orders, matchedOrders []*MarketOrder, log *zap.SugaredLogger) OptimizationMethod {
	return &AxeModel{
		Log: log.With("model", "AXE"),
	}
}

type AxeModel struct {
	treap *treap
	Log   *zap.SugaredLogger
}

func (m *AxeModel) Optimize(ctx context.Context, ks *Knapsack, orders []*MarketOrder) error {
	m.treap = initTreap(len(orders))
	for i := 0; i < len(orders); i++ {
		o := prepareOrder(orders[i], ks)
		m.treap.Push(o)
	}
	m.Log.Infof("successfully build priority queue from %d orders", len(orders))
	//if a.treap.size >= 1 {
	//	a.Log.Debugf("pq is %+v",a.treap.el[0])
	//}
	//if a.treap.size >= 2 {
	//	a.Log.Debugf("pq L: %+v" ,a.treap.el[1])
	//}
	//if a.treap.size >= 3 {
	//	a.Log.Debugf("pq R: %+v", a.treap.el[2])
	//}
	for {
		err := ks.Put(m.treap.Pop().o)
		if err != nil {
			if err == errExhausted {
				break
			}
			return err
		}
	}
	return nil
}

func estimateWeightFast(o *sonm.Order, ks *Knapsack) float64 {
	if len(ks.manager.freeBenchmarks) != len(o.Benchmarks.Values) {
		// todo what if len not equal?)
		return 0.0
	}

	var (
		ow  uint64
		ksw uint64
	)

	bm := ks.manager.freeBenchmarks
	v := o.Benchmarks.Values
	// than lesser resource amount than more valuable it is
	for i := 0; i < len(bm); i++ {
		ksw += bm[i]
		ow += v[i]
	}
	return float64(ow) / float64(ksw)
}

func prepareOrder(o *MarketOrder, ks *Knapsack) *axeOrder {
	d := big.NewInt(int64(o.Order.Duration))
	p := o.Order.Price.Unwrap()

	return &axeOrder{
		o: o.Order,
		w: estimateWeightFast(o.Order, ks),
		p: p.Quo(d, p),
	}
}

type axeOrder struct {
	w float64  // weight
	p *big.Int // price
	o *sonm.Order
}

// a <= b ?
func less(a, b *axeOrder) bool {
	if a.w > b.w && a.p.Cmp(b.p) < 0 {
		return true
	}
	return false
}

type treap struct {
	mu   sync.Mutex
	el   []*axeOrder
	size int
}

func initTreap(size int) *treap {
	return &treap{
		el: make([]*axeOrder, size),
	}
}

func (m *treap) Push(o *axeOrder) {
	m.mu.Lock()
	if m.size == 0 {
		m.el[0] = o
	} else {
		m.el[m.size] = o
		m.siftUp(m.size)
	}
	m.size++
	m.mu.Unlock()
}

// returns max-likely order which seller should get
func (m *treap) Pop() *axeOrder {
	m.mu.Lock()
	res := m.el[0]
	m.size--
	m.el[0] = m.el[m.size]
	m.siftDown(0)
	m.mu.Unlock()
	return res
}

func (m *treap) Flush() {
	m.mu.Lock()
	m.size = 0
	m.mu.Unlock()
}

func (m *treap) siftUp(i int) {
	for {
		if less(m.el[(i-1)/2], m.el[i]) {
			m.el[(i-1)/2], m.el[i] = m.el[i], m.el[(i-1)/2]
		} else {
			break
		}
	}
}

func (m *treap) siftDown(i int) {
	for {
		if 2*i+1 >= m.size {
			break
		}

		if less(m.el[i], m.el[2*i+1]) {
			m.el[i], m.el[2*i+1] = m.el[2*i+1], m.el[i]
		} else if 2*i+2 < m.size && less(m.el[i], m.el[2*i+2]) {
			m.el[i], m.el[2*i+1] = m.el[2*i+1], m.el[i]
		} else {
			break
		}
	}
}
