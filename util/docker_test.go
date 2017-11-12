package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePortBinding(t *testing.T) {
	binding, err := ParsePortBinding("80/tcp")
	assert.NoError(t, err)
	assert.Equal(t, binding.Network(), "tcp")
	assert.Equal(t, binding.Port(), uint16(80))
}

func TestParsePortBindingInvalidSeparator(t *testing.T) {
	binding, err := ParsePortBinding("80:tcp")
	assert.Error(t, err)
	assert.Nil(t, binding)
}

func TestParsePortBindingInvalidPort(t *testing.T) {
	binding, err := ParsePortBinding("port/tcp")
	assert.Error(t, err)
	assert.Nil(t, binding)
}

func TestParsePortBindingPortOutOfRange(t *testing.T) {
	binding, err := ParsePortBinding("100500/tcp")
	assert.Error(t, err)
	assert.Nil(t, binding)
}
