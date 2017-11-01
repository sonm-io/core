package hub

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"reflect"
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
	consul "github.com/hashicorp/consul/api"
	frd "github.com/sonm-io/core/fusrodah/hub"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
)

var (
	ErrInvalidOrderType = status.Errorf(codes.InvalidArgument, "invalid order type")
	ErrAskNotFound      = status.Errorf(codes.NotFound, "ask not found")
	ErrDeviceNotFound   = status.Errorf(codes.NotFound, "device not found")
	ErrMinerNotFound    = status.Errorf(codes.NotFound, "miner not found")
	ErrUnimplemented    = status.Errorf(codes.Unimplemented, "not implemented yet")
)

const tasksPrefix = "sonm/hub/tasks"
const leaderKey = "sonm/hub/leader"

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	// TODO (3Hren): Probably port pool should be associated with the gateway implicitly.
	ctx              context.Context
	cancel           context.CancelFunc
	gateway          *gateway.Gateway
	portPool         *gateway.PortPool
	grpcEndpoint     string
	grpcEndpointAddr string
	externalGrpc     *grpc.Server
	endpoint         string
	minerListener    net.Listener
	ethKey           *ecdsa.PrivateKey

	locatorEndpoint string
	locatorPeriod   time.Duration
	locatorClient   pb.LocatorClient

	localEndpoint string

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
	consul  Consul

	associatedHubs     map[string]struct{}
	associatedHubsLock sync.Mutex

	isLeader bool

	leaderClient     pb.HubClient
	leaderClientLock sync.Mutex

	eth    ETH
	market Market

	deviceProperties map[string]DeviceProperties

	// Scheduling.

	slots []*structs.Slot
}

type DeviceProperties map[string]float64

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.Empty) (*pb.PingReply, error) {
	log.G(h.ctx).Info("handling Ping request")
	return &pb.PingReply{}, nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
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
func (h *Hub) List(ctx context.Context, request *pb.Empty) (*pb.ListReply, error) {
	log.G(h.ctx).Info("handling List request")

	tasks, err := h.loadTaskInfo()
	if err != nil {
		return nil, err
	}

	reply := &pb.ListReply{
		Info: make(map[string]*pb.ListReply_ListValue),
	}
	for k := range h.miners {
		reply.Info[k] = new(pb.ListReply_ListValue)
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
func (h *Hub) Info(ctx context.Context, request *pb.ID) (*pb.InfoReply, error) {
	log.G(h.ctx).Info("handling Info request", zap.Any("req", request))
	client, ok := h.getMinerByID(request.GetId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no such miner")
	}

	resp, err := client.Client.Info(ctx, &pb.Empty{})
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

func (h *Hub) onRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.G(h.ctx).Debug("intercepting request")
	forwarded, r, err := h.tryForwardToLeader(ctx, req, info)
	if forwarded {
		return r, err
	}
	return handler(ctx, req)
}

func (h *Hub) tryForwardToLeader(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo) (bool, interface{}, error) {
	if h.isLeader {
		log.G(h.ctx).Info("isLeader is true")
		return false, nil, nil
	}
	log.G(h.ctx).Info("forwarding to leader", zap.String("method", info.FullMethod))
	h.leaderClientLock.Lock()
	cli := h.leaderClient
	h.leaderClientLock.Unlock()
	if cli != nil {
		t := reflect.ValueOf(h.leaderClient)
		parts := strings.Split(info.FullMethod, "/")
		methodName := parts[len(parts)-1]
		m := t.MethodByName(methodName)
		inValues := make([]reflect.Value, 0, 2)
		inValues = append(inValues, reflect.ValueOf(ctx), reflect.ValueOf(request))
		values := m.Call(inValues)
		return true, values[0].Interface(), values[1].Interface().(error)
	} else {
		return true, nil, status.Errorf(codes.Internal, "is not leader and no connection to hub leader")
	}
}

// StartTask schedules the Task on some miner
func (h *Hub) StartTask(ctx context.Context, request *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
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
	b, err := json.Marshal(info)
	if err != nil {
		miner.Client.Stop(ctx, &pb.ID{Id: taskID})
		return nil, status.Errorf(codes.Internal, "could not marshal task info %v", err)
	}

	kv := h.consul.KV()
	kvPair := consul.KVPair{Key: tasksPrefix + "/" + taskID, Value: b}
	_, err = kv.Put(&kvPair, &consul.WriteOptions{})
	if err != nil {
		miner.Client.Stop(ctx, &pb.ID{Id: taskID})
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

	// TODO: We no longer consume resources here. Instead allocation/usage checking required.
	//resources := request.GetRequirements().GetResources()
	//cpuCount := resources.GetCPUCores()
	//memoryCount := resources.GetMaxMemory()

	//var usage = resource.NewResources(int(cpuCount), int64(memoryCount))
	//if err := miner.Consume(taskID, &usage); err != nil {
	//	return nil, err
	//}

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
func (h *Hub) StopTask(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling StopTask request", zap.Any("req", request))

	taskID := request.Id
	task, err := h.getTask(taskID)
	if err != nil {
		return nil, err
	}

	miner, ok := h.getMinerByID(task.MinerId)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no miner with id %s", task.MinerId)
	}

	_, err = miner.Client.Stop(ctx, &pb.ID{Id: taskID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to stop the task %s", taskID)
	}

	miner.deregisterRoute(taskID)

	h.deleteTask(taskID)

	return &pb.Empty{}, nil
}

func (h *Hub) TaskList(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	h.mu.Lock()
	defer h.mu.Unlock()

	// map workerID to []Task
	reply := &pb.TaskListReply{Info: map[string]*pb.TaskListReply_TaskInfo{}}

	for workerID, worker := range h.miners {
		worker.status_mu.Lock()
		taskStatuses := pb.StatusMapReply{Statuses: worker.status_map}
		worker.status_mu.Unlock()

		// maps TaskID to TaskStatus
		info := &pb.TaskListReply_TaskInfo{Tasks: map[string]*pb.TaskStatusReply{}}

		for taskID := range taskStatuses.GetStatuses() {
			taskInfo, err := worker.Client.TaskDetails(ctx, &pb.ID{Id: taskID})
			if err != nil {
				return nil, err
			}

			info.Tasks[taskID] = taskInfo
		}

		reply.Info[workerID] = info

	}

	return reply, nil
}

func (h *Hub) MinerStatus(ctx context.Context, request *pb.ID) (*pb.StatusMapReply, error) {
	log.G(h.ctx).Info("handling MinerStatus request", zap.Any("req", request))

	miner := request.Id
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

func (h *Hub) TaskStatus(ctx context.Context, request *pb.ID) (*pb.TaskStatusReply, error) {
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

	req := &pb.ID{Id: taskID}
	reply, err := mincli.Client.TaskDetails(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "no status report for task %s", taskID)
	}

	reply.MinerID = mincli.ID()
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

func (h *Hub) ProposeDeal(ctx context.Context, r *pb.DealRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling ProposeDeal request", zap.Any("request", r))

	request, err := structs.NewDealRequest(r)
	if err != nil {
		return nil, err
	}

	order, err := structs.NewOrder(request.GetOrder())
	if err != nil {
		return nil, err
	}
	if !order.IsBid() {
		return nil, ErrInvalidOrderType
	}
	exists, err := h.market.OrderExists(order.GetID())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAskNotFound
	}
	resources, err := structs.NewResources(request.GetOrder().GetSlot().GetResources())
	if err != nil {
		return nil, err
	}
	usage := resource.NewResources(int(resources.GetCpuCores()), int64(resources.GetMemoryInBytes()))
	miner, err := h.findRandomMinerByUsage(&usage)
	if err != nil {
		return nil, err
	}
	if err := miner.Consume(OrderId(request.GetBidId()), &usage); err != nil {
		return nil, err
	}

	// TODO: Listen for ETH.
	// TODO: Start timeout for ETH approve deal.

	return &pb.Empty{}, nil
}

func (h *Hub) findRandomMinerByUsage(usage *resource.Resources) (*MinerCtx, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	rg := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := 0
	var result *MinerCtx = nil
	for _, miner := range h.miners {
		if err := miner.PollConsume(usage); err != nil {
			id++
			threshold := 1.0 / float64(id)
			if rg.Float64() < threshold {
				result = miner
			}
		}
	}

	if result == nil {
		return nil, ErrMinerNotFound
	}

	return result, nil
}

func (h *Hub) DiscoverHub(ctx context.Context, request *pb.DiscoverHubRequest) (*pb.Empty, error) {
	h.onNewHub(request.Endpoint)
	return &pb.Empty{}, nil
}

func (h *Hub) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Templates in go? Nevermind, just copy/paste.

	CPUs := map[string]*pb.CPUDeviceInfo{}
	for id, miner := range h.miners {
		for _, cpu := range miner.capabilities.CPU {
			hash := hex.EncodeToString(cpu.Hash())
			info, exists := CPUs[hash]
			if exists {
				info.Miners = append(info.Miners, id)
			} else {
				CPUs[hash] = &pb.CPUDeviceInfo{
					Miners: []string{id},
					Device: cpu.Marshal(),
				}
			}
		}
	}

	GPUs := map[string]*pb.GPUDeviceInfo{}
	for id, miner := range h.miners {
		for _, dev := range miner.capabilities.GPU {
			hash := hex.EncodeToString(dev.Hash())
			info, exists := GPUs[hash]
			if exists {
				info.Miners = append(info.Miners, id)
			} else {
				GPUs[hash] = &pb.GPUDeviceInfo{
					Miners: []string{id},
					Device: gpu.Marshal(dev),
				}
			}
		}
	}

	reply := &pb.DevicesReply{
		CPUs: CPUs,
		GPUs: GPUs,
	}

	return reply, nil
}

func (h *Hub) GetDeviceProperties(ctx context.Context, request *pb.ID) (*pb.GetDevicePropertiesReply, error) {
	log.G(h.ctx).Info("handling GetMinerProperties request", zap.Any("req", request))

	h.mu.Lock()
	defer h.mu.Unlock()

	properties, exists := h.deviceProperties[request.Id]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	return &pb.GetDevicePropertiesReply{Properties: properties}, nil
}

func (h *Hub) SetDeviceProperties(ctx context.Context, request *pb.SetDevicePropertiesRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling SetDeviceProperties request", zap.Any("req", request))

	h.mu.Lock()
	defer h.mu.Unlock()
	h.deviceProperties[request.ID] = DeviceProperties(request.Properties)

	return &pb.Empty{}, nil
}

func (h *Hub) Slots(ctx context.Context, request *pb.Empty) (*pb.SlotsReply, error) {
	log.G(h.ctx).Info("handling Slots request", zap.Any("request", request))

	h.mu.Lock()
	defer h.mu.Unlock()
	slots := make([]*pb.Slot, 0, len(h.slots))
	for _, slot := range h.slots {
		slots = append(slots, slot.Unwrap())
	}

	return &pb.SlotsReply{Slot: slots}, nil
}

func (h *Hub) InsertSlot(ctx context.Context, request *pb.Slot) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling InsertSlot request", zap.Any("request", request))

	// We do not perform any resource existence check here, because miners
	// can be added dynamically.
	slot, err := structs.NewSlot(request)
	if err != nil {
		return nil, err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// TODO: Check that such slot already exists.
	h.slots = append(h.slots, slot)

	return &pb.Empty{}, nil
}

func (h *Hub) RemoveSlot(ctx context.Context, request *pb.Slot) (*pb.Empty, error) {
	log.G(h.ctx).Info("RemoveSlot request", zap.Any("request", request))

	slot, err := structs.NewSlot(request)
	if err != nil {
		return nil, err
	}

	filtered := []*structs.Slot{}

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, s := range h.slots {
		if !s.Eq(slot) {
			filtered = append(filtered, s)
		}
	}

	return &pb.Empty{}, nil
}

// GetRegisteredWorkers returns a list of Worker IDs that  allowed to connect
// to the Hub.
func (h *Hub) GetRegisteredWorkers(ctx context.Context, empty *pb.Empty) (*pb.GetRegisteredWorkersReply, error) {
	log.G(h.ctx).Info("handling GetRegisteredWorkers request")

	// NOTE: it's a Stub implementation,  always return a list of the connected Workers
	// todo: implement me
	reply := &pb.GetRegisteredWorkersReply{
		Ids: []*pb.ID{},
	}

	h.mu.Lock()
	for minerID := range h.miners {
		reply.Ids = append(reply.Ids, &pb.ID{Id: minerID})
	}
	h.mu.Unlock()

	return reply, nil
}

// RegisterWorker allows Worker with given ID to connect to the Hub
func (h *Hub) RegisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	// todo: implement me
	log.G(h.ctx).Info("handling RegisterWorker request", zap.String("id", request.GetId()))
	return nil, ErrUnimplemented
}

// DeregisterWorkers deny Worker with given ID to connect to the Hub
func (h *Hub) DeregisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	// todo: implement me
	log.G(h.ctx).Info("handling DeregisterWorker request", zap.String("id", request.GetId()))
	return nil, ErrUnimplemented
}

// New returns new Hub
func New(ctx context.Context, cfg *HubConfig, version string) (*Hub, error) {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()
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

	var consulCli Consul
	if cfg.ConsulEnabled {
		consulCli, err = consul.NewClient(consul.DefaultConfig())
	} else {
		consulCli, err = newDevConsul(ctx)
	}
	if err != nil {
		return nil, err
	}

	eth, err := NewETH()
	if err != nil {
		return nil, err
	}

	market, err := NewMarket()
	if err != nil {
		return nil, err
	}

	h := &Hub{
		ctx:          ctx,
		cancel:       cancel,
		gateway:      gate,
		portPool:     portPool,
		externalGrpc: nil,

		miners: make(map[string]*MinerCtx),

		grpcEndpoint: cfg.Monitoring.Endpoint,
		endpoint:     cfg.Endpoint,
		ethKey:       ethKey,
		version:      version,

		locatorEndpoint: cfg.Locator.Address,
		locatorPeriod:   time.Second * time.Duration(cfg.Locator.Period),

		filters: []minerFilter{
			exactMatchFilter,
			resourcesFilter,
		},
		consul:         consulCli,
		associatedHubs: make(map[string]struct{}),

		eth:    eth,
		market: market,

		deviceProperties: make(map[string]DeviceProperties),
	}

	interceptor := h.onRequest
	grpcServer := grpc.NewServer(grpc.RPCCompressor(grpc.NewGZIPCompressor()), grpc.RPCDecompressor(grpc.NewGZIPDecompressor()), grpc.UnaryInterceptor(interceptor))
	h.externalGrpc = grpcServer

	h.localEndpoint, err = h.determineLocalEndpoint()
	if err != nil {
		return nil, err
	}
	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

func (h *Hub) leaderWatch() {
	log.G(h.ctx).Info("starting leader watch goroutine")
	var waitIdx uint64 = 0
	kv := h.consul.KV()
	var err error = nil
	for {
		kv_pair, _, err := kv.Get(leaderKey, &consul.QueryOptions{WaitIndex: waitIdx})
		if err != nil {
			break
		}
		if kv_pair == nil {
			time.Sleep(time.Second * 1)
			log.G(h.ctx).Info("leader key is empty. sleeping for 1 sec")
			continue
		}
		log.G(h.ctx).Info("leader watch: fetched leader", zap.Any("kvpair", kv_pair))
		log.G(h.ctx).Info("leader watch: fetched leader", zap.String("leader", string(kv_pair.Value)))
		h.leaderClientLock.Lock()
		h.leaderClient = nil
		h.leaderClientLock.Unlock()

		ep := string(kv_pair.Value)
		h.onNewHub(ep)
		conn, err := grpc.Dial(ep, grpc.WithInsecure(),
			grpc.WithCompressor(grpc.NewGZIPCompressor()),
			grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
		if err != nil {
			log.G(h.ctx).Warn("could not connect to hub", zap.String("endpoint", ep), zap.Error(err))
			time.Sleep(time.Duration(100 * 1000000))
			continue
		}
		h.leaderClientLock.Lock()
		h.leaderClient = pb.NewHubClient(conn)
		cli := h.leaderClient
		h.leaderClientLock.Unlock()
		cli.DiscoverHub(h.ctx, &pb.DiscoverHubRequest{Endpoint: h.localEndpoint})

		waitIdx = kv_pair.ModifyIndex
	}
	log.G(h.ctx).Error("leader watch failed", zap.Error(err))
	h.Close()
}

func (h *Hub) onNewHub(endpoint string) {
	h.associatedHubsLock.Lock()
	log.G(h.ctx).Info("new hub discovered", zap.String("endpoint", endpoint), zap.Any("known_hubs", h.associatedHubs))
	h.associatedHubs[endpoint] = struct{}{}

	h.associatedHubsLock.Unlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, miner := range h.miners {
		miner.Client.DiscoverHub(h.ctx, &pb.DiscoverHubRequest{Endpoint: endpoint})
	}
}

func (h *Hub) determineLocalEndpoint() (string, error) {
	if h.endpoint[0] == ':' {
		ifaces, err := net.Interfaces()
		if err != nil {
			return "", err
		}
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				return "", err
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && ip.IsGlobalUnicast() {
					ep := ip.String() + h.grpcEndpoint
					return ep, nil
				}
			}
		}
	} else {
		return h.endpoint, nil
	}
	return "", errors.New("unicast ip not found")
}

func (h *Hub) election() error {
	log.G(h.ctx).Info("starting leader election goroutine")
	go h.leaderWatch()
	var err error

	for {
		lock, err := h.consul.LockOpts(&consul.LockOptions{
			Key:   leaderKey,
			Value: []byte(h.localEndpoint),
		})
		if err != nil {
			log.G(h.ctx).Warn("could not create lock opts", zap.Error(err))
			break
		}

		log.G(h.ctx).Info("trying to aquire leader lock")
		followerCh, err := lock.Lock(nil)
		if err != nil {
			log.G(h.ctx).Info("could not acquire leader lock", zap.Error(err))
			break
		}
		log.G(h.ctx).Info("leader lock acquired")
		h.isLeader = true
		for {
			_, ok := <-followerCh
			if !ok {
				log.G(h.ctx).Info("leader lock released")
				break
			}
		}
		h.isLeader = false
	}
	log.G(h.ctx).Warn("election failed - closing hub", zap.Error(err))
	h.Close()
	return err
}

// TODO: Decomposed here to be able to easily comment when UDP capturing occurs :)
func (h *Hub) startDiscovery() error {
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

	workersEndpoint := ip.String() + ":" + workersPort
	clientEndpoint := ip.String() + ":" + clientPort
	h.grpcEndpointAddr = clientEndpoint

	srv, err := frd.NewServer(h.ethKey, workersEndpoint, clientEndpoint)
	if err != nil {
		return err
	}
	err = srv.Start()
	if err != nil {
		return err
	}
	srv.Serve()

	return nil
}

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {
	go h.election()

	h.startTime = time.Now()

	if err := h.startDiscovery(); err != nil {
		return err
	}

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

	// init locator connection and announce
	// address only on Leader
	if h.isLeader {
		err = h.initLocatorClient()
		if err != nil {
			return err
		}

		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			h.startLocatorAnnouncer()
		}()
	}

	h.wg.Wait()
	return nil
}

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.cancel()
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
	for address := range h.associatedHubs {
		log.G(h.ctx).Info("sending hub adderess", zap.String("hub_address", address))
		miner.Client.DiscoverHub(h.ctx, &pb.DiscoverHubRequest{Endpoint: address})
	}
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
	delete(h.miners, miner.ID())
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

func (h *Hub) initLocatorClient() error {
	conn, err := grpc.Dial(
		h.locatorEndpoint,
		grpc.WithInsecure(),
		grpc.WithTimeout(5*time.Second),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithCompressor(grpc.NewGZIPCompressor()))
	if err != nil {
		return err
	}

	h.locatorClient = pb.NewLocatorClient(conn)
	return nil
}

func (h *Hub) startLocatorAnnouncer() {
	tk := time.NewTicker(h.locatorPeriod)
	defer tk.Stop()

	h.announceAddress(h.ctx)

	for {
		select {
		case <-tk.C:
			h.announceAddress(h.ctx)
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) announceAddress(ctx context.Context) {
	req := &pb.AnnounceRequest{
		EthAddr: util.PubKeyToAddr(h.ethKey.PublicKey),
		IpAddr:  []string{h.grpcEndpointAddr},
	}

	log.G(ctx).Info("announcing Hub address",
		zap.String("eth", req.EthAddr),
		zap.String("addr", req.IpAddr[0]))

	_, err := h.locatorClient.Announce(ctx, req)
	if err != nil {
		log.G(ctx).Warn("cannot announce addresses to Locator", zap.Error(err))
	}
}
