package sonm

import (
	"math/big"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBigInt(t *testing.T) {
	v := big.NewInt(42000000000)
	price := NewBigInt(v)

	assert.Equal(t, v, price.Unwrap())
}

func TestNewBigIntFromString(t *testing.T) {
	price, err := NewBigIntFromString("42000000001")

	require.NotNil(t, price)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(42000000001), price.Unwrap())
}

func TestBigIntString(t *testing.T) {
	price := NewBigInt(big.NewInt(42000000002))
	assert.Equal(t, "42000000002", price.Unwrap().String())
}

func TestBigIntUnmarshal(t *testing.T) {
	in := []byte(`{"test": "100000000000000000000000"}`)
	r := make(map[string]*BigInt)

	err := json.Unmarshal(in, &r)
	assert.NoError(t, err)
	assert.Contains(t, r, "test")

	intVal, _ := NewBigIntFromString("100000000000000000000000")
	compare := r["test"].Cmp(intVal)
	assert.Equal(t, compare, 0)
}

func TestBigIntMarshal(t *testing.T) {
	intVal, _ := big.NewInt(0).SetString("100000000000000000000000", 10)
	in := map[string]*BigInt{
		"test": NewBigInt(intVal),
	}

	b, err := json.Marshal(&in)
	assert.NoError(t, err)

	assert.Equal(t, `{"test":"100000000000000000000000"}`, string(b))
}
