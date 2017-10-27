package locator

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLocator_Announce(t *testing.T) {
	lc := NewLocator(DefaultLocatorConfig())

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
	lc := NewLocator(DefaultLocatorConfig())

	n := &node{ethAddr: "123", ipAddr: []string{"111", "222"}}
	lc.putAnnounce(n)

	n2, err := lc.getResolve("123")
	assert.NoError(t, err)
	assert.Len(t, n2.ipAddr, 2)
}

func TestLocator_Resolve2(t *testing.T) {
	lc := NewLocator(DefaultLocatorConfig())

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
	}

	lc := NewLocator(conf)
	lc.putAnnounce(&node{ethAddr: "111"})
	lc.putAnnounce(&node{ethAddr: "222"})
	time.Sleep(1 * time.Second)
	assert.Len(t, lc.db, 2)
	lc.putAnnounce(&node{ethAddr: "333"})
	assert.Len(t, lc.db, 3)
	time.Sleep(1500 * time.Millisecond)
	assert.Len(t, lc.db, 1)
}
