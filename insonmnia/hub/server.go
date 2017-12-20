package hub

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"go.uber.org/zap"
	"golang.org/x/net/context"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/insonmnia/math"
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
	errDealNotFound     = status.Errorf(codes.NotFound, "deal not found")
	errTaskNotFound     = status.Errorf(codes.NotFound, "task not found")
	errImageForbidden   = status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")

	hubAPIPrefix = "/sonm.Hub/"

	// The following methods require TLS authentication and checking for client
	// and Hub's wallet equality.
	// The wallet is passed as peer metadata.
	hubManagementMethods = []string{
		"Status",
		"List",
		"Info",
		"TaskList",
		"Devices",
		"MinerDevices",
		"GetDeviceProperties",
		"SetDeviceProperties",
		"GetRegisteredWorkers",
		"RegisterWorker",
		"DeregisterWorker",
		"Slots",
		"InsertSlot",
		"RemoveSlot",
	}
)

type DealID string

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	// TODO (3Hren): Probably port pool should be associated with the gateway implicitly.
	cfg              *Config
	ctx              context.Context
	cancel           context.CancelFunc
	gateway          *gateway.Gateway
	portPool         *gateway.PortPool
	grpcEndpointAddr string
	externalGrpc     *grpc.Server
	minerListener    net.Listener

	ethKey  *ecdsa.PrivateKey
	ethAddr common.Address

	// locatorEndpoint string
	locatorPeriod time.Duration
	locatorClient pb.LocatorClient

	cluster       Cluster
	clusterEvents <-chan ClusterEvent

	miners   map[string]*MinerCtx
	minersMu sync.Mutex

	// TODO: rediscover jobs if Miner disconnected
	// TODO: store this data in some Storage interface

	waiter    errgroup.Group
	startTime time.Time
	version   string

	associatedHubs   map[string]struct{}
	associatedHubsMu sync.Mutex

	eth    ETH
	market pb.MarketClient

	// Device properties.
	// Must be synchronized with out Hub cluster.
	deviceProperties   map[string]DeviceProperties
	devicePropertiesMu sync.RWMutex

	// Scheduling.
	// Must be synchronized with out Hub cluster.
	slots   map[string]*structs.Slot
	slotsMu sync.RWMutex

	// Worker ACL.
	// Must be synchronized with out Hub cluster.
	acl   ACLStorage
	aclMu sync.RWMutex

	// Per-call ACL.
	// Must be synchronized with the Hub cluster.
	eventAuthorization *eventACL

	// Retroactive deals to tasks association. Tasks aren't popped when
	// completed to be able to save the history for the entire deal.
	// Note: this field is protected by tasksMu mutex.
	deals map[DealID]*DealMeta

	// Tasks
	tasks   map[string]*TaskInfo
	tasksMu sync.Mutex

	// TLS certificate rotator
	certRotator util.HitlessCertRotator
	// GRPC TransportCredentials supported our Auth
	creds credentials.TransportCredentials

	whitelist Whitelist
}

type DeviceProperties map[string]float64

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.Empty) (*pb.PingReply, error) {
	log.G(h.ctx).Info("handling Ping request")
	return &pb.PingReply{}, nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
	h.minersMu.Lock()
	minersCount := len(h.miners)
	h.minersMu.Unlock()

	uptime := uint64(time.Now().Sub(h.startTime).Seconds())

	reply := &pb.HubStatusReply{
		MinerCount: uint64(minersCount),
		Uptime:     uptime,
		Platform:   util.GetPlatformName(),
		Version:    h.version,
		EthAddr:    util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
	}

	return reply, nil
}

// List returns attached miners
func (h *Hub) List(ctx context.Context, request *pb.Empty) (*pb.ListReply, error) {
	log.G(h.ctx).Info("handling List request")

	reply := &pb.ListReply{
		Info: make(map[string]*pb.ListReply_ListValue),
	}
	for k := range h.miners {
		reply.Info[k] = new(pb.ListReply_ListValue)
	}
	for _, taskInfo := range h.tasks {
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

type routeMapping struct {
	containerPort string
	route         *route
}

func (h *Hub) onRequest(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.G(h.ctx).Debug("intercepting request")
	forwarded, r, err := h.tryForwardToLeader(ctx, request, info)
	if forwarded {
		return r, err
	}

	if err := h.eventAuthorization.authorize(ctx, method(info.FullMethod), request); err != nil {
		return nil, err
	}

	return handler(ctx, request)
}

func (h *Hub) tryForwardToLeader(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo) (bool, interface{}, error) {
	if h.cluster.IsLeader() {
		log.G(h.ctx).Info("isLeader is true")
		return false, nil, nil
	}
	log.G(h.ctx).Info("forwarding to leader", zap.String("method", info.FullMethod))
	cli, err := h.cluster.LeaderClient()
	if err != nil {
		log.G(h.ctx).Warn("failed to get leader client")
		return true, nil, err
	}
	if cli != nil {
		t := reflect.ValueOf(cli)
		parts := strings.Split(info.FullMethod, "/")
		methodName := parts[len(parts)-1]
		m := t.MethodByName(methodName)
		inValues := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request)}
		values := m.Call(inValues)
		var err error
		if !values[1].IsNil() {
			err = values[1].Interface().(error)
		}
		return true, values[0].Interface(), err
	} else {
		return true, nil, status.Errorf(codes.Internal, "is not leader and no connection to hub leader")
	}
}

func (h *Hub) PushTask(stream pb.Hub_PushTaskServer) error {
	log.G(h.ctx).Info("handling PushTask request")

	request, err := structs.NewImagePush(stream)
	if err != nil {
		return err
	}

	log.G(h.ctx).Info("pushing image", zap.Int64("size", request.ImageSize()))

	miner, _, err := h.findMinerByOrder(OrderId(request.DealId()))
	if err != nil {
		return err
	}

	// TODO: Check storage size.

	client, err := miner.Client.Load(stream.Context())
	if err != nil {
		return err
	}

	bytesCommitted := int64(0)
	clientCompleted := false
	// Intentionally block each time until miner responds to emulate congestion control.
	for {
		bytesRemaining := 0
		if !clientCompleted {
			chunk, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					clientCompleted = true

					log.G(h.ctx).Debug("client has closed its stream")
				} else {
					log.G(h.ctx).Error("failed to receive chunk from client", zap.Error(err))
					return err
				}
			}

			if chunk == nil {
				if err := client.CloseSend(); err != nil {
					log.G(h.ctx).Error("failed to close stream to miner", zap.Error(err))
					return err
				}
			} else {
				bytesRemaining = len(chunk.Chunk)
				if err := client.Send(chunk); err != nil {
					log.G(h.ctx).Error("failed to send chunk to miner", zap.Error(err))
					return err
				}
			}
		}

		for {
			progress, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					log.G(h.ctx).Debug("miner has closed its stream")
					if bytesCommitted == request.ImageSize() {
						stream.SetTrailer(client.Trailer())
						return nil
					} else {
						return status.Errorf(codes.Aborted, "miner closed its stream without committing all bytes")
					}
				} else {
					log.G(h.ctx).Error("failed to receive chunk from miner", zap.Error(err))
					return err
				}
			}

			bytesCommitted += progress.Size
			bytesRemaining -= int(progress.Size)
			log.G(h.ctx).Debug("progress", zap.Any("progress", progress), zap.Int64("bytesCommitted", bytesCommitted))

			if err := stream.Send(progress); err != nil {
				log.G(h.ctx).Error("failed to send chunk to client", zap.Error(err))
				return err
			}

			if bytesRemaining == 0 {
				break
			}
		}
	}
}

func (h *Hub) PullTask(request *pb.PullTaskRequest, stream pb.Hub_PullTaskServer) error {
	log.G(h.ctx).Info("handling PullTask request", zap.Any("request", request))

	ctx := log.WithLogger(h.ctx, log.G(h.ctx).With(zap.String("request", "pull task"), zap.String("id", uuid.New())))

	// TODO: Rename OrderId to DealId.
	miner, _, err := h.findMinerByOrder(OrderId(request.GetDealId()))
	if err != nil {
		return err
	}

	task, err := h.getTaskHistory(request.GetDealId(), request.GetTaskId())
	if err != nil {
		return err
	}

	imageID := fmt.Sprintf("%s:%s_%s", task.Image, request.GetDealId(), request.GetTaskId())

	log.G(ctx).Debug("pulling image", zap.String("imageID", imageID))

	client, err := miner.Client.Save(stream.Context(), &pb.SaveRequest{ImageID: imageID})
	header, err := client.Header()
	if err != nil {
		return err
	}
	stream.SetHeader(header)

	streaming := true
	for streaming {
		chunk, err := client.Recv()
		if chunk != nil {
			log.G(ctx).Debug("progress", zap.Int("chunkSize", len(chunk.Chunk)))

			if err := stream.Send(chunk); err != nil {
				return err
			}
		}
		if err != nil {
			if err == io.EOF {
				streaming = false
			} else {
				return err
			}
		}
	}

	return nil
}

func (h *Hub) getTaskHistory(dealID, taskID string) (*TaskInfo, error) {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	tasks, ok := h.deals[DealID(dealID)]
	if !ok {
		return nil, errDealNotFound
	}

	for _, task := range tasks.Tasks {
		if task.ID == taskID {
			return task, nil
		}
	}

	return nil, errTaskNotFound
}

func (h *Hub) StartTask(ctx context.Context, request *pb.HubStartTaskRequest) (*pb.HubStartTaskReply, error) {
	log.G(h.ctx).Info("handling StartTask request", zap.Any("request", request))

	taskRequest, err := structs.NewStartTaskRequest(request)
	if err != nil {
		return nil, err
	}

	return h.startTask(ctx, taskRequest)
}

func (h *Hub) generateTaskID() string {
	return uuid.New()
}

func (h *Hub) startTask(ctx context.Context, request *structs.StartTaskRequest) (*pb.HubStartTaskReply, error) {
	allowed, ref, err := h.whitelist.Allowed(h.ctx, request.Registry, request.Image, request.Auth)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, errImageForbidden
	}
	deal, err := h.eth.GetDeal(request.GetDeal().Id)
	if err != nil {
		return nil, err
	}

	dealID := DealID(deal.GetId())

	h.tasksMu.Lock()
	meta, ok := h.deals[dealID]
	h.tasksMu.Unlock()

	if !ok {
		// Hub knows nothing about this deal
		return nil, errDealNotFound
	}

	// Extract proper miner associated with the deal specified.
	miner, usage, err := h.findMinerByOrder(OrderId(meta.BidID))
	if err != nil {
		return nil, err
	}

	taskID := h.generateTaskID()
	startRequest := &pb.MinerStartRequest{
		OrderId:       request.GetDealId(),
		Id:            taskID,
		Registry:      reference.Domain(ref),
		Image:         reference.Path(ref),
		Auth:          request.GetAuth(),
		PublicKeyData: request.GetPublicKeyData(),
		CommitOnStop:  request.GetCommitOnStop(),
		Env:           request.GetEnv(),
		Resources: &pb.TaskResourceRequirements{
			CPUCores:   uint64(usage.NumCPUs),
			MaxMemory:  usage.Memory,
			GPUSupport: pb.GPUCount(math.Min(usage.NumGPUs, 2)),
		},
		RestartPolicy: &pb.ContainerRestartPolicy{
			Name:              "",
			MaximumRetryCount: 0,
		},
	}

	response, err := miner.Client.Start(ctx, startRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	info := TaskInfo{*request, *response, taskID, dealID, miner.uuid, nil}

	err = h.saveTask(DealID(request.GetDealId()), &info)
	if err != nil {
		miner.Client.Stop(ctx, &pb.ID{Id: taskID})
		return nil, err
	}

	routes := miner.registerRoutes(taskID, response.GetRoutes())

	// TODO: Synchronize routes with the cluster.
	reply := &pb.HubStartTaskReply{
		Id:      taskID,
		HubAddr: h.ethAddr.Hex(),
	}

	for _, route := range routes {
		reply.Endpoint = append(
			reply.Endpoint,
			fmt.Sprintf("%s->%s:%d", route.containerPort, route.route.Host, route.route.Port),
		)
	}

	return reply, nil
}

func (h *Hub) findMinerByOrder(id OrderId) (*MinerCtx, *resource.Resources, error) {
	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	for _, miner := range h.miners {
		for _, order := range miner.Orders() {
			if order == id {
				usage, err := miner.OrderUsage(id)
				if err != nil {
					return nil, nil, err
				}
				return miner, usage, nil
			}
		}
	}

	return nil, nil, ErrMinerNotFound
}

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling StopTask request", zap.Any("req", request))

	taskID := request.Id
	task, err := h.getTask(taskID)
	if err != nil {
		return nil, err
	}

	if err := h.stopTask(ctx, task); err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (h *Hub) stopTask(ctx context.Context, task *TaskInfo) error {
	miner, ok := h.getMinerByID(task.MinerId)
	if !ok {
		return status.Errorf(codes.NotFound, "no miner with id %s", task.MinerId)
	}

	_, err := miner.Client.Stop(ctx, &pb.ID{Id: task.ID})
	if err != nil {
		return status.Errorf(codes.NotFound, "failed to stop the task %s", task.ID)
	}

	miner.deregisterRoute(task.ID)

	h.deleteTask(task.ID)

	return nil
}

type dealInfo struct {
	ID             DealID
	Order          structs.Order
	TasksRunning   []TaskInfo
	TasksCompleted []TaskInfo
}

func (h *Hub) GetDealInfo(ctx context.Context, dealID *pb.ID) (*pb.DealInfoReply, error) {
	dealInfo, err := h.getDealInfo(DealID(dealID.Id))
	if err != nil {
		return nil, err
	}

	r := &pb.DealInfoReply{
		Id:             dealID,
		Order:          dealInfo.Order.Unwrap(),
		TasksRunning:   make([]*pb.ID, 0, len(dealInfo.TasksRunning)),
		TasksCompleted: make([]*pb.CompletedTask, 0, len(dealInfo.TasksCompleted)),
	}

	for _, taskInfo := range dealInfo.TasksRunning {
		r.TasksRunning = append(r.TasksRunning, &pb.ID{Id: taskInfo.ID})
	}

	for _, taskInfo := range dealInfo.TasksCompleted {
		r.TasksCompleted = append(r.TasksCompleted, &pb.CompletedTask{
			Id:    &pb.ID{Id: taskInfo.ID},
			Image: taskInfo.Image,
			EndTime: &pb.Timestamp{
				Seconds: taskInfo.EndTime.Unix(),
			},
		})
	}

	return r, nil
}

func (h *Hub) getDealInfo(dealID DealID) (*dealInfo, error) {
	meta, ok := h.deals[dealID]
	if !ok {
		return nil, errDealNotFound
	}

	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()
	dealInfo := &dealInfo{
		ID:             dealID,
		Order:          meta.Order,
		TasksRunning:   make([]TaskInfo, 0, len(h.tasks)),
		TasksCompleted: make([]TaskInfo, 0, len(meta.Tasks)),
	}

	for _, taskInfo := range h.tasks {
		dealInfo.TasksRunning = append(dealInfo.TasksRunning, *taskInfo)
	}

	for _, taskInfo := range meta.Tasks {
		dealInfo.TasksCompleted = append(dealInfo.TasksCompleted, *taskInfo)
	}

	return dealInfo, nil
}

//TODO: refactor - we can use h.tasks here
func (h *Hub) TaskList(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	// map workerID to []Task
	reply := &pb.TaskListReply{Info: map[string]*pb.TaskListReply_TaskInfo{}}

	for workerID, worker := range h.miners {
		worker.statusMu.Lock()
		taskStatuses := pb.StatusMapReply{Statuses: worker.statusMap}
		worker.statusMu.Unlock()

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

	mincli.statusMu.Lock()
	reply := pb.StatusMapReply{Statuses: mincli.statusMap}
	mincli.statusMu.Unlock()
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

	found, err := h.market.GetOrderByID(h.ctx, &pb.ID{Id: order.GetID()})
	if err != nil {
		return nil, err
	}

	if found == nil {
		return nil, ErrAskNotFound
	}

	resources, err := structs.NewResources(request.GetOrder().GetSlot().GetResources())
	if err != nil {
		return nil, err
	}
	usage := resource.NewResources(
		int(resources.GetCpuCores()),
		int64(resources.GetMemoryInBytes()),
		resources.GetGPUCount(),
	)

	miner, err := h.findRandomMinerByUsage(&usage)
	if err != nil {
		return nil, err
	}

	if err := miner.Consume(OrderId(request.GetBidId()), &usage); err != nil {
		return nil, err
	}

	h.waiter.Go(h.getDealWaiter(ctx, request, order))

	return &pb.Empty{}, nil
}

func (h *Hub) getDealWaiter(ctx context.Context, req *structs.DealRequest, order *structs.Order) func() error {
	return func() error {
		createdDeal, err := h.eth.WaitForDealCreated(req)
		if err != nil || createdDeal == nil {
			log.G(h.ctx).Warn(
				"cannot find created deal for current proposal",
				zap.String("bid_id", req.BidId),
				zap.String("ask_id", req.GetAskId()))
			return errors.New("cannot find created deal for current proposal")
		}

		err = h.eth.AcceptDeal(createdDeal.GetId())
		if err != nil {
			log.G(ctx).Warn("cannot accept deal",
				zap.String("deal_id", createdDeal.GetId()),
				zap.Error(err))
			return err
		}

		_, err = h.market.CancelOrder(h.ctx, &pb.Order{Id: req.GetAskId()})
		if err != nil {
			log.G(ctx).Warn("cannot cancel ask order from marketplace",
				zap.String("ask_id", req.GetAskId()),
				zap.Error(err))
		}

		dealID := DealID(createdDeal.GetId())

		h.tasksMu.Lock()
		defer h.tasksMu.Unlock()

		h.deals[dealID] = &DealMeta{Tasks: make([]*TaskInfo, 0), BidID: req.GetBidId()}
		h.eventAuthorization.insertDealCredentials(dealID, req.GetOrder().GetByuerID())

		go h.watchForDealClosed(dealID, req.GetOrder().GetByuerID())

		return nil
	}
}

func (h *Hub) watchForDealClosed(dealID DealID, buyerId string) {
	if err := h.eth.WaitForDealClosed(h.ctx, dealID, buyerId); err != nil {
		log.G(h.ctx).Error("failed to wait for closing deal",
			zap.String("dealID", string(dealID)),
			zap.Error(err),
		)
	}

	tasks, err := h.popDealHistory(dealID)
	if err != nil {
		return
	}

	log.S(h.ctx).Info("stopping at max %d tasks due to deal closing", len(tasks))
	for _, task := range tasks {
		if h.isTaskFinished(task.ID) {
			continue
		}

		if err := h.stopTask(h.ctx, task); err != nil {
			log.G(h.ctx).Error("failed to stop task",
				zap.String("dealID", string(dealID)),
				zap.String("taskID", task.ID),
				zap.Error(err),
			)
		}
	}

	h.eventAuthorization.removeDealCredentials(dealID)
}

func (h *Hub) isTaskFinished(id string) bool {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	_, ok := h.tasks[id]
	return !ok
}

func (h *Hub) findRandomMinerByUsage(usage *resource.Resources) (*MinerCtx, error) {
	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	rg := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := 0
	var result *MinerCtx = nil
	for _, miner := range h.miners {
		if err := miner.PollConsume(usage); err == nil {
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
	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	// Templates in go? Nevermind, just copy/paste.

	CPUs := map[string]*pb.CPUDeviceInfo{}
	for _, miner := range h.miners {
		h.collectMinerCPUs(miner, CPUs)
	}

	GPUs := map[string]*pb.GPUDeviceInfo{}
	for _, miner := range h.miners {
		h.collectMinerGPUs(miner, GPUs)
	}

	reply := &pb.DevicesReply{
		CPUs: CPUs,
		GPUs: GPUs,
	}

	return reply, nil
}

func (h *Hub) MinerDevices(ctx context.Context, request *pb.ID) (*pb.DevicesReply, error) {
	miner, ok := h.getMinerByID(request.Id)
	if !ok {
		return nil, ErrMinerNotFound
	}

	CPUs := map[string]*pb.CPUDeviceInfo{}
	h.collectMinerCPUs(miner, CPUs)

	GPUs := map[string]*pb.GPUDeviceInfo{}
	h.collectMinerGPUs(miner, GPUs)

	reply := &pb.DevicesReply{
		CPUs: CPUs,
		GPUs: GPUs,
	}

	return reply, nil
}

func (h *Hub) GetDeviceProperties(ctx context.Context, request *pb.ID) (*pb.GetDevicePropertiesReply, error) {
	log.G(h.ctx).Info("handling GetMinerProperties request", zap.Any("req", request))

	h.devicePropertiesMu.RLock()
	defer h.devicePropertiesMu.RUnlock()

	properties, exists := h.deviceProperties[request.Id]
	if !exists {
		return nil, ErrDeviceNotFound
	}

	return &pb.GetDevicePropertiesReply{Properties: properties}, nil
}

func (h *Hub) SetDeviceProperties(ctx context.Context, request *pb.SetDevicePropertiesRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling SetDeviceProperties request", zap.Any("req", request))

	h.devicePropertiesMu.Lock()
	defer h.devicePropertiesMu.Unlock()
	h.deviceProperties[request.ID] = DeviceProperties(request.Properties)
	err := h.cluster.Synchronize(h.deviceProperties)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

func (h *Hub) Slots(ctx context.Context, request *pb.Empty) (*pb.SlotsReply, error) {
	log.G(h.ctx).Info("handling Slots request")

	h.slotsMu.RLock()
	defer h.slotsMu.RUnlock()

	slots := make(map[string]*pb.Slot)
	for id, slot := range h.slots {
		slots[id] = slot.Unwrap()
	}

	return &pb.SlotsReply{Slots: slots}, nil
}

func (h *Hub) InsertSlot(ctx context.Context, request *pb.InsertSlotRequest) (*pb.ID, error) {
	log.G(h.ctx).Info("handling InsertSlot request", zap.Any("request", request))

	// We do not perform any resource existence check here, because miners
	// can be added dynamically.
	slot, err := structs.NewSlot(request.Slot)
	if err != nil {
		return nil, err
	}

	_, err = util.ParseBigInt(request.Price)
	if err != nil {
		return nil, err
	}

	// send slot to market
	ord := &pb.Order{
		OrderType:  pb.OrderType_ASK,
		Slot:       slot.Unwrap(),
		Price:      request.Price,
		SupplierID: util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
	}

	created, err := h.market.CreateOrder(h.ctx, ord)
	if err != nil {
		return nil, err
	}

	h.slotsMu.Lock()
	defer h.slotsMu.Unlock()

	h.slots[created.Id] = slot
	err = h.cluster.Synchronize(h.slots)
	if err != nil {
		return nil, err
	}

	return &pb.ID{Id: created.Id}, nil
}

func (h *Hub) RemoveSlot(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("RemoveSlot request", zap.Any("id", request.Id))

	h.slotsMu.Lock()
	defer h.slotsMu.Unlock()

	_, ok := h.slots[request.Id]
	if !ok {
		return nil, errSlotNotExists
	}

	_, err := h.market.CancelOrder(h.ctx, &pb.Order{Id: request.Id})
	if err != nil {
		return nil, err
	}

	delete(h.slots, request.Id)

	err = h.cluster.Synchronize(h.slots)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

// GetRegisteredWorkers returns a list of Worker IDs that are allowed to
// connect to the Hub.
func (h *Hub) GetRegisteredWorkers(ctx context.Context, empty *pb.Empty) (*pb.GetRegisteredWorkersReply, error) {
	log.G(h.ctx).Info("handling GetRegisteredWorkers request")

	var ids []*pb.ID

	h.acl.Each(func(cred string) bool {
		ids = append(ids, &pb.ID{Id: cred})
		return true
	})

	return &pb.GetRegisteredWorkersReply{Ids: ids}, nil
}

// RegisterWorker allows Worker with given ID to connect to the Hub
func (h *Hub) RegisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling RegisterWorker request", zap.String("id", request.GetId()))

	h.acl.Insert(request.Id)
	err := h.cluster.Synchronize(h.acl)
	if err != nil {
		return nil, err
	}

	return &pb.Empty{}, nil
}

// DeregisterWorkers deny Worker with given ID to connect to the Hub
func (h *Hub) DeregisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling DeregisterWorker request", zap.String("id", request.GetId()))

	if existed := h.acl.Remove(request.Id); !existed {
		log.G(h.ctx).Warn("attempt to deregister unregistered worker", zap.String("id", request.GetId()))
	} else {
		err := h.cluster.Synchronize(h.acl)
		if err != nil {
			return nil, err
		}
	}

	return &pb.Empty{}, nil
}

// New returns new Hub
func New(ctx context.Context, cfg *Config, version string, opts ...Option) (*Hub, error) {
	defaults := defaultHubOptions()
	for _, o := range opts {
		o(defaults)
	}

	if defaults.ethKey == nil {
		return nil, errors.New("cannot build Hub instance without private key")
	}

	if defaults.ctx == nil {
		defaults.ctx = context.Background()
	}

	var err error
	ctx, cancel := context.WithCancel(defaults.ctx)
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	ip := cfg.EndpointIP()
	clientPort, err := util.ParseEndpointPort(cfg.Cluster.Endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "error during parsing client endpoint")
	}
	grpcEndpointAddr := ip + ":" + clientPort

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

	if defaults.bcr == nil {
		defaults.bcr, err = blockchain.NewAPI(nil, nil)
		if err != nil {
			return nil, err
		}
	}

	ethWrapper, err := NewETH(ctx, defaults.ethKey, defaults.bcr, defaultDealWaitTimeout)
	if err != nil {
		return nil, err
	}

	if defaults.locator == nil {
		conn, err := util.MakeWalletAuthenticatedClient(ctx, defaults.creds, cfg.Locator.Endpoint)
		if err != nil {
			return nil, err
		}

		defaults.locator = pb.NewLocatorClient(conn)
	}

	if defaults.market == nil {
		conn, err := util.MakeWalletAuthenticatedClient(ctx, defaults.creds, cfg.Market.Endpoint)
		if err != nil {
			return nil, err
		}

		defaults.market = pb.NewMarketClient(conn)
	}

	if defaults.cluster == nil {
		defaults.cluster, defaults.clusterEvents, err = NewCluster(ctx, &cfg.Cluster, defaults.creds)
		if err != nil {
			return nil, err
		}
	}

	acl := NewACLStorage()
	if defaults.creds != nil {
		acl.Insert(defaults.ethAddr.Hex())
	}

	wl := NewWhitelist(ctx, &cfg.Whitelist)

	eventACL := newEventACL(ctx)

	h := &Hub{
		cfg:              cfg,
		ctx:              ctx,
		cancel:           cancel,
		gateway:          gate,
		portPool:         portPool,
		externalGrpc:     nil,
		grpcEndpointAddr: grpcEndpointAddr,

		ethKey:  defaults.ethKey,
		ethAddr: defaults.ethAddr,
		version: defaults.version,

		locatorPeriod: time.Second * time.Duration(cfg.Locator.Period),
		locatorClient: defaults.locator,

		eth:    ethWrapper,
		market: defaults.market,

		deals:            make(map[DealID]*DealMeta),
		tasks:            make(map[string]*TaskInfo),
		miners:           make(map[string]*MinerCtx),
		associatedHubs:   make(map[string]struct{}),
		deviceProperties: make(map[string]DeviceProperties),
		slots:            make(map[string]*structs.Slot),
		acl:              acl,

		eventAuthorization: eventACL,

		certRotator: defaults.rot,
		creds:       defaults.creds,

		cluster:       defaults.cluster,
		clusterEvents: defaults.clusterEvents,

		whitelist: wl,
	}

	dealAuthorization := map[string]DealMetaData{
		"TaskStatus": &taskFieldDealMetaData{hub: h},
		"StartTask":  &fieldDealMetaData{},
		"StopTask":   &taskFieldDealMetaData{hub: h},
		"TaskLogs":   &taskFieldDealMetaData{hub: h},
		"PushTask":   &contextDealMetaData{},
		"PullTask":   &contextDealMetaData{},
	}

	for event, metadata := range dealAuthorization {
		eventACL.addAuthorization(method(hubAPIPrefix+event), newDealAuthorization(ctx, metadata))
	}

	for _, event := range hubManagementMethods {
		eventACL.addAuthorization(method(hubAPIPrefix+event), newHubManagementAuthorization(ctx, h.ethAddr))
	}

	grpcServer := util.MakeGrpcServer(h.creds, grpc.UnaryInterceptor(h.onRequest))
	h.externalGrpc = grpcServer

	pb.RegisterHubServer(grpcServer, h)
	return h, nil
}

func (h *Hub) onNewHub(endpoint string) {
	h.associatedHubsMu.Lock()
	log.G(h.ctx).Info("new hub discovered", zap.String("endpoint", endpoint), zap.Any("known_hubs", h.associatedHubs))
	h.associatedHubs[endpoint] = struct{}{}

	h.associatedHubsMu.Unlock()

	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	for _, miner := range h.miners {
		miner.Client.DiscoverHub(h.ctx, &pb.DiscoverHubRequest{Endpoint: endpoint})
	}
}

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {
	h.startTime = time.Now()

	listener, err := net.Listen("tcp", h.cfg.Endpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", h.cfg.Endpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for connections from Miners", zap.Stringer("address", listener.Addr()))

	grpcL, err := net.Listen("tcp", h.cfg.Cluster.Endpoint)
	if err != nil {
		log.G(h.ctx).Error("failed to listen",
			zap.String("address", h.cfg.Cluster.Endpoint), zap.Error(err))
		listener.Close()
		return err
	}
	log.G(h.ctx).Info("listening for gRPC API connections", zap.Stringer("address", grpcL.Addr()))
	// TODO: fix this possible race: Close before Serve
	h.minerListener = listener

	h.waiter.Go(func() error {
		return h.externalGrpc.Serve(grpcL)
	})

	h.waiter.Go(func() error {
		for {
			conn, err := h.minerListener.Accept()
			if err != nil {
				return err
			}
			go h.handleInterconnect(h.ctx, conn)
		}
	})

	if err := h.cluster.RegisterAndLoadEntity("tasks", &h.tasks); err != nil {
		return err
	}
	if err := h.cluster.RegisterAndLoadEntity("device_properties", &h.deviceProperties); err != nil {
		return err
	}
	if err := h.cluster.RegisterAndLoadEntity("acl", h.acl); err != nil {
		return err
	}
	if err := h.cluster.RegisterAndLoadEntity("slots", &h.slots); err != nil {
		return err
	}
	log.G(h.ctx).Info("fetched entities",
		zap.Any("tasks", h.tasks),
		zap.Any("device_properties", h.deviceProperties),
		zap.Any("acl", h.acl),
		zap.Any("slots", h.slots))

	h.waiter.Go(h.runCluster)
	h.waiter.Go(h.listenClusterEvents)
	h.waiter.Go(h.startLocatorAnnouncer)

	h.waiter.Wait()

	return nil
}

func (h *Hub) runCluster() error {
	for {
		err := h.cluster.Run()
		log.G(h.ctx).Warn("cluster failure, retrying after 10 seconds", zap.Error(err))

		t := time.NewTimer(time.Second * 10)
		select {
		case <-h.ctx.Done():
			t.Stop()
			return nil
		case <-t.C:
			t.Stop()
		}
	}
}

func (h *Hub) listenClusterEvents() error {
	for {
		select {
		case event := <-h.clusterEvents:
			h.processClusterEvent(event)
		case <-h.ctx.Done():
			return nil
		}
	}
}

func (h *Hub) processClusterEvent(value interface{}) {
	log.G(h.ctx).Info("received cluster event", zap.Any("event", value))
	switch value := value.(type) {
	case NewMemberEvent:
		h.announceAddress()
	case LeadershipEvent:
		h.announceAddress()
	case map[string]*TaskInfo:
		log.G(h.ctx).Info("synchronizing tasks from cluster")
		h.tasksMu.Lock()
		defer h.tasksMu.Unlock()
		h.tasks = value
	case map[string]DeviceProperties:
		h.devicePropertiesMu.Lock()
		defer h.devicePropertiesMu.Unlock()
		h.deviceProperties = value
	case map[string]*structs.Slot:
		h.slotsMu.Lock()
		defer h.slotsMu.Unlock()
		h.slots = value
	case ACLStorage:
		h.acl = value
	default:
		log.G(h.ctx).Warn("received unknown cluster event",
			zap.Any("event", value),
			zap.String("type", reflect.TypeOf(value).String()))
	}
}

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.cancel()
	h.externalGrpc.Stop()
	h.minerListener.Close()
	if h.gateway != nil {
		h.gateway.Close()
	}
	if h.certRotator != nil {
		h.certRotator.Close()
	}
	h.waiter.Wait()
}

func (h *Hub) registerMiner(miner *MinerCtx) {
	h.minersMu.Lock()
	h.miners[miner.uuid] = miner
	h.minersMu.Unlock()
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

	h.minersMu.Lock()
	delete(h.miners, miner.ID())
	h.minersMu.Unlock()
}

func (h *Hub) getMinerByID(minerID string) (*MinerCtx, bool) {
	h.minersMu.Lock()
	defer h.minersMu.Unlock()
	m, ok := h.miners[minerID]
	return m, ok
}

func (h *Hub) saveTask(dealID DealID, info *TaskInfo) error {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()
	h.tasks[info.ID] = info

	taskIDs, ok := h.deals[dealID]
	if !ok {
		return errDealNotFound
	}

	taskIDs.Tasks = append(taskIDs.Tasks, info)
	h.deals[dealID] = taskIDs

	return h.cluster.Synchronize(h.tasks)
}

func (h *Hub) getTask(taskID string) (*TaskInfo, error) {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()
	info, ok := h.tasks[taskID]
	if !ok {
		return nil, errors.New("no such task")
	}
	return info, nil
}

func (h *Hub) deleteTask(taskID string) error {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()
	taskInfo, ok := h.tasks[taskID]
	if ok {
		delete(h.tasks, taskID)
		return h.cluster.Synchronize(h.tasks)
	}

	// Commit end time if such task exists in the history, if not - do nothing,
	// something terrible happened, but we just pretend nothing happened.
	taskHistory, ok := h.deals[taskInfo.DealId]
	if ok {
		for _, dealTaskInfo := range taskHistory.Tasks {
			if dealTaskInfo.ID == taskID {
				now := time.Now()
				dealTaskInfo.EndTime = &now
				break
			}
		}
	}
	return nil
}

func (h *Hub) popDealHistory(dealID DealID) ([]*TaskInfo, error) {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	tasks, ok := h.deals[dealID]
	if !ok {
		h.tasksMu.Unlock()
		return nil, errDealNotFound
	}
	delete(h.deals, dealID)

	return tasks.Tasks, nil
}

func (h *Hub) startLocatorAnnouncer() error {
	tk := time.NewTicker(h.locatorPeriod)
	defer tk.Stop()

	if err := h.announceAddress(); err != nil {
		log.G(h.ctx).Warn("cannot announce addresses to Locator", zap.Error(err))
	}

	for {
		select {
		case <-tk.C:
			if err := h.announceAddress(); err != nil {
				log.G(h.ctx).Warn("cannot announce addresses to Locator", zap.Error(err))
			}
		case <-h.ctx.Done():
			return nil
		}
	}
}

func (h *Hub) announceAddress() error {
	//TODO: is it really wrong to announce from several nodes simultaniously?
	if !h.cluster.IsLeader() {
		return nil
	}
	members, err := h.cluster.Members()
	if err != nil {
		return err
	}
	log.G(h.ctx).Info("got cluster members for locator announcement", zap.Any("members", members))

	endpoints := make([]string, 0)
	for _, member := range members {
		for _, ep := range member.endpoints {
			endpoints = append(endpoints, ep)
		}

	}
	req := &pb.AnnounceRequest{
		IpAddr: endpoints,
	}

	log.G(h.ctx).Info("announcing Hub address",
		zap.Stringer("eth", h.ethAddr),
		zap.Strings("addr", req.IpAddr))

	_, err = h.locatorClient.Announce(h.ctx, req)
	return err
}

func (h *Hub) collectMinerCPUs(miner *MinerCtx, dst map[string]*pb.CPUDeviceInfo) {
	for _, cpu := range miner.capabilities.CPU {
		hash := hex.EncodeToString(cpu.Hash())
		info, exists := dst[hash]
		if exists {
			info.Miners = append(info.Miners, miner.ID())
		} else {
			dst[hash] = &pb.CPUDeviceInfo{
				Miners: []string{miner.ID()},
				Device: cpu.Marshal(),
			}
		}
	}
}

func (h *Hub) collectMinerGPUs(miner *MinerCtx, dst map[string]*pb.GPUDeviceInfo) {
	for _, dev := range miner.capabilities.GPU {
		hash := hex.EncodeToString(dev.Hash())
		info, exists := dst[hash]
		if exists {
			info.Miners = append(info.Miners, miner.ID())
		} else {
			dst[hash] = &pb.GPUDeviceInfo{
				Miners: []string{miner.ID()},
				Device: gpu.Marshal(dev),
			}
		}
	}
}
