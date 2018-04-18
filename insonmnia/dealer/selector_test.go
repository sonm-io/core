package dealer

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

var input = []*sonm.Order{
	{Price: sonm.NewBigIntFromInt(100)},
	{Price: sonm.NewBigIntFromInt(20)},
	{Price: sonm.NewBigIntFromInt(10)},
	{Price: sonm.NewBigIntFromInt(200)},
	{Price: sonm.NewBigIntFromInt(50)},
}

func TestAskSelector_Select(t *testing.T) {
	m := NewAskSelector()
	out, err := m.Select(input)

	assert.NoError(t, err)
	assert.Equal(t, uint64(10), out.Price.Unwrap().Uint64())
}

func TestAskSelector_SelectNilEmpty(t *testing.T) {
	m := NewAskSelector()

	out, err := m.Select(nil)
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)

	out, err = m.Select([]*sonm.Order{})
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)
}

func TestBidSelector_Select(t *testing.T) {
	m := NewBidSelector()
	out, err := m.Select(input)

	assert.NoError(t, err)
	assert.Equal(t, uint64(200), out.Price.Unwrap().Uint64())
}

func TestBidSelector_SelectNilEmpty(t *testing.T) {
	m := NewBidSelector()

	out, err := m.Select(nil)
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)

	out, err = m.Select([]*sonm.Order{})
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)
}
