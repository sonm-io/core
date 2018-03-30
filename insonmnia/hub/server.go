package hub

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net"
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
	"github.com/sonm-io/core/insonmnia/math"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/npp"
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
		"TaskStatus",
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

type DeviceProperties map[string]float64

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	// TODO (3Hren): Probably port pool should be associated with the gateway implicitly.
	cfg          *Config
	ctx          context.Context
	cancel       context.CancelFunc
	externalGrpc *grpc.Server

	grpcListener net.Listener

	ethKey  *ecdsa.PrivateKey
	ethAddr common.Address

	announcer Announcer

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

	worker *miner.Miner
}

// New returns new Hub.
func New(ctx context.Context, cfg *Config, opts ...Option) (*Hub, error) {
	defaults := defaultHubOptions()
	for _, o := range opts {
		o(defaults)
	}

	if defaults.worker == nil {
		return nil, errors.New("cannot build Hub without worker")
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
		conn, err := xgrpc.NewClient(ctx, cfg.Locator.Endpoint, defaults.creds)
		if err != nil {
			return nil, err
		}

		defaults.locator = pb.NewLocatorClient(conn)
	}

	if defaults.market == nil {
		conn, err := xgrpc.NewClient(ctx, cfg.Market.Endpoint, defaults.creds)
		if err != nil {
			return nil, err
		}

		defaults.market = pb.NewMarketClient(conn)
	}

	if defaults.announcer == nil {
		a, err := newLocatorAnnouncer(
			defaults.ethKey,
			defaults.locator,
			cfg.Locator.UpdatePeriod,
			cfg)
		if err != nil {
			return nil, err
		}
		defaults.announcer = a
	}

	if len(cfg.Whitelist.PrivilegedAddresses) == 0 {
		cfg.Whitelist.PrivilegedAddresses = append(cfg.Whitelist.PrivilegedAddresses, defaults.ethAddr.Hex())
	}

	wl := NewWhitelist(ctx, &cfg.Whitelist)

	minerCtx, err := createMinerCtx(ctx, defaults.worker)
	if err != nil {
		return nil, err
	}

	hubState, err := newState(ctx, &cfg.Cluster, ethWrapper, defaults.market, minerCtx)
	if err != nil {
		return nil, err
	}

	h := &Hub{
		cfg:          cfg,
		ctx:          ctx,
		cancel:       cancel,
		externalGrpc: nil,

		ethKey:  defaults.ethKey,
		ethAddr: defaults.ethAddr,
		version: defaults.version,

		eth:    ethWrapper,
		market: defaults.market,

		certRotator: defaults.rot,
		creds:       defaults.creds,

		announcer: defaults.announcer,

		whitelist: wl,

		state:  hubState,
		worker: defaults.worker,
	}

	authorization := auth.NewEventAuthorization(h.ctx,
		auth.WithLog(log.G(ctx)),
		auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow("Handshake", "ProposeDeal").With(auth.NewNilAuthorization()),
		auth.Allow(hubManagementMethods...).With(auth.NewTransportAuthorization(h.ethAddr)),

		auth.Allow("TaskStatus").With(newMultiAuth(
			auth.NewTransportAuthorization(h.ethAddr),
			newDealAuthorization(ctx, hubState, newFromTaskDealExtractor(hubState)),
		)),
		auth.Allow("StopTask").With(newDealAuthorization(ctx, hubState, newFromTaskDealExtractor(hubState))),
		auth.Allow("JoinNetwork").With(newDealAuthorization(ctx, hubState, newFromNamedTaskDealExtractor(hubState, "TaskID"))),
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

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (h *Hub) Serve() error {
	h.startTime = time.Now()

	rendezvousEndpoints, err := h.cfg.NPP.Rendezvous.ConvertEndpoints()
	if err != nil {
		return err
	}

	relayEndpoints, err := h.cfg.NPP.Relay.ConvertEndpoints()
	if err != nil {
		return err
	}

	grpcL, err := npp.NewListener(h.ctx, h.cfg.Endpoint,
		npp.WithRendezvous(rendezvousEndpoints, h.creds),
		npp.WithRelay(relayEndpoints, h.ethKey),
		npp.WithLogger(log.G(h.ctx)),
	)
	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", h.cfg.Endpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for gRPC API connections", zap.Stringer("address", grpcL.Addr()))
	h.grpcListener = grpcL

	h.waiter.Go(h.listenAPI)

	h.waiter.Go(func() error {
		return h.state.RunMonitoring(h.ctx)
	})

	h.waiter.Go(h.startLocatorAnnouncer)

	h.waiter.Wait()

	return nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
	uptime := uint64(time.Now().Sub(h.startTime).Seconds())
	reply := &pb.HubStatusReply{
		Uptime:   uptime,
		Platform: util.GetPlatformName(),
		Version:  h.version,
		EthAddr:  util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
	}

	return reply, nil
}

func (h *Hub) onRequest(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

	return h.worker.Load(stream)
}

func (h *Hub) PullTask(request *pb.PullTaskRequest, stream pb.Hub_PullTaskServer) error {
	log.G(h.ctx).Info("handling PullTask request", zap.Any("request", request))

	if err := h.eventAuthorization.Authorize(stream.Context(), auth.Event(hubAPIPrefix+"PullTask"), request); err != nil {
		return err
	}

	ctx := log.WithLogger(h.ctx, log.G(h.ctx).With(zap.String("request", "pull task"), zap.String("id", uuid.New())))

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

	return h.worker.Save(&pb.SaveRequest{ImageID: imageID}, stream)
}

func (h *Hub) StartTask(ctx context.Context, request *pb.StartTaskRequest) (*pb.StartTaskReply, error) {
	log.G(h.ctx).Info("handling StartTask request", zap.Any("request", request))

	taskRequest, err := structs.NewStartTaskRequest(request)
	if err != nil {
		return nil, err
	}

	return h.startTask(ctx, taskRequest)
}

func (h *Hub) startTask(ctx context.Context, request *structs.StartTaskRequest) (*pb.StartTaskReply, error) {
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
	usage, err := h.state.minerCtx.OrderUsage(OrderID(meta.BidID))
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

	response, err := h.worker.Start(ctx, startRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	info := TaskInfo{*request, *response, taskID, dealID, "self", nil}

	err = h.state.SaveTask(DealID(request.GetDealId()), &info)
	if err != nil {
		h.worker.Stop(ctx, &pb.ID{Id: taskID})
		return nil, err
	}

	if err := h.state.Dump(); err != nil {
		log.G(h.ctx).Error("failed to dump state", zap.Error(err))
	}

	reply := &pb.StartTaskReply{
		Id:         taskID,
		HubAddr:    h.ethAddr.Hex(),
		NetworkIDs: response.NetworkIDs,
	}

	tasksGauge.Inc()

	return reply, nil
}

func (h *Hub) JoinNetwork(ctx context.Context, request *pb.HubJoinNetworkRequest) (*pb.NetworkSpec, error) {
	log.G(h.ctx).Info("handling JoinNetwork request", zap.Any("request", request))
	return h.worker.JoinNetwork(ctx, &pb.ID{Id: request.NetworkID})
}

func (h *Hub) generateTaskID() string {
	return uuid.New()
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
		taskDetails, err := h.worker.TaskDetails(ctx, &pb.ID{Id: t.ID})
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

func (h *Hub) Tasks(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling TaskList request")
	return h.state.GetTaskList(ctx)
}

func (h *Hub) TaskStatus(ctx context.Context, request *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(h.ctx).Info("handling TaskStatus request", zap.Any("req", request))
	return h.worker.TaskDetails(ctx, request)
}

func (h *Hub) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
	log.G(h.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	if err := h.eventAuthorization.Authorize(server.Context(), auth.Event(hubAPIPrefix+"TaskLogs"), request); err != nil {
		return err
	}

	return h.worker.TaskLogs(request, server)
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

	ethAddr, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return nil, err
	}

	orderID := OrderID(order.GetID())
	if err := h.state.minerCtx.Consume(orderID, &usage); err != nil {
		return nil, err
	}

	reservedDuration := time.Duration(10 * time.Minute)
	if err := h.state.ReserveOrder(orderID, "self", *ethAddr, reservedDuration); err != nil {
		h.state.minerCtx.Release(orderID)
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

	usage, err := h.state.minerCtx.OrderUsage(orderID)
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

func (h *Hub) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	cap, err := h.worker.Info(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return &pb.DevicesReply{
		CPUs: cap.Capabilities.Cpu,
		GPUs: cap.Capabilities.Gpu,
	}, nil
}

func (h *Hub) AskPlans(ctx context.Context, request *pb.Empty) (*pb.AskPlansReply, error) {
	log.G(h.ctx).Info("handling Slots request")
	return &pb.AskPlansReply{Slots: h.state.DumpSlots()}, nil
}

//TODO: Actually it is not slot, but AskPlan.
func (h *Hub) CreateAskPlan(ctx context.Context, request *pb.CreateAskPlanRequest) (*pb.ID, error) {
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
func (h *Hub) RemoveAskPlan(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
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

func (h *Hub) listenAPI() error {
	for {
		select {
		case <-h.ctx.Done():
			return h.ctx.Err()
		default:
		}

		if err := h.externalGrpc.Serve(h.grpcListener); err != nil {
			if _, ok := err.(npp.TransportError); ok {
				timer := time.NewTimer(1 * time.Second)
				select {
				case <-h.ctx.Done():
				case <-timer.C:
				}
				timer.Stop()
			}
		}
	}
}

// Close disposes all resources attached to the Hub
func (h *Hub) Close() {
	h.cancel()
	h.externalGrpc.Stop()
	if h.certRotator != nil {
		h.certRotator.Close()
	}
	h.worker.Close()
	h.waiter.Wait()
}

func (h *Hub) startLocatorAnnouncer() error {
	h.announcer.Start(h.ctx)
	return nil
}
