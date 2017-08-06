package miner

import (
	"golang.org/x/net/context"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"google.golang.org/grpc"
	"net"
)

type MinerBuilder struct {
	ctx       context.Context
	cfg       Config
	collector resource.Collector
	ip        net.IP
	ovs       Overseer
}

func (b *MinerBuilder) Context(ctx context.Context) {
	b.ctx = ctx
}

func (b *MinerBuilder) Config(config Config) {
	b.cfg = config
}

func (b *MinerBuilder) Collector(collector resource.Collector) {
	b.collector = collector
}

func (b *MinerBuilder) Address(ip net.IP) {
	b.ip = ip
}

func (b *MinerBuilder) Overseer(ovs Overseer) {
	b.ovs = ovs
}

func (b *MinerBuilder) Build() (miner *Miner, err error) {
	if b.ip == nil {
		addr, err := util.GetPublicIP()
		if err != nil {
			return nil, err
		}
		b.ip = addr
	}

	if b.collector == nil {
		b.collector = resource.New()
	}

	if b.ctx == nil {
		b.ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(b.ctx)

	if b.ovs == nil {
		b.ovs, err = NewOverseer(ctx)
		if err != nil {
			cancel()
			return nil, err
		}
	}

	resources, err := b.collector.OS()
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	deleter, err := initializeControlGroup(b.cfg.HubResources())
	if err != nil {
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

		resources: resources,

		pubAddress: b.ip.String(),
		hubAddress: b.cfg.HubEndpoint(),

		rl:             newReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
		nameMapping:    make(map[string]string),

		controlGroup: deleter,
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}
