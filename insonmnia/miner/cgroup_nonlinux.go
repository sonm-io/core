// +build !linux

package miner

import (
	"fmt"
	"runtime"
)

const (
	platformSupportCGroups = false
	parentCgroup           = ""
)

func initializeControlGroup() (cGroupDeleter, error) {
	return nil, fmt.Errorf("%s does not support Control Groups", runtime.GOOS)
}
