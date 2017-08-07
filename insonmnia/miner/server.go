package miner

import (
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/yamux"
	log "github.com/noxiouz/zapctx/ctxlog"

	pb "github.com/sonm-io/core/proto"

	"github.com/docker/docker/api/types/container"
	frd "github.com/sonm-io/core/fusrodah/miner"
	"github.com/sonm-io/core/insonmnia/resource"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	// Miner name for nice self-representation.
	name      string
	resources *resource.OS

	hubAddress string
	// NOTE: do not use static detection
	pubAddress string

	rl *reverseListener

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
	controlGroup   cGroupDeleter
}

func (m *Miner) saveContainerInfo(id string, info ContainerInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nameMapping[info.ID] = id
	m.containers[id] = &info
}

func (m *Miner) getTaskIdByContainerId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name, ok := m.nameMapping[id]
	return name, ok
}

func (m *Miner) deleteTaskMapping(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nameMapping, id)
}

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.G(m.ctx).Info("got ping request from Hub")
	return &pb.PingReply{}, nil
}

// Info returns runtime statistics collected from all containers working on this miner.
//
// This works the following way: a miner periodically collects various runtime statistics from all
// spawned containers that it knows about. For running containers metrics map the immediate
// state, for dead containers - their last memento.
func (m *Miner) Info(ctx context.Context, request *pb.MinerInfoRequest) (*pb.InfoReply, error) {
	log.G(m.ctx).Info("handling Info request", zap.Any("req", request))

	info, err := m.ovs.Info(ctx)
	if err != nil {
		return nil, err
	}

	var result = pb.InfoReply{
		Stats: make(map[string]*pb.InfoReplyStats),
	}

	for containerId, stats := range info {
		if id, ok := m.getTaskIdByContainerId(containerId); ok {
			// TODO: This transformation will be eliminated when protobuf unifying comes.
			result.Stats[id] = &pb.InfoReplyStats{
				CPU: &pb.InfoReplyStatsCpu{
					TotalUsage: stats.cpu.CPUUsage.TotalUsage,
				},
				Memory: &pb.InfoReplyStatsMemory{
					MaxUsage: stats.mem.MaxUsage,
				},
			}
		}
	}

	return &result, nil
}

// Handshake is the first frame received from a Hub.
//
// This is a self representation about initial resources this Miner provides.
// TODO: May be useful to register a channel to cover runtime resource changes.
func (m *Miner) Handshake(ctx context.Context, request *pb.MinerHandshakeRequest) (*pb.MinerHandshakeReply, error) {
	log.G(m.ctx).Info("handling Handshake request", zap.Any("req", request))

	resp := &pb.MinerHandshakeReply{
		Miner: m.name,
		Limits: &pb.Limits{
			Cores:  uint64(len(m.resources.CPU.List)),
			Memory: m.resources.Mem.Total,
		},
	}

	return resp, nil
}

func (m *Miner) scheduleStatusPurge(id string) {
	t := time.NewTimer(time.Second * 30)
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
	case newStatus := <-statusListener:
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

func transformResources(p *pb.ContainerResources) container.Resources {
	var resources = container.Resources{}
	if p != nil {
		resources.NanoCPUs = p.NanoCPUs
		resources.Memory = p.Memory
	}

	return resources
}

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.MinerStartRequest) (*pb.MinerStartReply, error) {
	var d = Description{
		Image:         request.Image,
		Registry:      request.Registry,
		Auth:          request.Auth,
		RestartPolicy: transformRestartPolicy(request.RestartPolicy),
		Resources:     transformResources(request.Resources),
	}
	log.G(m.ctx).Info("handling Start request", zap.Any("req", request))

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPOOLING}, request.Id)

	log.G(m.ctx).Info("spooling an image")
	err := m.ovs.Spool(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPAWNING}, request.Id)
	log.G(ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}

	m.saveContainerInfo(request.Id, containerInfo)

	go m.listenForStatus(statusListener, request.Id)

	var rpl = pb.MinerStartReply{
		Container: containerInfo.ID,
		Ports:     make(map[string]*pb.MinerStartReplyPort),
	}
	for port, v := range containerInfo.Ports {
		if len(v) > 0 {
			replyPort := &pb.MinerStartReplyPort{
				IP:   m.pubAddress,
				Port: v[0].HostPort,
			}
			rpl.Ports[string(port)] = replyPort
		}
	}

	return &rpl, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.StopTaskRequest) (*pb.StopTaskReply, error) {
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
	return &pb.StopTaskReply{}, nil
}

func (m *Miner) removeStatusChannel(idx int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.statusChannels, idx)
}

func (m *Miner) sendTasksStatus(server pb.Miner_TasksStatusServer) error {
	result := &pb.StatusMapReply{Statuses: make(map[string]*pb.TaskStatusReply)}
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, info := range m.containers {
		result.Statuses[id] = info.status
	}
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
			log.G(m.ctx).Info("tasks status server errored", zap.Error(err))
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

func (m *Miner) connectToHub(address string) {
	// Connect to the Hub
	var d = net.Dialer{
		DualStack: true,
	}
	conn, err := d.DialContext(m.ctx, "tcp", address)
	if err != nil {
		log.G(m.ctx).Error("failed to dial to the Hub", zap.String("addr", address), zap.Error(err))
		return
	}
	defer conn.Close()

	// HOLD reference
	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.G(m.ctx).Error("failed to create yamux.Server", zap.Error(err))
		return
	}
	defer session.Close()

	yaConn, err := session.Accept()
	if err != nil {
		log.G(m.ctx).Error("failed to Accept yamux.Stream", zap.Error(err))
		return
	}
	defer yaConn.Close()

	// Push the connection to a pool for grpcServer
	if err = m.rl.enqueue(yaConn); err != nil {
		log.G(m.ctx).Error("failed to enqueue yaConn for gRPC server", zap.Error(err))
		return
	}

	go func() {
		for {
			conn, err := session.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			rtt, err := session.Ping()
			if err != nil {
				log.G(m.ctx).Error("failed to Ping yamux.Session", zap.Error(err))
				return
			}
			log.G(m.ctx).Info("yamux.Ping OK", zap.Duration("rtt", rtt))
		case <-m.ctx.Done():
			return
		}
	}
}

// Serve starts discovery of Hubs,
// accepts incoming connections from a Hub
func (m *Miner) Serve() error {
	var grpcError error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcError = m.grpcServer.Serve(m.rl)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// if hub addr do not explicitly set via config we'll try to find it via discovery
		if m.hubAddress == "" {
			srv, err := frd.NewServer(nil)
			if err != nil {
				return
			}
			err = srv.Start()
			if err != nil {
				return
			}

			log.G(m.ctx).Debug("No hub IP, starting discovery")
			srv.Serve()
			m.hubAddress = srv.GetHubIp()
		} else {
			log.G(m.ctx).Debug("Using hub IP from config", zap.String("IP", m.hubAddress))
		}

		t := time.NewTicker(time.Second * 5)
		defer t.Stop()
		select {
		case <-m.ctx.Done():
			return
		case <-t.C:
			m.connectToHub(m.hubAddress)
		}
	}()
	wg.Wait()

	return grpcError
}

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	m.cancel()
	m.grpcServer.Stop()
	m.rl.Close()
	m.controlGroup.Delete()
}
