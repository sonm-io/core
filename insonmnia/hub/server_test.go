package hub

import (
	"context"
	"crypto/ecdsa"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/shirou/gopsutil/mem"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
)

func TestDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	miner, err := getTestMiner(ctrl)
	assert.NoError(t, err)
	hub := Hub{
		worker: miner,
		state: &state{
			minerCtx: &MinerCtx{
				miner: miner,
			},
		},
	}

	devices, err := hub.Devices(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(devices.CPUs))
}

func defaultMinerMockCfg(mock *gomock.Controller) *miner.MockConfig {
	cfg := miner.NewMockConfig(mock)
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

	c := []cpu.Device{{}}
	g := []*pb.GPUDevice{}
	m := &mem.VirtualMemoryStat{}

	h := &hardware.Hardware{CPU: c, GPU: g, Memory: m}
	print(h)

	hw.EXPECT().CPU().AnyTimes().Return(c, nil)
	hw.EXPECT().Memory().AnyTimes().Return(m, nil)
	hw.EXPECT().Info().AnyTimes().Return(h, nil)

	return hw
}

func getTestMiner(mock *gomock.Controller) (*miner.Miner, error) {

	cfg := defaultMinerMockCfg(mock)

	ovs := miner.NewMockOverseer(mock)

	ovs.EXPECT().Info(gomock.Any()).AnyTimes().Return(map[string]miner.ContainerMetrics{}, nil)
	hw := magicHardware(mock)

	return miner.NewMiner(cfg, miner.WithKey(getTestKey()), miner.WithOverseer(ovs), miner.WithHardware(hw))
}

var (
	key  = getTestKey()
	addr = util.PubKeyToAddr(key.PublicKey)
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := ethcrypto.GenerateKey()
	return k
}

func getTestMarket(ctrl *gomock.Controller) pb.MarketClient {
	m := pb.NewMockMarketClient(ctrl)

	ord := &pb.Order{
		Id:             "my-order-id",
		OrderType:      pb.OrderType_BID,
		PricePerSecond: pb.NewBigIntFromInt(1000),
		ByuerID:        addr.Hex(),
		Slot: &pb.Slot{
			Resources: &pb.Resources{},
		},
	}
	// TODO: fix this - it does not really call create order or cancel order
	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(ord, nil).AnyTimes()
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.Empty{}, nil).AnyTimes()
	return m
}

func getTestHubConfig() *Config {
	return &Config{
		Endpoint: "127.0.0.1:10002",
		Cluster: ClusterConfig{
			Store: StoreConfig{Type: "boltdb", Endpoint: "tmp/sonm/boltdb", Bucket: "sonm"},
		},
		Whitelist: WhitelistConfig{Enabled: new(bool)},
	}
}

func getTestCluster(ctrl *gomock.Controller) Cluster {
	cl := NewMockCluster(ctrl)
	cl.EXPECT().Synchronize(gomock.Any()).AnyTimes().Return(nil)
	cl.EXPECT().RegisterAndLoadEntity(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	return cl
}

func buildTestHub(ctrl *gomock.Controller) (*Hub, error) {
	market := getTestMarket(ctrl)
	clustr := getTestCluster(ctrl)
	config := getTestHubConfig()
	worker, _ := getTestMiner(ctrl)

	ctx := context.Background()

	bc := blockchain.NewMockBlockchainer(ctrl)
	bc.EXPECT().GetDealInfo(ctx, gomock.Any()).AnyTimes().Return(&pb.Deal{}, nil)

	return New(ctx, config, WithPrivateKey(key), WithMarket(market),
		WithCluster(clustr, nil), WithBlockchain(bc), WithWorker(worker))
}

//TODO: Move this to separate test for AskPlans
func TestHubCreateRemoveSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hu, err := buildTestHub(ctrl)
	assert.NoError(t, err)

	req := &pb.InsertSlotRequest{
		PricePerSecond: pb.NewBigIntFromInt(100),
		Slot: &pb.Slot{
			Duration:  uint64(structs.MinSlotDuration.Seconds()),
			Resources: &pb.Resources{},
		},
	}

	testCtx := context.Background()

	id, err := hu.InsertSlot(testCtx, req)
	assert.NoError(t, err)
	assert.True(t, id.Id != "", "ID must not be empty")

	actualSlots, err := hu.Slots(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.Slots), 1)

	_, err = hu.RemoveSlot(testCtx, id)
	assert.NoError(t, err)

	actualSlots, err = hu.Slots(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.Slots), 0)
}
