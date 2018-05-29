package disk

import (
	"context"
	"fmt"
	"syscall"

	"github.com/docker/docker/client"
)

// FreeDiskSpace returns free bytes for docker root path.
func FreeDiskSpace(ctx context.Context) (uint64, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return 0, fmt.Errorf("could not get docker client: %s", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		return 0, fmt.Errorf("could not get docker info: %s", err)
	}

	path := info.DockerRootDir
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		// if we cannot stat docker root - use system root
		if err := syscall.Statfs("/", &stat); err != nil {
			return 0, fmt.Errorf("could not perform statfs syscall: %s", err)
		}
	}

	// Available blocks * size per block = available space in bytes
	return stat.Bavail * uint64(stat.Bsize), nil
}
