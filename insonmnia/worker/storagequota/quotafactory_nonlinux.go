// +build !linux

package storagequota

import (
	"fmt"
	"runtime"

	"github.com/docker/docker/api/types"
)

// PlatformSupportsQuota says if the platform supports quota
var PlatformSupportsQuota = false

func NewQuotaTuner(info types.Info) (StorageQuotaTuner, error) {
	return nil, fmt.Errorf("Quota not supported by this platform: %s", runtime.GOOS)
}
