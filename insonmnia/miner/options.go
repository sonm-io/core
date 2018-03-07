package miner

import (
	"crypto/ecdsa"
	"net"

	"github.com/ccding/go-stun/stun"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pkg/errors"
	"github.com/sonm-io/core/insonmnia/hardware"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type options struct {
	ctx           context.Context
	hardware      hardware.Info
	nat           stun.NATType
	ovs           Overseer
	uuid          string
	ssh           SSH
	key           *ecdsa.PrivateKey
	publicIPs     []string
	locatorClient pb.LocatorClient
	listener      net.Listener
	insecure      bool
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

func WithHardware(hardwareInfo hardware.Info) Option {
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

func WithListener(listener net.Listener) Option {
	return func(opts *options) {
		opts.listener = listener
	}
}

func WithInsecure(val bool) Option {
	return func(opts *options) {
		opts.insecure = val
	}
}

func makeCgroupManager(cfg *ResourcesConfig) (cGroup, cGroupManager, error) {
	if !platformSupportCGroups || cfg == nil {
		return newNilCgroupManager()
	}
	return newCgroupManager(cfg.Cgroup, cfg.Resources)
}
