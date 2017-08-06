package miner

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerNewExtractsHubEndpoint(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().Times(1).Return("::1")
	m, err := New(ctx, cfg)

	assert.Nil(t, err)
	assert.Equal(t, "::1", m.hubAddress)
}
