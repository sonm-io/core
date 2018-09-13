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
	"github.com/sonm-io/core/insonmnia/worker/gpu"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/insonmnia/worker/volume"
	sonm "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// Task for a target application.
type Task struct {
	*sonm.TaskSpec
	Image        reference.Field
	Cgroup       string
	CgroupParent string
	TaskID       string
	Autoremove   bool

	GPUDevices []gpu.GPUID

	mounts []volume.Mount

	NetworkOptions *network.Network

	status      sonm.TaskStatusReply_Status
	ContainerID string
	StartAt     time.Time
	Ports       nat.PortMap
	PublicKey   PublicKey

	NetworkIDs []string
	dealID     *sonm.BigInt
	AskID      string
}

func (d *Task) ID() string {
	return d.TaskID
}

func (d *Task) Volumes() map[string]*sonm.Volume {
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
	return d.dealID.Unwrap()
}

func (d *Task) IsGPURequired() bool {
	return len(d.GPUDevices) > 0
}

func (d *Task) GpuDeviceIDs() []gpu.GPUID {
	return d.GPUDevices
}

func (d *Task) Networks() []*sonm.NetworkSpec {
	return d.Container.GetNetworks()
}

func (d *Task) FormatEnv() []string {
	vars := make([]string, 0, len(d.Container.Env))
	for k, v := range d.Container.Env {
		vars = append(vars, fmt.Sprintf("%s=%s", k, v))
	}

	return vars
}

func (d *Task) Expose() (nat.PortSet, nat.PortMap, error) {
	return nat.ParsePortSpecs(d.Container.Expose)
}

func (c *Task) IntoProto(ctx context.Context) *sonm.TaskStatusReply {
	ports := make(map[string]*sonm.Endpoints)
	for hostPort, binding := range c.Ports {
		addrs := make([]*sonm.SocketAddr, len(binding))
		for i, bind := range binding {
			port, err := strconv.ParseUint(bind.HostPort, 10, 16)
			if err != nil {
				log.G(ctx).Warn("cannot parse port from nat.PortMap",
					zap.Error(err), zap.String("value", bind.HostPort))
				continue
			}
			addrs[i] = &sonm.SocketAddr{Addr: bind.HostIP, Port: uint32(port)}
		}

		ports[string(hostPort)] = &sonm.Endpoints{Endpoints: addrs}
	}

	// According to the source code it cannot fail
	imageName, _ := c.Image.MarshalText()
	return &sonm.TaskStatusReply{
		Status:             c.status,
		ImageName:          string(imageName),
		PortMap:            ports,
		Uptime:             uint64(time.Now().Sub(c.StartAt).Nanoseconds()),
		Usage:              nil,
		AllocatedResources: nil,
		Tag:                c.Tag,
	}
}
