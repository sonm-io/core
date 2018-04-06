package miner

import (
	"crypto/ecdsa"

	"github.com/pkg/errors"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/dwh"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/util"
	"golang.org/x/net/context"
)

type options struct {
	ctx       context.Context
	ovs       Overseer
	ssh       SSH
	key       *ecdsa.PrivateKey
	publicIPs []string
	benchList benchmarks.BenchList
	storage   *state.Storage
	eth       blockchain.API
	dwh       dwh.DWH
}

func (o *options) setupNetworkOptions(cfg *Config) error {
	// Use public IPs from config (if provided).
	pubIPs := cfg.PublicIPs
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

func WithOverseer(ovs Overseer) Option {
	return func(opts *options) {
		opts.ovs = ovs
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

func WithBenchmarkList(list benchmarks.BenchList) Option {
	return func(opts *options) {
		opts.benchList = list
	}

}

func WithStateStorage(s *state.Storage) Option {
	return func(o *options) {
		o.storage = s
	}
}

func WithETH(e blockchain.API) Option {
	return func(o *options) {
		o.eth = e
	}
}

func WithDWH(d dwh.DWH) Option {
	return func(o *options) {
		o.dwh = d
	}
}
