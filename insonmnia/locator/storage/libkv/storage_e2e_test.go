package libkv_test

import (
	"testing"
	"time"

	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	engine "github.com/sonm-io/core/insonmnia/locator/storage/libkv"

	"encoding/json"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/libkv"
	ds "github.com/sonm-io/core/insonmnia/locator/datastruct"
	"github.com/stretchr/testify/require"
	"os"
)

var (
	nodeTTL = time.Hour
	prefix  = "node/"
)

func init() {
	boltdb.Register()
}

func TestStoragePut_ValidNodeGiven_NodeStoredSuccessfully(t *testing.T) {
	// arrange
	s, kv := NewStorage(t)

	addr := common.StringToAddress("123")

	// act
	err := s.Put(&ds.Node{EthAddr: addr})
	require.NoError(t, err, "cannot put node into storage")

	// assert
	ok, err := kv.Exists(prefix + addr.Hex())
	require.NoError(t, err, "cannot check if key exists")

	assert.True(t, ok)
}

func TestStorageByEthAddr_TwoIPAddressesInStorage_TwoIpAddressesReturned(t *testing.T) {
	// arrange
	s, kv := NewStorage(t)

	key := common.StringToAddress("123")
	expected := &ds.Node{EthAddr: key, IpAddr: []string{"111", "222"}, TS: time.Now()}

	value, err := json.Marshal(expected)
	spew.Dump(string(value))
	require.NoError(t, err)

	err = kv.Put(prefix+key.Hex(), value, nil)
	require.NoError(t, err, "cannot put node into storage")

	// act
	obtained, err := s.ByEthAddr(common.StringToAddress("123"))
	require.NoError(t, err)

	// assert
	assert.EqualValues(t, expected.IpAddr, obtained.IpAddr)
}

func TestStorageByEthAddr_InExistentAddressGiven_ErrorReturned(t *testing.T) {
	s, _ := NewStorage(t)

	_, err := s.ByEthAddr(common.StringToAddress("666"))
	assert.EqualError(t, err, "cannot get value at key: node/0x0000000000000000000000000000000000363636")
}

func TestStorageByEthAddr_ExpiredKeyGiven_ErrorReturned(t *testing.T) {
	// arrange
	s, kv := NewStorageWithParams(t, 50*time.Millisecond)

	key := common.StringToAddress("111")
	expected := &ds.Node{EthAddr: key, IpAddr: []string{"127.0.0.1"}, TS: time.Now()}

	value, err := json.Marshal(expected)
	require.NoError(t, err)

	err = kv.Put(prefix+key.Hex(), value, nil)
	require.NoError(t, err, "cannot put node into storage")

	time.Sleep(100 * time.Millisecond)

	// act
	_, err = s.ByEthAddr(key)

	// assert
	assert.EqualError(t, err, "value timed out at key: node/0x0000000000000000000000000000000000313131")
}

func NewStorage(t *testing.T) (*engine.Storage, store.Store) {
	kv, err := NewEngine()
	require.NoError(t, err, "cannot init storage engine")

	s, err := engine.NewStorage(nodeTTL, kv)
	require.NoError(t, err, "cannot init storage")

	return s, kv
}

func NewStorageWithParams(t *testing.T, ttl time.Duration) (*engine.Storage, store.Store) {
	kv, err := NewEngine()
	require.NoError(t, err, "cannot init storage engine")

	s, err := engine.NewStorage(ttl, kv)
	require.NoError(t, err, "cannot init storage")

	return s, kv
}

func NewEngine() (store.Store, error) {
	os.RemoveAll("/tmp/sonm/bolt.test")
	kv, err := libkv.NewStore(
		store.BOLTDB,
		[]string{"/tmp/sonm/bolt.test"},
		&store.Config{
			Bucket:            "sonm",
			ConnectionTimeout: 10 * time.Second,
		},
	)

	return kv, err
}
