package hub

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/gateway"
	"github.com/sonm-io/core/insonmnia/hardware/gpu"
	"github.com/sonm-io/core/insonmnia/math"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var (
	ErrMinerNotFound  = status.Errorf(codes.NotFound, "miner not found")
	errDealNotFound   = status.Errorf(codes.NotFound, "deal not found")
	errTaskNotFound   = status.Errorf(codes.NotFound, "task not found")
	errImageForbidden = status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")

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

	orderPublishThresholdETH = new(big.Int).Mul(big.NewInt(10), big.NewInt(params.Finney))
)

type DealID string

func (id DealID) String() string {
	return string(id)
}

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
	askPlans *AskPlans

	// Worker ACL.
	// Must be synchronized with out Hub cluster.
	acl ACLStorage

	// Per-call ACL.
	// Must be synchronized with the Hub cluster.
	eventAuthorization *auth.AuthRouter

	// Currently running deals.
	// Retroactive deals to tasks association. Tasks aren't popped when
	// completed to be able to save the history for the entire deal.
	// Note: this field is protected by tasksMu mutex.
	deals map[DealID]*DealMeta

	// Reserved orders.
	orderShelter *OrderShelter

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
	reply := &pb.ListReply{
		Info: make(map[string]*pb.ListReply_ListValue),
	}

	h.minersMu.Lock()
	for k := range h.miners {
		reply.Info[k] = new(pb.ListReply_ListValue)
	}
	h.minersMu.Unlock()

	for minerID := range reply.Info {
		infoReply, err := h.Info(ctx, &pb.ID{Id: minerID})
		if err == nil {
			list := reply.Info[minerID]
			for taskID := range infoReply.Usage {
				list.Values = append(list.Values, taskID)
			}
			reply.Info[minerID] = list
		}
	}

	return reply, nil
}

// Info returns aggregated runtime statistics for specified miners.
func (h *Hub) Info(ctx context.Context, request *pb.ID) (*pb.InfoReply, error) {
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
	route         *Route
}

func (h *Hub) onRequest(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !h.cluster.IsLeader() {
		return nil, status.Error(codes.PermissionDenied, "not a leader, check locator for leader endpoints")
	}

	return handler(ctx, request)
}

func (h *Hub) PushTask(stream pb.Hub_PushTaskServer) error {
	log.G(h.ctx).Info("handling PushTask request")

	if err := h.eventAuthorization.Authorize(stream.Context(), auth.Event(hubAPIPrefix+"PushTask"), nil); err != nil {
		return err
	}

	request, err := structs.NewImagePush(stream)
	if err != nil {
		return err
	}

	log.G(h.ctx).Info("pushing image", zap.Int64("size", request.ImageSize()))

	miner, _, err := h.findMinerByDeal(DealID(request.DealId()))
	if err != nil {
		log.G(h.ctx).Warn("unable to find miner by deal", zap.Error(err))
		return err
	}

	// TODO: Check storage size.

	client, err := miner.Client.Load(stream.Context())
	if err != nil {
		log.G(h.ctx).Warn("unable to forward push request to miner", zap.Error(err))
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

	if err := h.eventAuthorization.Authorize(stream.Context(), auth.Event(hubAPIPrefix+"PullTask"), request); err != nil {
		return err
	}

	ctx := log.WithLogger(h.ctx, log.G(h.ctx).With(zap.String("request", "pull task"), zap.String("id", uuid.New())))

	miner, _, err := h.findMinerByDeal(DealID(request.DealId))
	if err != nil {
		log.G(h.ctx).Warn("could not find miner by deal", zap.Error(err))
		return err
	}

	task, err := h.getTaskHistory(request.GetDealId(), request.GetTaskId())
	if err != nil {
		log.G(h.ctx).Warn("could not fetch task history by deal", zap.Error(err))
		return err
	}

	named, err := reference.ParseNormalizedNamed(task.ContainerID())
	if err != nil {
		log.G(h.ctx).Warn("could not parse image to reference", zap.Error(err), zap.String("image", task.ContainerID()))
		return err
	}

	tagged, err := reference.WithTag(named, fmt.Sprintf("%s_%s", request.GetDealId(), request.GetTaskId()))
	if err != nil {
		log.G(h.ctx).Warn("could not tag image", zap.Error(err), zap.String("image", task.ContainerID()))
		return err
	}
	imageID := tagged.String()

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
	allowed, ref, err := h.whitelist.Allowed(ctx, request.Container.Registry, request.Container.Image, request.Container.Auth)
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
	miner, usage, err := h.findMinerByOrder(OrderID(meta.BidID))
	if err != nil {
		return nil, err
	}

	taskID := h.generateTaskID()
	container := request.Container
	container.Registry = reference.Domain(ref)
	container.Image = reference.Path(ref)

	startRequest := &pb.MinerStartRequest{
		OrderId:   request.GetDealId(), // TODO: WTF?
		Id:        taskID,
		Container: container,
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

	routes := miner.registerRoutes(taskID, response.GetPortMap())

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

	tasksGauge.Inc()

	return reply, nil
}

func (h *Hub) findMinerByOrder(id OrderID) (*MinerCtx, *resource.Resources, error) {
	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	for _, miner := range h.miners {
		for _, order := range miner.Orders() {
			if order == id {
				usage, err := miner.OrderUsage(id)
				if err != nil {
					return nil, nil, err
				}
				return miner, &usage, nil
			}
		}
	}

	return nil, nil, ErrMinerNotFound
}

func (h *Hub) findMinerByDeal(id DealID) (*MinerCtx, *resource.Resources, error) {
	dealMeta, err := h.getDealMeta(id)
	if err != nil {
		log.G(h.ctx).Warn("unable to find deal meta by deal id", zap.Error(err))
		return nil, nil, err
	}

	return h.findMinerByOrder(OrderID(dealMeta.Order.Id))
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

	err = h.deleteTask(task.ID)
	if err != nil {
		log.G(ctx).Error("cannot delete task", zap.Error(err))
		return err
	}

	tasksGauge.Dec()

	return nil
}

func (h *Hub) GetDealInfo(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	meta, err := h.getDealMeta(DealID(id.Id))
	if err != nil {
		return nil, err
	}

	r := &pb.DealInfoReply{
		Id:        id,
		Order:     meta.Order.Unwrap(),
		Running:   &pb.StatusMapReply{Statuses: make(map[string]*pb.TaskStatusReply)},
		Completed: &pb.StatusMapReply{Statuses: make(map[string]*pb.TaskStatusReply)},
	}

	for _, t := range meta.Tasks {
		mctx, ok := h.getMinerByID(t.MinerId)
		if !ok {
			log.G(h.ctx).Warn("cannot get worker by id", zap.String("id", t.MinerId), zap.Error(err))
			continue
		}

		taskDetails, err := mctx.Client.TaskDetails(ctx, &pb.ID{Id: t.ID})
		if err != nil {
			log.G(h.ctx).Warn("cannot get task status",
				zap.String("workerID", t.MinerId), zap.String("taskID", t.ID), zap.Error(err))
			continue
		}

		if taskDetails.GetStatus() == pb.TaskStatusReply_RUNNING {
			r.Running.Statuses[t.ID] = taskDetails
		} else {
			r.Completed.Statuses[t.ID] = taskDetails
		}
	}

	return r, nil
}

func (h *Hub) getDealMeta(dealID DealID) (*DealMeta, error) {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	meta, ok := h.deals[dealID]
	if !ok {
		return nil, errDealNotFound
	}
	return meta, nil
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
				info.Tasks[taskID] = &pb.TaskStatusReply{Status: pb.TaskStatusReply_UNKNOWN}
			} else {
				info.Tasks[taskID] = taskInfo
			}
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
	log.G(h.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	if err := h.eventAuthorization.Authorize(server.Context(), auth.Event(hubAPIPrefix+"TaskLogs"), request); err != nil {
		return err
	}

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

	if !h.askPlans.HasOrder(request.AskId) {
		return nil, status.Errorf(codes.NotFound, "order not found")
	}

	bidOrder, err := h.market.GetOrderByID(h.ctx, &pb.ID{Id: request.GetBidId()})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "bid not found")
	}

	order, err := structs.NewOrder(bidOrder)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "bid order is malformed: %v", err)
	}

	askOrder, err := h.market.GetOrderByID(h.ctx, &pb.ID{Id: request.GetAskId()})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "ask not found")
	}

	if askOrder.GetByuerID() != "" {
		if askOrder.GetByuerID() != bidOrder.GetByuerID() {
			return nil, status.Errorf(codes.NotFound, "ask order is bound to special buyer, but IDs is not matching")
		}

		log.G(h.ctx).Info("handle proposal for bound order",
			zap.String("bidID", request.GetBidId()),
			zap.String("askID", request.GetAskId()),
		)
	}

	// Verify that bid's duration fits in ask.
	if bidOrder.GetDuration() > askOrder.GetDuration() {
		return nil, status.Errorf(codes.InvalidArgument, "bid's duration must fit in ask")
	}

	// Verify that bid price >= ask price, i.e we're not selling our resources
	// with lesser price than expected.
	if bidOrder.PricePerSecond.Cmp(askOrder.PricePerSecond) < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "BID price can not be less than ASK price")
	}

	// Verify that buyer has both enough money and allowance to have a deal.
	if err := h.eth.VerifyBuyerBalance(order); err != nil {
		return nil, err
	}
	if err := h.eth.VerifyBuyerAllowance(order); err != nil {
		return nil, err
	}

	resources, err := structs.NewResources(bidOrder.GetSlot().GetResources())
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

	ethAddr, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return nil, err
	}

	orderID := OrderID(order.GetID())
	if err := miner.Consume(orderID, &usage); err != nil {
		return nil, err
	}

	reservedDuration := time.Duration(10 * time.Minute)
	if err := h.orderShelter.Reserve(orderID, miner.ID(), *ethAddr, reservedDuration); err != nil {
		miner.Release(orderID)
		return nil, err
	}

	h.cluster.Synchronize(h.orderShelter.Dump())

	return &pb.Empty{}, nil
}

func (h *Hub) ApproveDeal(ctx context.Context, request *pb.ApproveDealRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling ApproveDeal request", zap.Any("request", request))

	bidOrder, err := h.market.GetOrderByID(h.ctx, &pb.ID{Id: request.GetBidID()})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "bid not found")
	}

	order, err := structs.NewOrder(bidOrder)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "bid order is malformed: %v", err)
	}

	// Ensure that deal is created in BC, if not - wait.
	dealID := DealID(request.GetDealID().Unwrap().String())
	buyerID := common.HexToAddress(bidOrder.GetByuerID())

	deal, err := h.eth.WaitForDealCreated(dealID, buyerID)
	if err != nil {
		log.G(h.ctx).Error("failed to find deal for approving",
			zap.Stringer("dealID", dealID),
			zap.String("bidID", request.GetBidID()),
			zap.String("askID", request.GetAskID()),
			zap.Error(err),
		)
		return nil, err
	}

	log.G(ctx).Info("received deal",
		zap.String("dealID", deal.GetId()),
		zap.String("dealPrice", deal.Price.Unwrap().String()),
		zap.String("orderPrice", order.PricePerSecond.Unwrap().String()),
	)

	if cmp := deal.Price.Cmp(pb.NewBigInt(order.GetTotalPrice())); cmp != 0 {
		return nil, fmt.Errorf("prices are not equal: %v != %v",
			deal.Price.Unwrap().String(), order.GetTotalPrice())
	}

	// Accept deal.
	err = h.eth.AcceptDeal(dealID.String())
	if err != nil {
		log.G(ctx).Error("failed to accept deal", zap.Stringer("dealID", dealID), zap.Error(err))
		return nil, err
	}

	// Commit reserved order.
	orderID := OrderID(order.GetID())
	reservedOrder, err := h.orderShelter.Commit(orderID)
	if err != nil {
		return nil, err
	}

	miner := h.GetMinerByID(reservedOrder.MinerID)
	if miner == nil {
		return nil, status.Errorf(codes.NotFound, "no miner with %s id found", reservedOrder.MinerID)
	}

	usage, err := miner.OrderUsage(orderID)
	if err != nil {
		return nil, err
	}

	// Cancel order from market.
	_, err = h.market.CancelOrder(h.ctx, &pb.Order{Id: request.GetAskID(), OrderType: pb.OrderType_ASK})
	if err != nil {
		log.G(ctx).Warn("cannot cancel ask order from marketplace",
			zap.String("askID", request.GetAskID()),
			zap.Error(err),
		)
	}

	dealMeta := &DealMeta{
		ID:      dealID,
		BidID:   request.GetBidID(),
		Order:   *order,
		Tasks:   make([]*TaskInfo, 0),
		MinerID: reservedOrder.MinerID,
		Usage:   usage,
		EndTime: time.Now().Add(order.GetDuration()),
	}

	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()
	h.deals[dealMeta.ID] = dealMeta

	dealsGauge.Inc()

	if err := h.cluster.Synchronize(h.deals); err != nil {
		log.G(h.ctx).Error("failed to synchronize deal with the cluster", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

func (h *Hub) waitForDealClosed(dealID DealID, buyerId string) error {
	return h.eth.WaitForDealClosed(h.ctx, dealID, buyerId)
}

// releaseDeal closes the specified deal freeing all associated resources.
func (h *Hub) releaseDeal(dealID DealID) error {
	tasks, err := h.popDealHistory(dealID)
	if err != nil {
		return err
	}

	log.S(h.ctx).Infof("stopping at max %d tasks due to deal closing", len(tasks))
	for _, task := range tasks {
		if h.isTaskFinished(task.ID) {
			continue
		}

		if err := h.stopTask(h.ctx, task); err != nil {
			log.G(h.ctx).Error("failed to stop task",
				zap.Stringer("dealID", dealID),
				zap.String("taskID", task.ID),
				zap.Error(err),
			)
		} else {
			tasksGauge.Dec()
		}
	}

	return nil
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

func (h *Hub) HasResources(resources *structs.Resources) bool {
	usage := resource.NewResources(
		int(resources.GetCpuCores()),
		int64(resources.GetMemoryInBytes()),
		resources.GetGPUCount(),
	)

	miner, err := h.findRandomMinerByUsage(&usage)
	return miner != nil && err == nil
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

	properties := h.deviceProperties[request.Id]
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

//TODO: It actually should be called AskPlans
func (h *Hub) Slots(ctx context.Context, request *pb.Empty) (*pb.SlotsReply, error) {
	log.G(h.ctx).Info("handling Slots request")
	return &pb.SlotsReply{Slots: h.askPlans.DumpSlots()}, nil
}

//TODO: Actually it is not slot, but AskPlan
func (h *Hub) InsertSlot(ctx context.Context, request *pb.InsertSlotRequest) (*pb.ID, error) {
	log.G(h.ctx).Info("handling InsertSlot request", zap.Any("request", request))

	slot, err := structs.NewSlot(request.Slot)
	if err != nil {
		return nil, err
	}

	ord := &pb.Order{
		OrderType:      pb.OrderType_ASK,
		Slot:           slot.Unwrap(),
		ByuerID:        request.BuyerID,
		PricePerSecond: request.PricePerSecond,
		SupplierID:     util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
	}
	order, err := structs.NewOrder(ord)
	if err != nil {
		return nil, err
	}

	id, err := h.askPlans.Add(h.ctx, order)
	if err != nil {
		return nil, err
	}

	err = h.cluster.Synchronize(h.askPlans.Dump())
	if err != nil {
		return nil, err
	}

	return &pb.ID{Id: id}, nil
}

//TODO: Actually it is not slot, but AskPlan
func (h *Hub) RemoveSlot(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("RemoveSlot request", zap.Any("id", request.Id))

	err := h.askPlans.Remove(h.ctx, request.Id)
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

	h.acl.Insert(common.HexToAddress(request.Id).Hex())
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

// New returns new Hub.
func New(ctx context.Context, cfg *Config, opts ...Option) (*Hub, error) {
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
		conn, err := xgrpc.NewWalletAuthenticatedClient(ctx, defaults.creds, cfg.Locator.Endpoint)
		if err != nil {
			return nil, err
		}

		defaults.locator = pb.NewLocatorClient(conn)
	}

	if defaults.market == nil {
		conn, err := xgrpc.NewWalletAuthenticatedClient(ctx, defaults.creds, cfg.Market.Endpoint)
		if err != nil {
			return nil, err
		}

		defaults.market = pb.NewMarketClient(conn)
	}

	if defaults.cluster == nil {
		defaults.cluster, defaults.clusterEvents, err = NewCluster(ctx, &cfg.Cluster, cfg.Endpoint, defaults.creds)
		if err != nil {
			return nil, err
		}
	}

	acl := NewACLStorage()
	if defaults.creds != nil {
		acl.Insert(defaults.ethAddr.Hex())
	}

	if len(cfg.Whitelist.PrivilegedAddresses) == 0 {
		cfg.Whitelist.PrivilegedAddresses = append(cfg.Whitelist.PrivilegedAddresses, defaults.ethAddr.Hex())
	}

	wl := NewWhitelist(ctx, &cfg.Whitelist)

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

		locatorPeriod: cfg.Locator.UpdatePeriod,
		locatorClient: defaults.locator,

		eth:    ethWrapper,
		market: defaults.market,

		deals:            make(map[DealID]*DealMeta),
		orderShelter:     nil,
		tasks:            make(map[string]*TaskInfo),
		miners:           make(map[string]*MinerCtx),
		associatedHubs:   make(map[string]struct{}),
		deviceProperties: make(map[string]DeviceProperties),
		acl:              acl,

		eventAuthorization: nil,

		certRotator: defaults.rot,
		creds:       defaults.creds,

		cluster:       defaults.cluster,
		clusterEvents: defaults.clusterEvents,

		whitelist: wl,
	}

	orderShelter := NewOrderShelter(h)
	h.orderShelter = orderShelter

	askPlans := NewAskPlans(h, defaults.market)
	h.askPlans = askPlans

	authorization := auth.NewEventAuthorization(h.ctx,
		auth.WithLog(log.G(ctx)),
		auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow("Handshake", "ProposeDeal").With(auth.NewNilAuthorization()),
		auth.Allow(hubManagementMethods...).With(auth.NewTransportAuthorization(h.ethAddr)),
		auth.Allow("TaskStatus", "StopTask").With(newDealAuthorization(ctx, h, newFromTaskDealExtractor(h))),
		auth.Allow("StartTask").With(newDealAuthorization(ctx, h, newFieldDealExtractor())),
		auth.Allow("TaskLogs").With(newDealAuthorization(ctx, h, newFromTaskDealExtractor(h))),
		auth.Allow("PushTask").With(newDealAuthorization(ctx, h, newContextDealExtractor())),
		auth.Allow("PullTask").With(newDealAuthorization(ctx, h, newRequestDealExtractor(func(request interface{}) (DealID, error) {
			return DealID(request.(*pb.PullTaskRequest).DealId), nil
		}))),
		auth.Allow("GetDealInfo").With(newDealAuthorization(ctx, h, newRequestDealExtractor(func(request interface{}) (DealID, error) {
			return DealID(request.(*pb.ID).GetId()), nil
		}))),
		auth.Allow("ApproveDeal").With(newOrderAuthorization(orderShelter, OrderExtractor(func(request interface{}) (OrderID, error) {
			return OrderID(request.(*pb.ApproveDealRequest).BidID), nil
		}))),
		auth.WithFallback(auth.NewDenyAuthorization()),
	)

	h.eventAuthorization = authorization

	logger := log.GetLogger(h.ctx)
	grpcServer := xgrpc.NewServer(logger,
		xgrpc.Credentials(h.creds),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.AuthorizationInterceptor(authorization),
		xgrpc.UnaryServerInterceptor(h.onRequest),
	)
	h.externalGrpc = grpcServer

	pb.RegisterHubServer(grpcServer, h)
	grpc_prometheus.Register(grpcServer)

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

	h.waiter.Go(func() error {
		return h.askPlans.Run(h.ctx)
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
	if err := h.cluster.RegisterAndLoadEntity("deals", &h.deals); err != nil {
		return err
	}

	askPlansData := AskPlansData{}
	if err := h.cluster.RegisterAndLoadEntity("ask_plans", &askPlansData); err != nil {
		return err
	}
	h.askPlans.RestoreFrom(askPlansData)

	reservedOrders := make(map[OrderID]ReservedOrder, 0)
	if err := h.cluster.RegisterAndLoadEntity("reserved_orders", &reservedOrders); err != nil {
		return err
	}
	h.orderShelter.RestoreFrom(reservedOrders)

	log.G(h.ctx).Info("fetched entities",
		zap.Any("tasks", h.tasks),
		zap.Any("device_properties", h.deviceProperties),
		zap.Any("acl", h.acl),
		zap.Any("ask_plans", h.askPlans),
		zap.Any("reserved_orders", h.orderShelter),
	)

	h.waiter.Go(h.runCluster)
	h.waiter.Go(h.listenClusterEvents)
	h.waiter.Go(h.startLocatorAnnouncer)
	h.waiter.Go(h.runAcceptedDealsWatcher)
	h.waiter.Go(h.runCluster)
	h.waiter.Go(h.watchDealsClosed)
	h.waiter.Go(func() error {
		return h.orderShelter.Run(h.ctx)
	})

	h.waiter.Wait()

	return nil
}

func (h *Hub) runDealsWatcher() error {
	timer := util.NewImmediateTicker(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			h.closeExpiredDeals()
		case <-h.ctx.Done():
			return nil
		}
	}
}

func (h *Hub) closeExpiredDeals() {
	now := time.Now()

	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	for dealID, dealMeta := range h.deals {
		if now.After(dealMeta.EndTime) {
			if err := h.eth.CloseDeal(dealID); err != nil {
				log.G(h.ctx).Error("failed to close deal using blockchain API",
					zap.Stringer("dealID", dealID),
					zap.Error(err),
				)
			}
		}
	}
}

func (h *Hub) runAcceptedDealsWatcher() error {
	timer := util.NewImmediateTicker(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.G(h.ctx).Debug("fetching accepted deals from the Blockchain")

			acceptedDeals, err := h.eth.GetAcceptedDeals(h.ctx)
			if err != nil {
				log.G(h.ctx).Warn("failed to fetch accepted deals from the Blockchain", zap.Error(err))
				continue
			}

			h.tasksMu.Lock()
			for _, acceptedDeal := range acceptedDeals {
				dealID := DealID(acceptedDeal.Id)

				deal, ok := h.deals[dealID]
				if !ok {
					continue
				}

				// Update deal expiration time according to the contract.
				deal.EndTime = acceptedDeal.EndTime.Unix()
			}

			h.cluster.Synchronize(h.deals)
			h.tasksMu.Unlock()
		case <-h.ctx.Done():
			return nil
		}
	}
}

// WatchDealsClosed watches ETH for currently closed deals.
func (h *Hub) watchDealsClosed() error {
	timer := util.NewImmediateTicker(30 * time.Second)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			log.G(h.ctx).Debug("fetching closed deals from the Blockchain")

			closedDeals, err := h.eth.GetClosedDeals(h.ctx)
			if err != nil {
				log.G(h.ctx).Warn("failed to fetch closed deals from the Blockchain", zap.Error(err))
				continue
			}

			for _, closedDeal := range closedDeals {
				dealID := DealID(closedDeal.Id)
				deal, ok := h.deals[dealID]
				if !ok {
					continue
				}

				orderID := OrderID(deal.Order.GetID())

				if err := h.releaseDeal(dealID); err != nil {
					log.G(h.ctx).Error("failed to release deal resources",
						zap.Stringer("dealID", dealID),
						zap.Stringer("orderID", orderID),
						zap.Error(err),
					)
					return err
				}

				miner, ok := h.getMinerByID(deal.MinerID)
				if !ok {
					continue
				}

				miner.Release(orderID)

				if err := h.publishOrder(orderID); err != nil {
					log.G(h.ctx).Error("failed to republish order on a market",
						zap.Stringer("dealID", dealID),
						zap.Stringer("orderID", orderID),
						zap.Error(err),
					)
				}
			}
		case <-h.ctx.Done():
			return nil
		}
	}
}

func (h *Hub) publishOrder(orderID OrderID) error {
	balance, err := h.eth.Balance()
	if err != nil {
		return err
	}

	if balance.Cmp(orderPublishThresholdETH) <= 0 {
		return fmt.Errorf("insufficient balance (%s <= %s)", balance.String(), orderPublishThresholdETH.String())
	}

	_, err = h.market.CreateOrder(h.ctx, &pb.Order{Id: orderID.String(), OrderType: pb.OrderType_ASK})
	return err
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
	default:
		if h.cluster.IsLeader() {
			return
		}
		switch value := value.(type) {
		case map[string]*TaskInfo:
			log.G(h.ctx).Info("synchronizing tasks from cluster")
			h.tasksMu.Lock()
			defer h.tasksMu.Unlock()
			h.tasks = value
		case map[string]DeviceProperties:
			h.devicePropertiesMu.Lock()
			defer h.devicePropertiesMu.Unlock()
			h.deviceProperties = value
		case AskPlansData:
			h.askPlans.RestoreFrom(value)
		case workerACLStorage:
			h.acl = &value
		case map[DealID]*DealMeta:
			h.tasksMu.Lock()
			defer h.tasksMu.Unlock()
			h.deals = value
			h.restoreResourceUsage()
		case map[OrderID]ReservedOrder:
			h.orderShelter.RestoreFrom(value)
		default:
			log.G(h.ctx).Warn("received unknown cluster event",
				zap.Any("event", value),
				zap.String("type", reflect.TypeOf(value).String()))
		}
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

func (h *Hub) SynchronizeAskPlans(data AskPlansData) error {
	return h.cluster.Synchronize(data)
}

func (h *Hub) registerMiner(miner *MinerCtx) {
	h.minersMu.Lock()
	h.miners[miner.uuid] = miner
	h.minersMu.Unlock()
	for address := range h.associatedHubs {
		log.G(h.ctx).Info("sending hub address", zap.String("hubAddress", address))
		miner.Client.DiscoverHub(h.ctx, &pb.DiscoverHubRequest{Endpoint: address})
	}

	h.minersMu.Lock()
	for dealID, dealMeta := range h.deals {
		if dealMeta.MinerID == miner.uuid {
			log.G(h.ctx).Debug("restoring resources consumption settings",
				zap.Stringer("dealID", dealID),
				zap.String("minerID", dealMeta.MinerID),
			)
			miner.Consume(OrderID(dealMeta.Order.GetID()), &dealMeta.Usage)
		}
	}
	h.minersMu.Unlock()
}

func (h *Hub) handleInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	miner, err := h.createMinerCtx(ctx, conn)
	if err != nil {
		log.G(h.ctx).Warn("failed to create miner context", zap.Error(err))
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

func (h *Hub) GetMinerByID(minerID string) *MinerCtx {
	miner, ok := h.getMinerByID(minerID)
	if miner != nil && ok {
		return miner
	} else {
		return nil
	}
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

	err := h.cluster.Synchronize(h.tasks)
	if err != nil {
		return err
	}
	return h.cluster.Synchronize(h.deals)
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
	}

	// Commit end time if such task exists in the history, if not - do nothing,
	// something terrible happened, but we just pretend nothing happened.
	taskHistory, ok := h.deals[taskInfo.DealId]
	if ok {
		for _, dealTaskInfo := range taskHistory.Tasks {
			if dealTaskInfo.ID == taskID {
				now := time.Now()
				dealTaskInfo.EndTime = &now
			}
		}
	}

	err := h.cluster.Synchronize(h.tasks)
	if err != nil {
		return err
	}

	return h.cluster.Synchronize(h.deals)
}

func (h *Hub) popDealHistory(dealID DealID) ([]*TaskInfo, error) {
	h.tasksMu.Lock()
	defer h.tasksMu.Unlock()

	tasks, ok := h.deals[dealID]
	if !ok {
		return nil, errDealNotFound
	}
	delete(h.deals, dealID)

	dealsGauge.Dec()

	err := h.cluster.Synchronize(h.deals)
	if err != nil {
		return nil, err
	}

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
	if !h.cluster.IsLeader() {
		return nil
	}

	members, err := h.cluster.Members()
	if err != nil {
		return err
	}

	log.G(h.ctx).Info("got cluster members for locator announcement", zap.Any("members", members))

	var (
		clientEndpoints []string
		workerEndpoints []string
	)
	for _, member := range members {
		clientEndpoints = append(clientEndpoints, member.Client...)
		workerEndpoints = append(workerEndpoints, member.Worker...)
	}
	req := &pb.AnnounceRequest{
		ClientEndpoints: clientEndpoints,
		WorkerEndpoints: workerEndpoints,
	}

	log.G(h.ctx).Info("announcing Hub addresses",
		zap.Stringer("eth", h.ethAddr),
		zap.Strings("client_endpoints", clientEndpoints),
		zap.Strings("worker_endpoints", workerEndpoints))

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

// NOTE: `tasksMu` must be held.
func (h *Hub) restoreResourceUsage() {
	log.G(h.ctx).Debug("synchronizing resource usage")

	h.minersMu.Lock()
	defer h.minersMu.Unlock()

	for dealID, dealInfo := range h.deals {
		miner, ok := h.miners[dealInfo.MinerID]
		if !ok {
			// Either miner has died or we have some kind of synchronization
			// error. Unfortunately we can't do anything meaningful here.
			log.G(h.ctx).Warn("detected worker inconsistency - found deal associated with unknown worker",
				zap.Stringer("dealID", dealID),
				zap.String("minerID", dealInfo.MinerID),
			)
			continue
		}

		// It's okay to ignore `AlreadyConsumed` errors here.
		miner.Consume(OrderID(dealInfo.Order.GetID()), &dealInfo.Usage)
	}

	for _, miner := range h.miners {
		for _, orderID := range miner.Orders() {
			orderExists := h.orderShelter.Exists(orderID)
			for _, dealInfo := range h.deals {
				if orderExists {
					break
				}
				if orderID == OrderID(dealInfo.Order.GetID()) {
					orderExists = true
				}
			}

			if !orderExists {
				miner.Release(orderID)
			}
		}
	}
}
