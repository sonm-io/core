package miner

import (
	"golang.org/x/net/context"

	"net"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type MinerBuilder struct {
	ctx      context.Context
	cfg      Config
	hardware hardware.HardwareInfo
	ip       net.IP
	ovs      Overseer
	uuid     string
	ssh      SSH
}

func (b *MinerBuilder) Context(ctx context.Context) *MinerBuilder {
	b.ctx = ctx
	return b
}

func (b *MinerBuilder) Config(config Config) *MinerBuilder {
	b.cfg = config
	return b
}

func (b *MinerBuilder) Hardware(hardware hardware.HardwareInfo) *MinerBuilder {
	b.hardware = hardware
	return b
}

func (b *MinerBuilder) Address(ip net.IP) *MinerBuilder {
	b.ip = ip
	return b
}

func (b *MinerBuilder) Overseer(ovs Overseer) *MinerBuilder {
	b.ovs = ovs
	return b
}

func (b *MinerBuilder) UUID(uuid string) *MinerBuilder {
	b.uuid = uuid
	return b
}

func (b *MinerBuilder) SSH(ssh SSH) *MinerBuilder {
	b.ssh = ssh
	return b
}

func (b *MinerBuilder) Build() (miner *Miner, err error) {
	if b.ctx == nil {
		b.ctx = context.Background()
	}

	if b.cfg == nil {
		return nil, errors.New("config is mandatory for MinerBuilder")
	}

	if b.hardware == nil {
		b.hardware = hardware.New()
	}

	if b.ip == nil {
		addr, err := util.GetPublicIP()
		if err != nil {
			return nil, err
		}
		b.ip = addr
	}

	ctx, cancel := context.WithCancel(b.ctx)
	if b.ovs == nil {
		b.ovs, err = NewOverseer(ctx, b.cfg.GPU())
		if err != nil {
			cancel()
			return nil, err
		}
	}

	if len(b.uuid) == 0 {
		b.uuid = uuid.New()
	}

	hardwareInfo, err := b.hardware.Info()

	if b.ssh == nil && b.cfg.SSH() != nil {
		b.ssh, err = NewSSH(b.cfg.SSH())
		if err != nil {
			cancel()
			return nil, err
		}
	}

	if err != nil {
		cancel()
		return nil, err
	}

	log.G(ctx).Info("collected Hardware info", zap.Any("hardware", hardwareInfo))

	grpcServer := grpc.NewServer()

	deleter, err := initializeControlGroup(b.cfg.HubResources())
	if err != nil {
		cancel()
		return nil, err
	}

	if !platformSupportCGroups && b.cfg.HubResources() != nil {
		log.G(ctx).Warn("your platform does not support CGroup, but the config has resources section")
	}

	m := &Miner{
		ctx:        ctx,
		cancel:     cancel,
		grpcServer: grpcServer,
		ovs:        b.ovs,

		name:      b.uuid,
		hardware:  hardwareInfo,
		resources: resource.NewPool(hardwareInfo),

		pubAddress: b.ip.String(),
		hubAddress: b.cfg.HubEndpoint(),

		rl:             newReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
		nameMapping:    make(map[string]string),

		controlGroup: deleter,
		ssh:          b.ssh,
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}

func NewMinerBuilder(cfg Config) MinerBuilder {
	b := MinerBuilder{}
	b.Config(cfg)
	return b
}
