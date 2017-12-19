package miner

import (
	"crypto/ecdsa"
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/shirou/gopsutil/mem"
	accounts "github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

var (
	key = getTestKey()
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := ethcrypto.GenerateKey()
	return k
}

func defaultMockCfg(mock *gomock.Controller) *MockConfig {
	cfg := NewMockConfig(mock)
	mockedwallet := util.PubKeyToAddr(getTestKey().PublicKey).Hex()
	cfg.EXPECT().HubEndpoint().AnyTimes().Return(mockedwallet + "@localhost:4242")
	cfg.EXPECT().HubResources().AnyTimes()
	cfg.EXPECT().Firewall().AnyTimes()
	cfg.EXPECT().GPU().AnyTimes()
	cfg.EXPECT().SSH().AnyTimes()
	cfg.EXPECT().ETH().AnyTimes().Return(&accounts.EthConfig{})
	cfg.EXPECT().PublicIPs().AnyTimes().Return([]string{"192.168.70.17", "46.148.198.133"})
	return cfg
}

func TestServerNewExtractsHubEndpoint(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	builder := MinerBuilder{key: key}
	builder.Config(cfg)

	m, err := builder.Build()
	cfg.EXPECT().GPU().AnyTimes()

	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Equal(t, "localhost:4242", m.hubAddress)
}

func TestServerNewFailsWhenFailedCollectResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)
	collector := hardware.NewMockHardwareInfo(mock)
	collector.EXPECT().Info().Times(1).Return(nil, errors.New(""))

	builder := MinerBuilder{key: key}
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
		CPU:    []cpu.Device{},
		Memory: &mem.VirtualMemoryStat{Total: 42},
	}, nil)

	builder := MinerBuilder{key: key}
	builder.Hardware(collector)
	builder.Config(cfg)
	m, err := builder.Build()

	assert.NotNil(t, m)
	require.Nil(t, err)
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

	builder := MinerBuilder{key: key}
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
		CPU:    []cpu.Device{{Cores: 2}},
		Memory: &mem.VirtualMemoryStat{Total: 2048},
	}, nil)

	builder := MinerBuilder{key: key}
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
	statusChan := make(chan pb.TaskStatusReply_Status)
	info := ContainerInfo{
		status: &pb.TaskStatusReply{Status: pb.TaskStatusReply_RUNNING},
		ID:     "deadbeef-cafe-dead-beef-cafedeadbeef",
	}
	ovs.EXPECT().Start(gomock.Any(), gomock.Any()).Times(1).Return(statusChan, info, nil)

	builder := MinerBuilder{key: key}
	m, err := builder.Config(cfg).Overseer(ovs).Build()
	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Start(context.Background(), &pb.MinerStartRequest{Id: "test", Resources: &pb.TaskResourceRequirements{}})
	require.NoError(t, err)
	require.NotNil(t, reply)

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

	description := Description{Env: vars}

	assert.Contains(t, description.FormatEnv(), "KEY1=value1")
	assert.Contains(t, description.FormatEnv(), "KEY2=VALUE2")
	assert.Contains(t, description.FormatEnv(), "KEY3=12345")
	assert.Contains(t, description.FormatEnv(), "KEY4=")
}
