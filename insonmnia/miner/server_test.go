package miner

import (
	"context"
	"errors"
	"github.com/cloudfoundry/gosigar"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServerNewExtractsHubEndpoint(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().Times(1).Return("::1")
	cfg.EXPECT().HubResources().AnyTimes()
	builder := MinerBuilder{}
	builder.Config(cfg)

	m, err := builder.Build()

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, "::1", m.hubAddress)
}

func TestServerNewFailsWhenFailedCollectResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := NewMockConfig(mock)
	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().Times(1).Return(nil, errors.New(""))

	builder := MinerBuilder{}
	builder.Collector(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestServerNewSavesResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().AnyTimes()
	cfg.EXPECT().HubResources().AnyTimes()
	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().Times(1).Return(&resource.OS{CPU: sigar.CpuList{}, Mem: sigar.Mem{Total: 42}}, nil)

	builder := MinerBuilder{}
	builder.Collector(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, uint64(42), m.resources.Mem.Total)
}

func TestMinerInfo(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().AnyTimes()
	cfg.EXPECT().HubResources().AnyTimes()

	ovs := NewMockOverseer(mock)
	info := make(map[string]ContainerMetrics)
	info["id1"] = ContainerMetrics{mem: types.MemoryStats{Usage: 42, MaxUsage: 43}}
	ovs.EXPECT().Info(context.Background()).AnyTimes().Return(info, nil)

	builder := MinerBuilder{}
	builder.Config(cfg)
	builder.Overseer(ovs)

	m, err := builder.Build()
	m.nameMapping["id1"] = "id1"
	ret, err := m.Info(builder.ctx, &pb.MinerInfoRequest{})

	assert.NotNil(t, ret)
	assert.Nil(t, err)
	assert.Equal(t, uint64(43), ret.Stats["id1"].Memory.MaxUsage)
}
