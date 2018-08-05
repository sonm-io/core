package volume

import (
	"fmt"
	"net"
	"path"
	"sync"
	"syscall"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/fs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-plugins-helpers/volume"
	"go.uber.org/zap"
)

const (
	BTFSDriverName = "btfs"
)

type BTFSDriver struct {
	listener net.Listener
	driver   *BTFSDockerDriver
	log      *zap.SugaredLogger
}

func NewBTFSDriver(options ...Option) (*BTFSDriver, error) {
	opts := newOptions()

	for _, option := range options {
		if err := option(opts); err != nil {
			return nil, err
		}
	}

	// TODO: Seems like these lines are duplicated...
	baseDir := path.Join(volume.DefaultDockerRootDirectory, BTFSDriverName)
	driverName := BTFSDriverName

	//rootDir := filepath.Join(baseDir, driverName)
	socketPath, err := fullSocketPath(opts.socketDir, driverName)
	if err != nil {
		return nil, err
	}

	listener, err := sockets.NewUnixSocket(socketPath, syscall.Getgid())
	if err != nil {
		return nil, err
	}

	// TODO: ... till now.

	m := &BTFSDriver{
		listener: listener,
		driver: &BTFSDockerDriver{
			baseDir: baseDir,
			volumes: map[string]*BTFSDockerVolume{},
			log:     opts.log.With("driver", BTFSDriverName),
		},
		log: opts.log.With("driver", BTFSDriverName),
	}

	go m.serve()

	return m, nil
}

func (m *BTFSDriver) serve() error {
	m.log.Infof("exposing BTFS volume plugin on %s", m.listener.Addr())
	defer m.log.Info("BTFS volume plugin has been stopped")

	return volume.NewHandler(m.driver).Serve(m.listener)
}

func (m *BTFSDriver) CreateVolume(name string, options map[string]string) (Volume, error) {
	m.log.Debugw("handling Worker.CreateVolume request", zap.String("name", name), zap.Any("options", options))
	return &BTFSVolume{}, nil
}

func (m *BTFSDriver) RemoveVolume(name string) error {
	m.log.Debug("handling Worker.RemoveVolume request")
	return nil
}

func (m *BTFSDriver) Close() error {
	m.log.Info("shutting down volume plugin")

	return m.listener.Close()
}

type BTFSVolume struct{}

func (m *BTFSVolume) Configure(mnt Mount, cfg *container.HostConfig) error {
	cfg.Mounts = append(cfg.Mounts, mount.Mount{
		Type:        mount.TypeVolume,
		Source:      mnt.Source,
		Target:      mnt.Target,
		ReadOnly:    mnt.ReadOnly(),
		Consistency: mount.ConsistencyDefault,

		BindOptions: nil,
		VolumeOptions: &mount.VolumeOptions{
			// Should volume be populated with the data from the target.
			NoCopy: false,
			Labels: map[string]string{}, // TODO: < DealID?
			DriverConfig: &mount.Driver{
				Name:    BTFSDriverName,
				Options: map[string]string{},
			},
		},
		TmpfsOptions: nil,
	})

	return nil
}

type BTFSDockerDriver struct {
	baseDir string

	mu      sync.Mutex
	volumes map[string]*BTFSDockerVolume
	log     *zap.SugaredLogger
}

func (m *BTFSDockerDriver) Create(request *volume.CreateRequest) error {
	m.log.Debugw("handling Volume.Create request", zap.Any("request", request))

	// Seems like we have to create torrent client per task... such a waste of
	// resources.
	cfg := torrent.NewDefaultClientConfig()
	// TODO: Fill these.
	cfg.DataDir = m.baseDir // TODO: Really?
	//cfg.UploadRateLimiter = nil
	//cfg.DownloadRateLimiter = nil
	// TODO: Assign to bridge to enable shaping.

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create torrent client: %v", err)
	}

	fs := torrentfs.New(client)
	node, err := fs.Root()
	if err != nil {
		return err
	}

	m.log.Debugw("NODE", zap.Any("node", node))

	m.mu.Lock()
	defer m.mu.Unlock()
	m.volumes[request.Name] = &BTFSDockerVolume{
		Client:     client,
		FS:         fs,
		MountPoint: "TODO",
	}

	return nil
}

func (m *BTFSDockerDriver) List() (*volume.ListResponse, error) {
	m.log.Debug("handling Volume.List request")

	panic("implement me")
}

func (m *BTFSDockerDriver) Get(request *volume.GetRequest) (*volume.GetResponse, error) {
	m.log.Debugw("handling Volume.Get request", zap.Any("request", request))

	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.volumes[request.Name]
	if !ok {
		return &volume.GetResponse{}, fmt.Errorf("volume %s not found", request.Name)
	}

	return &volume.GetResponse{Volume: &volume.Volume{Name: request.Name, Mountpoint: v.MountPoint}}, nil
}

func (m *BTFSDockerDriver) Remove(request *volume.RemoveRequest) error {
	m.log.Debugw("handling Volume.Remove request", zap.Any("request", request))

	panic("implement me")
}

func (m *BTFSDockerDriver) Path(request *volume.PathRequest) (*volume.PathResponse, error) {
	m.log.Debugw("handling Volume.Path request", zap.Any("request", request))

	panic("implement me")
}

func (m *BTFSDockerDriver) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	m.log.Debugw("handling Volume.Mount request", zap.Any("request", request))

	panic("implement me")
}

func (m *BTFSDockerDriver) Unmount(request *volume.UnmountRequest) error {
	m.log.Debugw("handling Volume.Unmount request", zap.Any("request", request))

	panic("implement me")
}

func (m *BTFSDockerDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local",
		},
	}
}

type BTFSDockerVolume struct {
	Client     *torrent.Client
	FS         *torrentfs.TorrentFS
	MountPoint string
}
