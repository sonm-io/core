package sonm

import (
	"encoding/json"
	"math/big"
	"testing"

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

func TestBigIntMarshalJSON(t *testing.T) {
	intVal, _ := big.NewInt(0).SetString("100000000000000000000000", 10)
	in := map[string]*BigInt{
		"test": NewBigInt(intVal),
	}

	b, err := json.Marshal(&in)
	assert.NoError(t, err)
	assert.Equal(t, `{"test":"100000000000000000000000"}`, string(b))
}

func TestBigInt_MarshalJSONByValue(t *testing.T) {
	// char 'd' has ascii code 100
	in := map[string]BigInt{"value": {Abs: []byte{'d'}}}

	b, err := json.Marshal(&in)
	assert.NoError(t, err)
	assert.Equal(t, `{"value":"100"}`, string(b))
}

func TestBigInt_UnmarshalJSONByValue(t *testing.T) {
	in := []byte(`{"test":"100"}`)
	r := make(map[string]BigInt)

	err := json.Unmarshal(in, &r)
	assert.NoError(t, err)
	assert.Contains(t, r, "test")

	val := r["test"]
	compare := NewBigIntFromInt(100).Cmp(&val)
	assert.Equal(t, 0, compare)
}

func TestBigInt_UnmarshalJSON(t *testing.T) {
	in := []byte(`{"test": "100000000000000000000000"}`)
	r := make(map[string]*BigInt)

	err := json.Unmarshal(in, &r)
	assert.NoError(t, err)
	assert.Contains(t, r, "test")

	intVal, _ := NewBigIntFromString("100000000000000000000000")
	compare := r["test"].Cmp(intVal)
	assert.Equal(t, 0, compare)
}

func TestBigInt_MarshalEmptyValue(t *testing.T) {
	in := map[string]BigInt{"value": {}}

	b, err := json.Marshal(&in)
	assert.NoError(t, err)
	assert.Equal(t, `{"value":"0"}`, string(b))
}
