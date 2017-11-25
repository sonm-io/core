package locator

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestLocator_Announce(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"))
	if err != nil {
		t.Error(err)
		return
	}

	lc.putAnnounce(&node{ethAddr: "123"})
	lc.putAnnounce(&node{ethAddr: "234"})
	lc.putAnnounce(&node{ethAddr: "345"})

	assert.Len(t, lc.db, 3)

	lc.putAnnounce(&node{ethAddr: "123"})
	lc.putAnnounce(&node{ethAddr: "123"})
	lc.putAnnounce(&node{ethAddr: "123"})

	assert.Len(t, lc.db, 3)
}

func TestLocator_Resolve(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"))
	if err != nil {
		t.Error(err)
		return
	}

	n := &node{ethAddr: "123", ipAddr: []string{"111", "222"}}
	lc.putAnnounce(n)

	n2, err := lc.getResolve("123")
	assert.NoError(t, err)
	assert.Len(t, n2.ipAddr, 2)
}

func TestLocator_Resolve2(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig(":9090"))
	if err != nil {
		t.Error(err)
		return
	}

	n := &node{ethAddr: "123", ipAddr: []string{"111", "222"}}
	lc.putAnnounce(n)

	n2, err := lc.getResolve("666")
	assert.Equal(t, err, errNodeNotFound)
	assert.Nil(t, n2)
}

func TestLocator_Expire(t *testing.T) {
	conf := &LocatorConfig{
		ListenAddr:    ":9090",
		NodeTTL:       2 * time.Second,
		CleanupPeriod: time.Second,
		Eth: EthConfig{
			PrivateKey: "d07fff36ef2c3d15144974c25d3f5c061ae830a81eefd44292588b3cea2c701c",
		},
	}

	lc, err := NewLocator(context.Background(), conf)
	if err != nil {
		t.Error(err)
		return
	}

	lc.putAnnounce(&node{ethAddr: "111"})
	lc.putAnnounce(&node{ethAddr: "222"})
	time.Sleep(1 * time.Second)
	assert.Len(t, lc.db, 2)
	lc.putAnnounce(&node{ethAddr: "333"})
	assert.Len(t, lc.db, 3)
	time.Sleep(1500 * time.Millisecond)
	assert.Len(t, lc.db, 1)
}

func TestLocator_AnnounceExternal(t *testing.T) {
	lc, err := NewLocator(context.Background(), DefaultConfig("localhost:9090"))
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		if err := lc.Serve(); err != nil {
			t.Errorf("Locator server failed: %s", err)
		}
	}()

	ethKey, err := crypto.HexToECDSA("44f2d47988142fda13c27d7da0990ba534b50d6bf7e9dbc9d13654c5795881ae")
	if err != nil {
		t.Error(errors.Wrap(err, "malformed ethereum private key"))
		return
	}

	cert, key, err := util.GenerateCert(ethKey)
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
