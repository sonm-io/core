package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mohae/deepcopy"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/hardware/disk"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/insonmnia/worker/salesman"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/debug"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xgrpc"

	// todo: drop alias
	bm "github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	workerAPIPrefix = "/sonm.WorkerManagement/"
	taskAPIPrefix   = "/sonm.Worker/"
)

var (
	workerManagementMethods = []string{
		workerAPIPrefix + "Status",
		workerAPIPrefix + "Tasks",
		workerAPIPrefix + "Devices",
		workerAPIPrefix + "FreeDevices",
		workerAPIPrefix + "AskPlans",
		workerAPIPrefix + "CreateAskPlan",
		workerAPIPrefix + "RemoveAskPlan",
		workerAPIPrefix + "PurgeAskPlans",
	}
)

// Worker holds information about jobs, make orders to Observer and communicates with Worker
type Worker struct {
	*options

	mu        sync.Mutex
	hardware  *hardware.Hardware
	resources *resource.Scheduler
	salesman  *salesman.Salesman

	eventAuthorization *auth.AuthRouter

	// Maps StartRequest's IDs to containers' IDs
	// TODO: It's doubtful that we should keep this map here instead in the Overseer.
	containers map[string]*ContainerInfo

	controlGroup  cgroups.CGroup
	cGroupManager cgroups.CGroupManager
	listener      *npp.Listener
	externalGrpc  *grpc.Server
	startTime     time.Time
}

func NewWorker(opts ...Option) (m *Worker, err error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	m = &Worker{
		options:    o,
		containers: make(map[string]*ContainerInfo),
	}

	if err := m.SetupDefaults(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupMaster(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupAuthorization(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupControlGroup(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupHardware(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.runBenchmarks(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupResources(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupSalesman(); err != nil {
		m.Close()
		return nil, err
	}

	if err := m.setupServer(); err != nil {
		m.Close()
		return nil, err
	}

	return m, nil
}

// Serve starts handling incoming API gRPC requests
func (m *Worker) Serve() error {
	defer m.Close()

	m.startTime = time.Now()
	if err := m.waitMasterApproved(); err != nil {
		return err
	}

	listener, err := npp.NewListener(m.ctx, m.cfg.Endpoint,
		npp.WithNPPBacklog(m.cfg.NPP.Backlog),
		npp.WithNPPBackoff(m.cfg.NPP.MinBackoffInterval, m.cfg.NPP.MaxBackoffInterval),
		npp.WithRendezvous(m.cfg.NPP.Rendezvous, m.creds),
		npp.WithRelay(m.cfg.NPP.Relay.Endpoints, m.key, log.G(m.ctx)),
		npp.WithLogger(log.G(m.ctx)),
	)
	if err != nil {
		log.G(m.ctx).Error("failed to listen", zap.String("address", m.cfg.Endpoint), zap.Error(err))
		return err
	}
	m.listener = listener

	log.G(m.ctx).Info("listening for gRPC API connections", zap.Stringer("address", listener.Addr()))
	err = m.externalGrpc.Serve(listener)

	return err
}

func (m *Worker) waitMasterApproved() error {
	if m.cfg.Development != nil && m.cfg.Development.DisableMasterApproval {
		return nil
	}
	selfAddr := m.ethAddr().Hex()
	log.S(m.ctx).Infof("waiting approval for %s from master %s", selfAddr, m.cfg.Master.Hex())
	expectedMaster := m.cfg.Master.Hex()
	ticker := util.NewImmediateTicker(time.Second)
	for {
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case <-ticker.C:
			addr, err := m.eth.Market().GetMaster(m.ctx, m.ethAddr())
			if err != nil {
				log.S(m.ctx).Warnf("failed to get master: %s, retrying...", err)
			}
			curMaster := addr.Hex()
			if curMaster == selfAddr {
				log.S(m.ctx).Infof("still no approval for %s from %s, continue waiting", m.ethAddr().Hex(), m.cfg.Master.Hex())
				continue
			}
			if curMaster != expectedMaster {
				return fmt.Errorf("received unexpected master %s", curMaster)
			}
			return nil
		}
	}
}

func (m *Worker) ethAddr() common.Address {
	return crypto.PubkeyToAddress(m.key.PublicKey)
}

func (m *Worker) setupMaster() error {
	if m.cfg.Development != nil && m.cfg.Development.DisableMasterApproval {
		return nil
	}
	log.S(m.ctx).Info("checking current master")
	addr, err := m.eth.Market().GetMaster(m.ctx, m.ethAddr())
	if err != nil {
		return err
	}
	if addr.Big().Cmp(m.ethAddr().Big()) == 0 {
		log.S(m.ctx).Infof("master is not confirmed or not set, sending request from %s to %s",
			m.ethAddr().Hex(), m.cfg.Master.Hex())
		err = m.eth.Market().RegisterWorker(m.ctx, m.key, m.cfg.Master)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Worker) setupAuthorization() error {
	managementAuthOptions := []auth.Authorization{
		auth.NewTransportAuthorization(m.ethAddr()),
		auth.NewTransportAuthorization(m.cfg.Master),
	}

	if m.cfg.Admin != nil {
		managementAuthOptions = append(managementAuthOptions, auth.NewTransportAuthorization(*m.cfg.Admin))
	}

	managementAuth := newAnyOfAuth(managementAuthOptions...)

	authorization := auth.NewEventAuthorization(m.ctx,
		auth.WithLog(log.G(m.ctx)),
		// Note: need to refactor auth router to support multiple prefixes for methods.
		// auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow(workerManagementMethods...).With(managementAuth),

		auth.Allow(taskAPIPrefix+"TaskStatus").With(newAnyOfAuth(
			managementAuth,
			newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m)),
		)),
		auth.Allow(taskAPIPrefix+"StopTask").With(newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m))),
		auth.Allow(taskAPIPrefix+"JoinNetwork").With(newDealAuthorization(m.ctx, m, newFromNamedTaskDealExtractor(m, "TaskID"))),
		auth.Allow(taskAPIPrefix+"StartTask").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
			return structs.DealID(request.(*pb.StartTaskRequest).GetDealID().Unwrap().String()), nil
		}))),
		auth.Allow(taskAPIPrefix+"TaskLogs").With(newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m))),
		auth.Allow(taskAPIPrefix+"PushTask").With(newDealAuthorization(m.ctx, m, newContextDealExtractor())),
		auth.Allow(taskAPIPrefix+"PullTask").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
			return structs.DealID(request.(*pb.PullTaskRequest).DealId), nil
		}))),
		auth.Allow(taskAPIPrefix+"GetDealInfo").With(newDealAuthorization(m.ctx, m, newRequestDealExtractor(func(request interface{}) (structs.DealID, error) {
			return structs.DealID(request.(*pb.ID).GetId()), nil
		}))),
		auth.WithFallback(auth.NewDenyAuthorization()),
	)

	m.eventAuthorization = authorization
	return nil
}

func (m *Worker) setupControlGroup() error {
	cgName := "sonm-worker-parent"
	cgResources := &specs.LinuxResources{}
	if m.cfg.Resources != nil {
		cgName = m.cfg.Resources.Cgroup
		cgResources = m.cfg.Resources.Resources
	}

	cgroup, cGroupManager, err := cgroups.NewCgroupManager(cgName, cgResources)
	if err != nil {
		return err
	}
	m.controlGroup = cgroup
	m.cGroupManager = cGroupManager
	return nil
}

func (m *Worker) setupHardware() error {
	// TODO: Do all the stuff inside hardware ctor
	hardwareInfo, err := hardware.NewHardware()
	if err != nil {
		return err
	}

	// check if memory is limited into cgroup
	if s, err := m.controlGroup.Stats(); err == nil {
		if s.MemoryLimit != 0 && s.MemoryLimit < hardwareInfo.RAM.Device.Total {
			hardwareInfo.RAM.Device.Available = s.MemoryLimit
		}
	}

	// apply info about GPUs, expose to logs
	m.plugins.ApplyHardwareInfo(hardwareInfo)
	hardwareInfo.SetNetworkIncoming(m.publicIPs)
	//TODO: configurable?
	hardwareInfo.Network.NetFlags.SetOutbound(true)
	m.hardware = hardwareInfo
	return nil
}

func (m *Worker) listenDeals(dealsCh <-chan *pb.Deal) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case deal := <-dealsCh:
			if deal.Status == pb.DealStatus_DEAL_CLOSED {
				if err := m.cancelDealTasks(deal); err != nil {
					log.S(m.ctx).Warnf("could not stop tasks for closed deal %s: %s", deal.GetId().Unwrap().String(), err)
				}
			}
		}
	}
}

func (m *Worker) cancelDealTasks(deal *pb.Deal) error {
	dealID := deal.GetId().Unwrap().String()
	var toDelete []*ContainerInfo

	m.mu.Lock()
	for key, container := range m.containers {
		if container.DealID == dealID {
			toDelete = append(toDelete, container)
			delete(m.containers, key)
		}
	}
	m.mu.Unlock()

	result := multierror.NewMultiError()
	for _, container := range toDelete {
		if err := m.ovs.OnDealFinish(m.ctx, container.ID); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

func (m *Worker) saveContainerInfo(id string, info ContainerInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.containers[id] = &info
}

func (m *Worker) GetContainerInfo(id string) (*ContainerInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.containers[id]
	return info, ok
}

func (m *Worker) getContainerIdByTaskId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.containers[id]
	if ok {
		return info.ID, ok
	}
	return "", ok
}

func (m *Worker) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	return m.hardware.IntoProto(), nil
}

// Status returns internal worker statistic
func (m *Worker) Status(ctx context.Context, _ *pb.Empty) (*pb.StatusReply, error) {
	uptime := uint64(time.Now().Sub(m.startTime).Seconds())

	rendezvousStatus := "not connected"
	nppMetrics := m.listener.Metrics()
	if nppMetrics.RendezvousAddr != nil {
		rendezvousStatus = nppMetrics.RendezvousAddr.String()
	}

	reply := &pb.StatusReply{
		Uptime:           uptime,
		Platform:         util.GetPlatformName(),
		Version:          m.version,
		EthAddr:          m.ethAddr().Hex(),
		TaskCount:        uint32(len(m.CollectTasksStatuses(pb.TaskStatusReply_RUNNING))),
		RendezvousStatus: rendezvousStatus,
	}

	return reply, nil
}

// FreeDevice provides information about unallocated resources
// that can be turned into ask-plans.
// TODO: Looks like DevicesReply is not really suitable here
func (m *Worker) FreeDevices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	resources, err := m.resources.GetFree()
	if err != nil {
		return nil, err
	}
	hardware, err := m.hardware.LimitTo(resources)
	if err != nil {
		return nil, err
	}

	return hardware.IntoProto(), nil
}

func (m *Worker) setStatus(status *pb.TaskStatusReply, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}

	m.containers[id].status = status.GetStatus()
	if status.Status == pb.TaskStatusReply_BROKEN || status.Status == pb.TaskStatusReply_FINISHED {
		m.resources.ReleaseTask(id)
	}
}

func (m *Worker) listenForStatus(statusListener chan pb.TaskStatusReply_Status, id string) {
	select {
	case newStatus, ok := <-statusListener:
		if !ok {
			return
		}
		m.setStatus(&pb.TaskStatusReply{Status: newStatus}, id)
	case <-m.ctx.Done():
		return
	}
}

func (m *Worker) PushTask(stream pb.Worker_PushTaskServer) error {
	log.G(m.ctx).Info("handling PushTask request")
	if err := m.eventAuthorization.Authorize(stream.Context(), auth.Event(taskAPIPrefix+"PushTask"), nil); err != nil {
		return err
	}

	request, err := structs.NewImagePush(stream)
	if err != nil {
		return err
	}
	log.G(m.ctx).Info("pushing image", zap.Int64("size", request.ImageSize()))

	result, err := m.ovs.Load(stream.Context(), newChunkReader(stream))
	if err != nil {
		return err
	}

	log.G(m.ctx).Info("image loaded, set trailer", zap.String("trailer", result.Status))
	stream.SetTrailer(metadata.Pairs("status", result.Status))
	return nil
}

func (m *Worker) PullTask(request *pb.PullTaskRequest, stream pb.Worker_PullTaskServer) error {
	log.G(m.ctx).Info("handling PullTask request", zap.Any("request", request))

	if err := m.eventAuthorization.Authorize(stream.Context(), auth.Event(taskAPIPrefix+"PullTask"), request); err != nil {
		return err
	}

	ctx := log.WithLogger(m.ctx, log.G(m.ctx).With(zap.String("request", "pull task"), zap.String("id", uuid.New())))

	task, err := m.TaskStatus(ctx, &pb.ID{Id: request.GetTaskId()})
	if err != nil {
		log.G(m.ctx).Warn("could not fetch task history by deal", zap.Error(err))
		return err
	}

	named, err := reference.ParseNormalizedNamed(task.GetImageName())
	if err != nil {
		log.G(m.ctx).Warn("could not parse image to reference", zap.Error(err), zap.String("image", task.GetImageName()))
		return err
	}

	tagged, err := reference.WithTag(named, fmt.Sprintf("%s_%s", request.GetDealId(), request.GetTaskId()))
	if err != nil {
		log.G(m.ctx).Warn("could not tag image", zap.Error(err), zap.String("image", task.GetImageName()))
		return err
	}
	imageID := tagged.String()

	log.G(ctx).Debug("pulling image", zap.String("imageID", imageID))

	info, rd, err := m.ovs.Save(stream.Context(), imageID)
	if err != nil {
		return err
	}
	defer rd.Close()

	stream.SendHeader(metadata.Pairs("size", strconv.FormatInt(info.Size, 10)))

	streaming := true
	buf := make([]byte, 1*1024*1024)
	for streaming {
		n, err := rd.Read(buf)
		if err != nil {
			if err == io.EOF {
				streaming = false
			} else {
				return err
			}
		}
		if err := stream.Send(&pb.Chunk{Chunk: buf[:n]}); err != nil {
			return err
		}
	}

	return nil
}

func (m *Worker) StartTask(ctx context.Context, request *pb.StartTaskRequest) (*pb.StartTaskReply, error) {
	log.G(m.ctx).Info("handling StartTask request", zap.Any("request", request))

	spec := request.GetSpec()
	registry := spec.GetRegistry()
	image := spec.GetContainer().GetImage()
	allowed, reference, err := m.whitelist.Allowed(ctx, image, registry.Auth())
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")
	}

	taskID := uuid.New()

	dealID := request.GetDealID()
	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}

	cgroup, err := m.salesman.CGroup(ask.ID)
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(spec.Container.SshKey)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid public key provided %v", err)
	}
	if spec.GetResources() == nil {
		spec.Resources = &pb.AskPlanResources{}
	}
	if spec.GetResources().GetGPU() == nil {
		spec.Resources.GPU = ask.Resources.GPU
	}

	hasher := &pb.AskPlanHasher{AskPlanResources: ask.GetResources()}
	err = spec.GetResources().GetGPU().Normalize(hasher)
	if err != nil {
		log.G(ctx).Error("could not normalize GPU resources", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "could not normalize GPU resources: %s", err)
	}

	//TODO: generate ID
	if err := m.resources.ConsumeTask(ask.ID, taskID, spec.Resources); err != nil {
		return nil, fmt.Errorf("could not start task: %s", err)
	}

	// This can be canceled by using "resourceHandle.commit()".
	//defer resourceHandle.release()

	mounts := make([]volume.Mount, 0)
	for _, spec := range spec.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			m.resources.ReleaseTask(taskID)
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(spec.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification: %s", err)
	}
	gpuids, err := m.hardware.GPUIDs(spec.GetResources().GetGPU())
	if err != nil {
		log.G(ctx).Error("failed to fetch GPU IDs ", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to fetch GPU IDs: %s", err)
	}

	if len(spec.GetContainer().GetExpose()) > 0 {
		if !ask.GetResources().GetNetwork().GetNetFlags().GetIncoming() {
			m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
			return nil, fmt.Errorf("incoming network is required due to explicit `expose` settings, but not allowed for `%s` deal", dealID.Unwrap())
		}
	}

	var d = Description{
		Reference:     reference,
		Auth:          spec.Registry.Auth(),
		RestartPolicy: spec.Container.RestartPolicy.Unwrap(),
		CGroupParent:  cgroup.Suffix(),
		Resources:     spec.Resources,
		DealId:        request.GetDealID().Unwrap().String(),
		TaskId:        taskID,
		CommitOnStop:  spec.Container.CommitOnStop,
		GPUDevices:    gpuids,
		Env:           spec.Container.Env,
		volumes:       spec.Container.Volumes,
		mounts:        mounts,
		networks:      networks,
	}

	// TODO: Detect whether it's the first time allocation. If so - release resources on error.

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPOOLING}, taskID)
	log.G(m.ctx).Info("spooling an image")
	if err := m.ovs.Spool(ctx, d); err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPAWNING}, taskID)
	log.G(m.ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = reference.String()
	containerInfo.DealID = dealID.Unwrap().String()

	var reply = pb.StartTaskReply{
		Id:         taskID,
		PortMap:    make(map[string]*pb.Endpoints, 0),
		NetworkIDs: containerInfo.NetworkIDs,
	}

	for internalPort, portBindings := range containerInfo.Ports {
		if len(portBindings) < 1 {
			continue
		}

		var socketAddrs []*pb.SocketAddr
		var pubPortBindings []nat.PortBinding

		for _, portBinding := range portBindings {
			hostPort := portBinding.HostPort
			hostPortInt, err := nat.ParsePort(hostPort)
			if err != nil {
				m.resources.ReleaseTask(taskID)
				return nil, err
			}

			for _, publicIP := range m.publicIPs {
				socketAddrs = append(socketAddrs, &pb.SocketAddr{
					Addr: publicIP,
					Port: uint32(hostPortInt),
				})

				pubPortBindings = append(pubPortBindings, nat.PortBinding{HostIP: publicIP, HostPort: hostPort})
			}
		}

		containerInfo.Ports[internalPort] = pubPortBindings

		reply.PortMap[string(internalPort)] = &pb.Endpoints{Endpoints: socketAddrs}
	}

	m.saveContainerInfo(taskID, containerInfo)

	go m.listenForStatus(statusListener, taskID)

	return &reply, nil
}

// Stop request forces to kill container
func (m *Worker) StopTask(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(ctx).Info("handling Stop request", zap.Any("req", request))

	m.mu.Lock()
	containerInfo, ok := m.containers[request.Id]
	m.mu.Unlock()

	if !ok {
		return nil, status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}

	if err := m.ovs.Stop(ctx, containerInfo.ID); err != nil {
		log.G(ctx).Error("failed to Stop container", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)

		return nil, status.Errorf(codes.Internal, "failed to stop container %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_FINISHED}, request.Id)

	return &pb.Empty{}, nil
}

func (m *Worker) Tasks(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(m.ctx).Info("handling Tasks request")
	return &pb.TaskListReply{Info: m.CollectTasksStatuses()}, nil
}

func (m *Worker) CollectTasksStatuses(statuses ...pb.TaskStatusReply_Status) map[string]*pb.TaskStatusReply {
	result := map[string]*pb.TaskStatusReply{}
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, info := range m.containers {
		if len(statuses) > 0 {
			for _, s := range statuses {
				if s == info.status {
					result[id] = info.IntoProto(m.ctx)
					break
				}
			}
		} else {
			result[id] = info.IntoProto(m.ctx)
		}
	}
	return result
}

// TaskLogs returns logs from container
func (m *Worker) TaskLogs(request *pb.TaskLogsRequest, server pb.Worker_TaskLogsServer) error {
	log.G(m.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	if err := m.eventAuthorization.Authorize(server.Context(), auth.Event(taskAPIPrefix+"TaskLogs"), request); err != nil {
		return err
	}
	log.G(m.ctx).Info("handling TaskLogs request", zap.Any("request", request))
	cid, ok := m.getContainerIdByTaskId(request.Id)
	if !ok {
		return status.Errorf(codes.NotFound, "no job with id %s", request.Id)
	}
	opts := types.ContainerLogsOptions{
		ShowStdout: request.Type == pb.TaskLogsRequest_STDOUT || request.Type == pb.TaskLogsRequest_BOTH,
		ShowStderr: request.Type == pb.TaskLogsRequest_STDERR || request.Type == pb.TaskLogsRequest_BOTH,
		Since:      request.Since,
		Timestamps: request.AddTimestamps,
		Follow:     request.Follow,
		Tail:       request.Tail,
		Details:    request.Details,
	}
	reader, err := m.ovs.Logs(server.Context(), cid, opts)
	if err != nil {
		return err
	}
	defer reader.Close()
	buffer := make([]byte, 100*1024)
	for {
		readCnt, err := reader.Read(buffer)
		if readCnt != 0 {
			server.Send(&pb.TaskLogsChunk{Data: buffer[:readCnt]})
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

//TODO: proper request
func (m *Worker) JoinNetwork(ctx context.Context, request *pb.WorkerJoinNetworkRequest) (*pb.NetworkSpec, error) {
	spec, err := m.plugins.JoinNetwork(request.NetworkID)
	if err != nil {
		return nil, err
	}
	return &pb.NetworkSpec{
		Type:    spec.NetworkType(),
		Options: spec.NetworkOptions(),
		Subnet:  spec.NetworkCIDR(),
		Addr:    spec.NetworkAddr(),
	}, nil
}

func (m *Worker) TaskStatus(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
	log.G(m.ctx).Info("starting TaskDetails status server")

	info, ok := m.GetContainerInfo(req.GetId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no task with id %s", req.GetId())
	}

	var metric ContainerMetrics
	// If a container has been stoped, ovs.Info has no metrics for such container
	if info.status == pb.TaskStatusReply_RUNNING {
		metrics, err := m.ovs.Info(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot get container metrics: %s", err.Error())
		}

		metric, ok = metrics[info.ID]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Cannot get metrics for container %s", req.GetId())
		}
	}

	reply := info.IntoProto(m.ctx)
	reply.Usage = metric.Marshal()
	// todo: fill `reply.AllocatedResources` field.

	return reply, nil
}

func (m *Worker) RunSSH() error {
	return m.ssh.Run()
}

// RunBenchmarks perform benchmarking of Worker's resources.
func (m *Worker) runBenchmarks() error {
	savedHardware := m.storage.HardwareHash()
	exitingHardware := m.hardware.Hash()

	log.G(m.ctx).Debug("hardware hashes",
		zap.String("saved", savedHardware),
		zap.String("exiting", exitingHardware))

	savedBenchmarks := m.storage.PassedBenchmarks()
	//TODO(@antmat): use plain list of benchmarks when it will be available (via following PR)
	requiredBenchmarks := m.benchmarks.MapByDeviceType()

	hwHashesMatched := exitingHardware == savedHardware
	benchMatched := m.isBenchmarkListMatches(requiredBenchmarks, savedBenchmarks)

	log.G(m.ctx).Debug("state matching",
		zap.Bool("hwHashesMatched", hwHashesMatched),
		zap.Bool("benchMatched", benchMatched))

	if benchMatched && hwHashesMatched {
		log.G(m.ctx).Debug("benchmarks list is matched, hardware is not changed, skip benchmarking this worker")
		// return back previously measured results for hardware
		m.hardware = m.storage.HardwareWithBenchmarks()
		m.hardware.SetDevicesFromBenches()
		return nil
	}

	passedBenchmarks := map[uint64]bool{}
	for dev, benches := range requiredBenchmarks {
		err := m.runBenchmarkGroup(dev, benches)
		if err != nil {
			return err
		}

		for _, b := range benches {
			passedBenchmarks[b.GetID()] = true
		}
	}
	m.hardware.SetDevicesFromBenches()

	if err := m.storage.SetPassedBenchmarks(passedBenchmarks); err != nil {
		return err
	}

	if err := m.storage.SetHardwareWithBenchmarks(m.hardware); err != nil {
		return err
	}

	return m.storage.SetHardwareHash(m.hardware.Hash())
}

func (m *Worker) setupResources() error {
	m.resources = resource.NewScheduler(m.ctx, m.hardware)
	return nil
}

func (m *Worker) setupSalesman() error {
	salesman, err := salesman.NewSalesman(
		salesman.WithLogger(log.S(m.ctx).With("source", "salesman")),
		salesman.WithStorage(m.storage),
		salesman.WithResources(m.resources),
		salesman.WithHardware(m.hardware),
		salesman.WithEth(m.eth),
		salesman.WithCGroupManager(m.cGroupManager),
		salesman.WithMatcher(m.matcher),
		salesman.WithEthkey(m.key),
		salesman.WithConfig(&m.cfg.Salesman),
	)
	if err != nil {
		return err
	}
	m.salesman = salesman

	ch := m.salesman.Run(m.ctx)
	go m.listenDeals(ch)
	return nil
}

func (m *Worker) setupServer() error {
	logger := log.GetLogger(m.ctx)
	grpcServer := xgrpc.NewServer(logger,
		xgrpc.Credentials(m.creds),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.AuthorizationInterceptor(m.eventAuthorization),
		xgrpc.VerifyInterceptor(),
	)
	m.externalGrpc = grpcServer

	pb.RegisterWorkerServer(grpcServer, m)
	pb.RegisterWorkerManagementServer(grpcServer, m)
	grpc_prometheus.Register(grpcServer)

	if m.cfg.Debug != nil {
		go debug.ServePProf(m.ctx, *m.cfg.Debug, log.G(m.ctx))
	}

	return nil
}

// isBenchmarkListMatches checks if already passed benchmarks is matches required benchmarks list.
//
// todo: test me
func (m *Worker) isBenchmarkListMatches(required map[pb.DeviceType][]*pb.Benchmark, exiting map[uint64]bool) bool {
	for _, benchs := range required {
		for _, bench := range benchs {
			if _, ok := exiting[bench.ID]; !ok {
				return false
			}
		}
	}

	return true
}

// runBenchmarkGroup executes group of benchmarks for given device type (CPU, GPU, Network, etc...).
// The results must be attached to worker's hardware capabilities inside this function (by magic).
func (m *Worker) runBenchmarkGroup(dev pb.DeviceType, benches []*pb.Benchmark) error {
	var hwBenches []map[uint64]*pb.Benchmark
	var gpuDevices []*pb.GPUDevice
	switch dev {
	case pb.DeviceType_DEV_CPU:
		hwBenches = []map[uint64]*pb.Benchmark{m.hardware.CPU.Benchmarks}
	case pb.DeviceType_DEV_RAM:
		hwBenches = []map[uint64]*pb.Benchmark{m.hardware.RAM.Benchmarks}
	case pb.DeviceType_DEV_GPU:
		for _, gpu := range m.hardware.GPU {
			hwBenches = append(hwBenches, gpu.Benchmarks)
			gpuDevices = append(gpuDevices, gpu.Device)
		}
	case pb.DeviceType_DEV_NETWORK_IN:
		hwBenches = []map[uint64]*pb.Benchmark{m.hardware.Network.BenchmarksIn}
	case pb.DeviceType_DEV_NETWORK_OUT:
		hwBenches = []map[uint64]*pb.Benchmark{m.hardware.Network.BenchmarksOut}
	case pb.DeviceType_DEV_STORAGE:
		hwBenches = []map[uint64]*pb.Benchmark{m.hardware.Storage.Benchmarks}
	default:
		return fmt.Errorf("unknown benchmark group \"%s\"", dev.String())
	}

	for _, bench := range benches {
		for idx, receiver := range hwBenches {
			if bench.GetID() == bm.CPUCores {
				bench.Result = uint64(m.hardware.CPU.Device.Cores)
			} else if bench.GetID() == bm.RamSize {
				bench.Result = m.hardware.RAM.Device.Total
			} else if bench.GetID() == bm.GPUCount {
				//GPU count is always 1 for each GPU device.
				bench.Result = uint64(1)
			} else if bench.GetID() == bm.GPUMem {
				bench.Result = gpuDevices[idx].Memory
			} else if bench.GetID() == bm.StorageSize {
				freeDiskSpace, err := disk.FreeDiskSpace(m.ctx)
				if err != nil {
					return err
				}
				bench.Result = freeDiskSpace
			} else if len(bench.GetImage()) != 0 {
				d, err := getDescriptionForBenchmark(bench)
				if err != nil {
					return fmt.Errorf("could not create description for benchmark: %s", err)
				}
				d.Env[bm.CPUCountBenchParam] = fmt.Sprintf("%d", m.hardware.CPU.Device.Cores)

				if gpuDevices != nil {
					gpuDev := gpuDevices[idx]
					d.Env[bm.GPUVendorParam] = gpuDev.VendorType().String()
					d.GPUDevices = []gpu.GPUID{gpu.GPUID(gpuDev.GetID())}
				}
				res, err := m.execBenchmarkContainer(bench, d)
				if err != nil {
					return err
				}
				bench.Result = res.Result
			} else {
				log.S(m.ctx).Warnf("skipping benchmark %s (setting explicitly to 0)", bench.Code)
				bench.Result = uint64(0)
			}
			receiver[bench.GetID()] = deepcopy.Copy(bench).(*pb.Benchmark)
		}
	}
	return nil
}

// execBenchmarkContainerWithResults executes benchmark as docker image,
// returns JSON output with measured values.
func (m *Worker) execBenchmarkContainerWithResults(d Description) (map[string]*bm.ResultJSON, error) {
	err := m.ovs.Spool(m.ctx, d)
	if err != nil {
		return nil, err
	}

	statusChan, statusReply, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		return nil, fmt.Errorf("cannot start container with benchmark: %v", err)
	}

	logOpts := types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
		Since:      strconv.FormatInt(time.Now().Unix(), 10),
	}

	reader, err := m.ovs.Logs(m.ctx, statusReply.ID, logOpts)
	if err != nil {
		return nil, fmt.Errorf("cannot create container log reader: %v", err)
	}
	defer reader.Close()

	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}

	select {
	case s := <-statusChan:
		if s == pb.TaskStatusReply_FINISHED || s == pb.TaskStatusReply_BROKEN {
			if _, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, reader); err != nil {
				return nil, fmt.Errorf("cannot read logs into buffer: %v", err)
			}
		}
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}

	resultsMap, err := parseBenchmarkResult(stdoutBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("cannot parse benchmark result: %v", err)
	}

	m.ovs.OnDealFinish(m.ctx, statusReply.ID)
	return resultsMap, nil
}

func (m *Worker) execBenchmarkContainer(ben *pb.Benchmark, des Description) (*bm.ResultJSON, error) {
	log.G(m.ctx).Debug("starting containered benchmark", zap.Any("benchmark", ben))
	res, err := m.execBenchmarkContainerWithResults(des)
	if err != nil {
		return nil, err
	}

	log.G(m.ctx).Debug("received raw benchmark results",
		zap.Uint64("bench_id", ben.GetID()),
		zap.Any("result", res))

	v, ok := res[ben.GetCode()]
	if !ok {
		return nil, fmt.Errorf("no results for benchmark id=%v found into container's output", ben.GetID())
	}

	return v, nil
}

func parseBenchmarkResult(data []byte) (map[string]*bm.ResultJSON, error) {
	v := &bm.ContainerBenchmarkResultsJSON{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return nil, err
	}

	if len(v.Results) == 0 {
		return nil, errors.New("results is empty")
	}

	return v.Results, nil
}

func getDescriptionForBenchmark(b *pb.Benchmark) (Description, error) {
	reference, err := reference.ParseNormalizedNamed(b.GetImage())
	if err != nil {
		return Description{}, err
	}
	return Description{
		autoremove: true,
		Reference:  reference,
		Env: map[string]string{
			bm.BenchIDEnvParamName: fmt.Sprintf("%d", b.GetID()),
		},
	}, nil
}

func (m *Worker) AskPlans(ctx context.Context, _ *pb.Empty) (*pb.AskPlansReply, error) {
	log.G(m.ctx).Info("handling AskPlans request")
	return &pb.AskPlansReply{AskPlans: m.salesman.AskPlans()}, nil
}

func (m *Worker) CreateAskPlan(ctx context.Context, request *pb.AskPlan) (*pb.ID, error) {
	log.G(m.ctx).Info("handling CreateAskPlan request", zap.Any("request", request))
	if len(request.GetID()) != 0 || !request.GetOrderID().IsZero() || !request.GetDealID().IsZero() {
		return nil, errors.New("creating ask plans with predefined id, order_id or deal_id are not supported")
	}
	id, err := m.salesman.CreateAskPlan(request)
	if err != nil {
		return nil, err
	}

	return &pb.ID{Id: id}, nil
}

func (m *Worker) RemoveAskPlan(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(m.ctx).Info("handling RemoveAskPlan request", zap.String("id", request.GetId()))

	if err := m.salesman.RemoveAskPlan(request.GetId()); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (m *Worker) PurgeAskPlans(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	plans := m.salesman.AskPlans()

	result := multierror.NewMultiError()
	for id := range plans {
		err := m.salesman.RemoveAskPlan(id)
		result = multierror.Append(result, err)
	}

	if result.ErrorOrNil() != nil {
		return nil, result.ErrorOrNil()
	}

	return &pb.Empty{}, nil
}

func (m *Worker) GetDealInfo(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	log.G(m.ctx).Info("handling GetDealInfo request")

	dealID, err := pb.NewBigIntFromString(id.Id)
	if err != nil {
		return nil, err
	}
	return m.getDealInfo(dealID)
}

func (m *Worker) getDealInfo(dealID *pb.BigInt) (*pb.DealInfoReply, error) {
	deal, err := m.salesman.Deal(dealID)
	if err != nil {
		return nil, err
	}

	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}
	resources := ask.GetResources()

	running := map[string]*pb.TaskStatusReply{}
	completed := map[string]*pb.TaskStatusReply{}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id, c := range m.containers {
		// task is ours
		if c.DealID == dealID.Unwrap().String() {
			task := c.IntoProto(m.ctx)

			// task is running or preparing to start
			if c.status == pb.TaskStatusReply_SPOOLING ||
				c.status == pb.TaskStatusReply_SPAWNING ||
				c.status == pb.TaskStatusReply_RUNNING {
				running[id] = task
			} else {
				completed[id] = task
			}
		}
	}

	return &pb.DealInfoReply{
		Deal:      deal,
		Running:   running,
		Completed: completed,
		Resources: resources,
	}, nil
}

func (m *Worker) AskPlanByTaskID(taskID string) (*pb.AskPlan, error) {
	planID, err := m.resources.AskPlanIDByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	return m.salesman.AskPlan(planID)
}

// todo: make the `worker.Init() error` method to kickstart all initial jobs for the Worker instance.
// (state loading, benchmarking, market sync).

// Close disposes all resources related to the Worker
func (m *Worker) Close() {
	log.G(m.ctx).Info("closing worker")

	if m.ssh != nil {
		m.ssh.Close()
	}
	if m.ovs != nil {
		m.ovs.Close()
	}
	if m.salesman != nil {
		m.salesman.Close()
	}
	if m.plugins != nil {
		m.plugins.Close()
	}
	if m.externalGrpc != nil {
		m.externalGrpc.Stop()
	}
	if m.certRotator != nil {
		m.certRotator.Close()
	}
}
