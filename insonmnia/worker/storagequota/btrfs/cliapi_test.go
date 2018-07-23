// +build linux

package btrfs

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		found, err := lookupQuotaInShowOutput(fixture.body, fixture.quotaID)
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
