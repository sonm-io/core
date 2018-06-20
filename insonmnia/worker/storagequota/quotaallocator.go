package storagequota

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sonm-io/core/insonmnia/worker/storagequota/btrfs"
)

type QuotaDescription struct {
	Bytes uint64
}

type Cleanup interface {
	Close() error
}

type StorageQuotaTuner interface {
	SetQuota(ctx context.Context, ID string, quotaID string, bytes uint64) (Cleanup, error)
}

type btrfsQuotaTuner struct{}

func QuotationSupported(dockerInfo types.Info) bool {
	return dockerInfo.Driver == "btrfs"
}

func (btrfsQuotaTuner) SetQuota(ctx context.Context, dockerClient *client.Client, ID string, quotaID string, bytes uint64) (Cleanup, error) {
	// TODO: add ROLLBACK to prevent quota leak

	// Assign
	info, err := dockerClient.Info(ctx)
	if err != nil {
		return nil, err
	}
	mountID, err := ioutil.ReadFile(filepath.Join(info.DockerRootDir, "image/btrfs/layerdb/mounts/", ID, "mount-id"))
	if err != nil {
		return nil, err
	}
	subvolumesPath := filepath.Join(info.DockerRootDir, "btrfs/subvolumes")
	// Enable quota
	if err = btrfs.API.QuotaEnable(ctx, subvolumesPath); err != nil {
		return nil, err
	}
	h := fnv.New64a()
	io.WriteString(h, quotaID)
	qgroupID := fmt.Sprintf("0/%d", h.Sum64())

	// Check if quota exists
	exists, err := btrfs.API.QuotaExists(ctx, qgroupID, subvolumesPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create qgroup
		if err = btrfs.API.QuotaCreate(ctx, qgroupID, subvolumesPath); err != nil {
			return nil, err
		}
		// Limit qgroup
		if err = btrfs.API.QuotaLimit(ctx, bytes, qgroupID, subvolumesPath); err != nil {
			return nil, err
		}
	}
	containerQuotaID, err := btrfs.API.GetQuotaID(ctx, filepath.Join(subvolumesPath, string(mountID)))
	if err != nil {
		return nil, err
	}
	if err = btrfs.API.QuotaAssign(ctx, containerQuotaID, qgroupID, subvolumesPath); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("NOT IMPLEMENTED")
}
