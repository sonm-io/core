package miner

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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/hashicorp/go-multierror"
	"github.com/mohae/deepcopy"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
	"github.com/sonm-io/core/insonmnia/miner/salesman"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"

	// todo: drop alias
	bm "github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
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
	taskAPIPrefix   = "/sonm.Hub/"
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
	}
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	*options

	mu        sync.Mutex
	hardware  *hardware.Hardware
	resources *resource.Scheduler
	salesman  *salesman.Salesman

	eventAuthorization *auth.AuthRouter

	// One-to-one mapping between container IDs and userland task names.
	//
	// The overseer operates with containers in terms of their ID, which does not change even during auto-restart.
	// However some requests pass an application (or task) name, which is more meaningful for user. To be able to
	// transform between these two identifiers this map exists.
	//
	// WARNING: This must be protected using `mu`.
	//
	// fixme: only write and delete on this struct, looks like we can
	// safety removes them.
	nameMapping map[string]string

	// Maps StartRequest's IDs to containers' IDs
	// TODO: It's doubtful that we should keep this map here instead in the Overseer.
	containers map[string]*ContainerInfo

	controlGroup  cgroups.CGroup
	cGroupManager cgroups.CGroupManager
	externalGrpc  *grpc.Server
	startTime     time.Time
}

func NewMiner(opts ...Option) (m *Miner, err error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	m = &Miner{
		options:     o,
		containers:  make(map[string]*ContainerInfo),
		nameMapping: make(map[string]string),
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

// Serve starts handling incoming API gRPC request and communicates
// with miners
func (m *Miner) Serve() error {
	m.startTime = time.Now()
	if err := m.waitMasterApproved(); err != nil {
		return err
	}

	listener, err := npp.NewListener(m.ctx, m.cfg.Endpoint,
		npp.WithRendezvous(m.cfg.NPP.Rendezvous.Endpoints, m.creds),
		npp.WithRelay(m.cfg.NPP.Relay.Endpoints, m.key),
		npp.WithLogger(log.G(m.ctx)),
	)
	if err != nil {
		log.G(m.ctx).Error("failed to listen", zap.String("address", m.cfg.Endpoint), zap.Error(err))
		return err
	}
	log.G(m.ctx).Info("listening for gRPC API connections", zap.Stringer("address", listener.Addr()))
	err = m.externalGrpc.Serve(listener)
	m.Close()

	return err
}

func (m *Miner) waitMasterApproved() error {
	if m.cfg.Master == nil {
		return nil
	}
	log.S(m.ctx).Info("waiting for master approval...")
	selfAddr := m.ethAddr().Hex()
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
				log.S(m.ctx).Info("still no approval, continue waiting")
				continue
			}
			if curMaster != expectedMaster {
				return fmt.Errorf("received unexpected master %s", curMaster)
			}
			return nil
		}
	}
}

func (m *Miner) ethAddr() common.Address {
	return util.PubKeyToAddr(m.key.PublicKey)
}

func (m *Miner) setupMaster() error {
	if m.cfg.Master != nil {
		log.S(m.ctx).Info("checking current master")
		addr, err := m.eth.Market().GetMaster(m.ctx, m.ethAddr())
		if err != nil {
			return err
		}
		if addr.Big().Cmp(m.ethAddr().Big()) == 0 {
			log.S(m.ctx).Infof("master is not set, sending request to %s", m.cfg.Master.Hex())
			err = <-m.eth.Market().RegisterWorker(m.ctx, m.key, *m.cfg.Master)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Miner) setupAuthorization() error {
	authorization := auth.NewEventAuthorization(m.ctx,
		auth.WithLog(log.G(m.ctx)),
		// Note: need to refactor auth router to support multiple prefixes for methods.
		// auth.WithEventPrefix(hubAPIPrefix),
		auth.Allow(workerManagementMethods...).With(auth.NewTransportAuthorization(m.ethAddr())),

		auth.Allow(taskAPIPrefix+"TaskStatus").With(newMultiAuth(
			auth.NewTransportAuthorization(m.ethAddr()),
			newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m)),
		)),
		auth.Allow(taskAPIPrefix+"StopTask").With(newDealAuthorization(m.ctx, m, newFromTaskDealExtractor(m))),
		auth.Allow(taskAPIPrefix+"JoinNetwork").With(newDealAuthorization(m.ctx, m, newFromNamedTaskDealExtractor(m, "TaskID"))),
		auth.Allow(taskAPIPrefix+"StartTask").With(newDealAuthorization(m.ctx, m, newFieldDealExtractor())),
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

func (m *Miner) setupControlGroup() error {
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

func (m *Miner) setupHardware() error {
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
	hardwareInfo.Network.Outbound = true
	m.hardware = hardwareInfo
	return nil
}

func (m *Miner) listenDeals(dealsCh <-chan *pb.Deal) {
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

func (m *Miner) cancelDealTasks(deal *pb.Deal) error {
	dealID := deal.GetId().Unwrap().String()
	var toDelete []string

	m.mu.Lock()
	for _, container := range m.containers {
		if container.DealID == dealID {
			toDelete = append(toDelete, container.ID)
		}
	}
	m.mu.Unlock()

	var result error
	for _, containerID := range toDelete {
		if err := m.ovs.Stop(m.ctx, containerID); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (m *Miner) saveContainerInfo(id string, info ContainerInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nameMapping[info.ID] = id
	m.containers[id] = &info
}

func (m *Miner) GetContainerInfo(id string) (*ContainerInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.containers[id]
	return info, ok
}

func (m *Miner) getContainerIdByTaskId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info, ok := m.containers[id]
	if ok {
		return info.ID, ok
	}
	return "", ok
}

func (m *Miner) deleteTaskMapping(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nameMapping, id)
}

func (m *Miner) Devices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	return m.hardware.IntoProto(), nil
}

// Status returns internal hub statistic
func (m *Miner) Status(ctx context.Context, _ *pb.Empty) (*pb.HubStatusReply, error) {
	uptime := uint64(time.Now().Sub(m.startTime).Seconds())
	reply := &pb.HubStatusReply{
		Uptime:    uptime,
		Platform:  util.GetPlatformName(),
		Version:   m.version,
		EthAddr:   m.ethAddr().Hex(),
		TaskCount: uint32(len(m.CollectTasksStatuses(pb.TaskStatusReply_RUNNING))),
	}

	return reply, nil
}

// FreeDevice provides information about unallocated resources
// that can be turned into ask-plans.
func (m *Miner) FreeDevices(ctx context.Context, request *pb.Empty) (*pb.DevicesReply, error) {
	// todo: this is stub, wait for Resource manager impl to use real data.
	return m.hardware.IntoProto(), nil
}

func (m *Miner) scheduleStatusPurge(id string) {
	t := time.NewTimer(time.Second * 3600)
	defer t.Stop()
	select {
	case <-t.C:
		m.mu.Lock()
		delete(m.containers, id)
		m.mu.Unlock()
	case <-m.ctx.Done():
		return
	}
}

func (m *Miner) setStatus(status *pb.TaskStatusReply, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.containers[id]
	if !ok {
		m.containers[id] = &ContainerInfo{}
	}

	m.containers[id].status = status.GetStatus()
	if status.Status == pb.TaskStatusReply_BROKEN || status.Status == pb.TaskStatusReply_FINISHED {
		m.resources.ReleaseTask(id)
		go m.scheduleStatusPurge(id)
	}
}

func (m *Miner) listenForStatus(statusListener chan pb.TaskStatusReply_Status, id string) {
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

func transformRestartPolicy(p *pb.ContainerRestartPolicy) container.RestartPolicy {
	var restartPolicy = container.RestartPolicy{}
	if p != nil {
		restartPolicy.Name = p.Name
		restartPolicy.MaximumRetryCount = int(p.MaximumRetryCount)
	}

	return restartPolicy
}

func (m *Miner) PushTask(stream pb.Hub_PushTaskServer) error {
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

func (m *Miner) PullTask(request *pb.PullTaskRequest, stream pb.Hub_PullTaskServer) error {
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

func (m *Miner) StartTask(ctx context.Context, request *pb.StartTaskRequest) (*pb.StartTaskReply, error) {
	log.G(m.ctx).Info("handling StartTask request", zap.Any("request", request))

	// TODO: get rid of this wrapper - just add validate method
	taskRequest, err := structs.NewStartTaskRequest(request)
	if err != nil {
		return nil, err
	}

	allowed, ref, err := m.whitelist.Allowed(ctx, request.Container.Registry, request.Container.Image, request.Container.Auth)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, status.Errorf(codes.PermissionDenied, "specified image is forbidden to run")
	}

	// TODO(sshaman1101): REFACTOR:   only check for whitelist there,
	// TODO(sshaman1101): REFACTOR:   move all deals and tasks related code into the Worker.

	container := request.Container
	container.Registry = reference.Domain(ref)
	container.Image = reference.Path(ref)

	return m.startTask(ctx, taskRequest)
}

// Start request from Hub makes Miner start a container
func (m *Miner) startTask(ctx context.Context, request *structs.StartTaskRequest) (*pb.StartTaskReply, error) {
	log.G(m.ctx).Info("handling Start request", zap.Any("request", request))

	taskID := uuid.New()

	//TODO: move to validate()
	if request.GetContainer() == nil {
		return nil, fmt.Errorf("container field is required")
	}

	dealID, err := pb.NewBigIntFromString(request.GetDealId())
	if err != nil {
		return nil, fmt.Errorf("could not parse deal id as big int: %s", err)
	}
	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}

	cgroup, err := m.salesman.CGroup(ask.ID)
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(request.Container.PublicKeyData)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid public key provided %v", err)
	}
	if request.GetResources() == nil {
		request.Resources = &pb.AskPlanResources{}
	}
	if request.GetResources().GetGPU() == nil {
		request.Resources.GPU = ask.Resources.GPU
	}

	err = request.GetResources().GetGPU().Normalize(m.hardware)
	if err != nil {
		log.G(ctx).Error("could not normalize GPU resources", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "could not normalize GPU resources: %s", err)
	}

	//TODO: generate ID
	if err := m.resources.ConsumeTask(ask.ID, taskID, request.Resources); err != nil {
		return nil, fmt.Errorf("could not start task: %s", err)
	}

	// This can be canceled by using "resourceHandle.commit()".
	//defer resourceHandle.release()

	mounts := make([]volume.Mount, 0)
	for _, spec := range request.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			m.resources.ReleaseTask(taskID)
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(request.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification: %s", err)
	}
	gpuids, err := m.hardware.GPUIDs(request.GetResources().GetGPU())
	if err != nil {
		log.G(ctx).Error("failed to fetch GPU IDs ", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to fetch GPU IDs: %s", err)
	}
	var d = Description{
		Image:         request.Container.Image,
		Registry:      request.Container.Registry,
		Auth:          request.Container.Auth,
		RestartPolicy: transformRestartPolicy(nil),
		CGroupParent:  cgroup.Suffix(),
		Resources:     request.Resources,
		DealId:        request.GetDealId(),
		TaskId:        taskID,
		CommitOnStop:  request.Container.CommitOnStop,
		GPUDevices:    gpuids,
		Env:           request.Container.Env,
		volumes:       request.Container.Volumes,
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
	log.G(ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, taskID)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = d.Image
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
func (m *Miner) StopTask(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(ctx).Info("handling Stop request", zap.Any("req", request))

	m.mu.Lock()
	containerInfo, ok := m.containers[request.Id]
	m.mu.Unlock()

	m.deleteTaskMapping(request.Id)

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

func (m *Miner) Tasks(ctx context.Context, request *pb.Empty) (*pb.TaskListReply, error) {
	log.G(m.ctx).Info("handling Tasks request")
	return &pb.TaskListReply{Info: m.CollectTasksStatuses()}, nil
}

func (m *Miner) CollectTasksStatuses(statuses ...pb.TaskStatusReply_Status) map[string]*pb.TaskStatusReply {
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
func (m *Miner) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
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
func (m *Miner) JoinNetwork(ctx context.Context, request *pb.HubJoinNetworkRequest) (*pb.NetworkSpec, error) {
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

func (m *Miner) TaskStatus(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
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

func (m *Miner) RunSSH() error {
	return m.ssh.Run()
}

// RunBenchmarks perform benchmarking of Worker's resources.
func (m *Miner) runBenchmarks() error {
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

	if err := m.storage.SetPassedBenchmarks(passedBenchmarks); err != nil {
		return err
	}

	if err := m.storage.SetHardwareWithBenchmarks(m.hardware); err != nil {
		return err
	}

	return m.storage.SetHardwareHash(m.hardware.Hash())
}

func (m *Miner) setupResources() error {
	m.resources = resource.NewScheduler(m.ctx, m.hardware)
	return nil
}

func (m *Miner) setupSalesman() error {
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

func (m *Miner) setupServer() error {
	logger := log.GetLogger(m.ctx)
	grpcServer := xgrpc.NewServer(logger,
		xgrpc.Credentials(m.creds),
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.AuthorizationInterceptor(m.eventAuthorization),
		xgrpc.VerifyInterceptor(),
	)
	m.externalGrpc = grpcServer

	pb.RegisterHubServer(grpcServer, m)
	pb.RegisterWorkerManagementServer(grpcServer, m)
	grpc_prometheus.Register(grpcServer)
	return nil
}

// isBenchmarkListMatches checks if already passed benchmarks is matches required benchmarks list.
//
// todo: test me
func (m *Miner) isBenchmarkListMatches(required map[pb.DeviceType][]*pb.Benchmark, exiting map[uint64]bool) bool {
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
func (m *Miner) runBenchmarkGroup(dev pb.DeviceType, benches []*pb.Benchmark) error {
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
			} else if len(bench.GetImage()) != 0 {
				d := getDescriptionForBenchmark(bench)
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
func (m *Miner) execBenchmarkContainerWithResults(d Description) (map[string]*bm.ResultJSON, error) {
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

	return resultsMap, nil
}

func (m *Miner) execBenchmarkContainer(ben *pb.Benchmark, des Description) (*bm.ResultJSON, error) {
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

func getDescriptionForBenchmark(b *pb.Benchmark) Description {
	return Description{
		autoremove: true,
		Image:      b.GetImage(),
		Env: map[string]string{
			bm.BenchIDEnvParamName: fmt.Sprintf("%d", b.GetID()),
		},
	}
}

func (m *Miner) AskPlans(ctx context.Context, _ *pb.Empty) (*pb.AskPlansReply, error) {
	log.G(m.ctx).Info("handling AskPlans request")
	return &pb.AskPlansReply{AskPlans: m.salesman.AskPlans()}, nil
}

func (m *Miner) CreateAskPlan(ctx context.Context, request *pb.AskPlan) (*pb.ID, error) {
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

func (m *Miner) RemoveAskPlan(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
	log.G(m.ctx).Info("handling RemoveAskPlan request", zap.String("id", request.GetId()))

	if err := m.salesman.RemoveAskPlan(request.GetId()); err != nil {
		return nil, err
	}
	return &pb.Empty{}, nil
}

func (m *Miner) GetDealInfo(ctx context.Context, id *pb.ID) (*pb.DealInfoReply, error) {
	log.G(m.ctx).Info("handling GetDealInfo request")

	dealID, err := pb.NewBigIntFromString(id.Id)
	if err != nil {
		return nil, err
	}
	return m.getDealInfo(dealID)
}

func (m *Miner) getDealInfo(dealID *pb.BigInt) (*pb.DealInfoReply, error) {
	deal, err := m.salesman.Deal(dealID)
	if err != nil {
		return nil, err
	}

	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}
	resources := ask.GetResources()

	running := &pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{},
	}

	completed := &pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{},
	}

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
				running.Statuses[id] = task
			} else {
				completed.Statuses[id] = task
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

func (m *Miner) AskPlanByTaskID(taskID string) (*pb.AskPlan, error) {
	planID, err := m.resources.AskPlanIDByTaskID(taskID)
	if err != nil {
		return nil, err
	}
	return m.salesman.AskPlan(planID)
}

// todo: make the `miner.Init() error` method to kickstart all initial jobs for the Worker instance.
// (state loading, benchmarking, market sync).

// Close disposes all resources related to the Worker
func (m *Miner) Close() {
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
