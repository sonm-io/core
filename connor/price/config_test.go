package price

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestParse_CoinMarkerCapProvider(t *testing.T) {
	raw := []byte(`
type: cmc
url: "http://127.0.0.1:8000/price.json"
what_to_mine_id: 101 # 151
update_interval: 3m
`)

	f := &SourceConfig{}
	err := yaml.Unmarshal(raw, f)
	require.NoError(t, err)

	cfg := f.Config().(*CoinMarketCapConfig)
	assert.Equal(t, "http://127.0.0.1:8000/price.json", cfg.URL)
	assert.Equal(t, 3*time.Minute, cfg.Interval)
	assert.Equal(t, 101, cfg.WhatToMineID)

	up, ok := f.Init(1).(Updateable)
	require.True(t, ok)
	assert.Equal(t, 3*time.Minute, up.Interval())
}

func TestParse_CoinMarketCapProvider_Validate(t *testing.T) {
	raw := []byte(`
type: cmc
url: "http://127.0.0.1:8000/price.json"
what_to_mine_id: 123 # 101, 151
update_interval: 1
`)

	f := &SourceConfig{}
	err := yaml.Unmarshal(raw, f)
	require.Error(t, err)
}

func TestParse_StaticProvider(t *testing.T) {
	raw := []byte(`
type: static
price: 134
`)

	f := &SourceConfig{}
	err := yaml.Unmarshal(raw, f)
	require.NoError(t, err)

	cfg := f.Config().(*StaticProviderConfig)
	assert.Equal(t, int64(134), cfg.Price)

	_, ok := f.Init(1).(Updateable)
	require.False(t, ok)
}

func TestParse_StaticProvider_Validate(t *testing.T) {
	raw := []byte(`
type: static
price: -124
`)

	f := &SourceConfig{}
	err := yaml.Unmarshal(raw, f)
	require.Error(t, err)
}
