package plugin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Provider unifies all possible providers for tuning.
type Provider interface {
	GPUProvider
	VolumeProvider
}

// GPUProvider describes an interface for applying GPU settings to the
// container.
type GPUProvider interface {
	GPU() bool
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
	volumes   map[string]volume.VolumeDriver
	gpuTuners map[sonm.GPUVendorType]gpu.Tuner
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
		log.G(ctx).Debug("initializing Volume plugin", zap.String("type", ty))

		driver, err := volume.NewVolumeDriver(ctx, ty,
			volume.WithPluginSocketDir(cfg.SocketDir),
			volume.WithOptions(options),
		)

		if err != nil {
			return nil, fmt.Errorf("cannnot initialize volume plugin \"%s\": %v", ty, err)
		}

		r.volumes[ty] = driver
	}

	for vendor, options := range cfg.GPUs {
		log.G(ctx).Debug("initializing GPU plugin",
			zap.String("vendor", vendor), zap.Any("options", options))

		vendorName := strings.ToUpper(vendor)
		t, ok := sonm.GPUVendorType_value[vendorName]
		if !ok {
			return nil, errors.New("unknown GPU vendor type")
		}

		typeID := sonm.GPUVendorType(t)

		tuner, err := gpu.New(ctx, typeID, gpu.WithSocketDir(cfg.SocketDir), gpu.WithOptions(options))
		if err != nil {
			return nil, fmt.Errorf("cannnot initialize GPU plugin for vendor\"%s\": %v", vendor, err)
		}

		r.gpuTuners[typeID] = tuner
	}

	return r, nil
}

// EmptyRepository constructs an empty repository. Used primarily in tests.
func EmptyRepository() *Repository {
	return &Repository{
		volumes:   make(map[string]volume.VolumeDriver),
		gpuTuners: make(map[sonm.GPUVendorType]gpu.Tuner),
	}
}

// Tune creates all plugin bound required for the given provider with further
// host config tuning.
func (r *Repository) Tune(provider Provider, cfg *container.HostConfig) (Cleanup, error) {
	// Do not specify GPU type right now,
	// just check that GPU is required
	if provider.GPU() {
		err := r.TuneGPU(provider, cfg)
		if err != nil {
			return nil, err
		}
	}

	return r.TuneVolumes(provider, cfg)
}

// HasGPU returns true if the Repository has at least one GPU plugin loaded
func (r *Repository) HasGPU() bool {
	return len(r.gpuTuners) > 0
}

// TuneGPU creates GPU bound required for the given provider with further
// host config tuning.
func (r *Repository) TuneGPU(provider GPUProvider, cfg *container.HostConfig) error {

	for _, tuner := range r.gpuTuners {
		err := tuner.Tune(cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

// TuneVolumes creates volumes required for the given provider with further
// host config tuning with mount settings.
func (r *Repository) TuneVolumes(provider VolumeProvider, cfg *container.HostConfig) (Cleanup, error) {
	cleanup := newNestedCleanup()

	for volumeName, options := range provider.Volumes() {
		mounts := provider.Mounts(volumeName)
		// No mounts - no volumes.
		if len(mounts) == 0 {
			continue
		}

		driver, ok := r.volumes[options.Driver]
		if !ok {
			cleanup.Close()
			return nil, fmt.Errorf("volume driver not supported: %s", options.Driver)
		}

		id := fmt.Sprintf("%s/%s", provider.ID(), volumeName)

		v, err := driver.CreateVolume(id, options.Settings)
		if err != nil {
			cleanup.Close()
			return nil, err
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
				cleanup.Close()
				return nil, err
			}
		}

		cleanup.Add(&volumeCleanup{driver: driver, id: id})
	}

	return &cleanup, nil
}

func (r *Repository) Close() error {
	errs := make([]error, 0)
	for ty, vol := range r.volumes {
		if err := vol.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s - %s", ty, err))
		}
	}

	for ty, g := range r.gpuTuners {
		if err := g.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s - %s", ty.String(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d plugins: %v", len(errs), errs)
	} else {
		return nil
	}
}
