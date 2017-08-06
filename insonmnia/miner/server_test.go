package miner

import (
	"context"
	"errors"
	"github.com/cloudfoundry/gosigar"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerNewExtractsHubEndpoint(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().Times(1).Return("::1")
	cfg.EXPECT().HubResources().AnyTimes()
	m, err := New(ctx, cfg)

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, "::1", m.hubAddress)
}

func TestServerNewFailsWhenFailedCollectResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	cfg := NewMockConfig(mock)
	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().Times(1).Return(nil, errors.New(""))
	m, err := newMiner(ctx, cfg, collector)

	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestServerNewSavesResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ctx := context.Background()
	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().AnyTimes()
	cfg.EXPECT().HubResources().AnyTimes()
	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().Times(1).Return(&resource.OS{CPU: sigar.CpuList{}, Mem: sigar.Mem{Total: 42}}, nil)
	m, err := newMiner(ctx, cfg, collector)

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, uint64(42), m.resources.Mem.Total)
}
