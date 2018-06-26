// +build windows

package gpu

import (
	"fmt"
	"runtime"
)

func deviceNumber(path string) (uint64, uint64, error) {
	return 0, 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}
