package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEmptyNode(t *testing.T) {
	node, err := ParseNode("")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestParseNode(t *testing.T) {
	node, err := ParseNode("uuid@127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, node)

	assert.Equal(t, &Node{Name: "uuid", Addr: "127.0.0.1"}, node)
}

func TestParseNodeEmptyName(t *testing.T) {
	node, err := ParseNode("@127.0.0.1")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestParseNodeEmptyAddr(t *testing.T) {
	node, err := ParseNode("uuid@")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestParseNodeNoSeparator(t *testing.T) {
	node, err := ParseNode("uuid")
	assert.Error(t, err)
	assert.Nil(t, node)
}

func TestParseNodeSuchMuchSeparator(t *testing.T) {
	node, err := ParseNode("sonm@what@127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, node)

	assert.Equal(t, &Node{Name: "sonm@what", Addr: "127.0.0.1"}, node)
}
