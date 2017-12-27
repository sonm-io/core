package miner

import (
	"crypto/ecdsa"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ccding/go-stun/stun"
	"github.com/ethereum/go-ethereum/common"
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

func (o *options) getHubConnectionInfo(cfg Config) (string, common.Address, error) {
	var (
		hubEndpoint                  = cfg.HubEndpoint()
		encounteredErrors            = make(map[string]error)
		hubSockAddr, hubEthAddr, err = auth.ParseEndpoint(cfg.HubEndpoint())
	)

	if err != nil {
		return "", common.Address{}, err
	}

	if strings.HasPrefix(hubSockAddr, ":") {
		// Only hub's port is provided.
		resolved, err := o.locatorClient.Resolve(o.ctx, &pb.ResolveRequest{EthAddr: hubEthAddr.Hex()})
		if err != nil {
			return "", common.Address{}, fmt.Errorf(
				"failed to resolve hub addr from %s: %s", hubEndpoint, err)
		}

		log.G(o.ctx).Info("resolved hub endpoints", zap.Any("endpoints", resolved.IpAddr))

		for _, addr := range resolved.IpAddr {
			addr = strings.Split(addr, ":")[0] + hubSockAddr

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
				return addr, hubEthAddr, nil
			}
		}

		return "", common.Address{}, fmt.Errorf("all hub endpoints are unreachable: %+v", encounteredErrors)
	}

	if _, _, err := net.SplitHostPort(hubSockAddr); err != nil {
		return "", common.Address{}, err
	}

	return hubSockAddr, hubEthAddr, err
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
