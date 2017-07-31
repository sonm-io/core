package hub

import (
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	pb "github.com/sonm-io/core/proto/hub"
	pbminer "github.com/sonm-io/core/proto/miner"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	miners map[string]*MinerCtx

	// TODO: rediscover jobs if Miner disconnected
	// TODO: store this data in some Storage interface
	tasksmu sync.Mutex
	tasks   map[string]string

	wg sync.WaitGroup
}

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.G(ctx).Info("reply to Ping")
	return &pb.PingReply{}, nil
}

// List returns attached miners
func (h *Hub) List(context.Context, *pb.ListRequest) (*pb.ListReply, error) {
	var info = make(map[string]*pb.ListReply_ListValue)
	h.mu.Lock()
	for k := range h.miners {
		info[k] = new(pb.ListReply_ListValue)
	}
	h.mu.Unlock()

	h.tasksmu.Lock()
	for k, v := range h.tasks {
		lr, ok := info[v]
		if ok {
			lr.Values = append(lr.Values, k)
			info[v] = lr
		}
	}
	h.tasksmu.Unlock()
	return &pb.ListReply{Info: info}, nil
}

// Info returns aggregated runtime statistics for all connected miners.
func (h *Hub) Info(ctx context.Context, req *pb.InfoRequest) (*pb.InfoReply, error) {
	h.mu.Lock()
	client, ok := h.miners[req.Miner]
	h.mu.Unlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner")
	}

	resp, err := client.Client.Info(ctx, &pbminer.InfoRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch info: %v", err)
	}

	// TODO: Reuse proto files with imports. Currently it's problematic, because of some reasons.
	var result = pb.InfoReply{
		Stats: make(map[string]*pb.InfoReplyStats),
	}

	for id, stats := range resp.Stats {
		result.Stats[id] = &pb.InfoReplyStats{
			CPU: &pb.InfoReplyStatsCpu{
				TotalUsage: stats.CPU.TotalUsage,
			},
			Memory: &pb.InfoReplyStatsMemory{
				MaxUsage: stats.Memory.MaxUsage,
			},
		}
	}

	return &result, nil
}

// StartTask schedules the Task on some miner
func (h *Hub) StartTask(ctx context.Context, request *pb.StartTaskRequest) (*pb.StartTaskReply, error) {
	log.G(ctx).Info("handling StartTask request", zap.Any("req", request))
	miner := request.Miner
	h.mu.Lock()
	mincli, ok := h.miners[miner]
	h.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner %s", miner)
	}

	uid := uuid.New()
	var startrequest = &pbminer.StartRequest{
		Id:       uid,
		Registry: request.Registry,
		Image:    request.Image,
	}

	resp, err := mincli.Client.Start(ctx, startrequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	h.tasksmu.Lock()
	h.tasks[uid] = miner
	h.tasksmu.Unlock()

	var reply = pb.StartTaskReply{
		Id: uid,
	}

	for k, v := range resp.Ports {
		reply.Endpoint = append(reply.Endpoint, fmt.Sprintf("%s->%s:%s", k, v.IP, v.Port))
	}

	return &reply, nil
}

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.StopTaskRequest) (*pb.StopTaskReply, error) {
	taskid := request.Id
	h.tasksmu.Lock()
	miner, ok := h.tasks[taskid]
	h.tasksmu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskid)
	}

	h.mu.Lock()
	mincli, ok := h.miners[miner]
	h.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner with task %s", miner)
	}

	var stoprequest = &pbminer.StopRequest{
		Id: taskid,
	}
	_, err := mincli.Client.Stop(ctx, stoprequest)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to stop the task %s", taskid)
	}

	h.tasksmu.Lock()
	delete(h.tasks, taskid)
	h.tasksmu.Unlock()

	return &pb.StopTaskReply{}, nil
}

func (h *Hub) MinerStatus(ctx context.Context, request *pb.MinerStatusRequest) (*pbminer.TasksStatusReply, error) {
	log.G(ctx).Info("handling MinerStatus request", zap.Any("req", request))
	miner := request.Miner
	h.mu.Lock()
	mincli, ok := h.miners[miner]
	h.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner %s", miner)
	}

	mincli.status_mu.Lock()
	reply := pbminer.TasksStatusReply{mincli.status_map}
	mincli.status_mu.Unlock()
	return &reply, nil

}

func (h *Hub) TaskStatus(ctx context.Context, request *pb.TaskStatusRequest) (*pb.TaskStatusReply, error) {
	taskid := request.Id
	h.tasksmu.Lock()
	miner, ok := h.tasks[taskid]
	h.tasksmu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskid)
	}

	h.mu.Lock()
	mincli, ok := h.miners[miner]
	h.mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner %s for task %s", miner, taskid)
	}

	mincli.status_mu.Lock()
	taskStatus, ok := mincli.status_map[taskid]
	mincli.status_mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no status report for task %s", taskid)
	}

	reply := pb.TaskStatusReply{taskStatus}
	return &reply, nil
}

// New returns new Hub
func New(ctx context.Context) (*Hub, error) {
	// TODO: add secure mechanism
	grpcServer := grpc.NewServer()
	h := &Hub{
		ctx:          ctx,
		externalGrpc: grpcServer,

		tasks:  make(map[string]string),
		miners: make(map[string]*MinerCtx),
	}
	pb.RegisterHubServer(grpcServer, h)

	return h, nil
}

// Serve starts handling incoming API gRPC request and communicates
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
	log.G(h.ctx).Info("listening for gRPC API connections", zap.Stringer("address", grpcL.Addr()))
	// TODO: fix this possible race: Close before Serve
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
			go h.handleInterconnect(h.ctx, conn)
		}
	}()
	h.wg.Wait()

	return nil
}

func (h *Hub) handleInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	miner, err := createMinerCtx(ctx, conn)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.miners[conn.RemoteAddr().String()] = miner
	h.mu.Unlock()

	miner.ping()
	miner.Close()

	h.mu.Lock()
	delete(h.miners, conn.RemoteAddr().String())
	h.mu.Unlock()
}

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.externalGrpc.Stop()
	h.minerListener.Close()
	h.wg.Wait()
}
