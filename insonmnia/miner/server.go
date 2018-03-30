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
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/miner/gpu"

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
	ctx context.Context

	cfg Config

	plugins *plugin.Repository

	hardware  *hardware.Hardware
	resources *resource.Pool

	hubKey *ecdsa.PublicKey

	publicIPs []string

	ovs Overseer

	mu sync.Mutex
	// One-to-one mapping between container IDs and userland task names.
	//
	// The overseer operates with containers in terms of their ID, which does not change even during auto-restart.
	// However some requests pass an application (or task) name, which is more meaningful for user. To be able to
	// transform between these two identifiers this map exists.
	//
	// WARNING: This must be protected using `mu`.
	nameMapping map[string]string

	// Maps StartRequest's IDs to containers' IDs
	// TODO: It's doubtful that we should keep this map here instead in the Overseer.
	containers map[string]*ContainerInfo

	channelCounter int
	controlGroup   cGroup
	cGroupManager  cGroupManager
	ssh            SSH
	state          *state
	benchmarkList  bm.BenchList
}

func NewMiner(cfg Config, opts ...Option) (m *Miner, err error) {
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

	if o.ctx == nil {
		o.ctx = context.Background()
	}

	if o.benchList == nil {
		o.benchList, err = bm.NewBenchmarksList(o.ctx, cfg.Benchmarks())
		if err != nil {
			return nil, err
		}
	}

	hardwareInfo, err := hardware.NewHardware()
	if err != nil {
		return nil, err
	}

	cgroup, cGroupManager, err := makeCgroupManager(cfg.HubResources())
	if err != nil {
		return nil, err
	}

	state, err := NewState(o.ctx, cfg)
	if err != nil {
		return nil, err
	}

	if !platformSupportCGroups && cfg.HubResources() != nil {
		log.G(o.ctx).Warn("your platform does not support CGroup, but the config has resources section")
	}

	if err := o.setupNetworkOptions(cfg); err != nil {
		return nil, errors.Wrap(err, "failed to set up network options")
	}

	log.G(o.ctx).Info("discovered public IPs", zap.Any("public IPs", o.publicIPs))

	plugins, err := plugin.NewRepository(o.ctx, cfg.Plugins())
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
		ctx: o.ctx,
		cfg: cfg,
		ovs: o.ovs,

		plugins: plugins,

		hardware:  hardwareInfo,
		resources: resource.NewPool(hardwareInfo),
		publicIPs: o.publicIPs,

		containers:  make(map[string]*ContainerInfo),
		nameMapping: make(map[string]string),

		controlGroup:  cgroup,
		cGroupManager: cGroupManager,
		ssh:           o.ssh,
		state:         state,
		benchmarkList: o.benchList,
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

func (m *Miner) getTaskIdByContainerId(id string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name, ok := m.nameMapping[id]
	return name, ok
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

// Ping works as Healthcheck for the Hub
func (m *Miner) Ping(ctx context.Context, _ *pb.Empty) (*pb.PingReply, error) {
	log.G(m.ctx).Info("got ping request from Hub")
	return &pb.PingReply{}, nil
}

// Info returns runtime statistics collected from all containers working on this miner.
//
// This works the following way: a miner periodically collects various runtime statistics from all
// spawned containers that it knows about. For running containers metrics map the immediate
// state, for dead containers - their last memento.
func (m *Miner) Info(ctx context.Context, request *pb.Empty) (*pb.InfoReply, error) {
	log.G(m.ctx).Info("handling Info request", zap.Any("req", request))

	info, err := m.ovs.Info(ctx)
	if err != nil {
		return nil, err
	}

	var result = &pb.InfoReply{
		Usage:        make(map[string]*pb.ResourceUsage),
		Capabilities: m.hardware.IntoProto(),
	}

	for containerID, stat := range info {
		if id, ok := m.getTaskIdByContainerId(containerID); ok {
			result.Usage[id] = stat.Marshal()
		}
	}

	return result, nil
}

// Handshake notifies Hub about Worker's hardware capabilities
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

	cgroup, resourceHandle, err := m.consume(request.GetOrderId(), resources)
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
		Resources:     resources.ToContainerResources(cgroup.Suffix()),
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

func (m *Miner) consume(orderId string, resources *structs.TaskResources) (cGroup, resourceHandle, error) {
	cgroup, err := m.cGroupManager.Attach(orderId, resources.ToCgroupResources())
	if err != nil && err != errCgroupAlreadyExists {
		return nil, nil, err
	}
	if err != errCgroupAlreadyExists {
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
	savedHardware := m.state.getHardwareHash()
	exitingHardware := m.hardware.Hash()

	log.G(m.ctx).Debug("hardware hashes",
		zap.String("saved", savedHardware),
		zap.String("exiting", exitingHardware))

	savedBenchmarks := m.state.getPassedBenchmarks()
	requiredBenchmarks := m.benchmarkList.List()

	hwHashesMatched := exitingHardware == savedHardware
	benchMatched := m.isBenchmarkListMatches(requiredBenchmarks, savedBenchmarks)

	log.G(m.ctx).Debug("state matching",
		zap.Bool("hwHashesMatched", hwHashesMatched),
		zap.Bool("benchMatched", benchMatched))

	if benchMatched && hwHashesMatched {
		log.G(m.ctx).Debug("benchmarks list is matched, hardware is not changed, skip benchmarking this worker")
		// return back previously measured results for hardware
		m.hardware = m.state.getHardwareWithBenchmarks()
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

	if err := m.state.setPassedBenchmarks(passedBenchmarks); err != nil {
		return err
	}

	if err := m.state.setHardwareWithBenchmarks(m.hardware); err != nil {
		return err
	}

	return m.state.setHardwareHash(m.hardware.Hash())
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
	for _, c := range m.hardware.CPU {
		for _, ben := range benches {
			// check for special cases
			if ben.GetID() == bm.CPUCores {
				c.Benchmark[bm.CPUCores] = &pb.Benchmark{
					ID:     bm.CPUCores,
					Result: uint64(c.Device.Cores),
				}

				continue
			}

			d := getDescriptionForBenchmark(ben)
			d.Env[bm.CPUCountBenchParam] = fmt.Sprintf("%d", c.Device.Cores)

			res, err := m.execBenchmarkContainer(ben, d)
			if err != nil {
				return err
			}

			// save benchmark resutls for current CPU unit
			c.Benchmark[ben.GetID()] = &pb.Benchmark{ID: ben.GetID(), Result: res.Result}
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
				Result: m.hardware.AvailableMemory(),
			}

			m.hardware.Memory.Benchmark[bm.RamSize] = b
			continue
		}

		d := getDescriptionForBenchmark(ben)
		res, err := m.execBenchmarkContainer(ben, d)
		if err != nil {
			return err
		}

		m.hardware.Memory.Benchmark[ben.GetID()] = &pb.Benchmark{ID: ben.GetID(), Result: res.Result}
		return nil
	}

	return nil
}

func (m *Miner) runGPUBenchGroup(benches []*pb.Benchmark) error {
	for _, dev := range m.hardware.GPU {
		for _, ben := range benches {
			if ben.GetID() == bm.GPUCount {
				dev.Benchmark[bm.GPUCount] = &pb.Benchmark{ID: ben.GetID(), Result: 1}
				continue
			}

			if ben.GetID() == bm.GPUMem {
				dev.Benchmark[bm.GPUMem] = &pb.Benchmark{
					ID: ben.GetID(),
					// todo(sshaman1101): add some method to retrieve GPU's memory
					Result: 100500,
				}
				continue
			}

			d := getDescriptionForBenchmark(ben)
			d.GPUDevices = []gpu.GPUID{gpu.GPUID(dev.Device.GetID())}

			res, err := m.execBenchmarkContainer(ben, d)
			if err != nil {
				return err
			}

			dev.Benchmark[ben.GetID()] = &pb.Benchmark{ID: ben.GetID(), Result: res.Result}
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

		m.hardware.Network.Benchmark[ben.GetID()] = &pb.Benchmark{ID: ben.GetID(), Result: res.Result}
	}

	return nil
}

func (m *Miner) runStorageBenchGroup(benches []*pb.Benchmark) error {
	// note: there is no storage sub-system for worker yet, so benchmark always returns zero
	m.hardware.Storage.Benchmark[bm.StorageSize] = &pb.Benchmark{ID: bm.StorageSize, Result: 0}
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
		return nil, fmt.Errorf("no results for benchmark id=%d found into container's output", ben.GetID())
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

// Close disposes all resources related to the Miner
func (m *Miner) Close() {
	log.G(m.ctx).Info("closing miner")

	m.ssh.Close()
	m.ovs.Close()
	m.controlGroup.Delete()
	m.plugins.Close()
}
