package sonm

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestConvertToPrice(t *testing.T) {
	tests := []struct {
		in       string
		expected *big.Int
	}{
		{
			in:       "100 USD/s",
			expected: big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(100)),
		},
		{
			in:       "2 USD/h",
			expected: big.NewInt(0).Quo(big.NewInt(params.Ether), big.NewInt(1800)),
		},
		{
			in:       "12USD/s",
			expected: big.NewInt(0).Mul(big.NewInt(params.Ether), big.NewInt(12)),
		},
	}

	for _, tt := range tests {
		pr := Price{}
		err := pr.LoadFromString(tt.in)
		require.NoError(t, err)
		assert.Equal(t, tt.expected, pr.PerSecond.Unwrap(), tt.in)
	}
}

func TestEthAddress_MarshalText(t *testing.T) {
	in := struct {
		Addr *EthAddress `yaml:"addr" json:"addr"`
	}{
		Addr: NewEthAddress(common.HexToAddress("0x123")),
	}

	ya, err := yaml.Marshal(in)
	require.NoError(t, err)

	assert.Equal(t, "addr: \"0x0000000000000000000000000000000000000123\"\n", string(ya),
		"YAML marshaller")

	js, err := json.Marshal(in)
	require.NoError(t, err)
	assert.Equal(t, `{"addr":"0x0000000000000000000000000000000000000123"}`, string(js),
		"JSON marshaller")
}

func TestEthAddress_MarshalTextByValue(t *testing.T) {
	in := struct {
		Addr EthAddress `yaml:"addr" json:"addr"`
	}{}

	tmp := NewEthAddress(common.HexToAddress("0x123"))
	in.Addr = *tmp

	ya, err := yaml.Marshal(in)
	require.NoError(t, err)

	assert.Equal(t, "addr: \"0x0000000000000000000000000000000000000123\"\n", string(ya),
		"YAML marshaller")

	js, err := json.Marshal(in)
	require.NoError(t, err)
	assert.Equal(t, `{"addr":"0x0000000000000000000000000000000000000123"}`, string(js),
		"JSON marshaller")
}

func TestEthAddress_UnmarshalText(t *testing.T) {
	in := []byte(`{"addr":"0x1234567891011121314151617181920212223242"}`)
	recv := struct {
		Addr *EthAddress `json:"addr" yaml:"addr"`
	}{}

	err := yaml.Unmarshal(in, &recv)
	require.NoError(t, err)
	assert.Equal(t, "0x1234567891011121314151617181920212223242", recv.Addr.Unwrap().Hex(),
		"YAML marshall")

	// reset prev value
	recv.Addr = nil
	err = json.Unmarshal(in, &recv)
	require.NoError(t, err)
	assert.Equal(t, "0x1234567891011121314151617181920212223242", recv.Addr.Unwrap().Hex(),
		"JSON marshall")
}
