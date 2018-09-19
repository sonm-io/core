// +build linux

package storage

import (
	"github.com/docker/docker/api/types"
)

// PlatformSupportsQuota says if the platform supports quota
var PlatformSupportsQuota = true

func NewQuotaTuner(info types.Info) (StorageQuotaTuner, error) {
	switch info.Driver {
	case "btrfs":
		return newBtrfsQuotaTuner(info)
	default:
		return nil, ErrDriverNotSupported{driver: info.Driver}
	}
}
