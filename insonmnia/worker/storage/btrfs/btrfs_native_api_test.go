// +build linux,btrfsnative

package btrfs

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseQroupID(t *testing.T) {
	require := require.New(t)
	id, err := parseQgroupID("1/555")
	require.NoError(err)
	require.NotZero(id)
}

func TestE2ENative(t *testing.T) {
	path := os.Getenv("BTRFS_PLAYGROUND_PATH")
	if path == "" {
		t.Skip("BTRFS_PLAYGROUND_PATH must be set for the test")
	}
	if os.Getenv("SUDO_USER") == "" {
		t.Skip("WARNING: root permissions required for that test")
	}

	var b btrfsNativeAPI
	t.Run(fmt.Sprintf("%T", b), func(t *testing.T) {
		testE2EOne(t, b, path)
	})
}
