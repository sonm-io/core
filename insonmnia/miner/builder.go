package miner

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"net"

	"github.com/ccding/go-stun/stun"
	"github.com/ethereum/go-ethereum/common"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/resource"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var (
	errInvalidEndpointFormat = errors.New("endpoint must be in <key>@<endpoint> format")
	errInvalidEthAddrFormat  = errors.New("invalid ETH address format")
)

type MinerBuilder struct {
	ctx           context.Context
	cfg           Config
	hardware      hardware.HardwareInfo
	nat           stun.NATType
	ovs           Overseer
	uuid          string
	ssh           SSH
	key           *ecdsa.PrivateKey
	locatorClient pb.LocatorClient
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

	publicIPs, err := b.getPublicIPs()
	if err != nil {
		return nil, err
	}

	log.G(b.ctx).Info("discovered public IPs",
		zap.Any("public IPs", publicIPs),
		zap.Any("nat", b.nat))

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

	if b.key == nil {
		cancel()
		return nil, fmt.Errorf("ethereum private key must be provided")
	}

	// The rotator will be stopped by ctx
	certRotator, TLSConf, err := util.NewHitlessCertRotator(ctx, b.key)
	if err != nil {
		cancel()
		return nil, err
	}

	hubSockaddr, hubEthAddr, err := b.getHubConnectionInfo()
	if err != nil {
		return nil, err
	}

	creds := util.NewWalletAuthenticator(util.NewTLS(TLSConf), hubEthAddr)
	grpcServer := util.MakeGrpcServer(creds)

	cgroup, cGroupManager, err := makeCgroupManager(b.cfg.HubResources())
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

		publicIPs:  publicIPs,
		natType:    b.nat,
		hubAddress: hubSockaddr,

		rl:             newReverseListener(1),
		containers:     make(map[string]*ContainerInfo),
		statusChannels: make(map[int]chan bool),
		nameMapping:    make(map[string]string),

		controlGroup:  cgroup,
		cGroupManager: cGroupManager,
		ssh:           b.ssh,

		connectedHubs: make(map[string]struct{}),

		certRotator: certRotator,
		creds:       creds,
	}

	pb.RegisterMinerServer(grpcServer, m)
	return m, nil
}

func (b *MinerBuilder) getPublicIPs() (pubIPs []string, err error) {
	// Discover IP if we're behind a NAT.
	if b.cfg.Firewall() != nil {
		log.G(b.ctx).Debug("discovering public IP address with NAT type, this might be slow")

		client := stun.NewClient()
		if b.cfg.Firewall().Server != "" {
			client.SetServerAddr(b.cfg.Firewall().Server)
		}

		nat, addr, err := client.Discover()
		if err != nil {
			return nil, err
		}

		pubIPs = append(pubIPs, addr.IP())
		b.nat = nat

		return SortedIPs(pubIPs), nil
	}

	b.nat = stun.NATNone

	// Use public IPs from config (if provided).
	pubIPs = b.cfg.PublicIPs()
	if len(pubIPs) > 0 {
		return SortedIPs(pubIPs), nil
	}

	// Scan interfaces if there's no config and no NAT.
	systemIPs, err := util.GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	for _, ip := range systemIPs {
		pubIPs = append(pubIPs, ip.String())
	}
	if len(pubIPs) > 0 {
		return SortedIPs(pubIPs), nil
	}

	return nil, errors.New("failed to get public IPs")
}

func (b *MinerBuilder) getHubConnectionInfo() (string, common.Address, error) {
	var hubEndpoint = b.cfg.HubEndpoint()
	if strings.Contains(hubEndpoint, "@") {
		return util.ParseEndpoint(b.cfg.HubEndpoint())
	} else {
		resolved, err := b.locatorClient.Resolve(b.ctx, &pb.ResolveRequest{EthAddr: hubEndpoint})
		if err != nil {
			return "", common.Address{}, fmt.Errorf(
				"failed to resolve hub addr from %s: %s", hubEndpoint, err)
		}

		log.G(b.ctx).Info("resolved hub endpoints", zap.Any("endpoints", resolved.IpAddr))

		for _, addr := range resolved.IpAddr {
			addr = strings.Replace(addr, "10001", "10002", 1)

			log.G(b.ctx).Debug("trying hub endpoint", zap.Any("endpoint", addr))

			dialer := net.Dialer{DualStack: true, Timeout: time.Second}
			testCC, err := dialer.DialContext(b.ctx, "tcp", addr)
			if err != nil {
				log.G(b.ctx).Debug(
					"hub endpoint is unreachable", zap.Any("endpoint", addr),
					zap.Any("error", err))
			} else {
				testCC.Close()
				hubSockaddr, hubEthAddr, err := util.ParseEndpoint(hubEndpoint + "@" + addr)
				if err == nil {
					return hubSockaddr, hubEthAddr, nil
				}
			}
		}

		return "", common.Address{}, errors.New("all hub endpoints are unreachable")
	}
}

func makeCgroupManager(cfg *ResourcesConfig) (cGroup, cGroupManager, error) {
	if !platformSupportCGroups || cfg == nil {
		return newNilCgroupManager()
	}
	return newCgroupManager(cfg.Cgroup, cfg.Resources)
}

func NewMinerBuilder(cfg Config, key *ecdsa.PrivateKey) (*MinerBuilder, error) {
	b := &MinerBuilder{key: key}
	b.Config(cfg)

	if b.ctx == nil {
		b.ctx = context.Background()
	}

	_, TLSConf, err := util.NewHitlessCertRotator(b.ctx, b.key)
	if err != nil {
		return nil, err
	}

	creds := util.NewTLS(TLSConf)

	locatorCC, err := util.MakeWalletAuthenticatedClient(b.ctx, creds, cfg.LocatorEndpoint())
	if err != nil {
		return nil, err
	}

	b.locatorClient = pb.NewLocatorClient(locatorCC)

	return b, nil
}
