package locator

import (
	"crypto/ecdsa"
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
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
		lc.put(&record{EthAddr: common.HexToAddress(addr)})
	}

	for _, addr := range put {
		rk, err := lc.get(common.HexToAddress(addr))
		assert.NoError(t, err)
		assert.Equal(t, rk.EthAddr, common.HexToAddress(addr))
	}
}

func TestLocator_Resolve(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &record{EthAddr: common.HexToAddress("123"), IPs: []string{"111", "222"}}
	lc.put(n)

	n2, err := lc.get(common.HexToAddress("123"))
	assert.NoError(t, err)
	assert.Len(t, n2.IPs, 2)
}

func TestLocator_Resolve2(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	n := &record{EthAddr: common.HexToAddress("123"), IPs: []string{"111", "222"}}
	lc.put(n)

	n2, err := lc.get(common.HexToAddress("666"))
	assert.Equal(t, err, errNodeNotFound)
	assert.Nil(t, n2)
}

func TestLocator_Expire(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig(":9090"), key)
	if err != nil {
		t.Error(err)
		return
	}

	lc.put(&record{EthAddr: common.HexToAddress("111")})
	time.Sleep(500 * time.Millisecond)
	rec, err := lc.get(common.HexToAddress("111"))
	assert.NoError(t, err)
	assert.Equal(t, rec.EthAddr, common.HexToAddress("111"))

	time.Sleep(1000 * time.Millisecond)
	rec, err = lc.get(common.HexToAddress("111"))
	assert.Error(t, err)
	assert.Nil(t, rec)
}

func TestLocator_AnnounceExternal(t *testing.T) {
	lc, err := NewLocator(context.Background(), testConfig("localhost:9090"), getTestKey())
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := lc.Serve(); err != nil {
			t.Errorf("Locator server failed: %s", err)
		}
	}()

	cert, crtKey, err := util.GenerateCert(key, time.Hour)
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

	conn, err := xgrpc.NewClient(context.Background(), "localhost:9090", creds)
	if err != nil {
		t.Error(err)
		return
	}

	locatorClient := pb.NewLocatorClient(conn)

	_, err = locatorClient.Announce(context.Background(), &pb.AnnounceRequest{IpAddr: []string{"192.168.0.0"}})
	assert.Error(t, err)

	_, err = locatorClient.Announce(context.Background(), &pb.AnnounceRequest{IpAddr: []string{"41.41.41.41:10001"}})
	if err != nil {
		t.Error(err)
		return
	}

	rec, err := lc.get(util.PubKeyToAddr(key.PublicKey))
	assert.NoError(t, err)

	assert.Equal(t, rec.EthAddr, util.PubKeyToAddr(key.PublicKey))
	assert.Equal(t, []string{"41.41.41.41:10001"}, rec.IPs)

	if err := conn.Close(); err != nil {
		t.Error(err)
	}
}

func TestLocator_SkipPrivateIP(t *testing.T) {
	cfg := testConfig("localhost:9191")
	cfg.Store.Endpoint += "-skip-private"

	lc, err := NewLocator(context.Background(), cfg, getTestKey())
	if err != nil {
		t.Error(err)
		return
	}

	defer os.Remove(cfg.Store.Endpoint)

	lc.onlyPublicIPs = true
	go func() {
		if err := lc.Serve(); err != nil {
			t.Errorf("Locator server failed: %s", err)
		}
	}()

	cert, crtKey, err := util.GenerateCert(key, time.Hour)
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

	conn, err := xgrpc.NewWalletAuthenticatedClient(context.Background(), creds, "localhost:9191")
	if err != nil {
		t.Error(err)
		return
	}

	locatorClient := pb.NewLocatorClient(conn)
	_, err = locatorClient.Announce(context.Background(), &pb.AnnounceRequest{IpAddr: []string{"192.168.0.0:10001"}})
	assert.Error(t, err)

	_, err = locatorClient.Announce(context.Background(),
		&pb.AnnounceRequest{IpAddr: []string{"42.42.42.42:10001", "192.168.0.0:10001"}})
	assert.NoError(t, err)
	rec, err := lc.get(util.PubKeyToAddr(key.PublicKey))
	assert.NoError(t, err)
	assert.Equal(t, []string{"42.42.42.42:10001"}, rec.IPs)

	if err := conn.Close(); err != nil {
		t.Error(err)
	}
}
