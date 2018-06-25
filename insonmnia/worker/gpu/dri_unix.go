// +build !windows

package gpu

import (
	"fmt"
	"syscall"
)

func deviceNumber(path string) (uint64, uint64, error) {
	stat := syscall.Stat_t{}
	err := syscall.Stat(path, &stat)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot get device major:minor numbers: %v", err)
	}

	major := uint64(stat.Rdev / 256)
	minor := uint64(stat.Rdev % 256)

	return major, minor, nil
}
