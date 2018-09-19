package plugin

import (
	"context"
	"fmt"
	"sort"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	minet "github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/insonmnia/worker/storage"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

const (
	bridgeNetwork = "bridge"
	tincNetwork   = "tinc"
	l2tpNetwork   = "l2tp"
)

// Provider unifies all possible providers for tuning.
type Provider interface {
	GPUProvider
	VolumeProvider
	NetworkProvider
	StorageQuotaProvider
}

// GPUProvider describes an interface for applying GPU settings to the
// container.
type GPUProvider interface {
	IsGPURequired() bool
	GpuDeviceIDs() []gpu.GPUID
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
	// DealID provides the deal ID for the volume is planned to be executed.
	DealID() string
	// Network provides an optional network name and ID indicating under which
	// network the volume plugin should be executed.
	Network() (string, string)
}

type StorageQuotaProvider interface {
	QuotaInBytes() uint64
	QuotaID() string
}

type NetworkProvider interface {
	Networks() []*structs.NetworkSpec
}

// Repository describes a place where all SONM plugins for Docker live.
type Repository struct {
	volumes           map[string]volume.VolumeDriver
	gpuTuners         map[sonm.GPUVendorType]gpu.Tuner
	networkTuners     map[string]minet.Tuner
	storageQuotaTuner storage.StorageQuotaTuner
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

	for ty, options := range cfg.Volumes.Drivers {
		log.G(ctx).Debug("initializing Volume plugin", zap.String("type", ty))

		driver, err := volume.NewVolumeDriver(ctx, ty,
			volume.WithPluginSocketDir(cfg.SocketDir),
			volume.WithOptions(options),
			volume.WithLogger(log.S(ctx)),
		)

		if err != nil {
			return nil, fmt.Errorf("cannot initialize volume plugin \"%s\": %v", ty, err)
		}

		r.volumes[ty] = driver
	}

	for vendor, options := range cfg.GPUs {
		log.G(ctx).Debug("initializing GPU plugin",
			zap.String("vendor", vendor), zap.Any("options", options))

		typeID, err := gpu.GetVendorByName(vendor)
		if err != nil {
			return nil, err
		}

		tuner, err := gpu.New(ctx, typeID, gpu.WithSocketDir(cfg.SocketDir), gpu.WithOptions(options))
		if err != nil {
			return nil, fmt.Errorf("cannot initialize GPU plugin for vendor\"%s\": %v", vendor, err)
		}

		r.gpuTuners[typeID] = tuner
	}

	if cfg.Overlay.Drivers.Tinc != nil {
		tincTuner, err := minet.NewTincTuner(ctx, cfg.Overlay.Drivers.Tinc)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tinc tuner - %v", err)
		}
		r.networkTuners[tincNetwork] = tincTuner
	}

	if cfg.Overlay.Drivers.L2TP != nil {
		l2tpTuner, err := minet.NewL2TPTuner(ctx, cfg.Overlay.Drivers.L2TP)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize l2tp tuner - %v", err)
		}
		r.networkTuners[l2tpNetwork] = l2tpTuner
	}

	if storage.PlatformSupportsQuota {
		// NOTE: not sure it's safe to do it here. Please, suggest better place
		docker, err := client.NewEnvClient()
		if err != nil {
			return nil, nil
		}
		defer docker.Close()

		info, err := docker.Info(ctx)
		if err != nil {
			return nil, err
		}

		log.G(ctx).Debug("fetched Docker info", zap.Any("info", info))

		r.storageQuotaTuner, err = storage.NewQuotaTuner(info)
		switch err.(type) {
		case nil:
			// pass
		case storage.ErrDriverNotSupported:
			log.G(ctx).Warn("storage quota is not supported by current Docker driver", zap.String("driver", info.Driver))
		default:
			return nil, err
		}
	}

	return r, nil
}

// EmptyRepository constructs an empty repository. Used primarily in tests.
func EmptyRepository() *Repository {
	return &Repository{
		volumes:       make(map[string]volume.VolumeDriver),
		gpuTuners:     make(map[sonm.GPUVendorType]gpu.Tuner),
		networkTuners: make(map[string]minet.Tuner),
	}
}

// Tune creates all plugin bound required for the given provider with further host config tuning.
func (r *Repository) Tune(ctx context.Context, provider Provider, hostCfg *container.HostConfig, netCfg *network.NetworkingConfig) (Cleanup, error) {
	log.G(ctx).Info("tuning container")
	// Do not specify GPU type right now,
	// just check that GPU is required
	if provider.IsGPURequired() {
		if err := r.TuneGPU(provider, hostCfg); err != nil {
			return nil, err
		}
	}
	cleanup := newNestedCleanup()
	c, err := r.TuneVolumes(ctx, provider, hostCfg)
	if err != nil {
		return nil, err
	}
	cleanup.Add(c)

	c, err = r.TuneNetworks(ctx, provider, hostCfg, netCfg)
	if err != nil {
		cleanup.Close()
		return nil, err
	}
	cleanup.Add(c)
	return &cleanup, nil
}

// HasGPU returns true if the Repository has at least one GPU plugin loaded
func (r *Repository) HasGPU() bool {
	return len(r.gpuTuners) > 0
}

func (r *Repository) collectGPUDevices() []*sonm.GPUDevice {
	var devs []*sonm.GPUDevice
	for _, tun := range r.gpuTuners {
		devs = append(devs, tun.Devices()...)
	}
	sort.Slice(devs, func(i, j int) bool {
		return devs[i].GetID() < devs[j].GetID()
	})
	return devs
}

// ApplyHardwareInfo exposing info about hardware units controlled by
// various plugins.
func (r *Repository) ApplyHardwareInfo(hw *hardware.Hardware) {
	for _, dev := range r.collectGPUDevices() {
		hw.GPU = append(hw.GPU, &sonm.GPU{
			Device:     dev,
			Benchmarks: make(map[uint64]*sonm.Benchmark),
		})
	}

	hw.Network.NetFlags.SetOverlay(len(r.networkTuners) > 0)
}

// TuneGPU creates GPU bound required for the given provider with further
// host config tuning.
func (r *Repository) TuneGPU(provider GPUProvider, cfg *container.HostConfig) error {
	for _, tuner := range r.gpuTuners {
		err := tuner.Tune(cfg, provider.GpuDeviceIDs())
		if err != nil {
			return err
		}
	}

	return nil
}

// TuneVolumes creates volumes required for the given provider with further
// host config tuning with mount settings.
func (r *Repository) TuneVolumes(ctx context.Context, provider VolumeProvider, cfg *container.HostConfig) (Cleanup, error) {
	cleanup := newNestedCleanup()

	for volumeName, options := range provider.Volumes() {
		networkName, networkID := provider.Network()
		options.Options[volume.OptionNetworkName] = networkName
		options.Options[volume.OptionNetworkID] = networkID

		mounts := provider.Mounts(volumeName)
		// No mounts - no volumes.
		if len(mounts) == 0 {
			continue
		}

		driver, ok := r.volumes[options.Type]
		if !ok {
			cleanup.Close()
			return nil, fmt.Errorf("volume driver not supported: %s", options.Type)
		}

		id := fmt.Sprintf("%s/%s", provider.ID(), volumeName)

		v, err := driver.CreateVolume(id, options.Options)
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

	for _, mount := range cfg.Mounts {
		mount.VolumeOptions.Labels["DealID"] = provider.DealID()
	}

	return &cleanup, nil
}

func (r *Repository) GetCleanup(ctx context.Context, provider Provider) (Cleanup, error) {
	cleanup := newNestedCleanup()

	c, err := r.GetNetworkCleaner(ctx, provider)
	if err != nil {
		return nil, err
	}
	cleanup.Add(c)

	c, err = r.GetVolumeCleaner(ctx, provider)
	if err != nil {
		cleanup.Close()
		return nil, err
	}

	cleanup.Add(c)

	return &cleanup, nil
}

func (r *Repository) GetVolumeCleaner(ctx context.Context, provider VolumeProvider) (Cleanup, error) {
	cleanup := newNestedCleanup()

	for volumeName, options := range provider.Volumes() {
		mounts := provider.Mounts(volumeName)

		if len(mounts) == 0 {
			continue
		}

		driver, ok := r.volumes[options.Type]
		if !ok {
			cleanup.Close()
			return nil, fmt.Errorf("volume driver not supported: %s", options.Type)
		}

		id := fmt.Sprintf("%s/%s", provider.ID(), volumeName)

		cleanup.Add(&volumeCleanup{driver: driver, id: id})
	}

	return &cleanup, nil
}

func (r *Repository) PostCreationTune(ctx context.Context, provider Provider, cleanup Cleanup, ID string) (Cleanup, error) {
	// NOTE: move it to r.TuneStorageQuota
	if r.storageQuotaTuner == nil || provider.QuotaInBytes() == 0 {
		return cleanup, nil
	}
	nCleanup := newNestedCleanup()
	nCleanup.Add(cleanup)
	cleanup, err := r.TuneStorageQuota(ctx, provider, ID)
	if err != nil {
		return nil, err
	}
	nCleanup.Add(cleanup)
	return &nCleanup, nil
}

func (r *Repository) TuneStorageQuota(ctx context.Context, provider StorageQuotaProvider, ID string) (Cleanup, error) {
	return r.storageQuotaTuner.SetQuota(ctx, ID, provider.QuotaID(), provider.QuotaInBytes())
}

func (r *Repository) TuneNetworks(ctx context.Context, provider NetworkProvider, hostCfg *container.HostConfig, netCfg *network.NetworkingConfig) (Cleanup, error) {
	log.G(ctx).Info("tuning networks")
	cleanup := newNestedCleanup()
	networks := provider.Networks()
	for _, net := range networks {
		tuner, ok := r.networkTuners[net.Type]
		if !ok {
			cleanup.Close()
			return nil, fmt.Errorf("network driver not supported: %s", net.Type)
		}
		c, err := tuner.Tune(ctx, net, hostCfg, netCfg)
		if err != nil {
			cleanup.Close()
			return nil, err
		}
		cleanup.Add(c)
	}
	return &cleanup, nil
}

func (r *Repository) GetNetworkCleaner(ctx context.Context, provider NetworkProvider) (Cleanup, error) {
	log.G(ctx).Info("attaching to networks")
	cleanup := newNestedCleanup()

	for _, net := range provider.Networks() {
		tuner, ok := r.networkTuners[net.Type]
		if !ok {
			cleanup.Close()
			return nil, fmt.Errorf("network driver not supported: %s", net.Type)
		}
		c, err := tuner.GetCleaner(ctx, net.NetID)
		if err != nil {
			cleanup.Close()
			return nil, err
		}
		cleanup.Add(c)
	}

	return &cleanup, nil
}

func (r *Repository) JoinNetwork(ID string) (*structs.NetworkSpec, error) {
	for _, net := range r.networkTuners {
		if net.Tuned(ID) {
			return net.GenerateInvitation(ID)
		}
	}
	return nil, fmt.Errorf("no such network %s", ID)
}

func (r *Repository) Close() error {
	errs := make([]error, 0)
	for ty, vol := range r.volumes {
		if err := vol.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %s", ty, err))
		}
	}

	for ty, g := range r.gpuTuners {
		if err := g.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %s", ty.String(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close %d plugins: %v", len(errs), errs)
	} else {
		return nil
	}
}
