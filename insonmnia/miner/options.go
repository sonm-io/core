package miner

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"time"

	"github.com/ccding/go-stun/stun"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/hardware"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type options struct {
	ctx           context.Context
	hardware      hardware.HardwareInfo
	nat           stun.NATType
	ovs           Overseer
	uuid          string
	ssh           SSH
	key           *ecdsa.PrivateKey
	publicIPs     []string
	locatorClient pb.LocatorClient
}

func (o *options) getHubConnectionInfo(cfg Config) (auth.Endpoint, error) {
	var (
		encounteredErrors = make(map[string]error)
		endpoints         []string
	)

	if cfg.HubResolveEndpoints() {
		resolved, err := o.locatorClient.Resolve(o.ctx,
			&pb.ResolveRequest{EthAddr: cfg.HubEthAddr(), EndpointType: pb.ResolveRequest_WORKER})
		if err != nil {
			return auth.Endpoint{}, fmt.Errorf("failed to resolve hub addr from %s: %s", cfg.HubEthAddr(), err)
		}

		log.G(o.ctx).Info("resolved hub endpoints", zap.Any("endpoints", resolved.Endpoints))

		endpoints = resolved.Endpoints
	} else {
		endpoints = cfg.HubEndpoints()
	}

	for _, addr := range endpoints {
		log.G(o.ctx).Debug("trying hub endpoint", zap.Any("endpoint", addr))

		dialer := net.Dialer{DualStack: true, Timeout: time.Second}
		testCC, err := dialer.DialContext(o.ctx, "tcp", addr)
		if err != nil {
			log.G(o.ctx).Debug(
				"hub endpoint is unreachable", zap.Any("endpoint", addr),
				zap.Any("error", err))
			encounteredErrors[addr] = err
		} else {
			testCC.Close()
			endpoint, err := auth.NewEndpoint(fmt.Sprintf("%s@%s", cfg.HubEthAddr(), addr))
			return *endpoint, err
		}
	}

	return auth.Endpoint{}, fmt.Errorf("all hub endpoints are unreachable: %+v", encounteredErrors)
}

func (o *options) setupNetworkOptions(cfg Config) error {
	var pubIPs []string

	// Discover IP if we're behind a NAT.
	if cfg.Firewall() != nil {
		log.G(o.ctx).Debug("discovering public IP address with NAT type, this might be slow")

		client := stun.NewClient()
		if cfg.Firewall().Server != "" {
			client.SetServerAddr(cfg.Firewall().Server)
		}

		nat, addr, err := client.Discover()
		if err != nil {
			return err
		}

		pubIPs = append(pubIPs, addr.IP())
		o.nat, o.publicIPs = nat, SortedIPs(pubIPs)

		return nil
	}

	o.nat = stun.NATNone

	// Use public IPs from config (if provided).
	pubIPs = cfg.PublicIPs()
	if len(pubIPs) > 0 {
		o.publicIPs = SortedIPs(pubIPs)
		return nil
	}

	// Scan interfaces if there's no config and no NAT.
	systemIPs, err := util.GetAvailableIPs()
	if err != nil {
		return err
	}

	for _, ip := range systemIPs {
		pubIPs = append(pubIPs, ip.String())
	}
	if len(pubIPs) > 0 {
		o.publicIPs = SortedIPs(pubIPs)
		return nil
	}

	return errors.New("failed to get public IPs")
}

type Option func(*options)

func WithContext(ctx context.Context) Option {
	return func(opts *options) {
		opts.ctx = ctx
	}
}

func WithHardware(hardwareInfo hardware.HardwareInfo) Option {
	return func(opts *options) {
		opts.hardware = hardwareInfo
	}
}

func WithNat(nat stun.NATType) Option {
	return func(opts *options) {
		opts.nat = nat
	}
}

func WithOverseer(ovs Overseer) Option {
	return func(opts *options) {
		opts.ovs = ovs
	}
}

func WithUUID(uuid string) Option {
	return func(opts *options) {
		opts.uuid = uuid
	}
}

func WithSSH(ssh SSH) Option {
	return func(opts *options) {
		opts.ssh = ssh
	}
}

func WithKey(key *ecdsa.PrivateKey) Option {
	return func(opts *options) {
		opts.key = key
	}
}

func WithLocatorClient(locatorClient pb.LocatorClient) Option {
	return func(opts *options) {
		opts.locatorClient = locatorClient
	}
}

func makeCgroupManager(cfg *ResourcesConfig) (cGroup, cGroupManager, error) {
	if !platformSupportCGroups || cfg == nil {
		return newNilCgroupManager()
	}
	return newCgroupManager(cfg.Cgroup, cfg.Resources)
}
