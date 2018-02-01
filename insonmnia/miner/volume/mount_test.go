package volume

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePermissionRW(t *testing.T) {
	perm, err := ParsePermission("rw")

	assert.Equal(t, RW, perm)
	require.NoError(t, err)
}

func TestParsePermissionRO(t *testing.T) {
	perm, err := ParsePermission("ro")

	assert.Equal(t, RO, perm)
	require.NoError(t, err)
}

func TestParsePermissionError(t *testing.T) {
	_, err := ParsePermission("??")

	assert.Error(t, err)
}

func TestNewMount(t *testing.T) {
	mount, err := NewMount("cifs:/mnt:rw")

	require.NoError(t, err)
	assert.Equal(t, Mount{Source: "cifs", Target: "/mnt", Permission: RW}, mount)
	assert.False(t, mount.ReadOnly())
}

func TestNewMountWithoutPerm(t *testing.T) {
	mount, err := NewMount("cifs:/mnt")

	require.NoError(t, err)
	assert.Equal(t, Mount{Source: "cifs", Target: "/mnt", Permission: RW}, mount)
	assert.False(t, mount.ReadOnly())
}

func TestNewMountOnlyTarget(t *testing.T) {
	mount, err := NewMount("/mnt")

	require.NoError(t, err)
	assert.Equal(t, Mount{Source: "", Target: "/mnt", Permission: RW}, mount)
	assert.False(t, mount.ReadOnly())
}

func TestNewMountInvalidSpec(t *testing.T) {
	mount, err := NewMount("whatever:cifs:/mnt:rw")

	assert.Equal(t, Mount{}, mount)
	assert.Error(t, err)
}

func TestNewMountInvalidSpecEmptySource(t *testing.T) {
	mount, err := NewMount(":/mnt:rw")

	assert.Equal(t, Mount{}, mount)
	assert.Error(t, err)
}

func TestNewMountInvalidPerm(t *testing.T) {
	mount, err := NewMount("cifs:/mnt:?")

	assert.Equal(t, Mount{}, mount)
	assert.Error(t, err)
}
