package sonm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
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

func TestBigInt_MarshalYaml(t *testing.T) {
	intVal, _ := big.NewInt(0).SetString("42", 10)
	in := map[string]*BigInt{
		"test": NewBigInt(intVal),
	}

	in2 := map[string]BigInt{
		"test": *NewBigInt(intVal),
	}

	text, err := yaml.Marshal(&in)
	assert.NoError(t, err)

	text2, err := yaml.Marshal(&in2)
	assert.NoError(t, err)

	assert.Equal(t, text, text2)
	assert.Equal(t, `test: "42"`+"\n", string(text))
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

func TestBigInt_ToPriceString(t *testing.T) {
	tests := []struct {
		in  float64
		out string
	}{
		{in: 1e18, out: "1"},
		{in: 1e15, out: "0.001"},
		{in: 1e20, out: "100"},
		{in: 1.2e18, out: "1.2"},
		{in: 1.3333333e18, out: "1.3333333"},
		{in: 1e12, out: "0.000001"},
		{in: 1e10, out: "0.00000001"},
		{in: 1e7, out: "0.00000000001"},
		{in: 1e4, out: "0.00000000000001"},
		{in: 1, out: "0.000000000000000001"},
		{in: 1234, out: "0.000000000000001234"},
	}

	for _, tt := range tests {
		bint, _ := big.NewFloat(tt.in).Int(nil)
		v := NewBigInt(bint)
		out := v.ToPriceString()
		assert.Equal(t, tt.out, out, fmt.Sprintf("expect %v == %s USD (result is \"%s\")", tt.in, tt.out, out))
	}
}
