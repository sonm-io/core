package sonm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestName(t *testing.T) {
	into := struct {
		Level IdentityLevel
	}{}

	input := []byte(`level: registered`)
	err := yaml.Unmarshal(input, &into)

	require.NoError(t, err)
	assert.Equal(t, IdentityLevel_REGISTERED, into.Level)
}

func TestBidOrderValidate(t *testing.T) {
	bid := &BidOrder{Tag: "this-string-is-too-long-for-tag-value"}
	err := bid.Validate()
	require.Error(t, err)

	bid.Tag = "short-and-valid"
	err = bid.Validate()
	require.NoError(t, err)
}
