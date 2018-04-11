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
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/cgroups"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/util"

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
	resources *resource.Pool
	ethkey    *ecdsa.PrivateKey
	publicIPs []string
	eth       blockchain.API
	dwh       dwh.DWH

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

	// In-mem state, can be safety reloaded
	// from the external sources.
	// Must be protected by `mu` mutex.
	Deals map[structs.DealID]*structs.DealMeta
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
		eth, err := blockchain.NewAPI()
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
		o.dwh = dwh.NewDumbDWH(o.ctx)
	}

	cgName := ""
	var cgRes *specs.LinuxResources
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

	// check if memory is limited into cgroup
	if s, err := cgroup.Stats(); err == nil {
		if s.MemoryLimit < hardwareInfo.RAM.Device.Total {
			hardwareInfo.RAM.Device.Available = s.MemoryLimit
		}
	} else {
		hardwareInfo.RAM.Device.Available = hardwareInfo.RAM.Device.Total
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
		resources: resource.NewPool(hardwareInfo),
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

	if err := m.loadExternalState(); err != nil {
		return nil, err
	}

	return m, nil
}

type resourceHandle interface {
	// Commit marks the handle that the resources consumed should not be
	// released.
	commit()
	// Release releases consumed resources.
	// Useful in conjunction with defer.
	release()
}

// NilResourceHandle is a resource handle that does nothing.
type nilResourceHandle struct{}

func (h *nilResourceHandle) commit() {}

func (h *nilResourceHandle) release() {}

type ownedResourceHandle struct {
	miner     *Miner
	usage     resource.Resources
	committed bool
}

func (h *ownedResourceHandle) commit() {
	h.committed = true
}

func (h *ownedResourceHandle) release() {
	if h.committed {
		return
	}

	h.miner.resources.Release(&h.usage)
	h.committed = true
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

	m.containers[id].status = status
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

	// TODO(sshaman1101): check for deals existence in a right way;
	// Note: orderID is dealID;
	if _, err := m.GetDealByID(structs.DealID(request.GetOrderId())); err != nil {
		return nil, err
	}

	if request.GetContainer() == nil {
		return nil, fmt.Errorf("container field is required")
	}

	resources, err := structs.NewTaskResources(request.GetResources())
	if err != nil {
		return nil, err
	}

	publicKey, err := parsePublicKey(request.Container.PublicKeyData)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid public key provided %v", err)
	}

	cGroup, resourceHandle, err := m.consume(request.GetOrderId(), resources)
	if err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "failed to start %v", err)
	}

	// This can be canceled by using "resourceHandle.commit()".
	defer resourceHandle.release()

	mounts := make([]volume.Mount, 0)
	for _, spec := range request.Container.Mounts {
		mount, err := volume.NewMount(spec)
		if err != nil {
			return nil, err
		}
		mounts = append(mounts, mount)
	}

	networks, err := structs.NewNetworkSpecs(request.Container.Networks)
	if err != nil {
		log.G(ctx).Error("failed to parse networking specification", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to parse networking specification - %s", err)
	}
	var d = Description{
		Image:         request.Container.Image,
		Registry:      request.Container.Registry,
		Auth:          request.Container.Auth,
		RestartPolicy: transformRestartPolicy(request.RestartPolicy),
		Resources:     resources.ToContainerResources(cGroup.Suffix()),
		DealId:        request.GetOrderId(),
		TaskId:        request.Id,
		CommitOnStop:  request.Container.CommitOnStop,
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
		return nil, status.Errorf(codes.Internal, "failed to Spool %v", err)
	}

	m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_SPAWNING}, request.Id)
	log.G(ctx).Info("spawning an image")
	statusListener, containerInfo, err := m.ovs.Start(m.ctx, d)
	if err != nil {
		log.G(ctx).Error("failed to spawn an image", zap.Error(err))
		m.setStatus(&pb.TaskStatusReply{Status: pb.TaskStatusReply_BROKEN}, request.Id)
		return nil, status.Errorf(codes.Internal, "failed to Spawn %v", err)
	}
	containerInfo.PublicKey = publicKey
	containerInfo.StartAt = time.Now()
	containerInfo.ImageName = d.Image

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

	resourceHandle.commit()

	return &reply, nil
}

func (m *Miner) consume(orderId string, resources *structs.TaskResources) (cgroups.CGroup, resourceHandle, error) {
	cgroup, err := m.cGroupManager.Attach(orderId, resources.ToCgroupResources())
	if err != nil && err != cgroups.ErrCgroupAlreadyExists {
		return nil, nil, err
	}

	if err != cgroups.ErrCgroupAlreadyExists {
		return cgroup, &nilResourceHandle{}, nil
	}

	usage := resources.ToUsage()
	if err := m.resources.Consume(&usage); err != nil {
		return nil, nil, err
	}

	handle := &ownedResourceHandle{
		miner:     m,
		usage:     usage,
		committed: false,
	}

	return cgroup, handle, nil
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
	m.resources.Release(&containerInfo.Resources)

	return &pb.Empty{}, nil
}

func (m *Miner) CollectTasksStatuses() map[string]*pb.TaskStatusReply {
	result := map[string]*pb.TaskStatusReply{}
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, info := range m.containers {
		result[id] = info.status
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
	if info.status.Status == pb.TaskStatusReply_RUNNING {
		metrics, err := m.ovs.Info(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot get container metrics: %s", err.Error())
		}

		metric, ok = metrics[info.ID]
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Cannot get metrics for container %s", req.GetId())
		}
	}

	portsStr, _ := json.Marshal(info.Ports)
	reply := &pb.TaskStatusReply{
		Status:    info.status.Status,
		ImageName: info.ImageName,
		Ports:     string(portsStr),
		Uptime:    uint64(time.Now().Sub(info.StartAt).Nanoseconds()),
		Usage:     metric.Marshal(),
		AvailableResources: &pb.AvailableResources{
			NumCPUs:      int64(info.Resources.NumCPUs),
			NumGPUs:      int64(info.Resources.NumGPUs),
			Memory:       uint64(info.Resources.Memory),
			Cgroup:       info.ID,
			CgroupParent: info.CgroupParent,
		},
	}

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
	requiredBenchmarks := m.benchmarkList.List()

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
	switch dev {
	case pb.DeviceType_DEV_CPU:
		return m.runCPUBenchGroup(benches)
	case pb.DeviceType_DEV_RAM:
		return m.runRAMBenchGroup(benches)
	case pb.DeviceType_DEV_GPU:
		return m.runGPUBenchGroup(benches)
	case pb.DeviceType_DEV_NETWORK:
		return m.runNetworkBenchGroup(benches)
	case pb.DeviceType_DEV_STORAGE:
		return m.runStorageBenchGroup(benches)
	default:
		return fmt.Errorf("unknown benchmark group \"%s\"", dev.String())
	}
}

func (m *Miner) runCPUBenchGroup(benches []*pb.Benchmark) error {
	for _, ben := range benches {
		// check for special cases
		if ben.GetID() == bm.CPUCores {
			m.hardware.CPU.Benchmarks[bm.CPUCores] = &pb.Benchmark{
				ID:     bm.CPUCores,
				Code:   ben.GetCode(),
				Result: uint64(m.hardware.CPU.Device.Cores),
			}

			continue
		}

		d := getDescriptionForBenchmark(ben)
		d.Env[bm.CPUCountBenchParam] = fmt.Sprintf("%d", m.hardware.CPU.Device.Cores)

		res, err := m.execBenchmarkContainer(ben, d)
		if err != nil {
			return err
		}

		// save benchmark resutls for current CPU unit
		m.hardware.CPU.Benchmarks[ben.GetID()] = &pb.Benchmark{
			ID:     ben.GetID(),
			Code:   ben.GetCode(),
			Result: res.Result,
		}
	}

	return nil
}

func (m *Miner) runRAMBenchGroup(benches []*pb.Benchmark) error {
	for _, ben := range benches {
		// special case, just copy available amount of mem as benchmark result.
		if ben.GetID() == bm.RamSize {
			b := &pb.Benchmark{
				ID:     bm.RamSize,
				Code:   ben.GetCode(),
				Result: m.hardware.RAM.Device.Total,
			}

			m.hardware.RAM.Benchmarks[bm.RamSize] = b
			continue
		}

		d := getDescriptionForBenchmark(ben)
		res, err := m.execBenchmarkContainer(ben, d)
		if err != nil {
			return err
		}

		m.hardware.RAM.Benchmarks[ben.GetID()] = &pb.Benchmark{
			ID:     ben.GetID(),
			Code:   ben.GetCode(),
			Result: res.Result,
		}
		return nil
	}

	return nil
}

func (m *Miner) runGPUBenchGroup(benches []*pb.Benchmark) error {
	for _, dev := range m.hardware.GPU {
		for _, ben := range benches {
			if ben.GetID() == bm.GPUCount {
				dev.Benchmarks[bm.GPUCount] = &pb.Benchmark{
					ID:     ben.GetID(),
					Code:   ben.GetCode(),
					Result: 1,
				}
				continue
			}

			if ben.GetID() == bm.GPUMem {
				dev.Benchmarks[bm.GPUMem] = &pb.Benchmark{
					ID:     ben.GetID(),
					Code:   ben.GetCode(),
					Result: dev.Device.Memory,
				}
				continue
			}

			d := getDescriptionForBenchmark(ben)
			d.GPUDevices = []gpu.GPUID{gpu.GPUID(dev.GetDevice().GetID())}

			res, err := m.execBenchmarkContainer(ben, d)
			if err != nil {
				return err
			}

			dev.Benchmarks[ben.GetID()] = &pb.Benchmark{
				ID:     ben.GetID(),
				Code:   ben.GetCode(),
				Result: res.Result,
			}
		}
	}

	return nil
}

func (m *Miner) runNetworkBenchGroup(benches []*pb.Benchmark) error {
	for _, ben := range benches {
		d := getDescriptionForBenchmark(ben)
		res, err := m.execBenchmarkContainer(ben, d)
		if err != nil {
			return err
		}

		m.hardware.Network.Benchmarks[ben.GetID()] = &pb.Benchmark{
			ID:     ben.GetID(),
			Code:   ben.GetCode(),
			Result: res.Result,
		}
	}

	return nil
}

func (m *Miner) runStorageBenchGroup(benches []*pb.Benchmark) error {
	// note: there is no storage sub-system for worker yet, so benchmark always returns zero
	for _, ben := range benches {
		m.hardware.Storage.Benchmarks[ben.GetID()] = &pb.Benchmark{
			ID:     ben.GetID(),
			Code:   ben.GetCode(),
			Result: 0,
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

	logOpts := types.ContainerLogsOptions{ShowStdout: true}
	reader, err := m.ovs.Logs(m.ctx, statusReply.ID, logOpts)
	if err != nil {
		return nil, fmt.Errorf("cannot create container log reader: %v", err)
	}
	defer reader.Close()

	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	_, err = stdcopy.StdCopy(&stdoutBuf, &stderrBuf, reader)
	if err != nil {
		return nil, fmt.Errorf("cannot read logs into buffer: %v", err)
	}

	resultsMap, err := parseBenchmarkResult(stdoutBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("cannot parse benchmark result: %v", err)
	}

	<-statusChan
	if err := m.ovs.Stop(m.ctx, statusReply.ID); err != nil {
		log.G(m.ctx).Warn("cannot stop benchmark container", zap.Error(err))
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
	return &pb.AskPlansReply{AskPlans: m.state.AskPlans()}, nil
}

func (m *Miner) CreateAskPlan(ctx context.Context, request *pb.AskPlan) (string, error) {
	log.G(m.ctx).Info("handling CreateAskPlan request", zap.Any("request", request))

	return m.state.CreateAskPlan(request)
}

func (m *Miner) RemoveAskPlan(ctx context.Context, id string) error {
	log.G(m.ctx).Info("handling RemoveAskPlan request", zap.String("id", id))

	if err := m.state.RemoveAskPlan(id); err != nil {
		return err
	}

	return nil
}

func (m *Miner) GetDealByID(id structs.DealID) (*structs.DealMeta, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	deal, ok := m.Deals[id]
	if !ok {
		return nil, fmt.Errorf("deal with id=%s does not found", id)
	}

	return deal, nil
}

// loadExternalState loads external state into in-mem storage
// (from blockchain, DWH, etc...). Must be called before
// boltdb state is loaded.
func (m *Miner) loadExternalState() error {
	log.G(m.ctx).Debug("loading initial state from external sources")
	if err := m.loadDeals(); err != nil {
		return err
	}

	if err := m.syncAskPlans(); err != nil {
		return err
	}

	return nil
}

func (m *Miner) loadDeals() error {
	log.G(m.ctx).Debug("loading opened deals")

	filter := dwh.DealsFilter{
		Author: util.PubKeyToAddr(m.ethkey.PublicKey),
	}

	deals, err := m.dwh.GetDeals(filter)
	if err != nil {
		return err
	}

	m.mu.Lock()
	for _, deal := range deals {
		m.Deals[structs.DealID(deal.GetId())] = structs.NewDealMeta(deal)
	}
	m.mu.Unlock()

	return nil
}

func (m *Miner) syncAskPlans() error {
	log.G(m.ctx).Debug("synchronizing ask-plans with remote marketplace")
	return nil
}

// todo: make the `miner.Init() error` method to kickstart all initial jobs for the Worker instance.
// (state loading, benchmarking, market sync).

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	log.G(m.ctx).Info("closing miner")

	m.ssh.Close()
	m.ovs.Close()
	m.controlGroup.Delete()
	m.plugins.Close()
}
