package hub

import (
	"context"
	"crypto/ecdsa"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
)

func defaultMinerMockCfg() *miner.Config {
	return &miner.Config{
		Endpoint:  "127.0.0.1:10002",
		Resources: &miner.ResourcesConfig{},
		SSH:       &miner.SSHConfig{},
		PublicIPs: []string{"192.168.70.17", "46.148.198.133"},
		Plugins:   plugin.Config{},
		Whitelist: miner.WhitelistConfig{Enabled: new(bool)},
	}
}

func getTestMiner(mock *gomock.Controller) (*miner.Miner, error) {
	cfg := defaultMinerMockCfg()

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

func buildTestHub(ctrl *gomock.Controller) (*Hub, error) {
	config := defaultMinerMockCfg()
	worker, _ := getTestMiner(ctrl)

	return New(config, WithPrivateKey(key), WithWorker(worker))
}

//TODO: Move this to separate test for AskPlans
func TestHubCreateRemoveSlot(t *testing.T) {
	t.Skip()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hu, err := buildTestHub(ctrl)
	assert.NoError(t, err)

	req := &pb.AskPlan{
		Duration:  &pb.Duration{Nanoseconds: structs.MinSlotDuration.Nanoseconds()},
		Resources: &pb.AskPlanResources{},
	}

	testCtx := context.Background()

	id, err := hu.CreateAskPlan(testCtx, req)
	assert.NoError(t, err)
	assert.True(t, id.Id != "", "ID must not be empty")

	actualSlots, err := hu.AskPlans(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.AskPlans), 1)

	_, err = hu.RemoveAskPlan(testCtx, id)
	assert.NoError(t, err)

	actualSlots, err = hu.AskPlans(testCtx, &pb.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, len(actualSlots.AskPlans), 0)
}
