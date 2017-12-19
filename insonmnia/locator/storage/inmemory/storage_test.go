package inmemory

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	ds "github.com/sonm-io/core/insonmnia/locator/datastruct"
)

var (
	nodeTTL       = time.Hour
	cleanUpPeriod = time.Minute
	logger        = zap.NewNop()
)

func TestStorage_Put(t *testing.T) {
	s := NewStorage(cleanUpPeriod, nodeTTL, logger)

	s.Put(&ds.Node{EthAddr: common.StringToAddress("123")})
	s.Put(&ds.Node{EthAddr: common.StringToAddress("234")})
	s.Put(&ds.Node{EthAddr: common.StringToAddress("345")})

	assert.Len(t, s.db, 3)

	s.Put(&ds.Node{EthAddr: common.StringToAddress("123")})
	s.Put(&ds.Node{EthAddr: common.StringToAddress("123")})
	s.Put(&ds.Node{EthAddr: common.StringToAddress("123")})

	assert.Len(t, s.db, 3)
}

func TestStorage_ByEthAddr_TwoIPAddressesInStorage_TwoIpAddressesReturned(t *testing.T) {
	s := NewStorage(cleanUpPeriod, nodeTTL, logger)

	expected := &ds.Node{EthAddr: common.StringToAddress("123"), IpAddr: []string{"111", "222"}}
	s.Put(expected)

	obtained, err := s.ByEthAddr(common.StringToAddress("123"))
	assert.NoError(t, err)

	assert.Len(t, obtained.IpAddr, 2)
}

func TestStorage_ByEthAddr_InExistentAddressGiven_ErrorReturned(t *testing.T) {
	s := NewStorage(cleanUpPeriod, nodeTTL, logger)

	expected := &ds.Node{EthAddr: common.StringToAddress("123"), IpAddr: []string{"111", "222"}}
	s.Put(expected)

	obtained, err := s.ByEthAddr(common.StringToAddress("666"))
	assert.Equal(t, err, errNodeNotFound)
	assert.Nil(t, obtained)
}

func TestStorage_Expire(t *testing.T) {
	s := NewStorage(time.Second, 2*time.Second, logger)

	s.Put(&ds.Node{EthAddr: common.StringToAddress("111")})
	s.Put(&ds.Node{EthAddr: common.StringToAddress("222")})
	time.Sleep(1 * time.Second)
	assert.Len(t, s.db, 2)

	s.Put(&ds.Node{EthAddr: common.StringToAddress("333")})
	assert.Len(t, s.db, 3)
	time.Sleep(1500 * time.Millisecond)
	assert.Len(t, s.db, 1)
}
