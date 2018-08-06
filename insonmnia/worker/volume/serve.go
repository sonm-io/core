package volume

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
	"go.uber.org/zap"
)

const defaultPluginSockDir = "/run/docker/plugins"

func fullSocketPath(dir, address string) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if filepath.IsAbs(address) {
		return address, nil
	}
	return filepath.Join(dir, address+".sock"), nil
}

type VolumeServer struct {
	name     string
	listener net.Listener
	log      *zap.SugaredLogger
}

func NewVolumeServer(name string, options ...Option) (*VolumeServer, error) {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	if err := os.MkdirAll(opts.socketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for %s plugin: %v", name, err)
	}

	path := filepath.Join(opts.socketDir, fmt.Sprintf("%s.sock", name))

	listener, err := sockets.NewUnixSocket(path, syscall.Getgid())
	if err != nil {
		return nil, err
	}

	m := &VolumeServer{
		name:     name,
		listener: listener,
		log:      opts.log,
	}

	return m, nil
}

func (m *VolumeServer) Serve(driver volume.Driver) error {
	m.log.Infof("exposing %s volume plugin on %s", m.name, m.listener.Addr())
	defer m.log.Infof("%s volume plugin has been stopped", m.name)

	return volume.NewHandler(driver).Serve(m.listener)
}

func (m *VolumeServer) Close() error {
	return m.listener.Close()
}
