package volume

import (
	"context"
	"fmt"

	"github.com/ContainX/docker-volume-netshare/netshare/drivers"
	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Volume specifies volume interface, that is mounted within Docker
// containers.
type Volume interface {
	// Configure mutates the specified host config applying required settings.
	Configure(mount Mount, cfg *container.HostConfig) error
}

type nilVolume struct {
}

func (nilVolume) Configure(mount Mount, cfg *container.HostConfig) error {
	return nil
}

// VolumeDriver specifies volume driver interface, providing abilities to
// create new volumes for containers.
type VolumeDriver interface {
	// CreateVolume creates a new volume using specified name and Option.
	CreateVolume(name string, options map[string]string) (Volume, error)
	// RemoveVolume removes an existing volume.
	RemoveVolume(name string) error
	// Close closes this driver, freeing all associated resources.
	Close() error
}

type nilVolumeDriver struct{}

func (nilVolumeDriver) CreateVolume(name string, options map[string]string) (Volume, error) {
	return &nilVolume{}, nil
}

func (nilVolumeDriver) RemoveVolume(name string) error {
	return nil
}

func (nilVolumeDriver) Close() error {
	return nil
}

// NewNilVolumeDriver constructs a new volume driver that does nothing.
//
// Used primarily in systems that are unable to manage Docker plugins for to
// be able to at least start the Worker.
func NewNilVolumeDriver() VolumeDriver {
	return &nilVolumeDriver{}
}

// NewVolumeDriver constructs a new volume driver.
func NewVolumeDriver(ctx context.Context, wg *errgroup.Group, ty string, options ...Option) (VolumeDriver, error) {
	ctx = log.WithLogger(ctx, log.G(ctx).With(zap.String("driver", ty)))

	switch ty {
	case drivers.CIFS.String():
		return NewCIFSVolumeDriver(ctx, wg, options...)
	default:
		return nil, fmt.Errorf("unknown volume driver: %s", ty)
	}
}
