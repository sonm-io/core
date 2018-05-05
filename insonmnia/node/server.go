package node

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/benchmarks"
	"github.com/sonm-io/core/insonmnia/npp"
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

type hubClientCreator func(ethAddr common.Address, netAddr string) (*hubClient, io.Closer, error)

type hubClient struct {
	pb.HubClient
	pb.WorkerManagementClient
}

// remoteOptions describe options related to remove gRPC services
type remoteOptions struct {
	ctx               context.Context
	key               *ecdsa.PrivateKey
	conf              *Config
	creds             credentials.TransportCredentials
	eth               blockchain.API
	dwh               pb.DWHClient
	hubCreator        hubClientCreator
	blockchainTimeout time.Duration
	nppDialer         *npp.Dialer
	benchList         benchmarks.BenchList
}

func newRemoteOptions(ctx context.Context, key *ecdsa.PrivateKey, cfg *Config, credentials credentials.TransportCredentials) (*remoteOptions, error) {
	nppDialerOptions := []npp.Option{
		npp.WithRendezvous(cfg.NPP.Rendezvous.Endpoints, credentials),
		npp.WithRelayClient(cfg.NPP.Relay.Endpoints, crypto.PubkeyToAddress(key.PublicKey)),
	}
	nppDialer, err := npp.NewDialer(ctx, nppDialerOptions...)
	if err != nil {
		return nil, err
	}

	hubFactory := func(ethAddr common.Address, netAddr string) (*hubClient, io.Closer, error) {
		addr := auth.NewAddrRaw(ethAddr, netAddr)
		conn, err := nppDialer.Dial(addr)
		if err != nil {
			return nil, nil, err
		}

		cc, err := xgrpc.NewClient(ctx, "-", auth.NewWalletAuthenticator(credentials, ethAddr), xgrpc.WithConn(conn))
		if err != nil {
			return nil, nil, err
		}

		m := &hubClient{
			pb.NewHubClient(cc),
			pb.NewWorkerManagementClient(cc),
		}

		return m, cc, nil
	}

	dwhCC, err := xgrpc.NewClient(ctx, cfg.DWH.Endpoint, credentials)
	if err != nil {
		return nil, err
	}

	eth, err := blockchain.NewAPI(blockchain.WithConfig(cfg.Blockchain))
	if err != nil {
		return nil, err
	}

	benchList, err := benchmarks.NewBenchmarksList(ctx, cfg.Benchmarks)
	if err != nil {
		return nil, err
	}

	return &remoteOptions{
		ctx:               ctx,
		key:               key,
		conf:              cfg,
		creds:             credentials,
		eth:               eth,
		dwh:               pb.NewDWHClient(dwhCC),
		blockchainTimeout: 180 * time.Second,
		hubCreator:        hubFactory,
		nppDialer:         nppDialer,
		benchList:         benchList,
	}, nil
}

// Node is LocalNode instance
type Node struct {
	cfg     *Config
	ctx     context.Context
	cancel  context.CancelFunc
	privKey *ecdsa.PrivateKey

	// servers for processing requests
	httpSrv *rest.Server
	srv     *grpc.Server

	// services, responsible for request handling
	hub    pb.WorkerManagementServer
	market pb.MarketServer
	deals  pb.DealManagementServer
	tasks  pb.TaskManagementServer
	master pb.MasterManagementServer
}

// New creates new Local Node instance
// also method starts internal gRPC client connections
// to the external services like Market and Hub
func New(ctx context.Context, config *Config, key *ecdsa.PrivateKey) (*Node, error) {
	ctx, cancel := context.WithCancel(ctx)
	_, TLSConfig, err := util.NewHitlessCertRotator(ctx, key)
	if err != nil {
		return nil, err
	}

	remoteCreds := util.NewTLS(TLSConfig)
	opts, err := newRemoteOptions(ctx, key, config, remoteCreds)
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

	masterMgmt := newMasterManagementAPI(opts)

	grpcServerOpts := []xgrpc.ServerOption{
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.UnaryServerInterceptor(hub.(*hubAPI).intercept),
		xgrpc.VerifyInterceptor(),
	}

	if !config.Node.AllowInsecureConnection {
		grpcServerOpts = append(grpcServerOpts, xgrpc.Credentials(remoteCreds))
	} else {
		log.G(ctx).Warn("using insecure grpc connection")
	}

	srv := xgrpc.NewServer(log.GetLogger(ctx), grpcServerOpts...)

	pb.RegisterWorkerManagementServer(srv, hub)
	log.G(ctx).Info("hub service registered", zap.String("endpt", config.Hub.Endpoint))

	pb.RegisterMarketServer(srv, market)
	log.G(ctx).Info("market service registered")

	pb.RegisterDealManagementServer(srv, deals)
	log.G(ctx).Info("deals service registered")

	pb.RegisterTaskManagementServer(srv, tasks)
	log.G(ctx).Info("tasks service registered")

	pb.RegisterMasterManagementServer(srv, masterMgmt)
	log.G(ctx).Info("master keys service registered")

	grpc_prometheus.Register(srv)

	return &Node{
		privKey: key,
		cfg:     config,
		ctx:     ctx,
		cancel:  cancel,
		srv:     srv,
		hub:     hub,
		market:  market,
		deals:   deals,
		tasks:   tasks,
		master:  masterMgmt,
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
	// todo: how to deal with background processing into brave new blockchain world?
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

	v4err := serve("tcp4", fmt.Sprintf("127.0.0.1:%d", n.cfg.Node.BindPort))
	v6err := serve("tcp6", fmt.Sprintf("[::1]:%d", n.cfg.Node.BindPort))

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

	lis6, err := net.Listen("tcp6", fmt.Sprintf("[::1]:%d", n.cfg.Node.HttpBindPort))
	if err == nil {
		log.S(n.ctx).Info("created ipv6 listener for http")
		options = append(options, rest.WithListener(lis6))
	}

	lis4, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", n.cfg.Node.HttpBindPort))
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
	err = srv.RegisterService((*pb.WorkerManagementServer)(nil), n.hub)
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
	err = srv.RegisterService((*pb.MasterManagementServer)(nil), n.master)
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
