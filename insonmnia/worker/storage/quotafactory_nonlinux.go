// +build !linux

package storage

import (
	"fmt"
	"runtime"

	"github.com/docker/docker/api/types"
)

// PlatformSupportsQuota says if the platform supports quota
var PlatformSupportsQuota = false

func NewQuotaTuner(info types.Info) (StorageQuotaTuner, error) {
	return nil, fmt.Errorf("quota not supported by this platform: %s", runtime.GOOS)
}
