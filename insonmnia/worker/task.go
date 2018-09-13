package worker

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/go-connections/nat"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Task for a target application.
type Task struct {
	pb.Container
	Reference    reference.Field
	Auth         string
	Resources    *pb.AskPlanResources
	CGroupParent string
	Cmd          []string
	TaskId       string
	DealId       *big.Int
	Autoremove   bool

	GPUDevices []gpu.GPUID

	mounts []volume.Mount

	NetworkOptions *network.Network
	NetworkSpecs   []*structs.NetworkSpec

	status       pb.TaskStatusReply_Status
	ContainerID  string
	ImageName    string
	StartAt      time.Time
	Ports        nat.PortMap
	PublicKey    PublicKey
	Cgroup       string
	CgroupParent string
	NetworkIDs   []string
	dealID       *pb.BigInt
	Tag          *pb.TaskTag
	AskID        string
}

func (d *Task) ID() string {
	return d.TaskId
}

func (d *Task) Volumes() map[string]*pb.Volume {
	return d.Container.Volumes
}

func (d *Task) Mounts(source string) []volume.Mount {
	return d.mounts
}

func (d *Task) Network() (string, string) {
	if d.NetworkOptions == nil {
		return "", ""
	}

	return d.NetworkOptions.Name, d.NetworkOptions.ID
}

func (d *Task) DealID() *big.Int {
	return d.DealId
}

func (d *Task) IsGPURequired() bool {
	return len(d.GPUDevices) > 0
}

func (d *Task) GpuDeviceIDs() []gpu.GPUID {
	return d.GPUDevices
}

func (d *Task) Networks() []*structs.NetworkSpec {
	return d.NetworkSpecs
}

func (d *Task) FormatEnv() []string {
	vars := make([]string, 0, len(d.Env))
	for k, v := range d.Env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}

	return vars
}

func (d *Task) Expose() (nat.PortSet, nat.PortMap, error) {
	return nat.ParsePortSpecs(d.Container.Expose)
}

func (c *Task) IntoProto(ctx context.Context) *pb.TaskStatusReply {
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
