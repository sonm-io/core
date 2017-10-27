package miner

import (
	"errors"
	"golang.org/x/net/context"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/insonmnia/hardware"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultMockCfg(mock *gomock.Controller) *MockConfig {
	cfg := NewMockConfig(mock)
	cfg.EXPECT().HubEndpoint().AnyTimes().Return("::1")
	cfg.EXPECT().HubResources().AnyTimes()
	cfg.EXPECT().Firewall().AnyTimes()
	cfg.EXPECT().GPU().AnyTimes()
	cfg.EXPECT().SSH().AnyTimes()
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
	collector := hardware.NewMockHardwareInfo(mock)
	collector.EXPECT().Info().Times(1).Return(nil, errors.New(""))

	builder := MinerBuilder{}
	builder.Hardware(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestServerNewSavesResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)
	collector := hardware.NewMockHardwareInfo(mock)
	collector.EXPECT().Info().Times(1).Return(&hardware.Hardware{
		CPU:    []cpu.InfoStat{},
		Memory: &mem.VirtualMemoryStat{Total: 42},
	}, nil)

	builder := MinerBuilder{}
	builder.Hardware(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.NotNil(t, m)
	assert.Nil(t, err)
	assert.Equal(t, uint64(42), m.resources.OS.Memory.Total)
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
	ret, err := m.Info(builder.ctx, &pb.Empty{})

	assert.NotNil(t, ret)
	assert.Nil(t, err)
	assert.Equal(t, uint64(43), ret.Usage["id1"].Memory.MaxUsage)
}

func TestMinerHandshake(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	info := make(map[string]ContainerMetrics)
	info["id1"] = ContainerMetrics{mem: types.MemoryStats{Usage: 42, MaxUsage: 43}}
	ovs.EXPECT().Info(context.Background()).AnyTimes().Return(info, nil)

	collector := hardware.NewMockHardwareInfo(mock)
	collector.EXPECT().Info().AnyTimes().Return(&hardware.Hardware{
		CPU:    []cpu.InfoStat{{Cores: 2}},
		Memory: &mem.VirtualMemoryStat{Total: 2048},
	}, nil)

	builder := MinerBuilder{}
	builder.Config(cfg)
	builder.Overseer(ovs)
	builder.Hardware(collector)
	builder.UUID("deadbeef-cafe-dead-beef-cafedeadbeef")

	m, err := builder.Build()
	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Handshake(context.Background(), &pb.MinerHandshakeRequest{Hub: "testHub"})
	assert.NotNil(t, reply)
	assert.Nil(t, err)
	assert.Equal(t, reply.Miner, "deadbeef-cafe-dead-beef-cafedeadbeef")
	assert.Equal(t, int32(2), reply.Capabilities.Cpu[0].Cores)
	assert.Equal(t, uint64(2048), reply.Capabilities.Mem.Total)
}

func TestMinerStart(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	ovs.EXPECT().Spool(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	status_chan := make(chan pb.TaskStatusReply_Status)
	info := ContainerInfo{
		status: &pb.TaskStatusReply{Status: pb.TaskStatusReply_RUNNING},
		ID:     "deadbeef-cafe-dead-beef-cafedeadbeef",
	}
	ovs.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1).Return(status_chan, info, nil)

	builder := MinerBuilder{}
	m, err := builder.Config(cfg).Overseer(ovs).Build()
	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Start(context.Background(), &pb.MinerStartRequest{Id: "test"})
	require.NotNil(t, reply)
	require.Nil(t, err)

	id, ok := m.getTaskIdByContainerId("deadbeef-cafe-dead-beef-cafedeadbeef")
	assert.True(t, ok)
	assert.Equal(t, id, "test")
}

func TestTransformEnvVars(t *testing.T) {
	vars := map[string]string{
		"key1": "value1",
		"KEY2": "VALUE2",
		"key3": "12345",
		"key4": "",
	}

	transformed := transformEnvVariables(vars)

	assert.Contains(t, transformed, "KEY1=value1")
	assert.Contains(t, transformed, "KEY2=VALUE2")
	assert.Contains(t, transformed, "KEY3=12345")
	assert.Contains(t, transformed, "KEY4=")
}
