package auth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAddr(t *testing.T) {
	addr, err := ParseAddr("8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090")
	require.NotNil(t, addr)
	require.NoError(t, err)

	eth, err := addr.ETH()
	require.NoError(t, err)

	net, err := addr.Addr()
	require.NoError(t, err)

	assert.Equal(t, common.HexToAddress("8125721C2413d99a33E351e1F6Bb4e56b6b633FD"), eth)
	assert.Equal(t, "127.0.0.1:9090", net)
}

func TestNewAddrOnlyNet(t *testing.T) {
	addr, err := ParseAddr("127.0.0.1:9090")
	require.NotNil(t, addr)
	require.NoError(t, err)

	net, err := addr.Addr()
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1:9090", net)
}

func TestNewAddrOnlyETH(t *testing.T) {
	addr, err := ParseAddr("8125721C2413d99a33E351e1F6Bb4e56b6b633FD")
	require.NotNil(t, addr)
	require.NoError(t, err)

	eth, err := addr.ETH()
	require.NoError(t, err)

	assert.Equal(t, common.HexToAddress("8125721C2413d99a33E351e1F6Bb4e56b6b633FD"), eth)
}

func TestNewAddrErr(t *testing.T) {
	endpoint, err := ParseAddr("@127.0.0.1:9090")
	require.Nil(t, endpoint)
	require.Error(t, err)

	endpoint, err = ParseAddr("WhatTheHell@127.0.0.1:9090")
	require.Nil(t, endpoint)
	require.Error(t, err)
}

func TestAddrMarshalText(t *testing.T) {
	addr, err := ParseAddr("8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090")
	require.NotNil(t, addr)
	require.NoError(t, err)

	data, err := addr.MarshalText()

	require.NoError(t, err)
	require.NotNil(t, data)

	// Note that "0x" prefix added.
	assert.Equal(t, []byte("0x8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090"), data)
}
