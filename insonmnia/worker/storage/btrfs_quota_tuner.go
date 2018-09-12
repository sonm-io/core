// +build linux

package storage

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/sonm-io/core/insonmnia/worker/storage/btrfs"
)

type btrfsQuotaTuner struct {
	dockerRootDir string
	subvolumesDir string
	btrfs.API
}

type btrfsQuotaCleaner struct {
	qgroupID         string
	containerQuotaID string
	path             string
	btrfs.API
}

func (b btrfsQuotaCleaner) Close() error {
	ctx := context.Background()
	b.API.QuotaRemove(ctx, b.containerQuotaID, b.qgroupID, b.path)
	// NOTE: it would not be removed if any subvolume is assigned to this quota
	b.API.QuotaDestroy(ctx, b.qgroupID, b.path)
	return nil
}

func newBtrfsQuotaTuner(info types.Info) (StorageQuotaTuner, error) {
	if info.Driver != "btrfs" {
		return nil, fmt.Errorf("%s is not supported", info.Driver)
	}

	btrfsAPI, err := btrfs.NewAPI()
	if err != nil {
		return nil, err
	}

	return btrfsQuotaTuner{
		dockerRootDir: info.DockerRootDir,
		subvolumesDir: filepath.Join(info.DockerRootDir, "btrfs/subvolumes"),
		API:           btrfsAPI,
	}, nil
}

func (b btrfsQuotaTuner) SetQuota(ctx context.Context, ID string, quotaID string, bytes uint64) (Cleanup, error) {
	// TODO: add ROLLBACK to prevent quota leak
	mountID, err := ioutil.ReadFile(filepath.Join(b.dockerRootDir, "image/btrfs/layerdb/mounts/", ID, "mount-id"))
	if err != nil {
		return nil, err
	}
	// Enable quota
	if err = b.API.QuotaEnable(ctx, b.subvolumesDir); err != nil {
		return nil, err
	}
	h := fnv.New32a()
	io.WriteString(h, quotaID)
	qgroupID := fmt.Sprintf("1/%d", h.Sum32())
	// Check if quota exists
	exists, err := b.API.QuotaExists(ctx, qgroupID, b.subvolumesDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create qgroup
		if err = b.API.QuotaCreate(ctx, qgroupID, b.subvolumesDir); err != nil {
			return nil, err
		}
		// Limit qgroup
		if err = b.API.QuotaLimit(ctx, bytes, qgroupID, b.subvolumesDir); err != nil {
			return nil, err
		}
	}
	containerQuotaID, err := b.API.GetQuotaID(ctx, filepath.Join(b.subvolumesDir, string(mountID)))
	if err != nil {
		return nil, err
	}
	if err = b.API.QuotaAssign(ctx, containerQuotaID, qgroupID, b.subvolumesDir); err != nil {
		return nil, err
	}
	cleaner := btrfsQuotaCleaner{
		qgroupID:         qgroupID,
		containerQuotaID: containerQuotaID,
		path:             b.subvolumesDir,
		API:              b.API,
	}
	return cleaner, nil
}
