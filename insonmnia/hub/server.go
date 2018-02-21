package hub

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"
	"net"
	"reflect"
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

	announcer     Announcer
	cluster       Cluster
	clusterEvents <-chan ClusterEvent

	// TODO: rediscover jobs if Miner disconnected.
	// TODO: store this data in some Storage interface.

	waiter    errgroup.Group
	startTime time.Time
	version   string

	eth    ETH
	market pb.MarketClient

	// TLS certificate rotator
	certRotator util.HitlessCertRotator
	// GRPC TransportCredentials supported our Auth
	creds credentials.TransportCredentials

	whitelist Whitelist

	state *state

	eventAuthorization *auth.AuthRouter
}

type DeviceProperties map[string]float64

// Ping should be used as Healthcheck for Hub
func (h *Hub) Ping(ctx context.Context, _ *pb.Empty) (*pb.PingReply, error) {
	return &pb.PingReply{}, nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
	clients, workers, err := collectEndpoints(h.cluster)
	if err != nil {
		log.G(h.ctx).Warn("cannot collect cluster endpoints", zap.Error(err))
		return nil, err
	}

	var (
		minersCount = h.state.MinersCount()
		uptime      = uint64(time.Now().Sub(h.startTime).Seconds())
		reply       = &pb.HubStatusReply{
			MinerCount:      uint64(minersCount),
			Uptime:          uptime,
			Platform:        util.GetPlatformName(),
			Version:         h.version,
			EthAddr:         util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
			ClientEndpoint:  clients[0],
			WorkerEndpoints: workers,
			AnnounceError:   h.announcer.ErrorMsg(),
		}
	)

	return reply, nil
}

// List returns attached miners
func (h *Hub) List(ctx context.Context, request *pb.Empty) (*pb.ListReply, error) {
	reply := &pb.ListReply{
		Info: make(map[string]*pb.ListReply_ListValue),
	}

	for _, minerID := range h.state.MinerIDs() {
		reply.Info[minerID] = new(pb.ListReply_ListValue)
	}

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
	client, ok := h.state.GetMinerByID(request.GetId())
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

	miner, _, err := h.state.GetMinerByDeal(DealID(request.DealId()))
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

	miner, _, err := h.state.GetMinerByDeal(DealID(request.DealId))
	if err != nil {
		log.G(h.ctx).Warn("could not find miner by deal", zap.Error(err))
		return err
	}

	task, err := h.state.GetTaskInfo(request.GetDealId(), request.GetTaskId())
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
	meta, err := h.state.GetDealMeta(dealID)
	if err != nil {
		// Hub knows nothing about this deal
		return nil, errDealNotFound
	}

	// Extract proper miner associated with the deal specified.
	miner, usage, err := h.state.GetMinerByOrder(OrderID(meta.BidID))
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

	err = h.state.SaveTask(DealID(request.GetDealId()), &info)
	if err != nil {
		miner.Client.Stop(ctx, &pb.ID{Id: taskID})
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
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

// StopTask sends termination request to a miner handling the task
func (h *Hub) StopTask(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling StopTask request", zap.Any("req", request))
	if err := h.state.StopTask(ctx, request.Id); err != nil {
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

func (h *Hub) GetDealInfo(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	meta, err := h.state.GetDealMeta(DealID(id.Id))
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
		mctx, ok := h.state.GetMinerByID(t.MinerId)
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

func (h *Hub) TaskList(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	return h.state.GetTaskList(ctx)
}

func (h *Hub) MinerStatus(ctx context.Context, request *pb.ID) (*pb.StatusMapReply, error) {
	log.G(h.ctx).Info("handling MinerStatus request", zap.Any("req", request))
	return h.state.GetMinerStatus(request.Id)
}

func (h *Hub) TaskStatus(ctx context.Context, request *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(h.ctx).Info("handling TaskStatus request", zap.Any("req", request))
	return h.state.GetTaskStatus(request.Id)
}

func (h *Hub) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
	log.G(h.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	if err := h.eventAuthorization.Authorize(server.Context(), auth.Event(hubAPIPrefix+"TaskLogs"), request); err != nil {
		return err
	}

	task, ok := h.state.getTaskByID(request.Id)
	if !ok {
		return errors.Errorf("no such task: %s", request.Id)
	}

	minerCtx, ok := h.state.GetMinerByID(task.MinerId)
	if !ok {
		return status.Errorf(codes.NotFound, "no miner %s for task %s", task.MinerId, request.Id)
	}

	client, err := minerCtx.Client.TaskLogs(server.Context(), request)
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

	if !h.state.HasOrder(request.AskId) {
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

	miner, err := h.state.GetRandomMinerByUsage(&usage)
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
	if err := h.state.ReserveOrder(orderID, miner.ID(), *ethAddr, reservedDuration); err != nil {
		miner.Release(orderID)
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

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
	reservedOrder, err := h.state.CommitOrder(orderID)
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

	h.state.SetDealMeta(dealMeta)

	dealsGauge.Inc()

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

func (h *Hub) waitForDealClosed(dealID DealID, buyerId string) error {
	return h.eth.WaitForDealClosed(h.ctx, dealID, buyerId)
}

func (h *Hub) DiscoverHub(ctx context.Context, request *pb.DiscoverHubRequest) (*pb.Empty, error) {
	h.onNewHub(request.Endpoint)
	return &pb.Empty{}, nil
}

func (h *Hub) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	return h.state.GetDevices()
}

func (h *Hub) MinerDevices(ctx context.Context, request *pb.ID) (*pb.DevicesReply, error) {
	return h.state.GetMinerDevices(request)
}

func (h *Hub) GetDeviceProperties(ctx context.Context, request *pb.ID) (*pb.GetDevicePropertiesReply, error) {
	log.G(h.ctx).Info("handling GetMinerProperties request", zap.Any("req", request))
	return h.state.GetDeviceProperties(request.Id)
}

func (h *Hub) SetDeviceProperties(ctx context.Context, request *pb.SetDevicePropertiesRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling SetDeviceProperties request", zap.Any("req", request))
	h.state.SetDeviceProperties(request.ID, request.Properties)

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

//TODO: It actually should be called AskPlans.
func (h *Hub) Slots(ctx context.Context, request *pb.Empty) (*pb.SlotsReply, error) {
	log.G(h.ctx).Info("handling Slots request")
	return &pb.SlotsReply{Slots: h.state.DumpSlots()}, nil
}

//TODO: Actually it is not slot, but AskPlan.
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

	id, err := h.state.AddSlot(h.ctx, order)
	if err != nil {
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.ID{Id: id}, nil
}

//TODO: Actually it is not slot, but AskPlan
func (h *Hub) RemoveSlot(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("RemoveSlot request", zap.Any("id", request.Id))

	err := h.state.RemoveSlot(h.ctx, request.Id)
	if err != nil {
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

// GetRegisteredWorkers returns a list of Worker IDs that are allowed to
// connect to the Hub.
func (h *Hub) GetRegisteredWorkers(ctx context.Context, empty *pb.Empty) (*pb.GetRegisteredWorkersReply, error) {
	log.G(h.ctx).Info("handling GetRegisteredWorkers request")
	return &pb.GetRegisteredWorkersReply{Ids: h.state.GetRegisteredWorkers()}, nil
}

// RegisterWorker allows Worker with given ID to connect to the Hub
func (h *Hub) RegisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling RegisterWorker request", zap.String("id", request.GetId()))
	h.state.ACLInsert(common.HexToAddress(request.Id).Hex())

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	return &pb.Empty{}, nil
}

// DeregisterWorkers deny Worker with given ID to connect to the Hub
func (h *Hub) DeregisterWorker(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling DeregisterWorker request", zap.String("id", request.GetId()))

	if existed := h.state.ACLRemove(request.Id); !existed {
		log.G(h.ctx).Warn("attempt to deregister unregistered worker", zap.String("id", request.GetId()))
	} else {
		if err := h.state.Dump(); err != nil {
			log.G(h.ctx).Error("failed to dump state", zap.Error(err))
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

	if defaults.announcer == nil {
		defaults.announcer = newLocatorAnnouncer(
			defaults.ethKey,
			defaults.locator,
			cfg.Locator.UpdatePeriod,
			defaults.cluster)
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

	acl := newWorkerACLStorage()
	if defaults.creds != nil {
		acl.Insert(defaults.ethAddr.Hex())
	}

	if len(cfg.Whitelist.PrivilegedAddresses) == 0 {
		cfg.Whitelist.PrivilegedAddresses = append(cfg.Whitelist.PrivilegedAddresses, defaults.ethAddr.Hex())
	}

	wl := NewWhitelist(ctx, &cfg.Whitelist)
	hubState, err := newState(ctx, acl, ethWrapper, defaults.market, defaults.cluster)
	if err != nil {
		return nil, err
	}

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

		eth:    ethWrapper,
		market: defaults.market,

		certRotator: defaults.rot,
		creds:       defaults.creds,

		announcer:     defaults.announcer,
		cluster:       defaults.cluster,
		clusterEvents: defaults.clusterEvents,

		whitelist: wl,

		state: hubState,
	}

	authorization := auth.NewEventAuthorization(h.ctx,
		auth.WithLog(log.G(ctx)),
		auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow("Handshake", "ProposeDeal").With(auth.NewNilAuthorization()),
		auth.Allow(hubManagementMethods...).With(auth.NewTransportAuthorization(h.ethAddr)),
		auth.Allow("TaskStatus", "StopTask").With(newDealAuthorization(ctx, hubState, newFromTaskDealExtractor(hubState))),
		auth.Allow("StartTask").With(newDealAuthorization(ctx, hubState, newFieldDealExtractor())),
		auth.Allow("TaskLogs").With(newDealAuthorization(ctx, hubState, newFromTaskDealExtractor(hubState))),
		auth.Allow("PushTask").With(newDealAuthorization(ctx, hubState, newContextDealExtractor())),
		auth.Allow("PullTask").With(newDealAuthorization(ctx, hubState, newRequestDealExtractor(func(request interface{}) (DealID, error) {
			return DealID(request.(*pb.PullTaskRequest).DealId), nil
		}))),
		auth.Allow("GetDealInfo").With(newDealAuthorization(ctx, hubState, newRequestDealExtractor(func(request interface{}) (DealID, error) {
			return DealID(request.(*pb.ID).GetId()), nil
		}))),
		auth.Allow("ApproveDeal").With(newOrderAuthorization(hubState, OrderExtractor(func(request interface{}) (OrderID, error) {
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
	log.G(h.ctx).Info("new hub discovered", zap.String("endpoint", endpoint))
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
		return h.state.RunMonitoring(h.ctx)
	})

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
		h.announcer.Once(h.ctx)
	case LeadershipEvent:
		h.announcer.Once(h.ctx)
	default:
		if h.cluster.IsLeader() {
			return
		}

		switch value := value.(type) {
		case stateJSON:
			log.G(h.ctx).Debug("received state", zap.Any("state", value))
			if err := h.state.Load(&value); err != nil {
				log.G(h.ctx).Error("failed to load state", zap.Any("state", value), zap.Error(err))
			}
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

func (h *Hub) handleInterconnect(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	log.G(ctx).Info("miner connected", zap.Stringer("remote", conn.RemoteAddr()))

	miner, err := h.createMinerCtx(ctx, conn)
	if err != nil {
		log.G(h.ctx).Warn("failed to create miner context", zap.Error(err))
		return
	}

	h.state.RegisterMiner(miner)

	go func() {
		miner.pollStatuses()
		miner.Close()
	}()

	miner.ping()
	miner.Close()

	h.state.DeleteMiner(miner.ID())
}

func (h *Hub) GetMinerByID(minerID string) *MinerCtx {
	miner, ok := h.state.GetMinerByID(minerID)
	if miner != nil && ok {
		return miner
	} else {
		return nil
	}
}

func (h *Hub) startLocatorAnnouncer() error {
	h.announcer.Start(h.ctx)
	return nil
}

func collectEndpoints(cluster Cluster) ([]string, []string, error) {
	members, err := cluster.Members()
	if err != nil {
		return nil, nil, err
	}

	var clients []string
	var workers []string
	for _, member := range members {
		clients = append(clients, member.Client...)
		workers = append(workers, member.Worker...)
	}

	return clients, workers, nil
}
