package salesman

import (
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type memStorage struct {
	data map[string]interface{}
}

func newMemStorage() *memStorage {
	return &memStorage{data: map[string]interface{}{}}
}

func (m *memStorage) Save(key string, value interface{}) error {
	m.data[key] = value
	return nil
}

func (m *memStorage) Load(key string, target interface{}) (bool, error) {
	val, ok := m.data[key]
	if ok {
		value := reflect.Indirect(reflect.ValueOf(target))
		value.Set(reflect.ValueOf(val))
		return true, nil
	}
	return false, nil
}

func getTestHardware(t *testing.T) *hardware.Hardware {
	x, err := hardware.NewHardware()
	require.NoError(t, err)

	x.RAM.Device.Available = 1024
	x.CPU.Device = &sonm.CPUDevice{
		ModelName: "Intel", Cores: 2, Sockets: 1,
	}
	x.Network.In = 100
	x.Network.Out = 200

	x.Storage.Device = &sonm.StorageDevice{
		BytesAvailable: 100500,
	}
	x.GPU = append(x.GPU, &sonm.GPU{
		Device:     &sonm.GPUDevice{ID: "1234", Memory: 123546},
		Benchmarks: map[uint64]*sonm.Benchmark{},
	})
	return x
}

func TestMemStorage(t *testing.T) {
	storage := newMemStorage()
	storage.Save("key", map[string]string{"item": "value"})
	target := map[string]string{}
	suc, err := storage.Load("key", &target)
	require.NoError(t, err)
	require.True(t, suc)
	require.Equal(t, 1, len(target))
	require.Equal(t, "value", target["item"])
}

func newTestSalesman(t *testing.T) (*Salesman, error) {
	hardware := getTestHardware(t)
	_, cgroup, _ := cgroups.NewNilCgroupManager()
	eth := NewDumEthAPI()
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	cfg := YAMLConfig{
		RegularBillPeriod:    24 * time.Hour,
		SpotBillPeriod:       time.Hour,
		SyncStepTimeout:      time.Minute * 2,
		SyncInterval:         time.Second * 10,
		MatcherRetryInterval: 10 * time.Second,
	}

	return NewSalesman(
		WithLogger(zap.S()),
		WithStorage(newMemStorage()),
		WithResources(resource.NewScheduler(zap.S(), hardware)),
		WithHardware(hardware),
		WithEth(eth),
		WithCGroupManager(cgroup),
		WithMatcher(eth),
		WithEthkey(key),
		WithConfig(&cfg),
	)
}

func TestNewSalesman(t *testing.T) {
	_, err := newTestSalesman(t)
	require.NoError(t, err)
}
