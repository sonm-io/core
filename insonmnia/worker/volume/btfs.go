package volume

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
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
	socketPath, err := fullSocketPath(opts.socketDir, BTFSDriverName)
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
			downloadRootDir: filepath.Join(volume.DefaultDockerRootDirectory, BTFSDriverName, "download"),
			mountRootDir:    filepath.Join(volume.DefaultDockerRootDirectory, BTFSDriverName, "mnt"),
			volumes:         map[string]*BTFSDockerVolume{},
			log:             opts.log.With("driver", BTFSDriverName),
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

	return &BTFSVolume{Options: options}, nil
}

func (m *BTFSDriver) RemoveVolume(name string) error {
	m.log.Debug("handling Worker.RemoveVolume request")
	return nil
}

func (m *BTFSDriver) Close() error {
	m.log.Info("shutting down volume plugin")

	return m.listener.Close()
}

type BTFSVolume struct {
	Options map[string]string
}

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
				Options: m.Options,
			},
		},
		TmpfsOptions: nil,
	})

	return nil
}

type BTFSDockerVolume struct {
	Client      *torrent.Client
	MountPoint  string
	FuseConn    *fuse.Conn
	Connections int
}

type BTFSDockerDriver struct {
	mountRootDir    string
	downloadRootDir string

	mu      sync.Mutex
	volumes map[string]*BTFSDockerVolume
	log     *zap.SugaredLogger
}

func (m *BTFSDockerDriver) Create(request *volume.CreateRequest) error {
	m.log.Debugw("handling Volume.Create request", zap.Any("request", request))

	// Seems like we have to create torrent client per task... such a waste of
	// resources.
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = filepath.Join(m.downloadRootDir, request.Name)
	cfg.NoUpload = true
	cfg.Debug = true
	// TODO: Assign to bridge to enable shaping - cfg.SetListenAddr(?).
	uri, ok := request.Options["magnet"]
	if !ok {
		return fmt.Errorf("`magnet` link is required")
	}

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create torrent client: %v", err)
	}

	if _, err := client.AddMagnet(uri); err != nil {
		return fmt.Errorf("failed to add magnet URI: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.volumes[request.Name] = &BTFSDockerVolume{
		Client:      client,
		MountPoint:  filepath.Join(m.mountRootDir, request.Name),
		FuseConn:    nil,
		Connections: 0,
	}

	return nil
}

func (m *BTFSDockerDriver) List() (*volume.ListResponse, error) {
	m.log.Debug("handling Volume.List request")

	m.mu.Lock()
	defer m.mu.Unlock()

	volumes := make([]*volume.Volume, 0, len(m.volumes))

	for id, v := range m.volumes {
		volumes = append(volumes, &volume.Volume{
			Name:       id,
			Mountpoint: v.MountPoint,
		})
	}

	response := &volume.ListResponse{
		Volumes: volumes,
	}

	return response, nil
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

	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok := m.volumes[request.Name]; ok {
		m.log.Infof("removing `%s` volume on `%s`", request.Name, v.MountPoint)
		delete(m.volumes, request.Name)
		return nil
	}

	return fmt.Errorf("volume `%s` not found", request.Name)
}

func (m *BTFSDockerDriver) Path(request *volume.PathRequest) (*volume.PathResponse, error) {
	m.log.Debugw("handling Volume.Path request", zap.Any("request", request))

	m.mu.Lock()
	defer m.mu.Unlock()

	if v, ok := m.volumes[request.Name]; ok {
		return &volume.PathResponse{Mountpoint: v.MountPoint}, nil
	}

	return nil, fmt.Errorf("volume `%s` not found", request.Name)
}

func (m *BTFSDockerDriver) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	m.log.Debugw("handling Volume.Mount request", zap.Any("request", request))

	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.volumes[request.Name]
	if !ok {
		return nil, fmt.Errorf("volume `%s` not found", request.Name)
	}

	if v.Connections == 0 {
		stat, err := os.Lstat(v.MountPoint)

		switch {
		case os.IsNotExist(err):
			if err := os.MkdirAll(v.MountPoint, 0755); err != nil {
				return nil, fmt.Errorf("failed to create %s directories for %s volume: %v", v.MountPoint, request.Name, err)
			}
		case err != nil:
			return nil, fmt.Errorf("failed to perform lstat %s volume: %v", request.Name, err)
		case stat != nil && !stat.IsDir():
			return nil, fmt.Errorf("failed to create %s directiries for %s volume: already exist and it's not a directory", v.MountPoint, request.Name)
		}

		conn, err := fuse.Mount(v.MountPoint)
		if err != nil {
			return nil, fmt.Errorf("failed to mount fuse for %s volume: %v", request.Name, err)
		}

		v.FuseConn = conn
		go fs.Serve(conn, torrentfs.New(v.Client))
	}

	v.Connections++

	return &volume.MountResponse{Mountpoint: v.MountPoint}, nil
}

func (m *BTFSDockerDriver) Unmount(request *volume.UnmountRequest) error {
	m.log.Debugw("handling Volume.Unmount request", zap.Any("request", request))

	m.mu.Lock()
	defer m.mu.Unlock()

	v, ok := m.volumes[request.Name]
	if !ok {
		return fmt.Errorf("volume `%s` not found", request.Name)
	}

	v.Connections--

	if v.Connections <= 0 {
		// Weird case when Docker calls "Unmount" before "Mount".
		v.Connections = 0

		if err := fuse.Unmount(v.MountPoint); err != nil {
			return fmt.Errorf("failed to unmount fuse for %s volume: %v", request.Name, err)
		}
	}

	return nil
}

func (m *BTFSDockerDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local",
		},
	}
}
