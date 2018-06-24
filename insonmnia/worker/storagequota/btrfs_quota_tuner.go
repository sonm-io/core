// +build linux

package storagequota

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/sonm-io/core/insonmnia/worker/storagequota/btrfs"
)

type btrfsQuotaTuner struct {
	dockerRootDir string
	subvolumesDir string
}

type btrfsQuotaCleaner struct {
	qgroupID         string
	containerQuotaID string
	path             string
}

func (b btrfsQuotaCleaner) Close() error {
	ctx := context.Background()
	btrfs.API.QuotaRemove(ctx, b.containerQuotaID, b.qgroupID, b.path)
	// NOTE: it would not be removed if any subvolume is assigned to this quota
	btrfs.API.QuotaDestroy(ctx, b.qgroupID, b.path)
	return nil
}

func newBtrfsQuotaTuner(info types.Info) (StorageQuotaTuner, error) {
	if info.Driver != "btrfs" {
		return nil, fmt.Errorf("%s is not supported", info.Driver)
	}
	return btrfsQuotaTuner{
		dockerRootDir: info.DockerRootDir,
		subvolumesDir: filepath.Join(info.DockerRootDir, "btrfs/subvolumes"),
	}, nil
}

func (b btrfsQuotaTuner) SetQuota(ctx context.Context, ID string, quotaID string, bytes uint64) (Cleanup, error) {
	// TODO: add ROLLBACK to prevent quota leak
	mountID, err := ioutil.ReadFile(filepath.Join(b.dockerRootDir, "image/btrfs/layerdb/mounts/", ID, "mount-id"))
	if err != nil {
		return nil, err
	}
	// Enable quota
	if err = btrfs.API.QuotaEnable(ctx, b.subvolumesDir); err != nil {
		return nil, err
	}
	h := fnv.New32a()
	io.WriteString(h, quotaID)
	qgroupID := fmt.Sprintf("1/%d", h.Sum32())
	// Check if quota exists
	exists, err := btrfs.API.QuotaExists(ctx, qgroupID, b.subvolumesDir)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create qgroup
		if err = btrfs.API.QuotaCreate(ctx, qgroupID, b.subvolumesDir); err != nil {
			return nil, err
		}
		// Limit qgroup
		if err = btrfs.API.QuotaLimit(ctx, bytes, qgroupID, b.subvolumesDir); err != nil {
			return nil, err
		}
	}
	containerQuotaID, err := btrfs.API.GetQuotaID(ctx, filepath.Join(b.subvolumesDir, string(mountID)))
	if err != nil {
		return nil, err
	}
	if err = btrfs.API.QuotaAssign(ctx, containerQuotaID, qgroupID, b.subvolumesDir); err != nil {
		return nil, err
	}
	cleaner := btrfsQuotaCleaner{
		qgroupID:         qgroupID,
		containerQuotaID: containerQuotaID,
		path:             b.subvolumesDir,
	}
	return cleaner, nil
}
