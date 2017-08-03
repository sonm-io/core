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

	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logger"
	pb "github.com/sonm-io/core/proto/miner"
	"github.com/sonm-io/core/util"

	frd "github.com/sonm-io/core/fusrodah/miner"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	hubAddress string
	// NOTE: do not use static detection
	pubAddress string

	rl *reverseListener

	ovs Overseer

	mu sync.Mutex
	// maps StartRequest's IDs to containers' IDs
	containers map[string]*ContainerInfo

	statusChannels map[int]chan bool
	channelCounter int
}

var _ pb.MinerServer = &Miner{}

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
func (m *Miner) Info(ctx context.Context, _ *pb.InfoRequest) (*pb.InfoReply, error) {
	info, err := m.ovs.Info(ctx)
	if err != nil {
		return nil, err
	}

	var result = pb.InfoReply{
		Stats: make(map[string]*pb.InfoReplyStats),
	}

	for id, stats := range info {
		result.Stats[id] = &pb.InfoReplyStats{
			CPU: &pb.InfoReplyStatsCpu{
				TotalUsage: stats.cpu.CPUUsage.TotalUsage,
			},
			Memory: &pb.InfoReplyStatsMemory{
				MaxUsage: stats.mem.MaxUsage,
			},
		}
	}

	return &result, nil
}

// Handshake reserves for the future usage
func (m *Miner) Handshake(context.Context, *pb.HandshakeRequest) (*pb.HandshakeReply, error) {
	return nil, status.Errorf(codes.Aborted, "not implemented")
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

func (m *Miner) setStatus(status *pb.TaskStatus, id string) {
	m.mu.Lock()
	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}
	m.containers[id].status = status
	if status.Status == pb.TaskStatus_BROKEN || status.Status == pb.TaskStatus_FINISHED {
		go m.scheduleStatusPurge(id)
	}
	for _, ch := range m.statusChannels {
		select {
		case ch <- true:
		case <-m.ctx.Done():
		}
	}
	m.mu.Unlock()
}

func (m *Miner) listenForStatus(statusListener chan pb.TaskStatus_Status, id string) {
	select {
	case newStatus := <-statusListener:
		m.setStatus(&pb.TaskStatus{newStatus}, id)
	case <-m.ctx.Done():
		return
	}
}

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.StartRequest) (*pb.StartReply, error) {
	var d = Description{
		Image:    request.Image,
		Registry: request.Registry,
		Auth:     request.Auth,
	}
	log.G(ctx).Info("handle Start request", zap.Any("req", request))

	m.setStatus(&pb.TaskStatus{pb.TaskStatus_SPOOLING}, request.Id)

	log.G(ctx).Info("spooling an image")
	err := m.ovs.Spool(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&pb.TaskStatus{pb.TaskStatus_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatus{pb.TaskStatus_SPAWNING}, request.Id)
	log.G(ctx).Info("spawning an image")
	statusListener, cinfo, err := m.ovs.Spawn(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatus{pb.TaskStatus_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	// TODO: clean it
	m.mu.Lock()
	m.containers[request.Id] = &cinfo
	m.mu.Unlock()
	go m.listenForStatus(statusListener, request.Id)

	var rpl = pb.StartReply{
		Container: cinfo.ID,
		Ports:     make(map[string]*pb.StartReplyPort),
	}
	for port, v := range cinfo.Ports {
		if len(v) > 0 {
			replyport := &pb.StartReplyPort{
				IP:   m.pubAddress,
				Port: v[0].HostPort,
			}
			rpl.Ports[string(port)] = replyport
		}
	}
	return &rpl, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.StopRequest) (*pb.StopReply, error) {
	log.G(ctx).Info("handle Stop request", zap.Any("req", request))
	m.mu.Lock()
	cinfo, ok := m.containers[request.Id]
	m.mu.Unlock()
	m.setStatus(&pb.TaskStatus{pb.TaskStatus_RUNNING}, request.Id)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.ovs.Stop(ctx, cinfo.ID); err != nil {
		log.G(ctx).Error("failed to Stop container", zap.Error(err))
		m.setStatus(&pb.TaskStatus{pb.TaskStatus_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to stop container %v", err)
	}
	m.setStatus(&pb.TaskStatus{pb.TaskStatus_FINISHED}, request.Id)
	return &pb.StopReply{}, nil
}

func (m *Miner) removeStatusChannel(idx int) {
	m.mu.Lock()
	delete(m.statusChannels, idx)
	m.mu.Unlock()
}

func (m *Miner) sendTasksStatus(server pb.Miner_TasksStatusServer) error {
	result := &pb.TasksStatusReply{Statuses: make(map[string]*pb.TaskStatus)}
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

		srv, err := frd.NewServer(nil)
		if err != nil {
			return
		}
		err = srv.Start()
		if err != nil {
			return
		}

		// if hub addr do not explicitly set via config we'll try to find it via discovery
		if m.hubAddress == "" {
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
}

// New returns new Miner
func New(ctx context.Context, config *MinerConfig) (*Miner, error) {
	loggr := logger.BuildLogger(config.Logger.Level, common.DevelopmentMode)
	ctx = log.WithLogger(ctx, loggr)

	addr, err := util.GetPublicIP()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	grpcServer := grpc.NewServer()
	ovs, err := NewOverseer(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	m := &Miner{
		ctx:        ctx,
		cancel:     cancel,
		grpcServer: grpcServer,
		ovs:        ovs,

		pubAddress: addr.String(),
		hubAddress: config.Miner.HubAddress,

		rl:             NewReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}
