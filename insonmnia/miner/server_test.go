package miner

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudfoundry/gosigar"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultMockCfg(mock *gomock.Controller) *MockConfig {
	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().AnyTimes().Return("::1")
	cfg.EXPECT().HubResources().AnyTimes()
	cfg.EXPECT().GPU().AnyTimes()
	return cfg
}

func TestServerNewExtractsHubEndpoint(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	builder := MinerBuilder{}
	builder.Config(cfg)

	m, err := builder.Build()
	cfg.EXPECT().GPU().AnyTimes()

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, "::1", m.hubAddress)
}

func TestServerNewFailsWhenFailedCollectResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)
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

	cfg := defaultMockCfg(mock)
	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().Times(1).Return(&resource.OS{CPU: sigar.CpuList{}, Mem: sigar.Mem{Total: 42}}, nil)

	builder := MinerBuilder{}
	builder.Collector(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, uint64(42), m.resources.OS.Mem.Total)
}

func TestMinerInfo(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	info := make(map[string]ContainerMetrics)
	info["id1"] = ContainerMetrics{mem: types.MemoryStats{Usage: 42, MaxUsage: 43}}
	ovs.EXPECT().Info(context.Background()).AnyTimes().Return(info, nil)

	builder := MinerBuilder{}
	builder.Config(cfg)
	builder.Overseer(ovs)

	m, err := builder.Build()
	t.Log(err)
	require.NotNil(t, m)
	require.Nil(t, err)

	m.nameMapping["id1"] = "id1"
	ret, err := m.Info(builder.ctx, &pb.MinerInfoRequest{})

	assert.NotNil(t, ret)
	assert.Nil(t, err)
	assert.Equal(t, uint64(43), ret.Stats["id1"].Memory.MaxUsage)
}

func TestMinerHandshake(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	info := make(map[string]ContainerMetrics)
	info["id1"] = ContainerMetrics{mem: types.MemoryStats{Usage: 42, MaxUsage: 43}}
	ovs.EXPECT().Info(context.Background()).AnyTimes().Return(info, nil)

	collector := resource.NewMockCollector(mock)
	collector.EXPECT().OS().AnyTimes().Return(&resource.OS{CPU: sigar.CpuList{List: make([]sigar.Cpu, 2)}, Mem: sigar.Mem{Total: 2048}}, nil)

	builder := MinerBuilder{}
	builder.Config(cfg)
	builder.Overseer(ovs)
	builder.Collector(collector)
	builder.UUID("deadbeef-cafe-dead-beef-cafedeadbeef")

	m, err := builder.Build()
	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Handshake(context.Background(), &pb.MinerHandshakeRequest{Hub: "testHub"})
	assert.NotNil(t, reply)
	assert.Nil(t, err)
	assert.Equal(t, reply, &pb.MinerHandshakeReply{Miner: "deadbeef-cafe-dead-beef-cafedeadbeef", Limits: &pb.Limits{Cores: 2, Memory: 2048}})
}

func TestMinerStart(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	ovs.EXPECT().Spool(context.Background(), Description{}).AnyTimes().Return(nil)
	status_chan := make(chan pb.TaskStatusReply_Status)
	info := ContainerInfo{
		status: &pb.TaskStatusReply{Status: pb.TaskStatusReply_RUNNING},
		ID:     "deadbeef-cafe-dead-beef-cafedeadbeef",
	}
	ovs.EXPECT().Start(context.Background(), Description{}).Times(1).Return(status_chan, info, nil)

	builder := MinerBuilder{}
	m, err := builder.Config(cfg).Overseer(ovs).Build()
	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Start(context.Background(), &pb.MinerStartRequest{Id: "test"})
	assert.NotNil(t, reply)
	assert.Nil(t, err)

	id, ok := m.getTaskIdByContainerId("deadbeef-cafe-dead-beef-cafedeadbeef")
	assert.True(t, ok)
	assert.Equal(t, id, "test")
}
