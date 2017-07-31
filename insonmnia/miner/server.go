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

	pb "github.com/sonm-io/core/proto/miner"
	"github.com/sonm-io/core/util"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	hubaddress string
	// NOTE: do not use static detection
	pubaddress string

	rl *reverseListener

	ovs Overseer

	mu sync.Mutex
	// maps StartRequest's IDs to containers' IDs
	containers map[string]*ContainerInfo

	statusChannels []chan bool
}

var _ pb.MinerServer = &Miner{}

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.GetLogger(ctx).Info("got ping request from Hub")
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

func (m *Miner) setStatus(status *pb.TaskStatus, id string) {
	m.mu.Lock()
	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}
	m.containers[id].status = status
	if status.Status == pb.TaskStatus_BROKEN || status.Status == pb.TaskStatus_FINISHED {
		go func() {
			//t := time.NewTicker(time.Second * 3600 * 24)
			t := time.NewTicker(time.Second * 30)
			defer t.Stop()
			for {
				select {
				case <-t.C:
					m.mu.Lock()
					delete(m.containers, id)
					m.mu.Unlock()
				case <-m.ctx.Done():
					return
				}
			}
		}()
	}
	for _, ch := range m.statusChannels {
		ch <- true
	}
	m.mu.Unlock()
}

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.StartRequest) (*pb.StartReply, error) {
	var d = Description{
		Image:    request.Image,
		Registry: request.Registry,
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
	go func() error {
		select {
		case newStatus := <-statusListener:
			m.setStatus(&pb.TaskStatus{newStatus}, request.Id)
		case <-m.ctx.Done():
			return m.ctx.Err()
		}
		return nil
	}()
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatus{pb.TaskStatus_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	// TODO: clean it
	m.mu.Lock()
	m.containers[request.Id] = &cinfo
	m.mu.Unlock()

	var rpl = pb.StartReply{
		Container: cinfo.ID,
		Ports:     make(map[string]*pb.StartReplyPort),
	}
	for port, v := range cinfo.Ports {
		if len(v) > 0 {
			replyport := &pb.StartReplyPort{
				IP:   m.pubaddress,
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

func (m *Miner) TasksStatus(server pb.Miner_TasksStatusServer) error {
	log.G(m.ctx).Info("starting tasks status server")
	m.mu.Lock()
	ch := make(chan bool)
	m.statusChannels = append(m.statusChannels, ch)
	idx := len(m.statusChannels)
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		if idx == len(m.statusChannels) {
			m.statusChannels = m.statusChannels[:idx]
		} else {
			m.statusChannels = append(m.statusChannels[:idx], m.statusChannels[idx+1:]...)
		}
		m.mu.Unlock()
	}()
	send := func() error {
		result := &pb.TasksStatusReply{Statuses: make(map[string]*pb.TaskStatus)}
		m.mu.Lock()
		for id, info := range m.containers {
			result.Statuses[id] = info.status
		}
		m.mu.Unlock()
		log.G(m.ctx).Info("sending result", zap.Any("info", m.containers), zap.Any("statuses", result.Statuses))
		return server.Send(result)
	}

	go func() error {
		for {
			select {
			case <-ch:
				send()
			case <-m.ctx.Done():
				return m.ctx.Err()
			}
		}
	}()
	for {
		_, err := server.Recv()
		if err != nil {
			log.G(m.ctx).Info("tasks status server errored", zap.Error(err))
			return err
		}
		log.G(m.ctx).Debug("handling tasks status request")
		err = send()
		if err != nil {
			log.G(m.ctx).Info("failed to send status update", zap.Error(err))
			return err
		}
	}
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
		// TODO: inject real discovery here
		var address = m.hubaddress
		for {
			m.connectToHub(address)
			select {
			case <-m.ctx.Done():
				return
			default:
				// TODO: backoff
				time.Sleep(5 * time.Second)
			}
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
func New(ctx context.Context, hubaddress string) (*Miner, error) {
	addr, err := util.GetPublicIP()
	if err != nil {
		return nil, err
	}

	pubaddress := addr.String()

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

		hubaddress: hubaddress,
		pubaddress: pubaddress,

		rl:             NewReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make([]chan bool, 0),
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}
