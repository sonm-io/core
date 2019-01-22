package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/logging"
	"github.com/sonm-io/core/proto"
)

var _ sonm.InspectServer = &InspectService{}

type ConfigProvider interface {
	// Config returns the configuration that can be encoded as JSON.
	Config() interface{}
}

type AuthSubscriber interface {
	Subscribe(addr common.Address) <-chan struct{}
}

type InspectService struct {
	ps             *process.Process
	dockerClient   *docker.Client
	configProvider ConfigProvider
	authWatcher    AuthSubscriber
	loggingWatcher *logging.WatcherCore
}

func NewInspectService(configProvider ConfigProvider, authWatcher AuthSubscriber, loggingWatcher *logging.WatcherCore) (*InspectService, error) {
	ps, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	dockerClient, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	m := &InspectService{
		ps:             ps,
		dockerClient:   dockerClient,
		configProvider: configProvider,
		loggingWatcher: loggingWatcher,
		authWatcher:    authWatcher,
	}

	return m, nil
}

func (m *InspectService) Config(ctx context.Context, request *sonm.InspectConfigRequest) (*sonm.InspectConfigResponse, error) {
	data, err := json.Marshal(m.configProvider.Config())
	if err != nil {
		return nil, err
	}

	return &sonm.InspectConfigResponse{
		Config: data,
	}, nil
}

func (m *InspectService) OpenFiles(ctx context.Context, request *sonm.InspectOpenFilesRequest) (*sonm.InspectOpenFilesResponse, error) {
	openFiles, err := m.ps.OpenFiles()
	if err != nil {
		return nil, err
	}

	filesStat := make([]*sonm.FileStat, len(openFiles))

	for id, stat := range openFiles {
		filesStat[id] = &sonm.FileStat{
			Fd:   stat.Fd,
			Path: stat.Path,
		}
	}

	return &sonm.InspectOpenFilesResponse{
		OpenFiles: filesStat,
	}, nil
}

func (m *InspectService) Network(ctx context.Context, request *sonm.InspectNetworkRequest) (*sonm.InspectNetworkResponse, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	connections, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	interfacesStat := make([]*sonm.InterfaceStat, 0, len(interfaces))
	for _, netIf := range interfaces {
		addrs := make([]string, 0, len(netIf.Addrs))
		for _, addr := range netIf.Addrs {
			addrs = append(addrs, addr.Addr)
		}

		interfacesStat = append(interfacesStat, &sonm.InterfaceStat{
			Mtu:          int32(netIf.MTU),
			Name:         netIf.Name,
			HardwareAddr: netIf.HardwareAddr,
			Flags:        netIf.Flags,
			Addrs:        addrs,
		})
	}

	connectionsStat := make([]*sonm.ConnectionStat, 0, len(connections))
	for _, conn := range connections {
		connectionsStat = append(connectionsStat, &sonm.ConnectionStat{
			Fd:     uint64(conn.Fd),
			Family: conn.Family,
			Type:   conn.Type,
			LocalAddr: &sonm.SocketAddr{
				Addr: conn.Laddr.IP,
				Port: conn.Laddr.Port,
			},
			RemoteAddr: &sonm.SocketAddr{
				Addr: conn.Raddr.IP,
				Port: conn.Raddr.Port,
			},
			Status: conn.Status,
			Pid:    conn.Pid,
		})
	}

	return &sonm.InspectNetworkResponse{
		Interfaces:  interfacesStat,
		Connections: connectionsStat,
	}, nil
}

func (m *InspectService) HostInfo(ctx context.Context, request *sonm.InspectHostInfoRequest) (*sonm.InspectHostInfoResponse, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	return &sonm.InspectHostInfoResponse{
		Hostname:             hostInfo.Hostname,
		Uptime:               hostInfo.Uptime,
		BootTime:             hostInfo.BootTime,
		ProcessesNumber:      hostInfo.Procs,
		Os:                   hostInfo.OS,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		KernelVersion:        hostInfo.KernelVersion,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,
		HostID:               hostInfo.HostID,
	}, nil
}

func (m *InspectService) DockerInfo(ctx context.Context, request *sonm.InspectDockerInfoRequest) (*sonm.InspectDockerInfoResponse, error) {
	// Not sure it's not changed during time.
	info, err := m.dockerClient.Info(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return &sonm.InspectDockerInfoResponse{
		Info: data,
	}, nil
}

func (m *InspectService) DockerNetwork(ctx context.Context, request *sonm.InspectDockerNetworkRequest) (*sonm.InspectDockerNetworkResponse, error) {
	info, err := m.dockerClient.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return &sonm.InspectDockerNetworkResponse{
		Info: data,
	}, nil
}

func (m *InspectService) DockerVolumes(ctx context.Context, request *sonm.InspectDockerVolumesRequest) (*sonm.InspectDockerVolumesResponse, error) {
	info, err := m.dockerClient.VolumeList(ctx, filters.Args{})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	return &sonm.InspectDockerVolumesResponse{
		Info: data,
	}, nil
}

func (m *InspectService) WatchLogs(request *sonm.InspectWatchLogsRequest, stream sonm.Inspect_WatchLogsServer) error {
	peerInfo, err := auth.FromContext(stream.Context())
	if err != nil {
		return err
	}

	txrx := make(chan string, 1024)
	defer close(txrx)

	id := m.loggingWatcher.Subscribe(txrx)
	defer m.loggingWatcher.Unsubscribe(id)

	expiredChan := m.authWatcher.Subscribe(peerInfo.Addr)

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case message := <-txrx:
			if err := stream.Send(&sonm.InspectWatchLogsChunk{Message: message}); err != nil {
				return err
			}
		case <-expiredChan:
			return fmt.Errorf("inspection capabilities has been expired, the peer %s has no access", peerInfo.Addr.Hex())
		}
	}
}

func (m *InspectService) Close() error {
	return m.dockerClient.Close()
}
