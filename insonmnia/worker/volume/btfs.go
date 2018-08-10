package volume

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"bazil.org/fuse"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sonm-io/core/util/xdocker"
	"go.uber.org/zap"
)

const (
	BTFSDriverName             = "btfs"
	BTFSImage                  = "sonm/btfs@sha256:b64b4a7849aa742049b039aeabcaff0fa58876c00df9abdf0321d3fff31789a2"
	DefaultDockerRootDirectory = volume.DefaultDockerRootDirectory
)

type BTFSDriver struct {
	server *VolumeServer
	log    *zap.SugaredLogger
}

func NewBTFSDriver(options ...Option) (*BTFSDriver, error) {
	server, err := NewVolumeServer(BTFSDriverName, options...)
	if err != nil {
		return nil, err
	}

	client, err := docker.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client for BTFS volume driver: %v", err)
	}

	driver, err := NewBTFSDockerDriver(client, server.log)
	if err != nil {
		return nil, fmt.Errorf("failed to create BTFS Docker driver: %v", err)
	}

	go server.Serve(driver)

	m := &BTFSDriver{
		server: server,
		log:    server.log.With("driver", BTFSDriverName),
	}

	return m, nil
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
	return m.server.Close()
}

type BTFSVolume struct {
	Options map[string]string
}

func (m *BTFSVolume) Configure(mnt Mount, cfg *container.HostConfig) error {
	cfg.Mounts = append(cfg.Mounts, mount.Mount{
		Type:        mount.TypeVolume,
		Source:      mnt.Source,
		Target:      mnt.Target,
		ReadOnly:    true,
		Consistency: mount.ConsistencyDefault,

		BindOptions: nil,
		VolumeOptions: &mount.VolumeOptions{
			NoCopy: false,
			Labels: map[string]string{},
			DriverConfig: &mount.Driver{
				Name:    BTFSDriverName,
				Options: m.Options,
			},
		},
		TmpfsOptions: nil,
	})

	return nil
}

func pullImage(ctx context.Context, client *docker.Client, ref string) error {
	images, err := client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return err
	}

	for _, summary := range images {
		if summary.ID == BTFSImage {
			return nil
		}
	}

	body, err := client.ImagePull(ctx, BTFSImage, types.ImagePullOptions{All: false})
	if err != nil {
		return fmt.Errorf("failed to pull %s image: %v", ref, err)
	}

	if err = xdocker.DecodeImagePull(body); err != nil {
		return fmt.Errorf("failed to pull %s image: %v", ref, err)
	}

	return nil
}

type BTFSDockerVolume struct {
	Client      *docker.Client
	MagnetURI   string
	MountPoint  string
	ID          string
	Connections int
	NetworkName string
	NetworkID   string
}

// BTFSDockerDriver is a Docker driver implementation for BTFS volume plugin.
type BTFSDockerDriver struct {
	client       *docker.Client
	mountRootDir string

	mu      sync.Mutex
	volumes map[string]*BTFSDockerVolume
	log     *zap.SugaredLogger
}

func NewBTFSDockerDriver(client *docker.Client, log *zap.SugaredLogger) (*BTFSDockerDriver, error) {
	ctx := context.Background()

	if err := pullImage(ctx, client, BTFSImage); err != nil {
		return nil, err
	}

	m := &BTFSDockerDriver{
		client:       client,
		mountRootDir: filepath.Join(DefaultDockerRootDirectory, BTFSDriverName, "mnt"),
		volumes:      map[string]*BTFSDockerVolume{},
		log:          log.With("driver", BTFSDriverName),
	}

	return m, nil
}

func (m *BTFSDockerDriver) Create(request *volume.CreateRequest) error {
	m.log.Debugw("handling Volume.Create request", zap.Any("request", request))

	uri, ok := request.Options["magnet"]
	if !ok {
		return fmt.Errorf("`magnet` URI is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.volumes[request.Name] = &BTFSDockerVolume{
		Client:      m.client,
		MagnetURI:   uri,
		MountPoint:  filepath.Join(m.mountRootDir, request.Name),
		Connections: 0,
		NetworkName: request.Options[OptionNetworkName],
		NetworkID:   request.Options[OptionNetworkID],
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

	return &volume.ListResponse{Volumes: volumes}, nil
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

		cfg := &container.Config{
			Image:   BTFSImage,
			Labels:  map[string]string{},
			Volumes: map[string]struct{}{},
			Cmd:     []string{"/usr/bin/btfs", v.MagnetURI, "/root/mnt", "-f"},
		}
		hostConfig := &container.HostConfig{
			CapAdd:          []string{"SYS_ADMIN"},
			Privileged:      true,
			SecurityOpt:     []string{"apparmor:unconfined"},
			PublishAllPorts: true,
			Mounts: []mount.Mount{
				{
					Type:        mount.TypeBind,
					Source:      v.MountPoint,
					Target:      "/root/mnt",
					ReadOnly:    false,
					Consistency: mount.ConsistencyDefault,

					BindOptions: &mount.BindOptions{
						Propagation: mount.PropagationRShared,
					},
					VolumeOptions: nil,
					TmpfsOptions:  nil,
				},
			},
		}

		networkConfig := &network.NetworkingConfig{}

		if len(v.NetworkName) > 0 && len(v.NetworkID) > 0 {
			m.log.Debugf("configuring %s volume with %s network", request.Name, v.NetworkName)

			networkConfig.EndpointsConfig = map[string]*network.EndpointSettings{
				v.NetworkName: {
					NetworkID: v.NetworkID,
				},
			}
		}

		ctx := context.Background()

		response, err := m.client.ContainerCreate(ctx, cfg, hostConfig, networkConfig, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create btfs container for %s volume: %v", request.Name, err)
		}
		if err := m.client.ContainerStart(ctx, response.ID, types.ContainerStartOptions{}); err != nil {
			return nil, fmt.Errorf("failed to start btfs container for %s volume: %v", request.Name, err)
		}

		v.ID = response.ID
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

		ctx := context.Background()
		if err := m.client.ContainerStop(ctx, v.ID, nil); err != nil {
			return fmt.Errorf("failed to stop btfs container for %s volume: %v", request.Name, err)
		}
		if err := m.client.ContainerRemove(ctx, v.ID, types.ContainerRemoveOptions{}); err != nil {
			return fmt.Errorf("failed to clean up btfs container for %s volume: %v", request.Name, err)
		}
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
