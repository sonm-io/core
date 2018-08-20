// Common Internet File System (CIFS) Docker integration.

package volume

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"syscall"

	"github.com/ContainX/docker-volume-netshare/netshare/drivers"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
)

type cifsVolumeDriver struct {
	drivers.CifsDriver

	listener net.Listener

	logger *zap.Logger
}

// NewCIFSVolumeDriver constructs and runs a new CIFS volume driver within the
// provided error group.
func NewCIFSVolumeDriver(ctx context.Context, options ...Option) (VolumeDriver, error) {
	opts := newOptions()

	for _, option := range options {
		if err := option(opts); err != nil {
			return nil, err
		}
	}

	baseDir := volume.DefaultDockerRootDirectory
	driverName := drivers.CIFS.String()

	rootDir := filepath.Join(baseDir, driverName)
	socketPath, err := fullSocketPath(opts.socketDir, driverName)
	if err != nil {
		return nil, err
	}

	listener, err := sockets.NewUnixSocket(socketPath, syscall.Getgid())
	if err != nil {
		return nil, err
	}

	driver := drivers.NewCIFSDriver(rootDir, &drivers.CifsCreds{}, "", "")
	handle := volume.NewHandler(driver)

	go func() {
		log.G(ctx).Info("CIFS volume plugin has been initialized")
		handle.Serve(listener)
	}()

	return &cifsVolumeDriver{driver, listener, log.G(ctx)}, nil
}

func (d *cifsVolumeDriver) CreateVolume(name string, options map[string]string) (Volume, error) {
	d.logger.Info("creating volume", zap.String("name", name))

	if version, ok := options["vers"]; ok {
		options[drivers.CifsOpts] = fmt.Sprintf("vers=\"%s\"", version)
	}

	request := &volume.CreateRequest{
		Name:    name,
		Options: options,
	}

	if err := d.Create(request); err != nil {
		return nil, err
	}

	return &cifsVolume{}, nil
}

func (d *cifsVolumeDriver) RemoveVolume(name string) error {
	d.logger.Info("removing volume", zap.String("name", name))

	request := &volume.RemoveRequest{
		Name: name,
	}

	if err := d.Remove(request); err != nil {
		return err
	}

	return nil
}

func (d *cifsVolumeDriver) Close() error {
	d.logger.Info("shutting down volume plugin")
	return d.listener.Close()
}

type cifsVolume struct {
}

func (v *cifsVolume) Configure(m Mount, cfg *container.HostConfig) error {
	cfg.Mounts = append(cfg.Mounts, mount.Mount{
		Type:        mount.TypeVolume,
		Source:      m.Source,
		Target:      m.Target,
		ReadOnly:    m.ReadOnly(),
		Consistency: mount.ConsistencyDefault,

		BindOptions: nil,
		VolumeOptions: &mount.VolumeOptions{
			// Should volume be populated with the data from the target.
			NoCopy: false,
			Labels: map[string]string{},
			DriverConfig: &mount.Driver{
				Name:    drivers.CIFS.String(),
				Options: map[string]string{},
			},
		},
		TmpfsOptions: nil,
	})

	return nil
}
