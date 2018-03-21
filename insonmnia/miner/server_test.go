package miner

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	accounts "github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	cfg.EXPECT().StorePath().AnyTimes().Return("/tmp/sonm/worker_test.boltdb")
	cfg.EXPECT().StoreBucket().AnyTimes().Return("sonm")
	return cfg
}

func TestMinerInfo(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	cfg := defaultMockCfg(mock)

	ovs := NewMockOverseer(mock)
	info := make(map[string]ContainerMetrics)
	info["id1"] = ContainerMetrics{mem: types.MemoryStats{Usage: 42, MaxUsage: 43}}
	ovs.EXPECT().Info(gomock.Any()).AnyTimes().Return(info, nil)

	m, err := NewMiner(cfg, WithKey(key), WithOverseer(ovs))
	t.Log(err)
	require.NotNil(t, m)
	require.Nil(t, err)

	m.nameMapping["id1"] = "id1"
	ret, err := m.Info(m.ctx, &pb.Empty{})

	assert.NotNil(t, ret)
	assert.Nil(t, err)
	assert.Equal(t, uint64(43), ret.Usage["id1"].Memory.MaxUsage)
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
