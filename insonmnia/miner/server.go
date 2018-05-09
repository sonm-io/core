package miner

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/mohae/deepcopy"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"

	// todo: drop alias
	bm "github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/insonmnia/resource"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Miner holds information about jobs, make orders to Observer and communicates with Hub
type Miner struct {
	ctx       context.Context
	cfg       *Config
	mu        sync.Mutex
	ovs       Overseer
	plugins   *plugin.Repository
	hardware  *hardware.Hardware
	salesman  *Salesman
	resources *resource.Scheduler
	ethkey    *ecdsa.PrivateKey
	publicIPs []string
	eth       blockchain.API
	dwh       pb.DWHClient

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
	ssh           SSH

	// external and in-mem storage
	state         *state.Storage
	benchmarkList bm.BenchList
}

func NewMiner(cfg *Config, opts ...Option) (m *Miner, err error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	if cfg == nil {
		return nil, errors.New("config is mandatory for MinerBuilder")
	}

	if o.key == nil {
		return nil, errors.New("private key is mandatory")
	}

	if o.storage == nil {
		return nil, errors.New("state storage is mandatory")
	}

	if o.ctx == nil {
		o.ctx = context.Background()
	}

	if o.eth == nil {
		eth, err := blockchain.NewAPI(blockchain.WithConfig(cfg.Blockchain))
		if err != nil {
			return nil, err
		}

		o.eth = eth
	}

	if o.benchList == nil {
		o.benchList, err = bm.NewBenchmarksList(o.ctx, cfg.Benchmarks)
		if err != nil {
			return nil, err
		}
	}

	if o.dwh == nil {
		if o.creds == nil {
			_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), o.key)
			if err != nil {
				return nil, err
			}
			o.creds = util.NewTLS(TLSConfig)
		}
		cc, err := xgrpc.NewClient(o.ctx, cfg.DWHEndpoint, o.creds)
		if err != nil {
			return nil, err
		}
		o.dwh = pb.NewDWHClient(cc)
	}

	cgName := "sonm-worker-parent"
	cgRes := &specs.LinuxResources{}
	if cfg.Resources != nil {
		cgName = cfg.Resources.Cgroup
		cgRes = cfg.Resources.Resources
	}

	cgroup, cGroupManager, err := cgroups.NewCgroupManager(cgName, cgRes)
	if err != nil {
		return nil, err
	}

	hardwareInfo, err := hardware.NewHardware()
	if err != nil {
		return nil, err
	}

	hardwareInfo.RAM.Device.Available = hardwareInfo.RAM.Device.Total
	// check if memory is limited into cgroup
	if s, err := cgroup.Stats(); err == nil {
		if s.MemoryLimit != 0 && s.MemoryLimit < hardwareInfo.RAM.Device.Total {
			hardwareInfo.RAM.Device.Available = s.MemoryLimit
		}
	}

	if err := o.setupNetworkOptions(cfg); err != nil {
		return nil, errors.Wrap(err, "failed to set up network options")
	}

	log.G(o.ctx).Info("discovered public IPs", zap.Any("public IPs", o.publicIPs))

	plugins, err := plugin.NewRepository(o.ctx, cfg.Plugins)
	if err != nil {
		return nil, err
	}

	// apply info about GPUs, expose to logs
	plugins.ApplyHardwareInfo(hardwareInfo)
	hardwareInfo.SetNetworkIncoming(o.publicIPs)
	//TODO: configurable?
	hardwareInfo.Network.Outbound = true

	log.G(o.ctx).Info("collected hardware info", zap.Any("hw", hardwareInfo))

	if o.ssh == nil {
		o.ssh = nilSSH{}
	}

	if o.ovs == nil {
		o.ovs, err = NewOverseer(o.ctx, plugins)
		if err != nil {
			return nil, err
		}
	}

	m = &Miner{
		ctx:    o.ctx,
		cfg:    cfg,
		ovs:    o.ovs,
		ethkey: o.key,

		plugins: plugins,

		hardware:  hardwareInfo,
		resources: nil,
		salesman:  nil,
		publicIPs: o.publicIPs,

		containers:  make(map[string]*ContainerInfo),
		nameMapping: make(map[string]string),

		controlGroup:  cgroup,
		cGroupManager: cGroupManager,
		ssh:           o.ssh,
		state:         o.storage,
		benchmarkList: o.benchList,
		eth:           o.eth,
		dwh:           o.dwh,
	}

	if err := m.RunBenchmarks(); err != nil {
		return nil, err
	}

	matcher := matcher.NewMatcher(&matcher.Config{
		Key: m.ethkey,
		//TODO: make configurable
		PollDelay: time.Second * 5,
		DWH:       m.dwh,
		Eth:       m.eth,
		//TODO: make configurable
		QueryLimit: 100,
	})

	//TODO: this is racy, because of post initialization of hardware via benchmarks
	m.resources = resource.NewScheduler(o.ctx, m.hardware)

	salesman, err := NewSalesman(o.ctx, o.storage, m.resources, m.hardware, o.eth, m.cGroupManager, matcher, o.key)
	if err != nil {
		return nil, err
	}
	m.salesman = salesman

	m.salesman.Run(o.ctx)

	return m, nil
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

// Hardware returns Worker's hardware capabilities
func (m *Miner) Hardware() *hardware.Hardware {
	return m.hardware
}

// FreeDevice provides information about unallocated resources
// that can be turned into ask-plans.
func (m *Miner) FreeDevice() *hardware.Hardware {
	// todo: this is stub, wait for Resource manager impl to use real data.
	return m.hardware
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

func (m *Miner) Load(stream pb.Hub_PushTaskServer) error {
	log.G(m.ctx).Info("handling Load request")

	result, err := m.ovs.Load(stream.Context(), newChunkReader(stream))
	if err != nil {
		return err
	}

	log.G(m.ctx).Info("image loaded, set trailer", zap.String("trailer", result.Status))
	stream.SetTrailer(metadata.Pairs("status", result.Status))
	return nil
}

func (m *Miner) Save(request *pb.SaveRequest, stream pb.Hub_PullTaskServer) error {
	log.G(m.ctx).Info("handling Save request", zap.Any("request", request))

	info, rd, err := m.ovs.Save(stream.Context(), request.ImageID)
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

// Start request from Hub makes Miner start a container
func (m *Miner) Start(ctx context.Context, request *pb.MinerStartRequest) (*pb.MinerStartReply, error) {
	log.G(m.ctx).Info("handling Start request", zap.Any("request", request))

	//TODO: move to validate()
	if request.GetContainer() == nil {
		return nil, fmt.Errorf("container field is required")
	}

	dealID, err := pb.NewBigIntFromString(request.DealID)
	if err != nil {
		return nil, fmt.Errorf("could not parse deal id as big int - %s", err)
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
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		m.resources.ReleaseTask(request.Id)
		return nil, status.Errorf(codes.Internal, "could not normalize GPU resources - %s", err)
	}

	//TODO: generate ID
	if err := m.resources.ConsumeTask(ask.ID, request.Id, request.Resources); err != nil {
		return nil, fmt.Errorf("could not start task - %s", err)
	}

	// This can be canceled by using "resourceHandle.commit()".
	//defer resourceHandle.release()

	mounts := make([]volume.Mount, 0)
	for _, spec := range request.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			m.resources.ReleaseTask(request.Id)
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(request.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		m.resources.ReleaseTask(request.Id)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification - %s", err)
	}
	gpuids, err := m.hardware.GPUIDs(request.GetResources().GetGPU())
	if err != nil {
		log.G(ctx).Error("failed to fetch GPU IDs ", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		m.resources.ReleaseTask(request.Id)
		return nil, status.Errorf(codes.Internal, "failed to fetch GPU IDs - %s", err)
	}
	var d = Description{
		Image:         request.Container.Image,
		Registry:      request.Container.Registry,
		Auth:          request.Container.Auth,
		RestartPolicy: transformRestartPolicy(request.RestartPolicy),
		CGroupParent:  cgroup.Suffix(),
		Resources:     request.Resources,
		DealId:        request.GetDealID(),
		TaskId:        request.Id,
		CommitOnStop:  request.Container.CommitOnStop,
		GPUDevices:    gpuids,
		Env:           request.Container.Env,
		volumes:       request.Container.Volumes,
		mounts:        mounts,
		networks:      networks,
	}

	// TODO: Detect whether it's the first time allocation. If so - release resources on error.

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPOOLING}, request.Id)
	log.G(m.ctx).Info("spooling an image")
	if err := m.ovs.Spool(ctx, d); err != nil {
		log.G(ctx).Error("failed to Spool an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		m.resources.ReleaseTask(request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPAWNING}, request.Id)
	log.G(ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		m.resources.ReleaseTask(request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = d.Image
	containerInfo.DealID = dealID.Unwrap().String()

	var reply = pb.MinerStartReply{
		Container:  containerInfo.ID,
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
				m.resources.ReleaseTask(request.Id)
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

	m.saveContainerInfo(request.Id, containerInfo)

	go m.listenForStatus(statusListener, request.Id)

	return &reply, nil
}

// Stop request forces to kill container
func (m *Miner) Stop(ctx context.Context, request *pb.ID) (*pb.Empty, error) {
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
	m.resources.ReleaseTask(request.Id)

	return &pb.Empty{}, nil
}

func (m *Miner) CollectTasksStatuses() map[string]*pb.TaskStatusReply {
	result := map[string]*pb.TaskStatusReply{}
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, info := range m.containers {
		result[id] = info.IntoProto()
	}
	return result
}

// TaskLogs returns logs from container
func (m *Miner) TaskLogs(request *pb.TaskLogsRequest, server pb.Hub_TaskLogsServer) error {
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
func (m *Miner) JoinNetwork(ctx context.Context, req *pb.ID) (*pb.NetworkSpec, error) {
	spec, err := m.plugins.JoinNetwork(req.Id)
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

func (m *Miner) TaskDetails(ctx context.Context, req *pb.ID) (*pb.TaskStatusReply, error) {
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

	reply := info.IntoProto()
	reply.Usage = metric.Marshal()
	// todo: fill `reply.AllocatedResources` field.

	return reply, nil
}

func (m *Miner) RunSSH() error {
	return m.ssh.Run()
}

// RunBenchmarks perform benchmarking of Worker's resources.
func (m *Miner) RunBenchmarks() error {
	savedHardware := m.state.HardwareHash()
	exitingHardware := m.hardware.Hash()

	log.G(m.ctx).Debug("hardware hashes",
		zap.String("saved", savedHardware),
		zap.String("exiting", exitingHardware))

	savedBenchmarks := m.state.PassedBenchmarks()
	//TODO(@antmat): use plain list of benchmarks when it will be available (via following PR)
	requiredBenchmarks := m.benchmarkList.MapByDeviceType()

	hwHashesMatched := exitingHardware == savedHardware
	benchMatched := m.isBenchmarkListMatches(requiredBenchmarks, savedBenchmarks)

	log.G(m.ctx).Debug("state matching",
		zap.Bool("hwHashesMatched", hwHashesMatched),
		zap.Bool("benchMatched", benchMatched))

	if benchMatched && hwHashesMatched {
		log.G(m.ctx).Debug("benchmarks list is matched, hardware is not changed, skip benchmarking this worker")
		// return back previously measured results for hardware
		m.hardware = m.state.HardwareWithBenchmarks()
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

	if err := m.state.SetPassedBenchmarks(passedBenchmarks); err != nil {
		return err
	}

	if err := m.state.SetHardwareWithBenchmarks(m.hardware); err != nil {
		return err
	}

	return m.state.SetHardwareHash(m.hardware.Hash())
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

	s := <-statusChan
	if s == pb.TaskStatusReply_FINISHED || s == pb.TaskStatusReply_BROKEN {
		if _, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, reader); err != nil {
			return nil, fmt.Errorf("cannot read logs into buffer: %v", err)
		}
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
		Image: b.GetImage(),
		Env: map[string]string{
			bm.BenchIDEnvParamName: fmt.Sprintf("%d", b.GetID()),
		},
	}
}

func (m *Miner) AskPlans(ctx context.Context) (*pb.AskPlansReply, error) {
	log.G(m.ctx).Info("handling AskPlans request")
	return &pb.AskPlansReply{AskPlans: m.salesman.AskPlans()}, nil
}

func (m *Miner) CreateAskPlan(ctx context.Context, askPlan *pb.AskPlan) (string, error) {
	log.G(m.ctx).Info("handling CreateAskPlan request", zap.Any("request", askPlan))
	if len(askPlan.GetID()) != 0 || !askPlan.GetOrderID().IsZero() || !askPlan.GetDealID().IsZero() {
		return "", errors.New("creating ask plans with predefined id, order_id or deal_id are not supported")
	}
	return m.salesman.CreateAskPlan(askPlan)
}

func (m *Miner) RemoveAskPlan(ctx context.Context, id string) error {
	log.G(m.ctx).Info("handling RemoveAskPlan request", zap.String("id", id))

	return m.salesman.RemoveAskPlan(id)
}

func (m *Miner) GetDealInfo(dealID *pb.BigInt) (*pb.DealInfoReply, error) {
	deal, err := m.salesman.Deal(dealID)
	if err != nil {
		return nil, err
	}

	ask, err := m.salesman.AskPlanByDeal(dealID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	running := &pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{},
	}

	completed := &pb.StatusMapReply{
		Statuses: map[string]*pb.TaskStatusReply{},
	}

	for id, c := range m.containers {
		// task is ours
		if c.DealID == dealID.Unwrap().String() {
			task := c.IntoProto()

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
		Resources: ask.GetResources(),
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

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	log.G(m.ctx).Info("closing miner")

	m.ssh.Close()
	m.ovs.Close()
	m.salesman.Close()
	m.plugins.Close()
}
