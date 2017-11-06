package commands

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/satori/uuid"
	pb "github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
)

func makeTestFilePath() string {
	f := uuid.NewV4().String() + ".yaml"
	return path.Join("/tmp", f)
}

func createTestYamlFile(body string) (string, error) {
	p := makeTestFilePath()
	err := ioutil.WriteFile(p, []byte(body), 0600)
	return p, err
}

func deleteTestYamlFile(p string) {
	os.Remove(p)
}

func TestLoadSlotYaml(t *testing.T) {
	p, err := createTestYamlFile(`
duration:
  since: 2017-12-01T00:00:00Z
  until: 2017-12-02T00:00:00Z

rating:
  buyer: 100
  supplier: 42

resources:
  cpu_cores: 1
  ram_bytes: 100000000
  gpu_count: SINGLE_GPU
  storage: 2000000000

  network:
    in: 100
    out: 200
    type: INCOMING

  properties:
    foo: 3.1415
    cycles: 42
`)
	assert.NoError(t, err)
	defer deleteTestYamlFile(p)

	slot, err := loadSlotFile(p)
	assert.NoError(t, err)

	ss := slot.Unwrap()
	assert.Equal(t, int64(100), ss.BuyerRating)
	assert.Equal(t, int64(42), ss.SupplierRating)
	assert.Equal(t, uint64(1), ss.Resources.CpuCores)
	assert.Equal(t, uint64(100000000), ss.Resources.RamBytes)
	assert.Equal(t, pb.GPUCount_SINGLE_GPU, ss.Resources.GpuCount)
	assert.Equal(t, uint64(2000000000), ss.Resources.Storage)
	assert.Equal(t, uint64(100), ss.Resources.NetTrafficIn)
	assert.Equal(t, uint64(200), ss.Resources.NetTrafficOut)
	assert.Equal(t, pb.NetworkType_INCOMING, ss.Resources.NetworkType)
	assert.Contains(t, ss.Resources.Properties, "foo")
	assert.Contains(t, ss.Resources.Properties, "cycles")
}

func TestLoadOrderYaml(t *testing.T) {
	p, err := createTestYamlFile(`
price: 145
order_type: BID
slot:
  duration:
    since: 2017-12-02T00:00:00Z
    until: 2017-12-03T00:00:00Z

  rating:
    buyer: 45
    supplier: 54

  resources:
    cpu_cores: 2
    ram_bytes: 100000000
    gpu_count: MULTIPLE_GPU
    storage: 2000000000

    network:
      in: 200
      out: 300
      type: INCOMING

    properties:
      foo: 3.14
      cycles: 42`)
	assert.NoError(t, err)
	defer deleteTestYamlFile(p)

	order, err := loadOrderFile(p)
	assert.NoError(t, err)

	ord := order.Unwrap()
	assert.Equal(t, int64(145), ord.Price)
	assert.Equal(t, pb.OrderType_BID, ord.OrderType)
}

func TestLoadPropsYaml(t *testing.T) {
	p, err := createTestYamlFile(`
foo: 3.14
cycles: 42`)
	assert.NoError(t, err)
	defer deleteTestYamlFile(p)

	props, err := loadPropsFile(p)
	assert.NoError(t, err)

	assert.Contains(t, props, "foo")
	assert.Contains(t, props, "cycles")
	assert.Equal(t, 3.14, props["foo"])
	assert.Equal(t, 42.0, props["cycles"])
}
