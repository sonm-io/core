package miner

import (
	"net"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sonm-io/insonmnia/proto/miner"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	grpcServer *grpc.Server

	rl *reverseListener

	ovs Overseer

	mu sync.Mutex
	// maps StartRequest's IDs to containers' IDs
	containers map[string]string
}

var _ pb.MinerServer = &Miner{}

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(context.Context, *pb.PingRequest) (*pb.PingReply, error) {
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
	containerid, err := m.ovs.Spawn(ctx, d)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	m.mu.Lock()
	m.containers[request.Id] = containerid
	m.mu.Unlock()
	return &pb.StartReply{Container: containerid}, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.StopRequest) (*pb.StopReply, error) {
	m.mu.Lock()
	containerid, ok := m.containers[request.Id]
	m.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.ovs.Stop(ctx, containerid); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to stop container %v", err)
	}
	return &pb.StopReply{}, nil
}

func (m *Miner) connectToHub(address string) error {
	// Connect to the Hub
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	// Push the connection to a pool for grcpServer
	if err = m.rl.enqueue(conn); err != nil {
		return err
	}
	return nil
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
	// TODO: spawn reconnector to the HUB in discovery
	wg.Wait()

	return grcpError
}

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	m.grpcServer.Stop()
	m.rl.Close()
}

// New returns new Miner
func New() (*Miner, error) {
	grpcServer := grpc.NewServer()
	m := &Miner{
		grpcServer: grpcServer,

		rl: NewReverseListener(1),
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}
