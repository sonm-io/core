package connor

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	c := &Connor{}
	p := &PoolModule{}
	r := &ProfitableModule{}

	trade := NewTraderModules(c, p, r)
	v := trade.FloatToBigInt(float64(1e10))
	fmt.Println(v)
}

func TestPoolModule_AdvancedPoolHashrateTracking(t *testing.T) {
	c := &Connor{}
}
