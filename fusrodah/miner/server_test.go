package miner

import (
	"encoding/json"
	"testing"

	"github.com/sonm-io/core/fusrodah"
	"github.com/stretchr/testify/assert"
)

func marshalTestData(wrk, cli string) []byte {
	dm := fusrodah.DiscoveryMessage{
		WorkerEndpoint: wrk,
		ClientEndpoint: cli,
	}
	b, _ := json.Marshal(&dm)
	return b
}

func TestUnmarshalDiscoveryMessage(t *testing.T) {
	srv, err := NewServer(nil)
	assert.NoError(t, err)

	msg := marshalTestData("1.1.1.1:1111", "2.2.2.2:2222")
	dm := srv.unmarshalDiscoveryMessage(msg)
	assert.Equal(t, dm.WorkerEndpoint, "1.1.1.1:1111")
	assert.Equal(t, dm.ClientEndpoint, "2.2.2.2:2222")
}

func TestUnmarshalDiscoveryMessageLegacyFormat(t *testing.T) {
	srv, err := NewServer(nil)
	assert.NoError(t, err)

	dm := srv.unmarshalDiscoveryMessage([]byte("1.2.3.4:1234"))
	assert.Equal(t, dm.WorkerEndpoint, "1.2.3.4:1234")
	assert.Equal(t, dm.ClientEndpoint, "")
}
