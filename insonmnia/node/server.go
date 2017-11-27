package node

import (
	"crypto/ecdsa"
	"net"

	log "github.com/noxiouz/zapctx/ctxlog"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// remoteOptions describe options related to remove gRPC service
type remoteOptions struct {
	ctx   context.Context
	key   *ecdsa.PrivateKey
	conf  Config
	creds credentials.TransportCredentials
}

// Node is LocalNode instance
type Node struct {
	ctx     context.Context
	conf    Config
	lis     net.Listener
	srv     *grpc.Server
	privKey *ecdsa.PrivateKey
}

// New creates new Local Node instance
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func New(ctx context.Context, c Config, key *ecdsa.PrivateKey) (*Node, error) {
	lis, err := net.Listen("tcp", c.ListenAddress())
	if err != nil {
		return nil, err
	}

	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	creds := util.NewTLS(TLSConfig)
	srv := util.MakeGrpcServer(creds)

	opts := &remoteOptions{
		key:   key,
		conf:  c,
		ctx:   ctx,
		creds: creds,
	}

	// register hub connection if hub addr is set
	if c.HubEndpoint() != "" {
		hub, err := newHubAPI(opts)
		if err != nil {
			return nil, err
		}
		pb.RegisterHubManagementServer(srv, hub)
		log.G(ctx).Info("hub service registered", zap.String("endpt", c.HubEndpoint()))
	}

	market, err := newMarketAPI(opts)
	if err != nil {
		return nil, err
	}
	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered", zap.String("endpt", c.MarketEndpoint()))

	deals, err := newDealsAPI(opts)
	if err != nil {
		return nil, err
	}
	pb.RegisterDealManagementServer(srv, deals)
	log.G(ctx).Info("deals service registered")

	tasks, err := newTasksAPI(opts)
	if err != nil {
		return nil, err
	}
	pb.RegisterTaskManagementServer(srv, tasks)
	log.G(ctx).Info("tasks service registered")

	return &Node{
		lis:     lis,
		conf:    c,
		ctx:     ctx,
		srv:     srv,
		privKey: key,
	}, nil
}

// Serve binds gRPC services and start it
func (n *Node) Serve() error {
	return n.srv.Serve(n.lis)
}
