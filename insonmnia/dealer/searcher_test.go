package dealer

import (
	"math/big"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

var ordersForTest = []*sonm.Order{
	{Id: "1", PricePerSecond: sonm.NewBigIntFromInt(1000), Slot: &sonm.Slot{Duration: 1}},
	{Id: "2", PricePerSecond: sonm.NewBigIntFromInt(500), Slot: &sonm.Slot{Duration: 1}},
	{Id: "3", PricePerSecond: sonm.NewBigIntFromInt(200), Slot: &sonm.Slot{Duration: 1}},
	{Id: "4", PricePerSecond: sonm.NewBigIntFromInt(100), Slot: &sonm.Slot{Duration: 1}},
}

func TestAskSearcher_filterByPrice(t *testing.T) {
	s := &askSearcher{}
	bal := big.NewInt(250)

	out, err := s.filterByPrice(ordersForTest, bal)
	assert.NoError(t, err)
	assert.Len(t, out, 2)

	assert.Equal(t, "3", out[0].Id)
	assert.Equal(t, "4", out[1].Id)
}

func TestAskSearcher_filterByAllowance(t *testing.T) {
	s := &askSearcher{}
	alw := big.NewInt(550)

	out, err := s.filterByAllowance(ordersForTest, alw)
	assert.NoError(t, err)
	assert.Len(t, out, 3)

	// first order must be filtered
	assert.Equal(t, "2", out[0].Id)
	assert.Equal(t, "3", out[1].Id)
	assert.Equal(t, "4", out[2].Id)
}

func TestAskSearcher_filterNil(t *testing.T) {
	s := &askSearcher{}
	_, err := s.filterByPriceAndAllowance(nil, nil, nil)
	assert.Error(t, err)
}

func TestAskSearcher_filtersChain(t *testing.T) {
	s := &askSearcher{}

	out, err := s.filterByPriceAndAllowance(ordersForTest, big.NewInt(550), big.NewInt(300))
	assert.NoError(t, err)
	assert.Len(t, out, 2)

	out, err = s.filterByPriceAndAllowance(ordersForTest, big.NewInt(220), big.NewInt(9999))
	assert.NoError(t, err)
	assert.Len(t, out, 2)

	out, err = s.filterByPriceAndAllowance(ordersForTest, big.NewInt(9999), big.NewInt(50))
	assert.EqualError(t, err, "no orders fit into allowance")

	out, err = s.filterByPriceAndAllowance(ordersForTest, big.NewInt(50), big.NewInt(9999))
	assert.EqualError(t, err, "no orders fit into available balance")
}

func TestNewSearchFilter(t *testing.T) {
	_, err := NewSearchFilter(nil, big.NewInt(1), big.NewInt(2))
	assert.EqualError(t, err, "order cannot be nil")

	_, err = NewSearchFilter(&sonm.Order{}, nil, big.NewInt(2))
	assert.EqualError(t, err, "balance cannot be nil")

	_, err = NewSearchFilter(&sonm.Order{}, big.NewInt(1), nil)
	assert.EqualError(t, err, "allowance cannot be nil")

	_, err = NewSearchFilter(&sonm.Order{}, big.NewInt(1), big.NewInt(2))
	assert.NoError(t, err)
}
