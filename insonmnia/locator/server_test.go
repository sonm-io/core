package locator

import (
	"crypto/ecdsa"
	"crypto/tls"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

var (
	key = getTestKey()
)

func getTestKey() *ecdsa.PrivateKey {
	k, _ := crypto.GenerateKey()
	return k
}

func TestLocator_Announce(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	lc.putAnnounce(&node{ethAddr: common.StringToAddress("123")})
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("234")})
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("345")})

	assert.Len(t, lc.db, 3)

	lc.putAnnounce(&node{ethAddr: common.StringToAddress("123")})
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("123")})
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("123")})

	assert.Len(t, lc.db, 3)
}

func TestLocator_Resolve(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &node{ethAddr: common.StringToAddress("123"), ipAddr: []string{"111", "222"}}
	lc.putAnnounce(n)

	n2, err := lc.getResolve(common.StringToAddress("123"))
	assert.NoError(t, err)
	assert.Len(t, n2.ipAddr, 2)
}

func TestLocator_Resolve2(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &node{ethAddr: common.StringToAddress("123"), ipAddr: []string{"111", "222"}}
	lc.putAnnounce(n)

	n2, err := lc.getResolve(common.StringToAddress("666"))
	assert.Equal(t, err, errNodeNotFound)
	assert.Nil(t, n2)
}

func TestLocator_Expire(t *testing.T) {
	conf := &LocatorConfig{
		ListenAddr:    ":9090",
		NodeTTL:       2 * time.Second,
		CleanupPeriod: time.Second,
	}

	lc, err := NewLocator(context.Background(), conf, key)
	if err != nil {
		t.Error(err)
		return
	}

	lc.putAnnounce(&node{ethAddr: common.StringToAddress("111")})
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("222")})
	time.Sleep(1 * time.Second)
	assert.Len(t, lc.db, 2)
	lc.putAnnounce(&node{ethAddr: common.StringToAddress("333")})
	assert.Len(t, lc.db, 3)
	time.Sleep(1500 * time.Millisecond)
	assert.Len(t, lc.db, 1)
}

func TestLocator_AnnounceExternal(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig("localhost:9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := lc.Serve(); err != nil {
			t.Errorf("Locator server failed: %s", err)
		}
	}()

	cert, key, err := util.GenerateCert(key)
	if err != nil {
		t.Error(err)
		return
	}

	crt, err := tls.X509KeyPair(cert, key)
	if err != nil {
		t.Error(err)
		return
	}

	creds := util.NewTLS(&tls.Config{Certificates: []tls.Certificate{crt}, InsecureSkipVerify: true})

	conn, err := util.MakeGrpcClient(context.Background(), "localhost:9090", creds)
	if err != nil {
		t.Error(err)
		return
	}

	locatorClient := pb.NewLocatorClient(conn)
	_, err = locatorClient.Announce(context.Background(), &pb.AnnounceRequest{IpAddr: []string{"42.42.42.42"}})
	if err != nil {
		t.Error(err)
		return
	}

	if len(lc.db) != 1 {
		t.Error("Failed to securely announce")
	}
}
