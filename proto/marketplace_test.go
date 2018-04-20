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

	input := []byte(`level: pseudonymous`)
	err := yaml.Unmarshal(input, &into)

	require.NoError(t, err)
	assert.Equal(t, IdentityLevel_PSEUDONYMOUS, into.Level)
}
