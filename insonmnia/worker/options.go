package worker

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/matcher"
	"github.com/sonm-io/core/insonmnia/state"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/multierror"
	"github.com/sonm-io/core/util/netutil"
	"github.com/sonm-io/core/util/xgrpc"
	"google.golang.org/grpc/credentials"
)

const (
	ethereumPrivateKeyKey = "ethereum_private_key"
	exportKeystorePath    = "/var/lib/sonm/worker_keystore"
)

type options struct {
	version     string
	cfg         *Config
	ctx         context.Context
	ovs         Overseer
	ssh         SSH
	key         *ecdsa.PrivateKey
	publicIPs   []string
	benchmarks  benchmarks.BenchList
	storage     *state.Storage
	eth         blockchain.API
	dwh         sonm.DWHClient
	creds       credentials.TransportCredentials
	certRotator util.HitlessCertRotator
	plugins     *plugin.Repository
	whitelist   Whitelist
	matcher     matcher.Matcher
}

func (m *options) validate() error {
	err := multierror.NewMultiError()

	if m.cfg == nil {
		err = multierror.Append(err, errors.New("config is mandatory for Worker options"))
	}

	if m.storage == nil {
		err = multierror.Append(err, errors.New("state storage is mandatory"))
	}

	return err.ErrorOrNil()
}

func (m *options) SetupDefaults() error {
	if err := m.validate(); err != nil {
		return err
	}

	if m.ctx == nil {
		m.ctx = context.Background()
	}

	if err := m.setupKey(); err != nil {
		return err
	}

	if err := m.setupBlockchainAPI(); err != nil {
		return err
	}

	if err := m.setupPlugins(); err != nil {
		return err
	}

	if err := m.setupCreds(); err != nil {
		return err
	}

	if err := m.setupDWH(); err != nil {
		return err
	}

	if err := m.setupWhitelist(); err != nil {
		return err
	}

	if err := m.setupMatcher(); err != nil {
		return err
	}

	if err := m.setupBenchmarks(); err != nil {
		return err
	}

	if err := m.setupNetworkOptions(); err != nil {
		return err
	}

	if err := m.setupSSH(); err != nil {
		return err
	}

	if err := m.setupOverseer(); err != nil {
		return err
	}

	return nil
}

func (m *options) setupKey() error {
	if m.key == nil {
		var data []byte
		loaded, err := m.storage.Load(ethereumPrivateKeyKey, &data)
		if err != nil {
			return err
		}
		if !loaded {
			key, err := crypto.GenerateKey()
			if err != nil {
				return err
			}
			if err := m.storage.Save(ethereumPrivateKeyKey, crypto.FromECDSA(key)); err != nil {
				return err
			}
			m.key = key
		} else {
			key, err := crypto.ToECDSA(data)
			if err != nil {
				return err
			}
			m.key = key
		}
	}
	if err := m.exportKey(); err != nil {
		return err
	}
	return nil
}

func (m *options) exportKey() error {
	if err := os.MkdirAll(exportKeystorePath, 0700); err != nil {
		return err
	}
	ks := keystore.NewKeyStore(exportKeystorePath, keystore.LightScryptN, keystore.LightScryptP)
	if !ks.HasAddress(crypto.PubkeyToAddress(m.key.PublicKey)) {
		_, err := ks.ImportECDSA(m.key, "sonm")
		return err
	}
	return nil
}

func (m *options) setupBlockchainAPI() error {
	if m.eth == nil {
		eth, err := blockchain.NewAPI(m.ctx, blockchain.WithConfig(m.cfg.Blockchain))
		if err != nil {
			return err
		}
		m.eth = eth
	}
	return nil
}

func (m *options) setupPlugins() error {
	plugins, err := plugin.NewRepository(m.ctx, m.cfg.Plugins)
	if err != nil {
		return err
	}
	m.plugins = plugins
	return nil
}

func (m *options) setupCreds() error {
	if m.creds == nil {
		if m.certRotator != nil {
			return errors.New("have certificate rotator in options, but do not have credentials")
		}
		certRotator, TLSConfig, err := util.NewHitlessCertRotator(m.ctx, m.key)
		if err != nil {
			return err
		}
		m.certRotator = certRotator
		m.creds = util.NewTLS(TLSConfig)
	}
	return nil
}

func (m *options) setupDWH() error {
	if m.dwh == nil {
		cc, err := xgrpc.NewClient(m.ctx, m.cfg.DWH.Endpoint, m.creds)
		if err != nil {
			return err
		}
		m.dwh = sonm.NewDWHClient(cc)
	}
	return nil
}

func (m *options) setupWhitelist() error {
	if m.whitelist == nil {
		cfg := m.cfg.Whitelist
		if len(cfg.PrivilegedAddresses) == 0 {
			cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, crypto.PubkeyToAddress(m.key.PublicKey).Hex())
			cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, m.cfg.Master.Hex())
			if m.cfg.Admin != nil {
				cfg.PrivilegedAddresses = append(cfg.PrivilegedAddresses, m.cfg.Admin.Hex())
			}
		}

		m.whitelist = NewWhitelist(m.ctx, &cfg)
	}
	return nil
}

func (m *options) setupMatcher() error {
	if m.matcher == nil {
		if m.cfg.Matcher != nil {
			matcher, err := matcher.NewMatcher(&matcher.Config{
				Key:        m.key,
				DWH:        m.dwh,
				Eth:        m.eth,
				PollDelay:  m.cfg.Matcher.PollDelay,
				QueryLimit: m.cfg.Matcher.QueryLimit,
				Log:        ctxlog.S(m.ctx),
			})
			if err != nil {
				return fmt.Errorf("cannot create matcher: %v", err)
			}
			m.matcher = matcher
		} else {
			m.matcher = matcher.NewDisabledMatcher()
		}
	}
	return nil
}

func (m *options) setupBenchmarks() error {
	if m.benchmarks == nil {
		benchList, err := benchmarks.NewBenchmarksList(m.ctx, m.cfg.Benchmarks)
		if err != nil {
			return err
		}
		m.benchmarks = benchList
	}
	return nil
}

func (m *options) setupNetworkOptions() error {
	// Use public IPs from config (if provided).
	publicIPs := m.cfg.PublicIPs
	if len(publicIPs) > 0 {
		m.publicIPs = SortedIPs(publicIPs)
		return nil
	}

	// Scan interfaces if there's no config and no NAT.
	rawPublicIPs, err := netutil.GetPublicIPs()
	if err != nil {
		return err
	}

	for _, ip := range rawPublicIPs {
		publicIPs = append(publicIPs, ip.String())
	}
	m.publicIPs = SortedIPs(publicIPs)

	return nil
}

func (m *options) setupSSH() error {
	if m.ssh == nil {
		m.ssh = nilSSH{}
	}
	return nil
}

func (m *options) setupOverseer() error {
	if m.ovs == nil {
		ovs, err := NewOverseer(m.ctx, m.plugins)
		if err != nil {
			return err
		}
		m.ovs = ovs
	}
	return nil
}

type Option func(*options)

func WithConfig(cfg *Config) Option {
	return func(opts *options) {
		opts.cfg = cfg
	}
}

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
		opts.benchmarks = list
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

func WithDWH(d sonm.DWHClient) Option {
	return func(o *options) {
		o.dwh = d
	}
}

func WithCreds(creds credentials.TransportCredentials) Option {
	return func(o *options) {
		o.creds = creds
	}
}

func WithVersion(v string) Option {
	return func(o *options) {
		o.version = v
	}
}

func WithCertRotator(certRotator util.HitlessCertRotator) Option {
	return func(o *options) {
		o.certRotator = certRotator
	}
}

func WithPlugins(plugins *plugin.Repository) Option {
	return func(o *options) {
		o.plugins = plugins
	}
}

func WithWhitelist(whitelist Whitelist) Option {
	return func(o *options) {
		o.whitelist = whitelist
	}
}

func WithMatcher(matcher matcher.Matcher) Option {
	return func(o *options) {
		o.matcher = matcher
	}
}
