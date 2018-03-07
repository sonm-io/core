package miner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/shirou/gopsutil/mem"
	accounts "github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

var (
	key = getTestKey()
	_   = setupTestResponder()
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := ethcrypto.GenerateKey()
	return k
}

func setupTestResponder() *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	l, _ := net.Listen("tcp", "127.0.0.1:4242")
	ts.Listener = l
	ts.Start()

	return ts
}

func defaultMockCfg(mock *gomock.Controller) *MockConfig {
	cfg := NewMockConfig(mock)
	mockedwallet := util.PubKeyToAddr(getTestKey().PublicKey).Hex()
	cfg.EXPECT().HubEndpoints().AnyTimes().Return([]string{"localhost:4242"})
	cfg.EXPECT().HubEthAddr().AnyTimes().Return(mockedwallet)
	cfg.EXPECT().HubResolveEndpoints().AnyTimes().Return(false)
	cfg.EXPECT().HubResources().AnyTimes()
	cfg.EXPECT().Firewall().AnyTimes()
	cfg.EXPECT().SSH().AnyTimes()
	cfg.EXPECT().ETH().AnyTimes().Return(&accounts.EthConfig{})
	cfg.EXPECT().LocatorEndpoint().AnyTimes().Return("127.0.0.1:9090")
	cfg.EXPECT().PublicIPs().AnyTimes().Return([]string{"192.168.70.17", "46.148.198.133"})
	cfg.EXPECT().Plugins().AnyTimes().Return(plugin.Config{})
	return cfg
}

func magicHardware(ctrl *gomock.Controller) hardware.Info {
	hw := hardware.NewMockInfo(ctrl)

	c := []cpu.Device{}
	g := []*pb.GPUDevice{}
	m := &mem.VirtualMemoryStat{}

	h := &hardware.Hardware{CPU: c, GPU: g, Memory: m}

	hw.EXPECT().CPU().AnyTimes().Return(c, nil)
	hw.EXPECT().Memory().AnyTimes().Return(m, nil)
	hw.EXPECT().Info().AnyTimes().Return(h, nil)

	return hw
}

func TestServerNewFailsWhenFailedCollectResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ovs := NewMockOverseer(mock)
	cfg := defaultMockCfg(mock)
	collector := hardware.NewMockInfo(mock)
	collector.EXPECT().Info().Times(1).Return(nil, errors.New(""))
	locator := pb.NewMockLocatorClient(mock)

	m, err := NewMiner(cfg, WithKey(key), WithHardware(collector), WithLocatorClient(locator),
		WithOverseer(ovs))

	assert.Nil(t, m)
	assert.Error(t, err)
}

func TestServerNewSavesResources(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	ovs := NewMockOverseer(mock)
	cfg := defaultMockCfg(mock)
	collector := hardware.NewMockInfo(mock)
	collector.EXPECT().Info().Times(1).Return(&hardware.Hardware{
		CPU:    []cpu.Device{},
		Memory: &mem.VirtualMemoryStat{Total: 42},
	}, nil)
	locator := pb.NewMockLocatorClient(mock)

	m, err := NewMiner(cfg, WithKey(key), WithHardware(collector),
		WithLocatorClient(locator), WithOverseer(ovs))

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
	ovs.EXPECT().Info(gomock.Any()).AnyTimes().Return(info, nil)
	locator := pb.NewMockLocatorClient(mock)
	hw := magicHardware(mock)

	m, err := NewMiner(cfg, WithKey(key), WithOverseer(ovs), WithLocatorClient(locator), WithHardware(hw))
	t.Log(err)
	require.NotNil(t, m)
	require.Nil(t, err)

	m.nameMapping["id1"] = "id1"
	ret, err := m.Info(m.ctx, &pb.Empty{})

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

	collector := hardware.NewMockInfo(mock)
	collector.EXPECT().Info().AnyTimes().Return(&hardware.Hardware{
		CPU:    []cpu.Device{{Cores: 2}},
		Memory: &mem.VirtualMemoryStat{Total: 2048},
	}, nil)
	locator := pb.NewMockLocatorClient(mock)

	m, err := NewMiner(cfg, WithKey(key), WithHardware(collector),
		WithOverseer(ovs), WithUUID("deadbeef-cafe-dead-beef-cafedeadbeef"), WithLocatorClient(locator))

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
	locator := pb.NewMockLocatorClient(mock)
	hw := magicHardware(mock)

	m, err := NewMiner(cfg, WithKey(key), WithOverseer(ovs),
		WithUUID("deadbeef-cafe-dead-beef-cafedeadbeef"), WithLocatorClient(locator), WithHardware(hw))

	require.NotNil(t, m)
	require.Nil(t, err)
	reply, err := m.Start(context.Background(), &pb.MinerStartRequest{Id: "test", Resources: &pb.TaskResourceRequirements{}, Container: &pb.Container{}})
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
		"keY":  "12345",
		"key4": "",
	}

	description := Description{Env: vars}

	assert.Contains(t, description.FormatEnv(), "key1=value1")
	assert.Contains(t, description.FormatEnv(), "KEY2=VALUE2")
	assert.Contains(t, description.FormatEnv(), "keY=12345")
	assert.Contains(t, description.FormatEnv(), "key4=")
}
