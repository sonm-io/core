package hub

import (
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/insonmnia/proto/hub"
	pbminer "github.com/sonm-io/insonmnia/proto/miner"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// TODO: make it configurable
	externalGRPCEndpoint         = ":10001"
	minerHubInterconnectEndpoint = ":10002"
)

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	ctx           context.Context
	externalGrpc  *grpc.Server
	minerListener net.Listener

	mu     sync.Mutex
	miners map[string]pbminer.MinerClient

	wg sync.WaitGroup
}

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.G(ctx).Info("reply to Ping")
	return &pb.PingReply{}, nil
}

// List returns attached miners
func (h *Hub) List(context.Context, *pb.ListRequest) (*pb.ListReply, error) {
	var lr pb.ListReply
	h.mu.Lock()
	for k := range h.miners {
		lr.Name = append(lr.Name, k)
	}
	h.mu.Unlock()
	return &lr, nil
}

// New returns new Hub
func New(ctx context.Context) (*Hub, error) {
	// TODO: add secure mechanism
	grpcServer := grpc.NewServer()
	h := &Hub{
		ctx:          ctx,
		externalGrpc: grpcServer,

		miners: make(map[string]pbminer.MinerClient),
	}
	pb.RegisterHubServer(grpcServer, h)

	return h, nil
}

// Serve starts handling incoming API gRCP request and communicates
// with miners
func (h *Hub) Serve() error {
	il, err := net.Listen("tcp", minerHubInterconnectEndpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", minerHubInterconnectEndpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for connections from Miners", zap.Stringer("address", il.Addr()))

	grpcL, err := net.Listen("tcp", externalGRPCEndpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen",
			zap.String("address", externalGRPCEndpoint), zap.Error(err))
		il.Close()
		return err
	}
	log.G(h.ctx).Info("listening for gRPC API conenctions", zap.Stringer("address", grpcL.Addr()))
	// TODO: fix this possible race
	h.minerListener = il

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.externalGrpc.Serve(grpcL)
	}()

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		for {
			conn, err := h.minerListener.Accept()
			if err != nil {
				return
			}
			go h.handlerInterconnect(h.ctx, conn)
		}
	}()
	h.wg.Wait()

	return nil
}

func (h *Hub) handlerInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	// TODO: secure connection
	dctx, cancel := context.WithTimeout(ctx, time.Second*5)
	cc, err := grpc.DialContext(dctx, "miner", grpc.WithInsecure(), grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
		return conn, nil
	}))
	cancel()
	if err != nil {
		log.G(ctx).Error("failed to connect to Miner's grpc server", zap.Error(err))
		return
	}
	defer cc.Close()
	log.G(ctx).Info("grpc.Dial successfully finished")
	minerClient := pbminer.NewMinerClient(cc)

	h.mu.Lock()
	h.miners[conn.RemoteAddr().String()] = minerClient
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.miners, conn.RemoteAddr().String())
		h.mu.Unlock()
	}()

	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for range t.C {
		log.G(ctx).Info("ping the Miner", zap.Stringer("remote", conn.RemoteAddr()))
		// TODO: identify miner via Authorization mechanism
		// TODO: implement retries
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		_, err = minerClient.Ping(ctx, &pbminer.PingRequest{})
		cancel()
		if err != nil {
			log.G(ctx).Error("failed to ping miner", zap.Error(err))
			return
		}
	}
}

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.externalGrpc.Stop()
	h.minerListener.Close()
	h.wg.Wait()
}
