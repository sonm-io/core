package volume

import (
	"os"
	"path/filepath"
)

const pluginSockDir = "/run/docker/plugins"

func fullSocketPath(dir, address string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if filepath.IsAbs(address) {
		return address, nil
	}
	return filepath.Join(dir, address+".sock"), nil
}
