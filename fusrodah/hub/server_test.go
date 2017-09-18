package hub

import (
	"encoding/json"
	"github.com/sonm-io/core/fusrodah"
	"github.com/stretchr/testify/assert"
	"testing"
)

// marshalDiscoveryMessage
func TestMarshalDiscoveryMessage(t *testing.T) {
	srv, err := NewServer(nil, "1.1.1.1:1111", "2.2.2.2:2222")
	assert.NoError(t, err)

	msg := srv.marshalDiscoveryMessage()
	msgStruct := fusrodah.DiscoveryMessage{}

	err = json.Unmarshal([]byte(msg), &msgStruct)
	assert.NoError(t, err)

	assert.Equal(t, "1.1.1.1:1111", msgStruct.WorkerEndpoint)
	assert.Equal(t, "2.2.2.2:2222", msgStruct.ClientEndpoint)
}
