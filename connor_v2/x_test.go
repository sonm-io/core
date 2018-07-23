package connor

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDivideOrders(t *testing.T) {
	c := &Connor{}

	ex1, _ := NewCorderFromParams("ETH", big.NewInt(1), 100)
	ex2, _ := NewCorderFromParams("ETH", big.NewInt(2), 200)
	ex3, _ := NewCorderFromParams("ETH", big.NewInt(3), 300)

	req0, _ := NewCorderFromParams("ETH", big.NewInt(1), 50)
	req1, _ := NewCorderFromParams("ETH", big.NewInt(1), 100)
	req2, _ := NewCorderFromParams("ETH", big.NewInt(2), 200)
	req3, _ := NewCorderFromParams("ETH", big.NewInt(3), 300)
	req4, _ := NewCorderFromParams("ETH", big.NewInt(4), 400)

	exiting := []*Corder{ex1, ex2, ex3}
	required := []*Corder{req0, req1, req2, req3, req4}

	set := c.divideOrdersSets(exiting, required)
	assert.Len(t, set.toRestore, 3)
	assert.Len(t, set.toCreate, 2)
}

func TestGetTargetCorders(t *testing.T) {
	// todo:
}
