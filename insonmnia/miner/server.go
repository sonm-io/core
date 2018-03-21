package miner

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/ccding/go-stun/stun"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx context.Context

	cfg Config

	plugins *plugin.Repository

	hardware  *hardware.Hardware
	resources *resource.Pool

	hubKey *ecdsa.PublicKey

	publicIPs []string
	natType   stun.NATType

	ovs Overseer

	mu sync.Mutex
	// One-to-one mapping between container IDs and userland task names.
	//
	// The overseer operates with containers in terms of their ID, which does not change even during auto-restart.
	// However some requests pass an application (or task) name, which is more meaningful for user. To be able to
	// transform between these two identifiers this map exists.
	//
	// WARNING: This must be protected using `mu`.
	nameMapping map[string]string

	// Maps StartRequest's IDs to containers' IDs
	// TODO: It's doubtful that we should keep this map here instead in the Overseer.
	containers map[string]*ContainerInfo

	statusChannels map[int]chan bool
	channelCounter int
	controlGroup   cGroup
	cGroupManager  cGroupManager
	ssh            SSH
	state          *state
}

func NewMiner(cfg Config, opts ...Option) (m *Miner, err error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if cfg == nil {
		return nil, errors.New("config is mandatory for MinerBuilder")
	}

	if o.key == nil {
		return nil, errors.New("private key is mandatory")
	}

	if o.ctx == nil {
		o.ctx = context.Background()
	}

	if o.hardware == nil {
		o.hardware = hardware.New()
	}

	hardwareInfo, err := o.hardware.Info()
	if err != nil {
		return nil, err
	}

	cgroup, cGroupManager, err := makeCgroupManager(cfg.HubResources())
	if err != nil {
		return nil, err
	}

	state, err := NewState(o.ctx, cfg)
	if err != nil {
		return nil, err
	}

	if !platformSupportCGroups && cfg.HubResources() != nil {
		log.G(o.ctx).Warn("your platform does not support CGroup, but the config has resources section")
	}

	if err := o.setupNetworkOptions(cfg); err != nil {
		return nil, errors.Wrap(err, "failed to set up network options")
	}

	log.G(o.ctx).Info("discovered public IPs",
		zap.Any("public IPs", o.publicIPs),
		zap.Any("nat", o.nat))

	plugins, err := plugin.NewRepository(o.ctx, cfg.Plugins())
	if err != nil {
		return nil, err
	}

	// apply info about GPUs, expose to logs
	plugins.ApplyHardwareInfo(hardwareInfo)
	log.G(o.ctx).Info("collected hardware info", zap.Any("hw", hardwareInfo))

	if o.ssh == nil {
		o.ssh = nilSSH{}
	}

	if o.ovs == nil {
		o.ovs, err = NewOverseer(o.ctx, plugins)
		if err != nil {
			return nil, err
		}
	}

	m = &Miner{
		ctx: o.ctx,

		cfg: cfg,

		ovs: o.ovs,

		plugins: plugins,

		hardware:  hardwareInfo,
		resources: resource.NewPool(hardwareInfo),

		publicIPs: o.publicIPs,
		natType:   o.nat,

		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
		nameMapping:    make(map[string]string),

		controlGroup:  cgroup,
		cGroupManager: cGroupManager,
		ssh:           o.ssh,
		state:         state,
	}

	return m, nil
}

type resourceHandle interface {
	// Commit marks the handle that the resources consumed should not be
	// released.
	commit()
	// Release releases consumed resources.
	// Useful in conjunction with defer.
	release()
}

// NilResourceHandle is a resource handle that does nothing.
type nilResourceHandle struct{}

func (h *nilResourceHandle) commit() {}

func (h *nilResourceHandle) release() {}

type ownedResourceHandle struct {
	miner     *Miner
	usage     resource.Resources
	committed bool
}

func (h *ownedResourceHandle) commit() {
	h.committed = true
}

func (h *ownedResourceHandle) release() {
	if h.committed {
		return
	}

	h.miner.resources.Release(&h.usage)
	h.committed = true
}

func (m *Miner) saveContainerInfo(id string, info ContainerInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nameMapping[info.ID] = id
	m.containers[id] = &info
}

func (m *Miner) GetContainerInfo(id string) (*ContainerInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.containers[id]
	return info, ok
}

func (m *Miner) getTaskIdByContainerId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name, ok := m.nameMapping[id]
	return name, ok
}

func (m *Miner) getContainerIdByTaskId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.containers[id]
	if ok {
		return info.ID, ok
	}
	return "", ok
}

func (m *Miner) deleteTaskMapping(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nameMapping, id)
}

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(ctx context.Context, _ *pb.Empty) (*pb.PingReply, error) {
	log.G(m.ctx).Info("got ping request from Hub")
	return &pb.PingReply{}, nil
}

// Info returns runtime statistics collected from all containers working on this miner.
//
// This works the following way: a miner periodically collects various runtime statistics from all
// spawned containers that it knows about. For running containers metrics map the immediate
// state, for dead containers - their last memento.
func (m *Miner) Info(ctx context.Context, request *pb.Empty) (*pb.InfoReply, error) {
	log.G(m.ctx).Info("handling Info request", zap.Any("req", request))

	info, err := m.ovs.Info(ctx)
	if err != nil {
		return nil, err
	}

	var result = &pb.InfoReply{
		Usage:        make(map[string]*pb.ResourceUsage),
		Capabilities: m.hardware.IntoProto(),
	}

	for containerID, stat := range info {
		if id, ok := m.getTaskIdByContainerId(containerID); ok {
			result.Usage[id] = stat.Marshal()
		}
	}

	return result, nil
}

// Handshake is the first frame received from a Hub.
//
// This is a self representation about initial resources this Miner provides.
// TODO: May be useful to register a channel to cover runtime resource changes.
func (m *Miner) Handshake(ctx context.Context, request *pb.MinerHandshakeRequest) (*pb.MinerHandshakeReply, error) {
	log.G(m.ctx).Info("handling Handshake request", zap.Any("req", request))

	resp := &pb.MinerHandshakeReply{
		Capabilities: m.hardware.IntoProto(),
		NatType:      pb.NewNATType(m.natType),
	}

	return resp, nil
}

func (m *Miner) scheduleStatusPurge(id string) {
	t := time.NewTimer(time.Second * 3600)
	defer t.Stop()
	select {
	case <-t.C:
		m.mu.Lock()
		delete(m.containers, id)
		m.mu.Unlock()
	case <-m.ctx.Done():
		return
	}
}

func (m *Miner) setStatus(status *pb.TaskStatusReply, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}

	m.containers[id].status = status
	if status.Status == pb.TaskStatusReply_BROKEN || status.Status == pb.TaskStatusReply_FINISHED {
		go m.scheduleStatusPurge(id)
	}
	for _, ch := range m.statusChannels {
		select {
		case ch <- true:
		case <-m.ctx.Done():
		}
	}
}

func (m *Miner) listenForStatus(statusListener chan pb.TaskStatusReply_Status, id string) {
	select {
	case newStatus, ok := <-statusListener:
		if !ok {
			return
		}
		m.setStatus(&pb.TaskStatusReply{Status: newStatus}, id)
	case <-m.ctx.Done():
		return
	}
}

func transformRestartPolicy(p *pb.ContainerRestartPolicy) container.RestartPolicy {
	var restartPolicy = container.RestartPolicy{}
	if p != nil {
		restartPolicy.Name = p.Name
		restartPolicy.MaximumRetryCount = int(p.MaximumRetryCount)
	}

	return restartPolicy
}

func (m *Miner) Load(stream pb.Miner_LoadServer) error {
	log.G(m.ctx).Info("handling Load request")

	result, err := m.ovs.Load(stream.Context(), newChunkReader(stream))
	if err != nil {
		return err
	}

	log.G(m.ctx).Info("image loaded, set trailer", zap.String("trailer", result.Status))
	stream.SetTrailer(metadata.Pairs("status", result.Status))
	return nil
}

func (m *Miner) Save(request *pb.SaveRequest, stream pb.Miner_SaveServer) error {
	log.G(m.ctx).Info("handling Save request", zap.Any("request", request))

	info, rd, err := m.ovs.Save(stream.Context(), request.ImageID)
	if err != nil {
		return err
	}
	defer rd.Close()

	stream.SendHeader(metadata.Pairs("size", strconv.FormatInt(info.Size, 10)))

	streaming := true
	buf := make([]byte, 1*1024*1024)
	for streaming {
		n, err := rd.Read(buf)
		if err != nil {
			if err == io.EOF {
				streaming = false
			} else {
				return err
			}
		}
		if err := stream.Send(&pb.Chunk{Chunk: buf[:n]}); err != nil {
			return err
		}
	}

	return nil
}

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.MinerStartRequest) (*pb.MinerStartReply, error) {
	log.G(m.ctx).Info("handling Start request", zap.Any("request", request))

	if request.GetContainer() == nil {
		return nil, fmt.Errorf("container field is required")
	}

	resources, err := structs.NewTaskResources(request.GetResources())
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(request.Container.PublicKeyData)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid public key provided %v", err)
	}

	cgroup, resourceHandle, err := m.consume(request.GetOrderId(), resources)
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "failed to start %v", err)
	}
	// This can be canceled by using "resourceHandle.commit()".
	defer resourceHandle.release()

	mounts := make([]volume.Mount, 0)
	for _, spec := range request.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(request.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification - %s", err)
	}
	var d = Description{
		Image:         request.Container.Image,
		Registry:      request.Container.Registry,
		Auth:          request.Container.Auth,
		RestartPolicy: transformRestartPolicy(request.RestartPolicy),
		Resources:     resources.ToContainerResources(cgroup.Suffix()),
		DealId:        request.GetOrderId(),
		TaskId:        request.Id,
		CommitOnStop:  request.Container.CommitOnStop,
		Env:           request.Container.Env,
		GPURequired:   resources.RequiresGPU(),
		volumes:       request.Container.Volumes,
		mounts:        mounts,
		networks:      networks,
	}

	// TODO: Detect whether it's the first time allocation. If so - release resources on error.

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPOOLING}, request.Id)

	log.G(m.ctx).Info("spooling an image")
	if err := m.ovs.Spool(ctx, d); err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPAWNING}, request.Id)
	log.G(ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = d.Image

	var reply = pb.MinerStartReply{
		Container:  containerInfo.ID,
		PortMap:    make(map[string]*pb.Endpoints, 0),
		NetworkIDs: containerInfo.NetworkIDs,
	}

	for internalPort, portBindings := range containerInfo.Ports {
		if len(portBindings) < 1 {
			continue
		}

		var socketAddrs []*pb.SocketAddr
		var pubPortBindings []nat.PortBinding

		for _, portBinding := range portBindings {
			hostPort := portBinding.HostPort
			hostPortInt, err := nat.ParsePort(hostPort)
			if err != nil {
				return nil, err
			}

			for _, publicIP := range m.publicIPs {
				socketAddrs = append(socketAddrs, &pb.SocketAddr{
					Addr: publicIP,
					Port: uint32(hostPortInt),
				})

				pubPortBindings = append(pubPortBindings, nat.PortBinding{HostIP: publicIP, HostPort: hostPort})
			}
		}

		containerInfo.Ports[internalPort] = pubPortBindings

		reply.PortMap[string(internalPort)] = &pb.Endpoints{Endpoints: socketAddrs}
	}

	m.saveContainerInfo(request.Id, containerInfo)

	go m.listenForStatus(statusListener, request.Id)

	resourceHandle.commit()

	return &reply, nil
}

func (m *Miner) consume(orderId string, resources *structs.TaskResources) (cGroup, resourceHandle, error) {
	cgroup, err := m.cGroupManager.Attach(orderId, resources.ToCgroupResources())
	if err != nil && err != errCgroupAlreadyExists {
		return nil, nil, err
	}
	if err != errCgroupAlreadyExists {
		return cgroup, &nilResourceHandle{}, nil
	}

	usage := resources.ToUsage()
	if err := m.resources.Consume(&usage); err != nil {
		return nil, nil, err
	}

	handle := &ownedResourceHandle{
		miner:     m,
		usage:     usage,
		committed: false,
	}

	return cgroup, handle, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(ctx).Info("handling Stop request", zap.Any("req", request))

	m.mu.Lock()
	containerInfo, ok := m.containers[request.Id]
	m.mu.Unlock()

	m.deleteTaskMapping(request.Id)

	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.ovs.Stop(ctx, containerInfo.ID); err != nil {
		log.G(ctx).Error("failed to Stop container", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)

		return nil, status.Errorf(codes.Internal, "failed to stop container %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_FINISHED}, request.Id)
	m.resources.Release(&containerInfo.Resources)

	return &pb.Empty{}, nil
}

func (m *Miner) removeStatusChannel(idx int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.statusChannels, idx)
}

func (m *Miner) CollectTasksStatuses() map[string]*pb.TaskStatusReply {
	result := map[string]*pb.TaskStatusReply{}
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, info := range m.containers {
		result[id] = info.status
	}
	return result
}

func (m *Miner) sendTasksStatus(server pb.Miner_TasksStatusServer) error {
	result := &pb.StatusMapReply{Statuses: make(map[string]*pb.TaskStatusReply)}
	result.Statuses = m.CollectTasksStatuses()
	log.G(m.ctx).Info("sending result", zap.Any("info", m.containers), zap.Any("statuses", result.Statuses))
	return server.Send(result)
}

func (m *Miner) sendUpdatesOnNotify(server pb.Miner_TasksStatusServer, ch chan bool) {
	for {
		select {
		case <-ch:
			err := m.sendTasksStatus(server)
			if err != nil {
				return
			}
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Miner) sendUpdatesOnRequest(server pb.Miner_TasksStatusServer) {
	for {
		_, err := server.Recv()
		if err != nil {
			log.G(m.ctx).Info("tasks status server returned an error", zap.Error(err))
			return
		}
		log.G(m.ctx).Debug("handling tasks status request")
		err = m.sendTasksStatus(server)
		if err != nil {
			log.G(m.ctx).Info("failed to send status update", zap.Error(err))
			return
		}
	}
}

// TaskLogs returns logs from container
func (m *Miner) TaskLogs(request *pb.TaskLogsRequest, server pb.Miner_TaskLogsServer) error {
	log.G(m.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	cid, ok := m.getContainerIdByTaskId(request.Id)
	if !ok {
		return status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}
	opts := types.ContainerLogsOptions{
		ShowStdout: request.Type == pb.TaskLogsRequest_STDOUT || request.Type == pb.TaskLogsRequest_BOTH,
		ShowStderr: request.Type == pb.TaskLogsRequest_STDERR || request.Type == pb.TaskLogsRequest_BOTH,
		Since:      request.Since,
		Timestamps: request.AddTimestamps,
		Follow:     request.Follow,
		Tail:       request.Tail,
		Details:    request.Details,
	}
	reader, err := m.ovs.Logs(server.Context(), cid, opts)
	if err != nil {
		return err
	}
	defer reader.Close()
	buffer := make([]byte, 100*1024)
	for {
		readCnt, err := reader.Read(buffer)
		if readCnt != 0 {
			server.Send(&pb.TaskLogsChunk{Data: buffer[:readCnt]})
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// TasksStatus returns the status of a task
func (m *Miner) TasksStatus(server pb.Miner_TasksStatusServer) error {
	log.G(m.ctx).Info("starting tasks status server")
	m.mu.Lock()
	ch := make(chan bool)
	m.channelCounter++
	m.statusChannels[m.channelCounter] = ch
	defer m.removeStatusChannel(m.channelCounter)
	m.mu.Unlock()

	go m.sendUpdatesOnNotify(server, ch)
	m.sendUpdatesOnRequest(server)

	return nil
}

//TODO: proper request
func (m *Miner) JoinNetwork(ctx context.Context, req *pb.ID) (*pb.NetworkSpec, error) {
	spec, err := m.plugins.JoinNetwork(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.NetworkSpec{
		Type:    spec.NetworkType(),
		Options: spec.NetworkOptions(),
		Subnet:  spec.NetworkCIDR(),
		Addr:    spec.NetworkAddr(),
	}, nil
}

func (m *Miner) TaskDetails(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(m.ctx).Info("starting TaskDetails status server")

	info, ok := m.GetContainerInfo(req.GetId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no task with id %s", req.GetId())
	}

	var metric ContainerMetrics
	// If a container has been stoped, ovs.Info has no metrics for such container
	if info.status.Status == pb.TaskStatusReply_RUNNING {
		metrics, err := m.ovs.Info(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot get container metrics: %s", err.Error())
		}

		metric, ok = metrics[info.ID]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Cannot get metrics for container %s", req.GetId())
		}
	}

	portsStr, _ := json.Marshal(info.Ports)
	reply := &pb.TaskStatusReply{
		Status:    info.status.Status,
		ImageName: info.ImageName,
		Ports:     string(portsStr),
		Uptime:    uint64(time.Now().Sub(info.StartAt).Nanoseconds()),
		Usage:     metric.Marshal(),
		AvailableResources: &pb.AvailableResources{
			NumCPUs:      int64(info.Resources.NumCPUs),
			NumGPUs:      int64(info.Resources.NumGPUs),
			Memory:       uint64(info.Resources.Memory),
			Cgroup:       info.ID,
			CgroupParent: info.CgroupParent,
		},
	}

	return reply, nil
}

func (m *Miner) RunSSH() error {
	return m.ssh.Run()
}

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	log.G(m.ctx).Info("closing miner")

	m.ssh.Close()
	m.ovs.Close()
	m.controlGroup.Delete()
	m.plugins.Close()
}
