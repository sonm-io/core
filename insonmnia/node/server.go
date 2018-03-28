package node

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type hubClientCreator func(addr string) (pb.HubClient, io.Closer, error)

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
	locatorCC, err := xgrpc.NewWalletAuthenticatedClient(ctx, creds, conf.LocatorEndpoint())
	if err != nil {
		return nil, err
	}

	marketCC, err := xgrpc.NewWalletAuthenticatedClient(ctx, creds, conf.MarketEndpoint())
	if err != nil {
		return nil, err
	}

	bcAPI, err := blockchain.NewAPI(nil, nil)
	if err != nil {
		return nil, err
	}

	hc := func(addr string) (pb.HubClient, io.Closer, error) {
		cc, err := xgrpc.NewClient(ctx, addr, creds)
		if err != nil {
			return nil, nil, err
		}

		return pb.NewHubClient(cc), cc, nil
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
	cfg     Config
	ctx     context.Context
	cancel  context.CancelFunc
	privKey *ecdsa.PrivateKey

	// servers for processing requests
	httpSrv *rest.Server
	srv     *grpc.Server

	// services, responsible for request handling
	hub    pb.HubManagementServer
	market pb.MarketServer
	deals  pb.DealManagementServer
	tasks  pb.TaskManagementServer
}

// New creates new Local Node instance
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func New(ctx context.Context, c Config, key *ecdsa.PrivateKey) (*Node, error) {
	ctx, cancel := context.WithCancel(ctx)
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

	logger := log.GetLogger(ctx)
	srv := xgrpc.NewServer(
		logger,
		// Intentionally constructing an unencrypted server.
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.UnaryServerInterceptor(hub.(*hubAPI).intercept),
	)

	pb.RegisterHubManagementServer(srv, hub)
	log.G(ctx).Info("hub service registered", zap.String("endpt", c.HubEndpoint()))

	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered", zap.String("endpt", c.MarketEndpoint()))

	pb.RegisterDealManagementServer(srv, deals)
	log.G(ctx).Info("deals service registered")

	pb.RegisterTaskManagementServer(srv, tasks)
	log.G(ctx).Info("tasks service registered")

	grpc_prometheus.Register(srv)

	return &Node{
		privKey: key,
		cfg:     c,
		ctx:     ctx,
		cancel:  cancel,
		srv:     srv,
		hub:     hub,
		market:  market,
		deals:   deals,
		tasks:   tasks,
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
	err := n.market.(*marketAPI).restartOrdersProcessing()
	if err != nil {
		// should it breaks startup?
		log.G(n.ctx).Warn("cannot restart order processing", zap.Error(err))
	}
	wg := errgroup.Group{}
	wg.Go(n.ServeHttp)
	wg.Go(n.ServeGRPC)

	return wg.Wait()
}

func (n *Node) ServeGRPC() error {
	wg := errgroup.Group{}

	serve := func(netFam, laddr string) error {
		lis, err := net.Listen(netFam, laddr)
		if err == nil {
			log.S(n.ctx).Infof("starting node %s listener on %s", netFam, lis.Addr().String())
			wg.Go(func() error {
				err := n.srv.Serve(lis)
				n.Close()
				return err
			})
		} else {
			log.S(n.ctx).Warnf("cannot create %s listener - %s", netFam, err)
		}
		return err
	}

	v4err := serve("tcp4", fmt.Sprintf("127.0.0.1:%d", n.cfg.BindPort()))
	v6err := serve("tcp6", fmt.Sprintf("[::1]:%d", n.cfg.BindPort()))

	if v4err != nil && v6err != nil {
		n.Close()
		return errors.New("neither ipv4 nor ipv6 localhost is available to bind")
	}
	return wg.Wait()
}

func (n *Node) ServeHttp() error {
	err := n.serveHttp()
	n.Close()
	return err
}

func (n *Node) serveHttp() error {
	aesKey := []byte{}
	h := sha256.New()
	h.Write(n.privKey.D.Bytes())
	aesKey = h.Sum(aesKey)
	decenc, err := rest.NewAESDecoderEncoder(aesKey)
	if err != nil {
		return err
	}

	options := []rest.Option{rest.WithContext(n.ctx), rest.WithDecoder(decenc), rest.WithEncoder(decenc), rest.WithInterceptor(n.hub.(*hubAPI).intercept)}

	lis6, err := net.Listen("tcp6", fmt.Sprintf("[::1]:%d", n.cfg.HttpBindPort()))
	if err == nil {
		log.S(n.ctx).Info("created ipv6 listener for http")
		options = append(options, rest.WithListener(lis6))
	}

	lis4, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", n.cfg.HttpBindPort()))
	if err == nil {
		log.S(n.ctx).Info("created ipv4 listener for http")
		options = append(options, rest.WithListener(lis4))
	}

	if lis4 == nil && lis6 == nil {
		return errors.New("could not listen http")
	}
	srv, err := rest.NewServer(options...)
	if err != nil {
		return err
	}
	err = srv.RegisterService((*pb.HubManagementServer)(nil), n.hub)
	if err != nil {
		return err
	}
	err = srv.RegisterService((*pb.MarketServer)(nil), n.market)
	if err != nil {
		return err
	}
	err = srv.RegisterService((*pb.DealManagementServer)(nil), n.deals)
	if err != nil {
		return err
	}
	err = srv.RegisterService((*pb.TaskManagementServer)(nil), n.tasks)
	if err != nil {
		return err
	}
	n.httpSrv = srv
	return srv.Serve()
}

func (n *Node) Close() {
	n.cancel()
	if n.httpSrv != nil {
		n.httpSrv.Close()
	}
	if n.srv != nil {
		n.srv.Stop()
	}
}
