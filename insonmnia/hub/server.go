package hub

import (
	"fmt"
	"net"
	"sync"

	"go.uber.org/zap"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/common"
	"github.com/sonm-io/core/insonmnia/logger"
	pb "github.com/sonm-io/core/proto/hub"
	pbminer "github.com/sonm-io/core/proto/miner"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	frd "github.com/sonm-io/core/fusrodah/hub"

	"github.com/sonm-io/core/util"
)

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	ctx           context.Context
	grpcEndpoint  string
	externalGrpc  *grpc.Server
	minerEndpoint string
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
	log.G(ctx).Info("handling Ping request")
	return &pb.PingReply{}, nil
}

// List returns attached miners
func (h *Hub) List(ctx context.Context, request *pb.ListRequest) (*pb.ListReply, error) {
	log.G(ctx).Info("handling List request")
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
func (h *Hub) Info(ctx context.Context, request *pb.InfoRequest) (*pb.InfoReply, error) {
	log.G(ctx).Info("handling Info request", zap.Any("req", request))
	client, ok := h.getMinerByID(request.Miner)
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
	mincli, ok := h.getMinerByID(miner)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner %s", miner)
	}

	taskID := uuid.New()
	var startrequest = &pbminer.StartRequest{
		Id:       taskID,
		Registry: request.Registry,
		Image:    request.Image,
		Auth:     request.Auth,
	}

	resp, err := mincli.Client.Start(ctx, startrequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	h.setMinerTaskID(miner, taskID)

	var reply = pb.StartTaskReply{
		Id: taskID,
	}

	for k, v := range resp.Ports {
		reply.Endpoint = append(reply.Endpoint, fmt.Sprintf("%s->%s:%s", k, v.IP, v.Port))
	}

	return &reply, nil
}

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.StopTaskRequest) (*pb.StopTaskReply, error) {
	log.G(ctx).Info("handling StopTask request", zap.Any("req", request))
	taskID := request.Id
	minerID, ok := h.getMinerByTaskID(taskID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskID)
	}

	mincli, ok := h.getMinerByID(minerID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner with task %s", minerID)
	}

	stoprequest := &pbminer.StopRequest{Id: taskID}
	_, err := mincli.Client.Stop(ctx, stoprequest)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to stop the task %s", taskID)
	}

	h.deleteTaskByID(taskID)

	return &pb.StopTaskReply{}, nil
}

func (h *Hub) MinerStatus(ctx context.Context, request *pb.MinerStatusRequest) (*pbminer.TasksStatusReply, error) {
	log.G(ctx).Info("handling MinerStatus request", zap.Any("req", request))

	miner := request.Miner
	mincli, ok := h.getMinerByID(miner)
	if !ok {
		log.G(ctx).Error("miner not found", zap.String("miner", miner))
		return nil, status.Errorf(codes.NotFound, "no such miner %s", miner)
	}

	var reply pbminer.TasksStatusReply
	err := mincli.fetchStatuses()

	if err != nil {
		log.G(ctx).Error("cannot get status", zap.Error(err))
		return nil, status.Errorf(codes.DataLoss, "cannot get status")
	}

	mincli.status_mu.Lock()
	reply = pbminer.TasksStatusReply{Statuses: mincli.status_map}
	mincli.status_mu.Unlock()
	return &reply, nil
}

func (h *Hub) TaskStatus(ctx context.Context, request *pb.TaskStatusRequest) (*pb.TaskStatusReply, error) {
	log.G(ctx).Info("handling TaskStatus request", zap.Any("req", request))
	taskID := request.Id
	minerID, ok := h.getMinerByTaskID(taskID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskID)
	}

	mincli, ok := h.getMinerByID(minerID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner %s for task %s", minerID, taskID)
	}

	mincli.status_mu.Lock()
	taskStatus, ok := mincli.status_map[taskID]
	mincli.status_mu.Unlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no status report for task %s", taskID)
	}

	reply := pb.TaskStatusReply{Status: taskStatus}
	return &reply, nil
}

// New returns new Hub
func New(ctx context.Context, config *HubConfig) (*Hub, error) {
	// TODO: add secure mechanism

	loggr := logger.BuildLogger(config.Logger.Level, common.DevelopmentMode)
	ctx = log.WithLogger(ctx, loggr)

	grpcServer := grpc.NewServer()
	h := &Hub{
		ctx:          ctx,
		externalGrpc: grpcServer,

		tasks:  make(map[string]string),
		miners: make(map[string]*MinerCtx),

		grpcEndpoint:  config.Hub.GRPCEndpoint,
		minerEndpoint: config.Hub.MinerEndpoint,
	}
	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {

	ip, err := util.GetPublicIP()
	if err != nil {
		return err
	}
	srv, err := frd.NewServer(nil, ip.String()+h.minerEndpoint)
	if err != nil {
		return err
	}
	err = srv.Start()
	if err != nil {
		return err
	}
	srv.Serve()

	il, err := net.Listen("tcp", h.minerEndpoint)

	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", h.minerEndpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for connections from Miners", zap.Stringer("address", il.Addr()))

	grpcL, err := net.Listen("tcp", h.grpcEndpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen",
			zap.String("address", h.grpcEndpoint), zap.Error(err))
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

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.externalGrpc.Stop()
	h.minerListener.Close()
	h.wg.Wait()
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

func (h *Hub) getMinerByID(minerID string) (*MinerCtx, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m, ok := h.miners[minerID]
	return m, ok
}

func (h *Hub) getMinerByTaskID(taskID string) (string, bool) {
	h.tasksmu.Lock()
	defer h.tasksmu.Unlock()
	miner, ok := h.tasks[taskID]
	return miner, ok
}

func (h *Hub) setMinerTaskID(minerID, taskID string) {
	h.tasksmu.Lock()
	defer h.tasksmu.Unlock()
	h.tasks[taskID] = minerID
}

func (h *Hub) deleteTaskByID(taskID string) {
	h.tasksmu.Lock()
	defer h.tasksmu.Unlock()
	delete(h.tasks, taskID)
}
