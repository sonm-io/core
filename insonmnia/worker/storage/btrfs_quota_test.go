// +build linux

package storage

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sonm-io/core/util/xdocker"
	"github.com/stretchr/testify/require"
)

func TestBTRFSQuota(t *testing.T) {
	if os.Getenv("SUDO_USER") == "" {
		t.Skip("sudo required for the test")
	}

	ctx := context.Background()
	require := require.New(t)
	dclient, err := xdocker.NewClient()
	require.NoError(err)

	info, err := dclient.Info(ctx)
	require.NoError(err)
	tuner, err := NewQuotaTuner(info)
	require.NoError(err)
	_ = tuner

	config := &container.Config{
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		Tty:          true,
		Image:        "busybox",
		Cmd:          strings.Split("dd if=/dev/zero of=/FILE bs=1024 count=10000", " "),
	}

	// imageInspect, _, err := dclient.ImageInspectWithRaw(ctx, "busybox")
	// require.NoError(err)

	hostConfig := &container.HostConfig{}
	networkingConfig := &network.NetworkingConfig{}

	type Container struct {
		ID string
	}

	var containers = make([]Container, 0)
	cleanups := make([]Cleanup, 0)
	defer func() {
		for _, c := range cleanups {
			c.Close()
		}

		for _, c := range containers {
			dclient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
		}
	}()

	limit := uint64(20 * 1024 * 1024)
	for _, name := range []string{"aaa", "bbb", "ccc"} {
		resp, err := dclient.ContainerCreate(ctx, config, hostConfig, networkingConfig, name)
		require.NoError(err)
		containers = append(containers, Container{ID: resp.ID})
		cleanup, err := tuner.SetQuota(ctx, resp.ID, "xxxx", limit)
		require.NoError(err)
		cleanups = append(cleanups, cleanup)
		require.NoError(dclient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}))
	}

	written := int64(0)
	failed := 0
	for _, container := range containers {
		rdcloser, stat, err := dclient.CopyFromContainer(ctx, container.ID, "/FILE")
		if err == nil {
			written += stat.Size
			rdcloser.Close()
		} else {
			failed++
		}
	}
	require.Equal(1, failed)
	require.True(written > 0)
	require.True(written < int64(limit))
}
