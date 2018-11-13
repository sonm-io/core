package disk

import (
	"context"
	"fmt"
	"syscall"

	"github.com/docker/docker/client"
)

type Info struct {
	AvailableBytes uint64
	FreeBytes      uint64
}

// FreeDiskSpace returns free bytes for docker root path.
func FreeDiskSpace(ctx context.Context) (*Info, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("could not get docker client: %s", err)
	}

	info, err := cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get docker info: %s", err)
	}

	path := info.DockerRootDir
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		// if we cannot stat docker root - use system root
		if err := syscall.Statfs("/", &stat); err != nil {
			return nil, fmt.Errorf("could not perform statfs syscall: %s", err)
		}
	}

	// blocks * size per block = total size in bytes
	return &Info{
		AvailableBytes: stat.Bavail * uint64(stat.Bsize),
		FreeBytes:      stat.Bfree * uint64(stat.Bsize),
	}, nil
}
