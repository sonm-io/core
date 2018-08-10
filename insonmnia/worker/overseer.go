package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/xdocker"
	"go.uber.org/zap"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	pb "github.com/sonm-io/core/proto"
)

const overseerTag = "sonm.overseer"
const dealIDTag = "sonm.dealid"
const dieEvent = "die"

// Description for a target application.
type Description struct {
	pb.Container
	Reference    reference.Reference
	Auth         string
	Resources    *pb.AskPlanResources
	CGroupParent string
	Cmd          []string
	TaskId       string
	DealId       string
	Autoremove   bool

	GPUDevices []gpu.GPUID

	mounts []volume.Mount

	network  *network.Network
	networks []*structs.NetworkSpec
}

func (d *Description) ID() string {
	return d.TaskId
}

type descriptionAlias Description

type descriptionMarshaller struct {
	*descriptionAlias
	Networks []*structs.NetworkSpec
	RefField reference.Field `json:"Reference"`
}

func (d Description) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(&descriptionMarshaller{
		descriptionAlias: (*descriptionAlias)(&d),
		Networks:         d.networks,
		RefField:         reference.AsField(d.Reference),
	})

	return b, err
}

func (d *Description) UnmarshalJSON(data []byte) error {
	aux := &descriptionMarshaller{descriptionAlias: (*descriptionAlias)(d)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	d.Reference = aux.RefField.Reference()
	d.networks = aux.Networks
	return nil
}

func (d *Description) Volumes() map[string]*pb.Volume {
	return d.Container.Volumes
}

func (d *Description) Mounts(source string) []volume.Mount {
	return d.mounts
}

func (d *Description) Network() (string, string) {
	if d.network == nil {
		return "", ""
	}

	return d.network.Name, d.network.ID
}

func (d *Description) DealID() string {
	return d.DealId
}

func (d *Description) IsGPURequired() bool {
	return len(d.GPUDevices) > 0
}

func (d *Description) GpuDeviceIDs() []gpu.GPUID {
	return d.GPUDevices
}

func (d *Description) Networks() []*structs.NetworkSpec {
	return d.networks
}

func (d *Description) FormatEnv() []string {
	vars := make([]string, 0, len(d.Env))
	for k, v := range d.Env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}

	return vars
}

func (d *Description) Expose() (nat.PortSet, nat.PortMap, error) {
	return nat.ParsePortSpecs(d.Container.Expose)
}

// ContainerInfo is a brief information about containers
type ContainerInfo struct {
	status       pb.TaskStatusReply_Status
	ID           string
	ImageName    string
	StartAt      time.Time
	Ports        nat.PortMap
	PublicKey    ssh.PublicKey
	Cgroup       string
	CgroupParent string
	NetworkIDs   []string
	DealID       string
	TaskId       string
	Tag          *pb.TaskTag
	AskID        string
}

func (c *ContainerInfo) IntoProto(ctx context.Context) *pb.TaskStatusReply {
	ports := make(map[string]*pb.Endpoints)
	for hostPort, binding := range c.Ports {
		addrs := make([]*pb.SocketAddr, len(binding))
		for i, bind := range binding {
			port, err := strconv.ParseUint(bind.HostPort, 10, 16)
			if err != nil {
				log.G(ctx).Warn("cannot parse port from nat.PortMap",
					zap.Error(err), zap.String("value", bind.HostPort))
				continue
			}
			addrs[i] = &pb.SocketAddr{Addr: bind.HostIP, Port: uint32(port)}
		}

		ports[string(hostPort)] = &pb.Endpoints{Endpoints: addrs}
	}

	return &pb.TaskStatusReply{
		Status:             c.status,
		ImageName:          c.ImageName,
		PortMap:            ports,
		Uptime:             uint64(time.Now().Sub(c.StartAt).Nanoseconds()),
		Usage:              nil,
		AllocatedResources: nil,
		Tag:                c.Tag,
	}
}

// ContainerMetrics are metrics collected from Docker about running containers
type ContainerMetrics struct {
	cpu types.CPUStats
	mem types.MemoryStats
	net map[string]types.NetworkStats
}

func (m *ContainerMetrics) Marshal() *pb.ResourceUsage {
	network := make(map[string]*pb.NetworkUsage)
	for i, n := range m.net {
		network[i] = &pb.NetworkUsage{
			TxBytes:   n.TxBytes,
			RxBytes:   n.RxBytes,
			TxPackets: n.TxPackets,
			RxPackets: n.RxPackets,
			TxErrors:  n.TxErrors,
			RxErrors:  n.RxErrors,
		}
	}

	return &pb.ResourceUsage{
		Cpu: &pb.CPUUsage{
			Total: m.cpu.CPUUsage.TotalUsage,
		},
		Memory: &pb.MemoryUsage{
			MaxUsage: m.mem.MaxUsage,
		},
		Network: network,
	}
}

type ExecConnection types.HijackedResponse

// Overseer watches all worker's applications.
type Overseer interface {
	// Load loads an image from the specified reader to the Docker.
	Load(ctx context.Context, rd io.Reader) (imageLoadStatus, error)

	// Save saves an image from the Docker into the returned reader.
	Save(ctx context.Context, imageID string) (types.ImageInspect, io.ReadCloser, error)

	// Spool prepares an application for its further start.
	//
	// For Docker containers this is an equivalent of pulling from the registry.
	Spool(ctx context.Context, d Description) error

	// Start attempts to start an application using the specified description.
	//
	// After successful starting an application becomes a target for accepting request, but not guarantees
	// to complete them.
	Start(ctx context.Context, description Description) (chan pb.TaskStatusReply_Status, ContainerInfo, error)

	// Attach attemps to attach to a running application with a specified description
	Attach(ctx context.Context, ID string, description Description) (chan pb.TaskStatusReply_Status, error)

	// Exec a given command in running container
	Exec(ctx context.Context, Id string, cmd []string, env []string, isTty bool, wCh <-chan ssh.Window) (types.HijackedResponse, error)

	// Stop terminates the container.
	Stop(ctx context.Context, containerID string) error

	// Makes all cleanup related to closed deal
	OnDealFinish(ctx context.Context, containerID string) error

	// Info returns runtime statistics collected from all running containers.
	//
	// Depending on the implementation this can be cached.
	Info(ctx context.Context) (map[string]ContainerMetrics, error)

	// Fetch logs of the container
	Logs(ctx context.Context, id string, opts types.ContainerLogsOptions) (io.ReadCloser, error)

	// Close terminates all associated asynchronous operations and prepares the Overseer for shutting down.
	Close() error
}

type overseer struct {
	ctx    context.Context
	cancel context.CancelFunc

	plugins *plugin.Repository

	client *client.Client

	registryAuth map[string]string

	// protects containers map
	mu         sync.Mutex
	containers map[string]*containerDescriptor
	statuses   map[string]chan pb.TaskStatusReply_Status
}

func (o *overseer) supportGPU() bool {
	return o.plugins.HasGPU()
}

// NewOverseer creates new overseer
func NewOverseer(ctx context.Context, plugins *plugin.Repository) (Overseer, error) {
	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	ovr := &overseer{
		ctx:        ctx,
		cancel:     cancel,
		plugins:    plugins,
		client:     dockerClient,
		containers: make(map[string]*containerDescriptor),
		statuses:   make(map[string]chan pb.TaskStatusReply_Status),
	}

	go ovr.collectStats()
	go ovr.watchEvents()

	return ovr, nil
}

func (o *overseer) Info(ctx context.Context) (map[string]ContainerMetrics, error) {
	info := make(map[string]ContainerMetrics)

	o.mu.Lock()
	for _, container := range o.containers {
		metrics := ContainerMetrics{
			cpu: container.stats.CPUStats,
			mem: container.stats.MemoryStats,
			net: container.stats.Networks,
		}

		info[container.ID] = metrics
	}
	o.mu.Unlock()

	return info, nil
}

func (o *overseer) Close() error {
	o.cancel()
	return nil
}

func (o *overseer) handleStreamingEvents(ctx context.Context, sinceUnix int64, filterArgs filters.Args) (last int64, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// No messages handled
	// so in the worst case restart from Since
	last = sinceUnix
	options := types.EventsOptions{
		Since:   strconv.FormatInt(sinceUnix, 10),
		Filters: filterArgs,
	}
	log.G(ctx).Info("subscribe to Docker events", zap.String("since", options.Since))
	messages, errors := o.client.Events(ctx, options)
	for {
		select {
		case err = <-errors:
			return last, err

		case message := <-messages:
			last = message.TimeNano

			switch message.Status {
			case dieEvent:
				id := message.Actor.ID
				log.G(ctx).Info("container has died", zap.String("id", id))

				var c *containerDescriptor
				o.mu.Lock()
				c, containerFound := o.containers[id]
				s, statusFound := o.statuses[id]
				// We intentionally do not delete container from the map to save history for deal.
				// It will be removed from that map after deal finishes and corresponding container would be deleted
				delete(o.statuses, id)
				o.mu.Unlock()

				if !containerFound {
					// NOTE: it could be orphaned container from our previous launch
					log.G(ctx).Warn("unknown container with sonm tag will be removed", zap.String("id", id))
					containerRemove(o.ctx, o.client, id)
					continue
				}
				if statusFound {
					s <- pb.TaskStatusReply_BROKEN
					close(s)
				}
				if c.description.CommitOnStop {
					log.G(ctx).Info("trying to upload container")
					err := c.upload(ctx)
					if err != nil {
						log.G(ctx).Error("failed to commit container", zap.String("id", id), zap.Error(err))
					}
				}
				if err := c.Cleanup(); err != nil {
					log.G(ctx).Error("failed to clean up container", zap.String("id", id), zap.Error(err))
				}
			default:
				log.G(ctx).Warn("received unknown event", zap.String("status", message.Status))
			}
		}
	}
}

func (o *overseer) watchEvents() {
	backoff := NewBackoffTimer(time.Second, time.Second*32)
	defer backoff.Stop()

	sinceUnix := time.Now().Unix()

	filterArgs := filters.NewArgs()
	filterArgs.Add("event", dieEvent)
	filterArgs.Add("label", overseerTag)

	var err error
	for {
		sinceUnix, err = o.handleStreamingEvents(o.ctx, sinceUnix, filterArgs)
		switch err {
		// NOTE: seems no nil-error case needed
		// case nil:
		// 	// pass
		case context.Canceled, context.DeadlineExceeded:
			log.G(o.ctx).Info("event listening has been cancelled")
			return
		default:
			log.G(o.ctx).Warn("failed to attach to a Docker events stream. Retry later")
			select {
			case <-backoff.C():
				//pass
			case <-o.ctx.Done():
				log.G(o.ctx).Info("event listening has been cancelled during sleep")
				return
			}
		}
	}
}

func (o *overseer) collectStats() {
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			ids := stringArrayPool.Get().([]string)
			o.mu.Lock()
			for id := range o.containers {
				ids = append(ids, id)
			}
			o.mu.Unlock()
			for _, id := range ids {
				if len(id) == 0 {
					continue
				}

				resp, err := o.client.ContainerStats(o.ctx, id, false)
				if err != nil {
					log.G(o.ctx).Warn("failed to get Stats", zap.String("id", id), zap.Error(err))
					continue
				}
				var stats types.StatsJSON
				err = json.NewDecoder(resp.Body).Decode(&stats)
				switch err {
				case nil:
					// pass
				case io.EOF:
					// pass
				default:
					log.G(o.ctx).Warn("failed to decode container Stats", zap.String("id", id), zap.Error(err))
				}
				resp.Body.Close()

				log.G(o.ctx).Debug("received container stats", zap.String("id", id), zap.Any("stats", stats))
				o.mu.Lock()
				if container, ok := o.containers[id]; ok {
					container.stats = stats
				}
				o.mu.Unlock()
			}
			stringArrayPool.Put(ids[:0])
		case <-o.ctx.Done():
			return
		}
	}
}

func (o *overseer) Load(ctx context.Context, rd io.Reader) (imageLoadStatus, error) {
	response, err := o.client.ImageLoad(ctx, rd, true)

	if err != nil {
		log.G(o.ctx).Error("failed to load an image", zap.Error(err))
		return imageLoadStatus{}, err
	}

	defer response.Body.Close()

	return decodeImageLoad(response.Body)
}

func (o *overseer) Save(ctx context.Context, imageID string) (types.ImageInspect, io.ReadCloser, error) {
	imageInspect, _, err := o.client.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return types.ImageInspect{}, nil, err
	}

	rd, err := o.client.ImageSave(ctx, []string{imageID})
	if err != nil {
		return types.ImageInspect{}, nil, err
	}

	return imageInspect, rd, nil
}

func (o *overseer) Spool(ctx context.Context, d Description) error {
	log.G(ctx).Info("pull the application image")
	// TODO: maybe add sonm labels to make filtration easier
	summaries, err := o.client.ImageList(ctx, types.ImageListOptions{All: true})
	if err != nil {
		return err
	}
	refStr := d.Reference.String()
	for _, summary := range summaries {
		if summary.ID == refStr {
			log.S(ctx).Infof("application image %s is already present", d.Reference.String())
			return nil
		}
	}
	options := types.ImagePullOptions{
		All:          false,
		RegistryAuth: d.Auth,
	}

	body, err := o.client.ImagePull(ctx, refStr, options)
	if err != nil {
		log.G(ctx).Error("ImagePull failed", zap.String("ref", refStr), zap.Error(err))
		return err
	}

	if err = xdocker.DecodeImagePull(body); err != nil {
		log.G(ctx).Error("failed to pull an image", zap.Error(err))
		return err
	}

	return nil
}

func (o *overseer) Attach(ctx context.Context, ID string, d Description) (chan pb.TaskStatusReply_Status, error) {
	cont, err := attachContainer(ctx, o.client, ID, d, o.plugins)
	if err != nil {
		log.S(ctx).Debugf("failed to attach to container %s", err)
		return nil, err
	}
	cont.ID = ID
	log.S(ctx).Debugf("attached to running container %s", ID)

	o.mu.Lock()
	o.containers[ID] = cont
	status := make(chan pb.TaskStatusReply_Status, 1)
	o.statuses[ID] = status
	o.mu.Unlock()

	return status, nil
}
func (o *overseer) Start(ctx context.Context, description Description) (status chan pb.TaskStatusReply_Status, cinfo ContainerInfo, err error) {
	if description.IsGPURequired() && !o.supportGPU() {
		err = fmt.Errorf("GPU required but not supported or disabled")
		return
	}

	// TODO: Well, we should refactor those dozens of arguments.
	// Note: maybe will be better to make the "newContainer()" func as part of the overseer struct
	// ( in that case we can access docker client and plugins repo from the Ovs instance. )
	pr, err := newContainer(ctx, o.client, description, o.plugins)
	if err != nil {
		log.S(ctx).Debugf("failed to create container")
		return
	}
	log.S(ctx).Debugf("created container %s", pr.ID)

	o.mu.Lock()
	o.containers[pr.ID] = pr
	status = make(chan pb.TaskStatusReply_Status, 1)
	o.statuses[pr.ID] = status
	o.mu.Unlock()

	if err = pr.startContainer(ctx); err != nil {
		log.S(ctx).Warnf("failed to start container %s", pr.ID)
		return
	}
	log.S(ctx).Debugf("started container %s", pr.ID)

	cjson, err := o.client.ContainerInspect(ctx, pr.ID)
	if err != nil {
		log.S(ctx).Warnf("failed to inspect container %s", pr.ID)
		// NOTE: I don't think it can fail
		return
	}

	var networkIDs []string
	for k := range cjson.NetworkSettings.Networks {
		networkIDs = append(networkIDs, k)
	}

	cinfo = ContainerInfo{
		status:       pb.TaskStatusReply_RUNNING,
		ID:           cjson.ID,
		Ports:        cjson.NetworkSettings.Ports,
		Cgroup:       string(cjson.HostConfig.Cgroup),
		CgroupParent: string(cjson.HostConfig.CgroupParent),
		NetworkIDs:   networkIDs,
		TaskId:       description.TaskId,
	}

	return status, cinfo, nil
}

func (o *overseer) Exec(ctx context.Context, id string, cmd []string, env []string, isTty bool, wCh <-chan ssh.Window) (ret types.HijackedResponse, err error) {
	o.mu.Lock()
	descriptor, dok := o.containers[id]
	o.mu.Unlock()
	if !dok {
		err = fmt.Errorf("no such container %s", id)
		return
	}
	ret, err = descriptor.execCommand(ctx, cmd, env, isTty, wCh)
	return
}

func (o *overseer) Stop(ctx context.Context, containerid string) error {
	o.mu.Lock()

	descriptor, dok := o.containers[containerid]
	status, sok := o.statuses[containerid]
	delete(o.statuses, containerid)
	o.mu.Unlock()

	if sok {
		status <- pb.TaskStatusReply_FINISHED
		close(status)
	}

	if !dok {
		return fmt.Errorf("no such container %s", containerid)
	}

	return descriptor.Kill(ctx)
}

func (o *overseer) OnDealFinish(ctx context.Context, containerID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	descriptor, ok := o.containers[containerID]
	delete(o.containers, containerID)
	status, sok := o.statuses[containerID]
	delete(o.statuses, containerID)
	if sok {
		status <- pb.TaskStatusReply_FINISHED
		close(status)
	}
	if !ok {
		return fmt.Errorf("unknown container %s", containerID)
	}
	result := multierror.NewMultiError()
	if err := descriptor.Kill(ctx); err != nil {
		result = multierror.Append(result, err)
	}
	//This is needed in case container is uploaded into external registry
	if descriptor.description.CommitOnStop {
		if err := descriptor.upload(ctx); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if err := descriptor.Cleanup(); err != nil {
		result = multierror.Append(result, err)
	}
	if err := descriptor.Remove(ctx); err != nil {
		result = multierror.Append(result, err)
	}
	return result.ErrorOrNil()
}

func (o *overseer) Logs(ctx context.Context, id string, opts types.ContainerLogsOptions) (io.ReadCloser, error) {
	return o.client.ContainerLogs(ctx, id, opts)
}
