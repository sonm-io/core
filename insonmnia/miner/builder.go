package miner

import (
	"golang.org/x/net/context"

	"net"

	"github.com/ccding/go-stun/stun"
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
	nat      stun.NATType
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

	log.G(b.ctx).Debug("building a miner", zap.Any("config", b.cfg))

	if b.hardware == nil {
		b.hardware = hardware.New()
	}

	if b.ip == nil {
		if b.cfg.Firewall() == nil {
			log.G(b.ctx).Debug("discovering public IP address ...")
			addr, err := util.GetPublicIP()
			if err != nil {
				return nil, err
			}

			b.ip = addr
			b.nat = stun.NATNone
		} else {
			log.G(b.ctx).Debug("discovering public IP address with NAT type, this may take a long ...")

			client := stun.NewClient()
			if b.cfg.Firewall().Server != "" {
				client.SetServerAddr(b.cfg.Firewall().Server)
			}
			nat, addr, err := client.Discover()
			if err != nil {
				return nil, err
			}
			b.ip = net.ParseIP(addr.IP())
			b.nat = nat
		}

		log.G(b.ctx).Info("discovered public IP address",
			zap.Any("addr", b.ip),
			zap.Any("nat", b.nat),
		)
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

	grpcServer := grpc.NewServer(grpc.RPCCompressor(grpc.NewGZIPCompressor()), grpc.RPCDecompressor(grpc.NewGZIPDecompressor()))

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
		natType:    b.nat,
		hubAddress: b.cfg.HubEndpoint(),

		rl:             newReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
		nameMapping:    make(map[string]string),

		controlGroup: deleter,
		ssh:          b.ssh,

		connectedHubs: make(map[string]struct{}),
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}

func NewMinerBuilder(cfg Config) MinerBuilder {
	b := MinerBuilder{}
	b.Config(cfg)
	return b
}
