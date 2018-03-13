package dealer

import (
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

var input = []*sonm.Order{
	{PricePerSecond: sonm.NewBigIntFromInt(100)},
	{PricePerSecond: sonm.NewBigIntFromInt(20)},
	{PricePerSecond: sonm.NewBigIntFromInt(10)},
	{PricePerSecond: sonm.NewBigIntFromInt(200)},
	{PricePerSecond: sonm.NewBigIntFromInt(50)},
}

func TestBidMatcher_Match(t *testing.T) {
	m := NewBidMatcher()
	out, err := m.Match(input)

	assert.NoError(t, err)
	assert.Equal(t, uint64(10), out.PricePerSecond.Unwrap().Uint64())
}

func TestBidMatcher_MatchNilEmpty(t *testing.T) {
	m := NewBidMatcher()

	out, err := m.Match(nil)
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)

	out, err = m.Match([]*sonm.Order{})
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)
}

func TestAskMatcher_Match(t *testing.T) {
	m := NewAskMatcher()
	out, err := m.Match(input)

	assert.NoError(t, err)
	assert.Equal(t, uint64(200), out.PricePerSecond.Unwrap().Uint64())
}

func TestAskMatcher_MatchNilEmpty(t *testing.T) {
	m := NewAskMatcher()

	out, err := m.Match(nil)
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)

	out, err = m.Match([]*sonm.Order{})
	assert.EqualError(t, err, "no orders provided")
	assert.Nil(t, out)
}
