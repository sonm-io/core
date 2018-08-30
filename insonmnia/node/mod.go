package node

import (
	"context"
	"crypto/ecdsa"
	"fmt"

	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/debug"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/credentials"
)

type Node struct {
	cfg    *Config
	server *Server
}

// New constructs a new Node instance.
//
// The specified context should outlive the returned Node value, because it is
// used for certificate refreshing.
func New(ctx context.Context, cfg *Config, options ...Option) (*Node, error) {
	opts := newOptions()
	for _, o := range options {
		o(opts)
	}

	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load Ethereum keys: %v", err)
	}

	transportCredentials, err := newTLS(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	remoteOptions, err := newRemoteOptions(ctx, cfg, key, transportCredentials, opts.log.Sugar())
	if err != nil {
		return nil, err
	}

	log := opts.log
	services := newServices(remoteOptions)

	serverOptions := []ServerOption{
		WithGRPCServer(),
		WithRESTServer(),
		WithGRPCServerMetrics(),
		WithServerLog(log),
	}

	if cfg.Node.AllowInsecureConnection {
		// This is intentional.
		// Enabling insecure mode is disrespectful.
		log.Warn("--- INSECURE SERVER MODE ACTIVATED, YOUR CONNECTIONS WILL **NOT** BE ENCRYPTED ---")
	} else {
		serverOptions = append(serverOptions,
			WithGRPCSecure(transportCredentials, key),
			WithRESTSecure(key),
		)
	}

	if cfg.SSH != nil {
		serverOptions = append(serverOptions, WithSSH(*cfg.SSH, transportCredentials, remoteOptions.eth.Market(), log.Sugar()))
	}

	server, err := newServer(cfg.Node, services, serverOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to build Node instance: %s", err)
	}

	m := &Node{
		cfg:    cfg,
		server: server,
	}

	return m, nil
}

func (m *Node) Serve(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		if m.cfg.Debug == nil {
			return nil
		}

		return debug.ServePProf(ctx, *m.cfg.Debug, m.server.log.Desugar())
	})

	wg.Go(func() error { return m.server.Serve(ctx) })

	return wg.Wait()
}

// NewTLS constructs new transport credentials using specified private key.
// The credentials will be automatically refreshed while the given context
// is active.
// Indented to be used in top-level function with long-living background
// context.
func newTLS(ctx context.Context, privateKey *ecdsa.PrivateKey) (credentials.TransportCredentials, error) {
	_, tlsConfig, err := util.NewHitlessCertRotator(ctx, privateKey)
	return util.NewTLS(tlsConfig), err
}
