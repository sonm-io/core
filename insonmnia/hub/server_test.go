package hub

import (
	"context"
	"crypto/ecdsa"
	"os"
	"path"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
)

func defaultMinerMockCfg(mock *gomock.Controller) *miner.MockConfig {
	cfg := miner.NewMockConfig(mock)
	mockedwallet := util.PubKeyToAddr(getTestKey().PublicKey).Hex()
	cfg.EXPECT().HubEndpoints().AnyTimes().Return([]string{"localhost:4242"})
	cfg.EXPECT().HubEthAddr().AnyTimes().Return(mockedwallet)
	cfg.EXPECT().HubResolveEndpoints().AnyTimes().Return(false)
	cfg.EXPECT().HubResources().AnyTimes()
	cfg.EXPECT().SSH().AnyTimes()
	cfg.EXPECT().ETH().AnyTimes().Return(&accounts.EthConfig{})
	cfg.EXPECT().LocatorEndpoint().AnyTimes().Return("127.0.0.1:9090")
	cfg.EXPECT().PublicIPs().AnyTimes().Return([]string{"192.168.70.17", "46.148.198.133"})
	cfg.EXPECT().Plugins().AnyTimes().Return(plugin.Config{})
	cfg.EXPECT().StorePath().AnyTimes().Return("/tmp/sonm/worker.boltdb")
	cfg.EXPECT().StoreBucket().AnyTimes().Return("sonm")

	return cfg
}

func getTestMiner(mock *gomock.Controller) (*miner.Miner, error) {
	cfg := defaultMinerMockCfg(mock)

	ovs := miner.NewMockOverseer(mock)
	ovs.EXPECT().Info(gomock.Any()).AnyTimes().Return(map[string]miner.ContainerMetrics{}, nil)

	bl := benchmarks.NewMockBenchList(mock)
	bl.EXPECT().List().AnyTimes().Return(map[pb.DeviceType][]*pb.Benchmark{})

	return miner.NewMiner(
		cfg,
		miner.WithKey(getTestKey()),
		miner.WithOverseer(ovs),
		miner.WithBenchmarkList(bl),
	)
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
	p := path.Join(os.TempDir(), "hub_test.boltdb")

	return &Config{
		Endpoint:  "127.0.0.1:10002",
		Store:     StoreConfig{Endpoint: p, Bucket: "sonm"},
		Whitelist: WhitelistConfig{Enabled: new(bool)},
	}
}

func buildTestHub(ctrl *gomock.Controller) (*Hub, error) {
	market := getTestMarket(ctrl)
	config := getTestHubConfig()
	worker, _ := getTestMiner(ctrl)

	ctx := context.Background()

	bc := blockchain.NewMockBlockchainer(ctrl)
	bc.EXPECT().GetDealInfo(ctx, gomock.Any()).AnyTimes().Return(&pb.Deal{}, nil)

	return New(ctx, config, WithPrivateKey(key), WithMarket(market), WithBlockchain(bc), WithWorker(worker))
}

//TODO: Move this to separate test for AskPlans
func TestHubCreateRemoveSlot(t *testing.T) {
	t.Skip()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hu, err := buildTestHub(ctrl)
	assert.NoError(t, err)

	req := &pb.CreateAskPlanRequest{
		PricePerSecond: pb.NewBigIntFromInt(100),
		Slot: &pb.Slot{
			Duration:  uint64(structs.MinSlotDuration.Seconds()),
			Resources: &pb.Resources{},
		},
	}

	testCtx := context.Background()

	id, err := hu.CreateAskPlan(testCtx, req)
	assert.NoError(t, err)
	assert.True(t, id.Id != "", "ID must not be empty")

	actualSlots, err := hu.AskPlans(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.Slots), 1)

	_, err = hu.RemoveAskPlan(testCtx, id)
	assert.NoError(t, err)

	actualSlots, err = hu.AskPlans(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.Slots), 0)
}
