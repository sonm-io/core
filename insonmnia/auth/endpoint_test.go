package auth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEndpoint(t *testing.T) {
	endpoint, err := NewEndpoint("8125721C2413d99a33E351e1F6Bb4e56b6b633FD@127.0.0.1:9090")
	require.NotNil(t, endpoint)
	require.NoError(t, err)

	assert.Equal(t, common.HexToAddress("8125721C2413d99a33E351e1F6Bb4e56b6b633FD"), endpoint.EthAddress)
	assert.Equal(t, "127.0.0.1:9090", endpoint.Endpoint)
}

func TestNewEndpointErr(t *testing.T) {
	endpoint, err := NewEndpoint("127.0.0.1:9090")
	require.Nil(t, endpoint)
	require.Error(t, err)

	endpoint, err = NewEndpoint("@127.0.0.1:9090")
	require.Nil(t, endpoint)
	require.Error(t, err)

	endpoint, err = NewEndpoint("WhatTheHell@127.0.0.1:9090")
	require.Nil(t, endpoint)
	require.Error(t, err)

	endpoint, err = NewEndpoint("8125721C2413d99a33E351e1F6Bb4e56b6b633FD")
	require.Nil(t, endpoint)
	require.Error(t, err)
}
