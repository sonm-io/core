package hub

import (
	"context"
	"crypto/ecdsa"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/hardware/cpu"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
)

func TestDevices(t *testing.T) {
	// GPU characteristics shared between miners.
	gpuDevice, err := gpu.NewDevice("a", "b", 1488, 660)
	assert.NoError(t, err)

	hub := Hub{
		miners: map[string]*MinerCtx{
			"miner1": {
				uuid: "miner1",
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 64}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
			"miner2": {
				uuid: "miner2",
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 65}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
		},
	}

	devices, err := hub.Devices(context.Background(), &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(devices.CPUs), 2)
	assert.Equal(t, len(devices.GPUs), 1)
}

func TestMinerDevices(t *testing.T) {
	gpuDevice, err := gpu.NewDevice("a", "b", 1488, 660)
	assert.NoError(t, err)

	hub := Hub{
		miners: map[string]*MinerCtx{
			"miner1": {
				uuid: "miner1",
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 64}},
					GPU: []gpu.Device{gpuDevice},
				},
			},

			"miner2": {
				uuid: "miner2",
				capabilities: &hardware.Hardware{
					CPU: []cpu.Device{{CPU: 65}},
					GPU: []gpu.Device{gpuDevice},
				},
			},
		},
	}

	devices, err := hub.MinerDevices(context.Background(), &pb.ID{Id: "miner1"})
	assert.NoError(t, err)
	assert.Equal(t, len(devices.CPUs), 1)
	assert.Equal(t, len(devices.GPUs), 1)

	devices, err = hub.MinerDevices(context.Background(), &pb.ID{Id: "span"})
	assert.Error(t, err)
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
		Id:        "my-order-id",
		OrderType: pb.OrderType_BID,
		Price:     "1000",
		ByuerID:   addr.Hex(),
		Slot: &pb.Slot{
			Resources: &pb.Resources{},
		},
	}
	m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(ord, nil).MinTimes(1)
	//m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).AnyTimes().
	//	Return(&pb.GetOrdersReply{Orders: []*pb.Order{ord}}, nil)
	m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).AnyTimes().
		Return(&pb.Empty{}, nil).MinTimes(1)
	return m
}

func getTestHubConfig() *Config {
	return &Config{
		Endpoint: "127.0.0.1:10002",
		Cluster: ClusterConfig{
			Endpoint: "127.0.0.1:10001",
			Failover: false,
			Store:    StoreConfig{Type: "boltdb", Endpoint: "tmp/sonm/boltdb", Bucket: "sonm"},
		},
	}
}

func getTestCluster(ctrl *gomock.Controller) Cluster {
	cl := NewMockCluster(ctrl)
	cl.EXPECT().Synchronize(gomock.Any()).AnyTimes().Return(nil)
	return cl
}

func buildTestHub(ctrl *gomock.Controller) (*Hub, error) {
	market := getTestMarket(ctrl)
	clustr := getTestCluster(ctrl)
	config := getTestHubConfig()

	bc := blockchain.NewMockBlockchainer(ctrl)
	bc.EXPECT().GetDealInfo(gomock.Any()).AnyTimes().Return(&pb.Deal{}, nil)

	return New(context.Background(), config, "",
		WithPrivateKey(key), WithMarket(market), WithCluster(clustr, nil), WithBlockchain(bc))
}

func TestHubCreateRemoveSlot(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hu, err := buildTestHub(ctrl)
	assert.NoError(t, err)

	req := &pb.InsertSlotRequest{
		Price: "100",
		Slot: &pb.Slot{
			Resources: &pb.Resources{},
		},
	}

	testCtx := context.Background()

	id, err := hu.InsertSlot(testCtx, req)
	assert.NoError(t, err)
	assert.True(t, id.Id != "", "ID must not be empty")
	assert.Equal(t, len(hu.slots), 1)

	_, err = hu.RemoveSlot(testCtx, id)
	assert.NoError(t, err)
	assert.Equal(t, len(hu.slots), 0)
}
