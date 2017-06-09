package miner

import (
	"testing"

	"github.com/stretchr/testify/require"

	"golang.org/x/net/context"
)

func TestOvsSpool(t *testing.T) {
	ctx := context.Background()
	ovs, err := NewOverseer(ctx)
	require.NoError(t, err, "failed to create Overseer")
	err = ovs.Spool(ctx, Description{Registry: "docker.io", Image: "alpine"})
	require.NoError(t, err, "failed to pull an image")
	err = ovs.Spool(ctx, Description{Registry: "docker2.io", Image: "alpine"})
	require.NotNil(t, err)
}
