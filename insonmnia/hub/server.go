package hub

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	pb "github.com/sonm-io/core/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/crypto"
	frd "github.com/sonm-io/core/fusrodah/hub"

	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/util"
	"strconv"
	"strings"
)

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	ctx           context.Context
	gateway       *gateway.Gateway
	grpcEndpoint  string
	externalGrpc  *grpc.Server
	endpoint      string
	minerListener net.Listener
	ethKey        *ecdsa.PrivateKey

	mu     sync.Mutex
	miners map[string]*MinerCtx

	// TODO: rediscover jobs if Miner disconnected
	// TODO: store this data in some Storage interface
	tasksmu sync.Mutex
	tasks   map[string]string

	wg        sync.WaitGroup
	startTime time.Time
}

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingReply, error) {
	log.G(h.ctx).Info("handling Ping request")
	return &pb.PingReply{}, nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.HubStatusRequest) (*pb.HubStatusReply, error) {
	h.mu.Lock()
	minersCount := len(h.miners)
	h.mu.Unlock()

	uptime := time.Now().Unix() - h.startTime.Unix()

	reply := &pb.HubStatusReply{
		MinerCount: uint64(minersCount),
		Uptime:     uint64(uptime),
	}

	return reply, nil
}

// List returns attached miners
func (h *Hub) List(ctx context.Context, request *pb.ListRequest) (*pb.ListReply, error) {
	log.G(h.ctx).Info("handling List request")
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
func (h *Hub) Info(ctx context.Context, request *pb.HubInfoRequest) (*pb.InfoReply, error) {
	log.G(h.ctx).Info("handling Info request", zap.Any("req", request))
	client, ok := h.getMinerByID(request.Miner)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner")
	}

	resp, err := client.Client.Info(ctx, &pb.MinerInfoRequest{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch info: %v", err)
	}

	return resp, nil
}

// StartTask schedules the Task on some miner
func (h *Hub) StartTask(ctx context.Context, request *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	log.G(h.ctx).Info("handling StartTask request", zap.Any("req", request))
	minerID := request.Miner
	miner, ok := h.getMinerByID(minerID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner %s", minerID)
	}

	taskID := uuid.New()
	var startRequest = &pb.MinerStartRequest{
		Id:            taskID,
		Registry:      request.Registry,
		Image:         request.Image,
		Auth:          request.Auth,
		PublicKeyData: request.PublicKeyData,
		// TODO: Fill restart policy and resources fields.
	}

	resp, err := miner.Client.Start(ctx, startRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	for k, v := range resp.Ports {
		mapping := strings.Split(k, "/")
		protocol := mapping[1]
		realPort, err := strconv.ParseUint(v.Port, 10, 16)
		if err != nil {
			log.G(h.ctx).Warn("failed to convert real port to uint16", zap.String("port", v.Port))
			continue
		}
		miner.router.RegisterRoute(taskID, protocol, v.IP, uint16(realPort))
	}

	// TODO: Reply with remapped ip:port.

	h.setMinerTaskID(minerID, taskID)

	var reply = pb.HubStartTaskReply{
		Id: taskID,
	}

	for k, v := range resp.Ports {
		reply.Endpoint = append(reply.Endpoint, fmt.Sprintf("%s->%s:%s", k, v.IP, v.Port))
	}

	return &reply, nil
}

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.StopTaskRequest) (*pb.StopTaskReply, error) {
	log.G(h.ctx).Info("handling StopTask request", zap.Any("req", request))
	taskID := request.Id
	minerID, ok := h.getMinerByTaskID(taskID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskID)
	}

	miner, ok := h.getMinerByID(minerID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner with task %s", minerID)
	}

	_, err := miner.Client.Stop(ctx, &pb.StopTaskRequest{Id: taskID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to stop the task %s", taskID)
	}

	h.deleteTaskByID(taskID)

	return &pb.StopTaskReply{}, nil
}

func (h *Hub) MinerStatus(ctx context.Context, request *pb.HubStatusMapRequest) (*pb.StatusMapReply, error) {
	log.G(h.ctx).Info("handling MinerStatus request", zap.Any("req", request))

	miner := request.Miner
	mincli, ok := h.getMinerByID(miner)
	if !ok {
		log.G(ctx).Error("miner not found", zap.String("miner", miner))
		return nil, status.Errorf(codes.NotFound, "no such miner %s", miner)
	}

	mincli.status_mu.Lock()
	reply := pb.StatusMapReply{Statuses: mincli.status_map}
	mincli.status_mu.Unlock()
	return &reply, nil
}

func (h *Hub) TaskStatus(ctx context.Context, request *pb.TaskStatusRequest) (*pb.TaskStatusReply, error) {
	log.G(h.ctx).Info("handling TaskStatus request", zap.Any("req", request))
	taskID := request.Id
	minerID, ok := h.getMinerByTaskID(taskID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such task %s", taskID)
	}

	mincli, ok := h.getMinerByID(minerID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner %s for task %s", minerID, taskID)
	}

	req := &pb.TaskStatusRequest{Id: taskID}
	reply, err := mincli.Client.TaskDetails(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no status report for task %s", taskID)
	}

	return reply, nil
}

func (h *Hub) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
	minerID, ok := h.getMinerByTaskID(request.Id)
	if !ok {
		return status.Errorf(codes.NotFound, "no such task %s", request.Id)
	}

	mincli, ok := h.getMinerByID(minerID)
	if !ok {
		return status.Errorf(codes.NotFound, "no miner %s for task %s", minerID, request.Id)
	}

	client, err := mincli.Client.TaskLogs(server.Context(), request)
	if err != nil {
		return err
	}
	for {
		chunk, err := client.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		server.Send(chunk)
	}
}

// New returns new Hub
func New(ctx context.Context, cfg *HubConfig) (*Hub, error) {
	ethKey, err := crypto.HexToECDSA(cfg.Eth.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "malformed ethereum private key")
	}

	gate, err := gateway.NewGateway(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: add secure mechanism
	grpcServer := grpc.NewServer()
	h := &Hub{
		ctx:          ctx,
		gateway:      gate,
		externalGrpc: grpcServer,

		tasks:  make(map[string]string),
		miners: make(map[string]*MinerCtx),

		grpcEndpoint: cfg.Monitoring.Endpoint,
		endpoint:     cfg.Endpoint,
		ethKey:       ethKey,
	}
	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {
	h.startTime = time.Now()

	ip, err := util.GetPublicIP()
	if err != nil {
		return err
	}

	srv, err := frd.NewServer(h.ethKey, ip.String()+h.endpoint)
	if err != nil {
		return err
	}
	err = srv.Start()
	if err != nil {
		return err
	}
	srv.Serve()

	listener, err := net.Listen("tcp", h.endpoint)

	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", h.endpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for connections from Miners", zap.Stringer("address", listener.Addr()))

	grpcL, err := net.Listen("tcp", h.grpcEndpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen",
			zap.String("address", h.grpcEndpoint), zap.Error(err))
		listener.Close()
		return err
	}
	log.G(h.ctx).Info("listening for gRPC API connections", zap.Stringer("address", grpcL.Addr()))
	// TODO: fix this possible race: Close before Serve
	h.minerListener = listener

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
	h.gateway.Close()
	h.wg.Wait()
}

func (h *Hub) handleInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	miner, err := createMinerCtx(ctx, conn, h.gateway)
	if err != nil {
		return
	}

	h.mu.Lock()
	h.miners[conn.RemoteAddr().String()] = miner
	h.mu.Unlock()

	go func() {
		miner.pollStatuses()
		miner.Close()
	}()
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
