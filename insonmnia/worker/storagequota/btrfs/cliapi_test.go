// +build linux

package btrfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupQuotaInShowOutpout(t *testing.T) {
	assert := assert.New(t)
	const longBody = `
WARNING: qgroup data inconsistent, rescan recommended
qgroupid         rfer         excl     max_rfer     max_excl parent  child
--------         ----         ----     --------     -------- ------  -----
0/5         580.00KiB    580.00KiB         none         none ---     ---
0/323        16.00KiB     16.00KiB         none         none ---     ---
0/324        16.00KiB     16.00KiB         none         none ---     ---
0/325        81.54MiB    208.00KiB         none         none ---     ---
0/326        81.55MiB     64.00KiB         none         none ---     ---
0/327        70.67MiB     48.00KiB         none         none ---     ---
0/328        70.67MiB     48.00KiB         none         none ---     1/100
0/329        70.67MiB     48.00KiB         none         none ---     ---
0/330        70.67MiB     80.00KiB         none         none ---     ---
0/331        70.67MiB    256.00KiB         none         none ---     ---
0/332        70.67MiB     16.00KiB         none         none ---     ---
0/333        70.67MiB     16.00KiB         none         none ---     ---
1/100           0.00B        0.00B         none         none 0/328   ---
`
	fixtures := []struct {
		found   bool
		err     error
		body    []byte
		quotaID string
	}{
		{
			found:   true,
			body:    []byte(longBody),
			quotaID: "1/100",
		},
		{
			found:   false,
			body:    []byte(longBody),
			quotaID: "1/101",
		},
		{
			found:   false,
			body:    []byte(""),
			quotaID: "",
		},
	}
	for _, fixture := range fixtures {
		found, err := lookupQuotaInShowOutpout(fixture.body, fixture.quotaID)
		assert.Equal(fixture.found, found)
		assert.Equal(fixture.err, err)
	}
}

func TestLookupIDForSubvolumeWithPath(t *testing.T) {
	assert := assert.New(t)
	const longBody = `
qgroupid         rfer         excl
--------         ----         ----
0/270        16.00KiB     16.00KiB
`
	const shortBody = `
qgroupid         rfer         excl
--------         ----         ----
`
	fixtures := []struct {
		ID   string
		err  error
		body []byte
	}{
		{
			ID:   "0/270",
			body: []byte(longBody),
		},
		{
			ID:   "",
			body: []byte(shortBody),
			err:  errors.New("not found"),
		},
	}
	for _, fixture := range fixtures {
		ID, err := lookupIDForSubvolumeWithPath(fixture.body)
		assert.Equal(fixture.ID, ID)
		assert.Equal(fixture.err, err)
	}
}

func TestE2E(t *testing.T) {
	path := os.Getenv("BTRFS_PLAYGROUND_PATH")
	if path == "" {
		t.Skip("BTRFS_PLAYGROUND_PATH must be set for the test")
	}
	if os.Getenv("SUDO_USER") == "" {
		t.Skip("WARNING: root permissions required for that test")
	}

	var b btrfsCLI
	t.Run(fmt.Sprintf("%T", b), func(t *testing.T) {
		testE2EOne(t, b, path)
	})
}

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
		// t.Logf("%s\n%s\n", state, postState)
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
		require.NoError(createSubvolume(ctx, subvolumePath))
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
			require.NoError(err)
			nn, wrErr := io.CopyN(f, devZero, partToWrite)
			require.NoError(wrErr)
			require.Equal(nn, partToWrite)
			written += nn
			f.Close()
		case 1:
			require.NoError(err)
			nn, wrErr := io.CopyN(f, devZero, partToWrite)
			require.Error(wrErr)
			require.True(nn <= Limit-written)
			written += nn
			f.Close()
		default:
			require.Error(err)
		}
	}
	require.True(written < Limit)

	nn, err := io.CopyN(freeFile, devZero, 10*Limit)
	require.NoError(err)
	require.Equal(nn, int64(10*Limit))

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

func createSubvolume(ctx context.Context, subvolumePath string) error {
	_, err := exec.CommandContext(ctx, "btrfs", "subvolume", "create", subvolumePath).Output()
	return err
}

func destroySubvolume(ctx context.Context, subvolumePath string) error {
	_, err := exec.CommandContext(ctx, "btrfs", "subvolume", "delete", subvolumePath).Output()
	return err
}
