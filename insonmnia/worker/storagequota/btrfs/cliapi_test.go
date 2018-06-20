package btrfs

import (
	"errors"
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
		found, err := lookupQuotaInShowOutpout(fixture.body, fixture.quotaID)
		assert.Equal(fixture.found, found)
		assert.Equal(fixture.err, err)
	}
}

func TestLookupIDForSubvolumeWithPath(t *testing.T) {
	assert := assert.New(t)
	const longBody = `
ID	gen	top level	path
--	---	---------	----
323	996	5		btrfs/subvolumes/d727e93c203c5eea60df8067ab0f28cea2fae7127b05ebfc896d12bdf63f8948-init
324	996 5		btrfs/subvolumes/978e9fbb1281b0000dcbb758a99d3729f5ca39f593f5b5003d7feb041bed983e
325	1008	5		btrfs/subvolumes/6b6f4ef74a041a82830756c7a1aee62a91548198e2ee7e7cca87e41bd89cf0bf
326	1010	5		btrfs/subvolumes/570a1e4305f38422dfdcba25e670a04e7a3fc7a8da743b989e1a9469f99d48dc
327	1012	5		btrfs/subvolumes/b561c990bfaef4d60576dccb9665e092d0b1c158948ff6444ff4a50a4bdc6f27
328	1014	5		btrfs/subvolumes/40b44d13464dec7edf9cba24f8d0f61373d08488dc3648b7b41b6d7ba0dc5fea
329	1022	5		btrfs/subvolumes/ef009bb57192c3faba5d5008fec989df4e53ea900e9a165e2992daf7db20f363
330	1017	5		btrfs/subvolumes/d59d8eb3e1abb19a3920c9e512b0e0a6cdafc000b20eda010eed35403618c55c-init
331	1018	5		btrfs/subvolumes/d59d8eb3e1abb19a3920c9e512b0e0a6cdafc000b20eda010eed35403618c55c
332	1023	5		btrfs/subvolumes/538a08e49e3b320630f9524d1897ea76d51c1417417eba4066ff5fe8b2142cf3-init
333	1023	5		btrfs/subvolumes/538a08e49e3b320630f9524d1897ea76d51c1417417eba4066ff5fe8b2142cf3
`
	fixtures := []struct {
		ID            string
		err           error
		body          []byte
		subvolumePath string
	}{
		{
			ID:            "331",
			body:          []byte(longBody),
			subvolumePath: "btrfs/subvolumes/d59d8eb3e1abb19a3920c9e512b0e0a6cdafc000b20eda010eed35403618c55c",
		},
		{
			ID:            "",
			body:          []byte(longBody),
			err:           errors.New("not found"),
			subvolumePath: "btrfs/subvolumes/d59d8eb3e1abb19a3920c9e512b0e0a6cdafc000b20eda010eed35403618c",
		},
	}
	for _, fixture := range fixtures {
		ID, err := lookupIDForSubvolumeWithPath(fixture.body, fixture.subvolumePath)
		assert.Equal(fixture.ID, ID)
		assert.Equal(fixture.err, err)
	}
}
