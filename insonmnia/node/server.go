package node

import (
	"crypto/ecdsa"
	"net"
	"time"

	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type hubClientCreator func(addr string) (pb.HubClient, error)

// remoteOptions describe options related to remove gRPC services
type remoteOptions struct {
	ctx                context.Context
	key                *ecdsa.PrivateKey
	conf               Config
	creds              credentials.TransportCredentials
	locator            pb.LocatorClient
	market             pb.MarketClient
	eth                blockchain.Blockchainer
	hubCreator         hubClientCreator
	dealApproveTimeout time.Duration
	dealCreateTimeout  time.Duration
}

func newRemoteOptions(ctx context.Context, key *ecdsa.PrivateKey, conf Config, creds credentials.TransportCredentials) (*remoteOptions, error) {
	locatorCC, err := util.MakeWalletAuthenticatedClient(ctx, creds, conf.LocatorEndpoint())
	if err != nil {
		return nil, err
	}

	marketCC, err := util.MakeWalletAuthenticatedClient(ctx, creds, conf.MarketEndpoint())
	if err != nil {
		return nil, err
	}

	bcAPI, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	hc := func(addr string) (pb.HubClient, error) {
		cc, err := util.MakeGrpcClient(ctx, addr, creds)
		if err != nil {
			return nil, err
		}

		return pb.NewHubClient(cc), nil
	}

	return &remoteOptions{
		key:                key,
		conf:               conf,
		ctx:                ctx,
		creds:              creds,
		locator:            pb.NewLocatorClient(locatorCC),
		market:             pb.NewMarketClient(marketCC),
		eth:                bcAPI,
		dealApproveTimeout: 900 * time.Second,
		dealCreateTimeout:  180 * time.Second,
		hubCreator:         hc,
	}, nil
}

// Node is LocalNode instance
type Node struct {
	lis     net.Listener
	srv     *grpc.Server
	ctx     context.Context
	privKey *ecdsa.PrivateKey
	// processorRestarter must start together with node .Serve (not .New).
	// This func must fetch orders from the Market and restart it background processing.
	processorRestarter func() error
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

	remoteCreds := util.NewTLS(TLSConfig)
	opts, err := newRemoteOptions(ctx, key, c, remoteCreds)
	if err != nil {
		return nil, err
	}

	hub := newHubAPI(opts)

	market, err := newMarketAPI(opts)
	if err != nil {
		return nil, err
	}

	deals, err := newDealsAPI(opts)
	if err != nil {
		return nil, err
	}

	tasks, err := newTasksAPI(opts)
	if err != nil {
		return nil, err
	}

	addr := util.PubKeyToAddr(key.PublicKey)
	creds := util.NewWalletAuthenticator(util.NewTLS(TLSConfig), addr)
	srv := util.MakeGrpcServer(creds, grpc.UnaryInterceptor(hub.(*hubAPI).intercept))

	pb.RegisterHubManagementServer(srv, hub)
	log.G(ctx).Info("hub service registered", zap.String("endpt", c.HubEndpoint()))

	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered", zap.String("endpt", c.MarketEndpoint()))

	pb.RegisterDealManagementServer(srv, deals)
	log.G(ctx).Info("deals service registered")

	pb.RegisterTaskManagementServer(srv, tasks)
	log.G(ctx).Info("tasks service registered")

	return &Node{
		lis:                lis,
		ctx:                ctx,
		srv:                srv,
		processorRestarter: market.(*marketAPI).restartOrdersProcessing(),
	}, nil
}

type serverStreamMDForwarder struct {
	grpc.ServerStream
}

func (s *serverStreamMDForwarder) Context() context.Context {
	return util.ForwardMetadata(s.ServerStream.Context())
}

func (n *Node) InterceptStreamRequest(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, &serverStreamMDForwarder{ss})
}

// Serve binds gRPC services and start it
func (n *Node) Serve() error {
	// restart background processing
	if n.processorRestarter != nil {
		err := n.processorRestarter()
		if err != nil {
			// should it breaks startup?
			log.G(n.ctx).Error("cannot restart order processing", zap.Error(err))
		}
	}

	return n.srv.Serve(n.lis)
}
