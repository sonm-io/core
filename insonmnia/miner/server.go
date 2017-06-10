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
	pb "github.com/sonm-io/insonmnia/proto/miner"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx        context.Context
	cancel     context.CancelFunc
	grpcServer *grpc.Server

	hubaddress string

	rl *reverseListener

	ovs Overseer

	mu sync.Mutex
	// maps StartRequest's IDs to containers' IDs
	containers map[string]string
}

var _ pb.MinerServer = &Miner{}

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.GetLogger(ctx).Info("got ping request from Hub")
	return &pb.PingReply{}, nil
}

// Handshake reserves for the future usage
func (m *Miner) Handshake(context.Context, *pb.HandshakeRequest) (*pb.HandshakeReply, error) {
	return nil, status.Errorf(codes.Aborted, "not implemented")
}

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.StartRequest) (*pb.StartReply, error) {
	var d = Description{
		Image:    request.Image,
		Registry: request.Registry,
	}
	log.G(ctx).Info("handle Start request", zap.Any("req", request))

	log.G(ctx).Info("spooling an image")
	err := m.ovs.Spool(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	log.G(ctx).Info("spawning an image")
	containerid, err := m.ovs.Spawn(ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}

	// TODO: clean it
	m.mu.Lock()
	m.containers[request.Id] = containerid
	m.mu.Unlock()
	return &pb.StartReply{Container: containerid}, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.StopRequest) (*pb.StopReply, error) {
	log.G(ctx).Info("handle Stop request", zap.Any("req", request))
	m.mu.Lock()
	containerid, ok := m.containers[request.Id]
	m.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.ovs.Stop(ctx, containerid); err != nil {
		log.G(ctx).Error("failed to Stop container", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to stop container %v", err)
	}
	return &pb.StopReply{}, nil
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

	// Push the connection to a pool for grcpServer
	if err = m.rl.enqueue(yaConn); err != nil {
		log.G(m.ctx).Error("failed to enqueue yaConn for gRPC server", zap.Error(err))
		return
	}

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
	var grcpError error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		grcpError = m.grpcServer.Serve(m.rl)
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

	return grcpError
}

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	m.cancel()
	m.grpcServer.Stop()
	m.rl.Close()
}

// New returns new Miner
func New(ctx context.Context, hubaddress string) (*Miner, error) {
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

		rl: NewReverseListener(1),
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}
