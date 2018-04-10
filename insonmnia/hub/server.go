package hub

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/miner"
	"github.com/sonm-io/core/insonmnia/npp"
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

const (
	hubAPIPrefix = "/sonm.Hub/"
)

var (
	errImageForbidden = status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")

	// The following methods require TLS authentication and checking for client
	// and Hub's wallet equality.
	// The wallet is passed as peer metadata.
	hubManagementMethods = []string{
		"Status",
		"Tasks",
		"Devices",
		"AskPlans",
		"CreateAskPlan",
		"RemoveAskPlan",
	}
)

// Hub collects miners, send them orders to spawn containers, etc.
type Hub struct {
	cfg          *miner.Config
	ctx          context.Context
	cancel       context.CancelFunc
	externalGrpc *grpc.Server

	grpcListener net.Listener

	ethKey  *ecdsa.PrivateKey
	ethAddr common.Address

	waiter    errgroup.Group
	startTime time.Time
	version   string

	// TLS certificate rotator
	certRotator util.HitlessCertRotator
	// GRPC TransportCredentials supported our Auth
	creds credentials.TransportCredentials

	whitelist Whitelist

	eventAuthorization *auth.AuthRouter

	worker *miner.Miner
}

// New returns new Hub.
func New(cfg *miner.Config, opts ...Option) (*Hub, error) {
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
		defaults.bcr, err = blockchain.NewAPI_DEPRECATED()
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.Whitelist.PrivilegedAddresses) == 0 {
		cfg.Whitelist.PrivilegedAddresses = append(cfg.Whitelist.PrivilegedAddresses, defaults.ethAddr.Hex())
	}

	wl := NewWhitelist(ctx, &cfg.Whitelist)

	if err := defaults.worker.RunBenchmarks(); err != nil {
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

		certRotator: defaults.rot,
		creds:       defaults.creds,

		whitelist: wl,
		worker:    defaults.worker,
	}

	authorization := auth.NewEventAuthorization(h.ctx,
		auth.WithLog(log.G(ctx)),
		auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow("ProposeDeal", "ApproveDeal").With(auth.NewNilAuthorization()),

		auth.Allow(hubManagementMethods...).With(auth.NewTransportAuthorization(h.ethAddr)),

		auth.Allow("TaskStatus").With(newMultiAuth(
			auth.NewTransportAuthorization(h.ethAddr),
			newDealAuthorization(ctx, h, newFromTaskDealExtractor(h)),
		)),
		auth.Allow("StopTask").With(newDealAuthorization(ctx, h, newFromTaskDealExtractor(h))),
		auth.Allow("JoinNetwork").With(newDealAuthorization(ctx, h, newFromNamedTaskDealExtractor(h, "TaskID"))),
		auth.Allow("StartTask").With(newDealAuthorization(ctx, h, newFieldDealExtractor())),
		auth.Allow("TaskLogs").With(newDealAuthorization(ctx, h, newFromTaskDealExtractor(h))),
		auth.Allow("PushTask").With(newDealAuthorization(ctx, h, newContextDealExtractor())),
		auth.Allow("PullTask").With(newDealAuthorization(ctx, h, newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
			return structs.DealID(request.(*pb.PullTaskRequest).DealId), nil
		}))),
		auth.Allow("GetDealInfo").With(newDealAuthorization(ctx, h, newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
			return structs.DealID(request.(*pb.ID).GetId()), nil
		}))),
		auth.WithFallback(auth.NewDenyAuthorization()),
	)

	h.eventAuthorization = authorization

	logger := log.GetLogger(h.ctx)
	grpcServer := xgrpc.NewServer(logger,
		xgrpc.Credentials(h.creds),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.AuthorizationInterceptor(authorization),
		xgrpc.VerifyInterceptor(),
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

	grpcL, err := npp.NewListener(h.ctx, h.cfg.Endpoint,
		npp.WithRendezvous(h.cfg.NPP.Rendezvous.Endpoints, h.creds),
		npp.WithRelay(h.cfg.NPP.Relay.Endpoints, h.ethKey),
		npp.WithLogger(log.G(h.ctx)),
	)
	if err != nil {
		log.G(h.ctx).Error("failed to listen", zap.String("address", h.cfg.Endpoint), zap.Error(err))
		return err
	}
	log.G(h.ctx).Info("listening for gRPC API connections", zap.Stringer("address", grpcL.Addr()))
	h.grpcListener = grpcL

	h.waiter.Go(h.listenAPI)

	h.waiter.Wait()

	return nil
}

// Status returns internal hub statistic
func (h *Hub) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
	uptime := uint64(time.Now().Sub(h.startTime).Seconds())
	reply := &pb.HubStatusReply{
		Uptime:    uptime,
		Platform:  util.GetPlatformName(),
		Version:   h.version,
		EthAddr:   util.PubKeyToAddr(h.ethKey.PublicKey).Hex(),
		TaskCount: uint32(len(h.worker.CollectTasksStatuses())),
	}

	return reply, nil
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

	task, err := h.worker.TaskDetails(ctx, &pb.ID{Id: request.GetTaskId()})
	if err != nil {
		log.G(h.ctx).Warn("could not fetch task history by deal", zap.Error(err))
		return err
	}

	named, err := reference.ParseNormalizedNamed(task.GetImageName())
	if err != nil {
		log.G(h.ctx).Warn("could not parse image to reference", zap.Error(err), zap.String("image", task.GetImageName()))
		return err
	}

	tagged, err := reference.WithTag(named, fmt.Sprintf("%s_%s", request.GetDealId(), request.GetTaskId()))
	if err != nil {
		log.G(h.ctx).Warn("could not tag image", zap.Error(err), zap.String("image", task.GetImageName()))
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

	// TODO(sshaman1101): REFACTOR:   only check for whitelist there,
	// TODO(sshaman1101): REFACTOR:   move all deals and tasks related code into the Worker.

	taskID := h.generateTaskID()
	container := request.Container
	container.Registry = reference.Domain(ref)
	container.Image = reference.Path(ref)

	startRequest := &pb.MinerStartRequest{
		OrderId:   request.GetDealId(), // TODO: WTF?
		Id:        taskID,
		Container: container,
		RestartPolicy: &pb.ContainerRestartPolicy{
			Name:              "",
			MaximumRetryCount: 0,
		},
	}

	response, err := h.worker.Start(ctx, startRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start %v", err)
	}

	reply := &pb.StartTaskReply{
		Id:         taskID,
		HubAddr:    h.ethAddr.Hex(),
		NetworkIDs: response.NetworkIDs,
	}

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
	return h.worker.Stop(ctx, request)
}

func (h *Hub) GetDealInfo(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	log.G(h.ctx).Info("handling GetDealInfo request")

	deal, err := h.worker.GetDealByID(structs.DealID(id.GetId()))
	if err != nil {
		return nil, err
	}

	reply := &pb.DealInfoReply{
		Id:    &pb.ID{Id: deal.Deal.GetId()},
		Order: deal.BidOrder.Unwrap(),
	}

	return reply, nil
}

func (h *Hub) Tasks(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(h.ctx).Info("handling Tasks request")
	return &pb.TaskListReply{Info: h.worker.CollectTasksStatuses()}, nil
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

// ProposeDeal is deprecated.
func (h *Hub) ProposeDeal(ctx context.Context, r *pb.DealRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling ProposeDeal request", zap.Any("request", r))
	return nil, nil
}

// ApproveDeal is deprecated.
func (h *Hub) ApproveDeal(ctx context.Context, request *pb.ApproveDealRequest) (*pb.Empty, error) {
	log.G(h.ctx).Info("handling ApproveDeal request", zap.Any("request", request))
	return nil, nil
}

func (h *Hub) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	return h.worker.Hardware().IntoProto(), nil
}

func (h *Hub) AskPlans(ctx context.Context, _ *pb.Empty) (*pb.AskPlansReply, error) {
	return h.worker.AskPlans(ctx)
}

func (h *Hub) CreateAskPlan(ctx context.Context, request *pb.CreateAskPlanRequest) (*pb.ID, error) {
	id, err := h.worker.CreateAskPlan(ctx, request)
	if err != nil {
		return nil, err
	}

	return &pb.ID{Id: id}, nil
}

func (h *Hub) RemoveAskPlan(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	if err := h.worker.RemoveAskPlan(ctx, request.GetId()); err != nil {
		return nil, err
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
