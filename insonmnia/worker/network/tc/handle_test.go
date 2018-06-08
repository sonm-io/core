package tc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleUInt32(t *testing.T) {
	assert.Equal(t, uint32(0x00010000), NewHandle(1, 0).UInt32())
	assert.Equal(t, uint32(0x00100000), NewHandle(16, 0).UInt32())
}

func TestHandleNoneUInt32(t *testing.T) {
	assert.Equal(t, uint32(0), Handle(HandleNone).UInt32())
	assert.Equal(t, uint32(0), NewHandle(0, 0).UInt32())
}

func TestHandleRootUInt32(t *testing.T) {
	assert.Equal(t, uint32(0xffffffff), Handle(HandleRoot).UInt32())
	assert.Equal(t, uint32(0xffffffff), NewHandle(0xffff, 0xffff).UInt32())
}

func TestHandleIngressUInt32(t *testing.T) {
	assert.Equal(t, uint32(0xfffffff1), Handle(HandleIngress).UInt32())
	assert.Equal(t, uint32(0xfffffff1), NewHandle(0xffff, 0xfff1).UInt32())
}

func TestHandleString(t *testing.T) {
	assert.Equal(t, "1:0", NewHandle(1, 0).String())
	assert.Equal(t, "10:0", NewHandle(16, 0).String())
}

func TestHandleNoneString(t *testing.T) {
	assert.Equal(t, "none", Handle(HandleNone).String())
	assert.Equal(t, "none", NewHandle(0, 0).String())
}

func TestHandleRootString(t *testing.T) {
	assert.Equal(t, "root", Handle(HandleRoot).String())
	assert.Equal(t, "root", NewHandle(0xffff, 0xffff).String())
}

func TestHandleIngressString(t *testing.T) {
	assert.Equal(t, "ingress", Handle(HandleIngress).String())
	assert.Equal(t, "ingress", NewHandle(0xffff, 0xfff1).String())
}

func TestHandleWithMinor(t *testing.T) {
	assert.Equal(t, "10:1", NewHandle(16, 0).WithMinor(1).String())
}
