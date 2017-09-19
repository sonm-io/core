package hub

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/crypto"
	frd "github.com/sonm-io/core/fusrodah/hub"

	"encoding/hex"
	"encoding/json"
	consul "github.com/hashicorp/consul/api"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

var (
	ErrBidRequired      = status.Errorf(codes.InvalidArgument, "bid field is required")
	ErrInvalidOrderType = status.Errorf(codes.InvalidArgument, "invalid order type")
)

const tasksPrefix = "sonm/hub/tasks"
const leaderKey = "sonm/hub/leader"

var LeaderStepDown = errors.New("leader stepped down")

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	// TODO (3Hren): Probably port pool should be associated with the gateway implicitly.
	ctx           context.Context
	gateway       *gateway.Gateway
	portPool      *gateway.PortPool
	grpcEndpoint  string
	externalGrpc  *grpc.Server
	endpoint      string
	minerListener net.Listener
	ethKey        *ecdsa.PrivateKey

	mu     sync.Mutex
	miners map[string]*MinerCtx

	// TODO: rediscover jobs if Miner disconnected
	// TODO: store this data in some Storage interface
	//tasksmu sync.Mutex
	//tasks   map[string]string

	wg        sync.WaitGroup
	startTime time.Time
	version   string

	// Scheduling.
	filters []minerFilter
	consul  *consul.Client

	isLeader bool

	leaderClient     pb.HubClient
	leaderClientLock sync.Mutex

	stopCh chan struct{}
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
		Platform:   util.GetPlatformName(),
		Version:    h.version,
		EthAddr:    hex.EncodeToString(crypto.FromECDSA(h.ethKey)),
	}

	return reply, nil
}

func (h *Hub) loadTaskInfo() ([]*TaskInfo, error) {
	kv := h.consul.KV()
	tasks, _, err := kv.List(tasksPrefix, &consul.QueryOptions{})
	reply := make([]*TaskInfo, 0, len(tasks))
	if err != nil {
		return nil, err
	}

	for _, kvpair := range tasks {
		taskInfo := TaskInfo{}
		err = json.Unmarshal(kvpair.Value, &taskInfo)
		if err != nil {
			kv.Delete(kvpair.Key, &consul.WriteOptions{})
			log.G(h.ctx).Warn("inconsistent key found", zap.Error(err))
			continue
		} else {
			reply = append(reply, &taskInfo)
		}
	}
	return reply, nil
}

// List returns attached miners
func (h *Hub) List(ctx context.Context, request *pb.ListRequest) (*pb.ListReply, error) {
	log.G(h.ctx).Info("handling List request")

	tasks, err := h.loadTaskInfo()
	if err != nil {
		return nil, err
	}

	reply := &pb.ListReply{
		Info: make(map[string]*pb.ListReply_ListValue),
	}
	for _, taskInfo := range tasks {
		list, ok := reply.Info[taskInfo.MinerId]
		if !ok {
			reply.Info[taskInfo.MinerId] = &pb.ListReply_ListValue{
				Values: make([]string, 0),
			}
			list = reply.Info[taskInfo.MinerId]
		}
		list.Values = append(list.Values, taskInfo.ID)
	}

	return reply, nil
}

// Info returns aggregated runtime statistics for specified miners.
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

func decodePortBinding(v string) (string, string, error) {
	mapping := strings.Split(v, "/")
	if len(mapping) != 2 {
		return "", "", errors.New("failed to decode Docker port mapping")
	}

	return mapping[0], mapping[1], nil
}

type extRoute struct {
	containerPort string
	route         *route
}

type minerFilter func(miner *MinerCtx, requirements *pb.TaskRequirements) (bool, error)

// ExactMatchFilter checks for exact match.
//
// Returns true if there are no miners specified in the requirements. In that
// case we must apply hardware filtering for discovering miners that can start
// the task.
// Otherwise only specified miners become targets to start the task.
func exactMatchFilter(miner *MinerCtx, requirements *pb.TaskRequirements) (bool, error) {
	if len(requirements.GetMiners()) == 0 {
		return true, nil
	}

	for _, minerID := range requirements.GetMiners() {
		if minerID == miner.ID() {
			return true, nil
		}
	}

	return false, nil
}

func resourcesFilter(miner *MinerCtx, requirements *pb.TaskRequirements) (bool, error) {
	resources := requirements.GetResources()
	if resources == nil {
		return false, status.Errorf(codes.InvalidArgument, "resources section is required")
	}

	cpuCount := resources.GetCPUCores()
	memoryCount := resources.GetMaxMemory()

	var usage = resource.NewResources(int(cpuCount), int64(memoryCount))
	if err := miner.PollConsume(&usage); err != nil {
		return false, err
	}

	return true, nil
}

func (h *Hub) selectMiner(request *pb.HubStartTaskRequest) (*MinerCtx, error) {
	requirements := request.GetRequirements()
	if requirements == nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing requirements")
	}

	// Filter out miners that aren't met the requirements.
	h.mu.Lock()
	defer h.mu.Unlock()
	miners := []*MinerCtx{}
	for _, miner := range h.miners {
		ok, err := h.applyFilters(miner, requirements)
		if err != nil {
			return nil, err
		}

		if ok {
			miners = append(miners, miner)
		}
	}

	// Select random miner from the list.
	if len(miners) == 0 {
		return nil, status.Errorf(codes.NotFound, "failed to find miner to match specified requirements")
	}

	return miners[rand.Int()%len(miners)], nil
}

func (h *Hub) applyFilters(miner *MinerCtx, req *pb.TaskRequirements) (bool, error) {
	for _, filter := range h.filters {
		ok, err := filter(miner, req)
		if err != nil {
			return false, err
		}

		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// StartTask schedules the Task on some miner
func (h *Hub) StartTask(ctx context.Context, request *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	if !h.isLeader {
		h.leaderClientLock.Lock()
		cli := h.leaderClient
		h.leaderClientLock.Unlock()
		if cli != nil {
			return h.leaderClient.StartTask(ctx, request)
		} else {
			return nil, status.Errorf(codes.Internal, "is not leader and no connection to hub leader")
		}
	}

	log.G(h.ctx).Info("handling StartTask request", zap.Any("req", request))

	taskID := uuid.New()
	miner, err := h.selectMiner(request)
	if err != nil {
		return nil, err
	}

	var startRequest = &pb.MinerStartRequest{
		Id:            taskID,
		Registry:      request.Registry,
		Image:         request.Image,
		Auth:          request.Auth,
		PublicKeyData: request.PublicKeyData,
		CommitOnStop:  request.CommitOnStop,
		Env:           request.Env,
		Usage:         request.Requirements.GetResources(),
		RestartPolicy: &pb.ContainerRestartPolicy{
			Name:              "",
			MaximumRetryCount: 0,
		},
	}

	resp, err := miner.Client.Start(ctx, startRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	info := TaskInfo{*request, *resp, taskID, miner.uuid}
	json, err := json.Marshal(info)
	if err != nil {
		miner.Client.Stop(ctx, &pb.StopTaskRequest{Id: taskID})
		return nil, status.Errorf(codes.Internal, "could not marshal task info %v", err)
	}
	kv := h.consul.KV()
	kvPair := consul.KVPair{Key: tasksPrefix + "/" + taskID, Value: json}
	_, err = kv.Put(&kvPair, &consul.WriteOptions{})
	if err != nil {
		miner.Client.Stop(ctx, &pb.StopTaskRequest{Id: taskID})
		return nil, status.Errorf(codes.Internal, "could not store task info %v", err)
	}

	//TODO: save routes in consul
	routes := []extRoute{}
	for k, v := range resp.Ports {
		_, protocol, err := decodePortBinding(k)
		if err != nil {
			log.G(h.ctx).Warn("failed to decode miner's port mapping",
				zap.String("mapping", k),
				zap.Error(err),
			)
			continue
		}

		realPort, err := strconv.ParseUint(v.Port, 10, 16)
		if err != nil {
			log.G(h.ctx).Warn("failed to convert real port to uint16",
				zap.Error(err),
				zap.String("port", v.Port),
			)
			continue
		}

		route, err := miner.router.RegisterRoute(taskID, protocol, v.IP, uint16(realPort))
		if err != nil {
			log.G(h.ctx).Warn("failed to register route", zap.Error(err))
			continue
		}
		routes = append(routes, extRoute{
			containerPort: k,
			route:         route,
		})
	}

	//h.setMinerTaskID(miner.ID(), taskID)

	resources := request.GetRequirements().GetResources()
	cpuCount := resources.GetCPUCores()
	memoryCount := resources.GetMaxMemory()

	var usage = resource.NewResources(int(cpuCount), int64(memoryCount))
	if err := miner.Consume(taskID, &usage); err != nil {
		return nil, err
	}

	var reply = pb.HubStartTaskReply{
		Id: taskID,
	}

	for _, route := range routes {
		reply.Endpoint = append(
			reply.Endpoint,
			fmt.Sprintf("%s->%s:%d", route.containerPort, route.route.Host, route.route.Port),
		)
	}

	return &reply, nil
}

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.StopTaskRequest) (*pb.StopTaskReply, error) {
	log.G(h.ctx).Info("handling StopTask request", zap.Any("req", request))
	if !h.isLeader {
		h.leaderClientLock.Lock()
		cli := h.leaderClient
		h.leaderClientLock.Unlock()
		if cli != nil {
			return h.leaderClient.StopTask(ctx, request)
		} else {
			return nil, status.Errorf(codes.Internal, "is not leader and no connection to hub leader")
		}
	}

	taskID := request.Id
	task, err := h.getTask(taskID)
	if err != nil {
		return nil, err
	}

	miner, ok := h.getMinerByID(task.MinerId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner with id %s", task.MinerId)
	}

	_, err = miner.Client.Stop(ctx, &pb.StopTaskRequest{Id: taskID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to stop the task %s", taskID)
	}

	miner.deregisterRoute(taskID)
	miner.Retain(taskID)

	h.deleteTask(taskID)

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
	task, err := h.getTask(taskID)
	if err != nil {
		return nil, err
	}

	mincli, ok := h.getMinerByID(task.MinerId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner %s for task %s", task.MinerId, taskID)
	}

	req := &pb.TaskStatusRequest{Id: taskID}
	reply, err := mincli.Client.TaskDetails(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no status report for task %s", taskID)
	}

	// todo: fill this field into miner method, use Miner.name (uuid) instead of addr
	reply.MinerID = minerID
	return reply, nil
}

func (h *Hub) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
	task, err := h.getTask(request.Id)
	if err != nil {
		return err
	}

	mincli, ok := h.getMinerByID(task.MinerId)
	if !ok {
		return status.Errorf(codes.NotFound, "no miner %s for task %s", task.MinerId, request.Id)
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

func (h *Hub) ProposeDeal(ctx context.Context, request *pb.DealRequest) (*pb.DealReply, error) {
	log.G(h.ctx).Info("handling ProposeDeal request", zap.Any("req", request))

	order := request.GetOrder()
	if order == nil {
		return nil, ErrBidRequired
	}
	if order.OrderType != pb.OrderType_BID {
		return nil, ErrInvalidOrderType
	}
	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

// New returns new Hub
func New(ctx context.Context, cfg *HubConfig, version string) (*Hub, error) {
	ethKey, err := crypto.HexToECDSA(cfg.Eth.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "malformed ethereum private key")
	}

	var gate *gateway.Gateway
	var portPool *gateway.PortPool
	if cfg.GatewayConfig != nil {
		gate, err = gateway.NewGateway(ctx)
		if err != nil {
			return nil, err
		}

		if len(cfg.GatewayConfig.Ports) != 2 {
			return nil, errors.New("gateway ports must be a range of two values")
		}

		portRangeFrom := cfg.GatewayConfig.Ports[0]
		portRangeSize := cfg.GatewayConfig.Ports[1] - portRangeFrom
		portPool = gateway.NewPortPool(portRangeFrom, portRangeSize)
	}

	consul, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	// TODO: add secure mechanism
	grpcServer := grpc.NewServer(grpc.RPCCompressor(grpc.NewGZIPCompressor()), grpc.RPCDecompressor(grpc.NewGZIPDecompressor()))
	h := &Hub{
		ctx:          ctx,
		gateway:      gate,
		portPool:     portPool,
		externalGrpc: grpcServer,

		//tasks:  make(map[string]string),
		miners: make(map[string]*MinerCtx),

		grpcEndpoint: cfg.Monitoring.Endpoint,
		endpoint:     cfg.Endpoint,
		ethKey:       ethKey,
		version:      version,

		filters: []minerFilter{
			exactMatchFilter,
			resourcesFilter,
		},
		consul: consul,
	}
	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

func (h *Hub) leaderWatch() {
	var waitIdx uint64 = 0
	kv := h.consul.KV()
	var err error = nil
	for {
		h.leaderClientLock.Lock()
		h.leaderClient = nil
		h.leaderClientLock.Unlock()
		kv_pair, _, err := kv.Get(leaderKey, &consul.QueryOptions{WaitIndex: waitIdx})
		if err != nil {
			break
		}
		waitIdx = kv_pair.ModifyIndex
		conn, err := grpc.Dial(string(kv_pair.Value))
		if err != nil {
			break
		}
		h.leaderClientLock.Lock()
		h.leaderClient = pb.NewHubClient(conn)
		h.leaderClientLock.Unlock()
	}
	log.G(h.ctx).Error("leader watch failed", zap.Error(err))
	h.Close()
}

func (h *Hub) election() error {
	var err error
	for {
		lock, err := h.consul.LockOpts(&consul.LockOptions{
			Key:   leaderKey,
			Value: []byte(h.endpoint),
		})
		if err != nil {
			break
		}

		followerCh, err := lock.Lock(h.stopCh)
		if err != nil {
			break
		}
		h.isLeader = true
		for {
			_, ok := <-followerCh
			if !ok {
				break
			}
		}
		h.isLeader = false
	}
	h.Close()
	return err
}

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {
	go h.election()

	h.startTime = time.Now()

	ip, err := util.GetPublicIP()
	if err != nil {
		return err
	}

	workersPort, err := util.ParseEndpointPort(h.endpoint)
	if err != nil {
		return err
	}

	clientPort, err := util.ParseEndpointPort(h.grpcEndpoint)
	if err != nil {
		return err
	}

	workersEndpt := ip.String() + ":" + workersPort
	clientEndpt := ip.String() + ":" + clientPort

	srv, err := frd.NewServer(h.ethKey, workersEndpt, clientEndpt)
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

// Close disposes all capabilitiesCurrent attached to the Hub
func (h *Hub) Close() {
	h.stopCh <- struct{}{}
	h.externalGrpc.Stop()
	h.minerListener.Close()
	if h.gateway != nil {
		h.gateway.Close()
	}
	h.wg.Wait()
}

func (h *Hub) registerMiner(miner *MinerCtx) {
	h.mu.Lock()
	h.miners[miner.uuid] = miner
	h.mu.Unlock()
}

func (h *Hub) handleInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	miner, err := h.createMinerCtx(ctx, conn)
	if err != nil {
		return
	}

	h.registerMiner(miner)

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

func (h *Hub) getTask(taskID string) (*TaskInfo, error) {
	kv := h.consul.KV()
	taskData, _, err := kv.Get(tasksPrefix+"/"+taskID, &consul.QueryOptions{})
	if err != nil {
		return nil, err
	}

	taskInfo := TaskInfo{}
	err = json.Unmarshal(taskData.Value, &taskInfo)
	if err != nil {
		return nil, err
	}
	return &taskInfo, nil
}

func (h *Hub) deleteTask(taskID string) error {
	kv := h.consul.KV()
	_, err := kv.Delete(tasksPrefix+"/"+taskID, &consul.WriteOptions{})
	if err != nil {
		return err
	}

	return nil
}
