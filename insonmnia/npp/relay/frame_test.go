package relay

import (
	"bytes"
	"testing"

	"github.com/sonm-io/core/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrameRoundTrip(t *testing.T) {
	message := &sonm.HandshakeResponse{
		Error:       42,
		Description: "oh, boy",
	}

	var err error
	wr := bytes.NewBuffer([]byte{})
	err = sendFrame(wr, message)
	require.NoError(t, err)

	messageBack := &sonm.HandshakeResponse{}
	err = recvFrame(bytes.NewBuffer(wr.Bytes()), messageBack)
	require.NoError(t, err)

	assert.Equal(t, message, messageBack)
}
