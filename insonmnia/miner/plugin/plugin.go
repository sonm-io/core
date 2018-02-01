package plugin

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/proto"
)

// Provider unifies all possible providers for tuning.
type Provider interface {
	GPUProvider
	VolumeProvider
}

// GPUProvider describes an interface for applying GPU settings to the
// container.
type GPUProvider interface {
}

// VolumeProvider describes an interface for applying volumes to the container.
type VolumeProvider interface {
	// ID returns a unique identifier that will be used as a new volume name.
	ID() string
	// Volumes returns volumes specified for configuring.
	Volumes() map[string]*sonm.Volume
	// Mounts returns all mounts whose source equals to the volume name
	// provided.
	Mounts(source string) []volume.Mount
}

// Repository describes a place where all SONM plugins for Docker live.
type Repository struct {
	volumes map[string]volume.VolumeDriver
}

// NewRepository constructs a new repository for SONM plugins from the
// specified config.
//
// Plugins will be attempted to run inside the given wait group immediately
// during the call of this function. Any error that can be recovered at the
// initialization stage will be returned. Other errors will interrupt the wait
// group, forcing making the entire plugin system to halt.
func NewRepository(ctx context.Context, cfg Config) (*Repository, error) {
	r := EmptyRepository()

	log.G(ctx).Info("initializing SONM plugins")

	for ty, options := range cfg.Volumes.Volumes {
		driver, err := volume.NewVolumeDriver(ctx, ty,
			volume.WithPluginSocketDir(cfg.SocketPath),
			volume.WithOptions(options),
		)

		if err != nil {
			return nil, err
		}

		r.volumes[ty] = driver
	}

	return r, nil
}

// EmptyRepository constructs an empty repository. Used primarily in tests.
func EmptyRepository() *Repository {
	return &Repository{
		volumes: make(map[string]volume.VolumeDriver),
	}
}

// Tune creates all plugin bound required for the given provider with further
// host config tuning.
func (r *Repository) Tune(provider Provider, cfg *container.HostConfig) error {
	if err := r.TuneGPU(provider, cfg); err != nil {
		return err
	}
	if err := r.TuneVolumes(provider, cfg); err != nil {
		return err
	}

	return nil
}

// TuneGPU creates GPU bound required for the given provider with further
// host config tuning.
func (r *Repository) TuneGPU(provider GPUProvider, cfg *container.HostConfig) error {
	return nil
}

// TuneVolumes creates volumes required for the given provider with further
// host config tuning with mount settings.
func (r *Repository) TuneVolumes(provider VolumeProvider, cfg *container.HostConfig) error {
	for volumeName, options := range provider.Volumes() {
		mounts := provider.Mounts(volumeName)
		// No mounts - no volumes.
		if len(mounts) == 0 {
			continue
		}

		driver, ok := r.volumes[options.Driver]
		if !ok {
			return fmt.Errorf("volume driver not supported: %s", options.Driver)
		}

		id := fmt.Sprintf("%s/%s", provider.ID(), volumeName)

		v, err := driver.CreateVolume(id, options.Settings)
		if err != nil {
			return err
		}

		for _, mount := range mounts {
			// We don't trust the provider implementation.
			if mount.Source != volumeName {
				continue
			}

			mount = volume.Mount{
				Source:     id,
				Target:     mount.Target,
				Permission: mount.Permission,
			}

			if err := v.Configure(mount, cfg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Repository) Close() error {
	errors := make([]error, 0)
	for ty, vol := range r.volumes {
		if err := vol.Close(); err != nil {
			errors = append(errors, fmt.Errorf("%s - %s", ty, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d plugins: %v", len(errors), errors)
	} else {
		return nil
	}
}
