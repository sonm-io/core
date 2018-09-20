// +build linux

package btrfs

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func testE2EOne(t *testing.T, b API, path string) {
	require := require.New(t)
	ctx := context.Background()

	require.NoError(b.QuotaEnable(ctx, path))
	// one more time
	require.NoError(b.QuotaEnable(ctx, path))

	state, err := getBTRFSState(ctx, path)
	require.NoError(err)
	defer func() {
		postState, postErr := getBTRFSState(ctx, path)
		require.NoError(postErr)
		require.Equal(state, postState)
	}()

	// Create QUOTA
	var qgroupID = "1/999"
	exists, err := b.QuotaExists(ctx, qgroupID, path)
	require.NoError(err)
	require.False(exists)

	require.NoError(b.QuotaCreate(ctx, qgroupID, path))
	defer func() {
		require.NoError(b.QuotaDestroy(ctx, qgroupID, path))
	}()

	exists, err = b.QuotaExists(ctx, qgroupID, path)
	require.NoError(err)
	require.True(exists)

	const Limit = 10 * 1024 * 1024
	require.NoError(b.QuotaLimit(ctx, Limit, qgroupID, path))

	// Create subvolumes
	type subvolume struct {
		path    string
		quotaID string
	}
	var subvolumes = make([]subvolume, 0)

	for _, dir := range []string{"xyz", "abc", "def", "ghi"} {
		subvolumePath := filepath.Join(path, dir)
		require.NoError(createSubvolume(ctx, t, subvolumePath))
		quotaID, err := b.GetQuotaID(ctx, subvolumePath)
		require.NoError(err)
		require.NotEmpty(quotaID)
		subvolumes = append(subvolumes, subvolume{path: subvolumePath, quotaID: quotaID})
	}

	defer func() {
		for _, subvolume := range subvolumes {
			require.NoError(b.QuotaDestroy(ctx, subvolume.quotaID, subvolume.path))
			require.NoError(destroySubvolume(ctx, subvolume.path))
		}
	}()

	// Assign limit to 3 containers
	const lastUnassigned = 3
	for _, subvolume := range subvolumes[:lastUnassigned] {
		require.NoError(b.QuotaAssign(ctx, subvolume.quotaID, qgroupID, path))
	}

	devZero, err := os.Open("/dev/zero")
	require.NoError(err)
	defer devZero.Close()

	// Write 100 KB to the last subvolume
	freeFile, err := os.Create(filepath.Join(subvolumes[lastUnassigned].path, "FILE"))
	require.NoError(err)
	defer freeFile.Close()
	_, err = io.CopyN(freeFile, devZero, 10*Limit)
	require.NoError(err)

	written := int64(0)
	partToWrite := int64(Limit * 0.8)
	for j, subvolume := range subvolumes[:lastUnassigned] {
		f, err := os.Create(filepath.Join(subvolume.path, "QFILE"))
		switch j {
		case 0:
			// The first containe can create and write data less than limit
			require.NoError(err)
			nn, wrErr := io.CopyN(f, devZero, partToWrite)
			require.NoError(wrErr)
			require.Equal(nn, partToWrite)
			written += nn
			f.Close()
		case 1:
			// The second one can create file, but amount of available space is less
			// than limit - space occupied by the first container
			require.NoError(err)
			nn, wrErr := io.CopyN(f, devZero, partToWrite)
			require.Error(wrErr)
			require.True(nn <= Limit-written)
			written += nn
			f.Close()
		default:
			// The third can NOT even create a file
			require.Error(err)
		}
	}
	require.True(written < Limit)

	// Even after quotas we can append
	// as much data as we want to a free container.
	// Our quotas does not affect it.
	nn, err := io.CopyN(freeFile, devZero, 5*Limit)
	require.NoError(err)
	require.Equal(int64(5*Limit), nn)

	// Write to an empty subvolume again
	// after cleaning others
	for j, subvolume := range subvolumes[:lastUnassigned] {
		switch j {
		case 0, 1:
			require.NoError(os.RemoveAll(filepath.Join(subvolume.path, "QFILE")))
		default:
			require.NoError(
				ioutil.WriteFile(filepath.Join(subvolume.path, "QQFILE"), make([]byte, partToWrite), 0x777),
			)
		}
	}
}

func getBTRFSState(ctx context.Context, path string) ([]byte, error) {
	return exec.CommandContext(ctx, "btrfs", "qgroup", "show", "-r", "-e", "-c", "-p", path).Output()
}

func createSubvolume(ctx context.Context, t *testing.T, subvolumePath string) error {
	output, err := exec.CommandContext(ctx, "btrfs", "subvolume", "create", subvolumePath).CombinedOutput()
	if err != nil {
		t.Logf("%s", output)
	}
	return err
}

func destroySubvolume(ctx context.Context, subvolumePath string) error {
	_, err := exec.CommandContext(ctx, "btrfs", "subvolume", "delete", subvolumePath).Output()
	return err
}
