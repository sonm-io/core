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
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	put := []string{
		"123",
		"234",
		"345",
	}

	for _, addr := range put {
		lc.put(&record{EthAddr: common.StringToAddress(addr)})
	}

	for _, addr := range put {
		rk, err := lc.get(common.StringToAddress(addr))
		assert.NoError(t, err)
		assert.Equal(t, rk.EthAddr, common.StringToAddress(addr))
	}
}

func TestLocator_Resolve(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &record{EthAddr: common.StringToAddress("123"), IPs: []string{"111", "222"}}
	lc.put(n)

	n2, err := lc.get(common.StringToAddress("123"))
	assert.NoError(t, err)
	assert.Len(t, n2.IPs, 2)
}

func TestLocator_Resolve2(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &record{EthAddr: common.StringToAddress("123"), IPs: []string{"111", "222"}}
	lc.put(n)

	n2, err := lc.get(common.StringToAddress("666"))
	assert.Equal(t, err, errNodeNotFound)
	assert.Nil(t, n2)
}

func TestLocator_Expire(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	lc.put(&record{EthAddr: common.StringToAddress("111")})
	time.Sleep(500 * time.Millisecond)
	rec, err := lc.get(common.StringToAddress("111"))
	assert.NoError(t, err)
	assert.Equal(t, rec.EthAddr, common.StringToAddress("111"))

	time.Sleep(1000 * time.Millisecond)
	rec, err = lc.get(common.StringToAddress("111"))
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestLocator_AnnounceExternal(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig("localhost:9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := lc.Serve(); err != nil {
			t.Errorf("Locator server failed: %s", err)
		}
	}()

	cert, crtKey, err := util.GenerateCert(key, time.Hour*4)
	if err != nil {
		t.Error(err)
		return
	}

	crt, err := tls.X509KeyPair(cert, crtKey)
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

	rec, err := lc.get(util.PubKeyToAddr(key.PublicKey))
	assert.NoError(t, err)

	assert.Equal(t, rec.EthAddr, util.PubKeyToAddr(key.PublicKey))
}
