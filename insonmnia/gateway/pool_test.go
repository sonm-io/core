package gateway

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPool(t *testing.T) {
	p := NewPortPool(10, 2)

	port, err := p.Assign("0")
	assert.NoError(t, err)
	assert.True(t, port == 10 || port == 11)

	port, err = p.Assign("1")
	assert.NoError(t, err)
	assert.True(t, port == 10 || port == 11)

	port, err = p.Assign("2")
	assert.Error(t, err)
	assert.Equal(t, uint16(0), port)

}

func TestPoolRetainWhileEmpty(t *testing.T) {
	p := NewPortPool(10, 0)

	err := p.Retain("0")
	assert.Error(t, err)
}

func TestPoolDoubleAssign(t *testing.T) {
	p := NewPortPool(10, 2)

	port, err := p.Assign("0")
	assert.NoError(t, err)
	assert.True(t, port == 10 || port == 11)

	port, err = p.Assign("0")
	assert.Error(t, err)
	assert.True(t, port == 0)
}

func TestPoolAssignRetainAssign(t *testing.T) {
	p := NewPortPool(10, 1)

	port, err := p.Assign("0")
	assert.NoError(t, err)
	assert.True(t, port == 10)

	err = p.Retain("0")
	assert.NoError(t, err)

	port, err = p.Assign("0")
	assert.NoError(t, err)
	assert.True(t, port == 10)
}
