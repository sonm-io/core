package miner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/sonm-io/core/insonmnia/miner/plugin"
	"github.com/sonm-io/core/insonmnia/miner/volume"
	"github.com/sonm-io/core/insonmnia/structs"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/miner/gpu"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
)

const overseerTag = "sonm.overseer"
const dieEvent = "die"

// Description for a target application.
type Description struct {
	Registry      string
	Image         string
	Auth          string
	RestartPolicy container.RestartPolicy
	Resources     container.Resources
	Cmd           []string
	Env           map[string]string
	TaskId        string
	DealId        string
	CommitOnStop  bool

	GPURequired bool

	volumes map[string]*pb.Volume
	mounts  []volume.Mount

	networks []structs.Network
}

func (d *Description) ID() string {
	return d.TaskId
}

func (d *Description) Volumes() map[string]*pb.Volume {
	return d.volumes
}

func (d *Description) Mounts(source string) []volume.Mount {
	return d.mounts
}

func (d *Description) IsGPURequired() bool {
	_, ok := d.Env["GPU"]
	return ok
}

func (d *Description) GpuDeviceIDs() []gpu.GPUID {
	// warn: hack just for testing, allows to set GPU device via task's ENV params
	v, ok := d.Env["GPU"]
	if !ok {
		return []gpu.GPUID{}
	}

	return []gpu.GPUID{gpu.GPUID(v)}
}

func (d *Description) Networks() []structs.Network {
	return d.networks
}

func (d *Description) FormatEnv() []string {
	vars := make([]string, 0, len(d.Env))
	for k, v := range d.Env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}

	return vars
}

// ContainerInfo is a brief information about containers
type ContainerInfo struct {
	status       *pb.TaskStatusReply
	ID           string
	ImageName    string
	StartAt      time.Time
	Ports        nat.PortMap
	Resources    resource.Resources
	PublicKey    ssh.PublicKey
	Cgroup       string
	CgroupParent string
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

// Overseer watches all miner's applications.
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

	// Exec a given command in running container
	Exec(ctx context.Context, Id string, cmd []string, env []string, isTty bool, wCh <-chan ssh.Window) (types.HijackedResponse, error)

	// Stop terminates the container.
	Stop(ctx context.Context, containerID string) error

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
				delete(o.containers, id)
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
					go func() {
						log.G(ctx).Info("trying to upload container")
						err := c.upload()
						if err != nil {
							log.G(ctx).Error("failed to commit container", zap.String("id", id), zap.Error(err))
						}

						if err := c.Cleanup(); err != nil {
							log.G(ctx).Error("failed to clean up container", zap.String("id", id), zap.Error(err))
						}
						c.cancel()
					}()
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
	source := types.ImageImportSource{
		Source:     rd,
		SourceName: "-",
	}

	options := types.ImageImportOptions{
		Tag:     "",
		Message: "",
	}

	response, err := o.client.ImageImport(ctx, source, "", options)
	if err != nil {
		log.G(o.ctx).Error("failed to load an image", zap.Error(err))
		return imageLoadStatus{}, err
	}

	defer response.Close()

	return decodeImageLoad(response)
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
	options := types.ImagePullOptions{
		All:          false,
		RegistryAuth: d.Auth,
	}

	refStr := filepath.Join(d.Registry, d.Image)

	body, err := o.client.ImagePull(ctx, refStr, options)
	if err != nil {
		log.G(ctx).Error("ImagePull failed", zap.String("ref", refStr), zap.Error(err))
		return err
	}

	if err = decodeImagePull(body); err != nil {
		log.G(ctx).Error("failed to pull an image", zap.Error(err))
		return err
	}

	return nil
}

func (o *overseer) Start(ctx context.Context, description Description) (status chan pb.TaskStatusReply_Status, cinfo ContainerInfo, err error) {
	// TODO: do we really need this check in that place?
	// TODO: maybe will be better to check somewhere into the "newContainer()" method?
	if description.GPURequired {
		if !o.supportGPU() {
			err = fmt.Errorf("GPU required but not supported or disabled")
			return
		}
	}

	// TODO: Well, we should refactor those dozens of arguments.
	// Note: maybe will be better to make the "newContainer()" func as part of the overseer struct
	// ( in that case we can access docker client and plugins repo from the Ovs instance. )
	pr, err := newContainer(ctx, o.client, description, o.plugins)
	if err != nil {
		return
	}

	o.mu.Lock()
	o.containers[pr.ID] = pr
	status = make(chan pb.TaskStatusReply_Status)
	o.statuses[pr.ID] = status
	o.mu.Unlock()

	if err = pr.startContainer(); err != nil {
		return
	}

	cjson, err := o.client.ContainerInspect(ctx, pr.ID)
	if err != nil {
		// NOTE: I don't think it can fail
		return
	}

	var cpuCount int
	if description.Resources.NanoCPUs > 0 {
		cpuCount = int(description.Resources.NanoCPUs / 1000000000)
	} else if description.Resources.CPUQuota > 0 && description.Resources.CPUPeriod > 0 {
		cpuCount = int(description.Resources.CPUQuota / description.Resources.CPUPeriod)
	} else if description.Resources.CPUCount > 0 {
		cpuCount = int(description.Resources.CPUCount)
	} else {
		cpuCount = 1
	}

	var gpuCount = 0
	if description.GPURequired {
		gpuCount = -1
	}

	cinfo = ContainerInfo{
		status:       &pb.TaskStatusReply{Status: pb.TaskStatusReply_RUNNING},
		ID:           cjson.ID,
		Ports:        cjson.NetworkSettings.Ports,
		Resources:    resource.NewResources(cpuCount, description.Resources.Memory, gpuCount),
		Cgroup:       string(cjson.HostConfig.Cgroup),
		CgroupParent: string(cjson.HostConfig.CgroupParent),
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
	ret, err = descriptor.execCommand(cmd, env, isTty, wCh)
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

	return descriptor.Kill()
}

func (o *overseer) Logs(ctx context.Context, id string, opts types.ContainerLogsOptions) (io.ReadCloser, error) {
	return o.client.ContainerLogs(ctx, id, opts)
}
