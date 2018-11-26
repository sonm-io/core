package tensor

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ethereum/go-ethereum/common"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/blockchain"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	JupyterImage = "sonm.io/jupyter:latest"
)

type Server struct {
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	cfg         *Config
	key         *ecdsa.PrivateKey
	certRotator util.HitlessCertRotator
	creds       credentials.TransportCredentials
	logger      *zap.Logger
	blockchain  blockchain.API
	grpc        *grpc.Server
	http        *rest.Server
}

func NewServer(ctx context.Context, cfg *Config, key *ecdsa.PrivateKey) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &Server{
		ctx:    ctx,
		cancel: cancel,
		cfg:    cfg,
		key:    key,
		logger: log.GetLogger(ctx),
	}, nil
}

func (m *Server) Serve() error {
	//m.logger.Info("starting with backend", zap.String("endpoint", m.cfg.Storage.Endpoint))
	var err error
	//m.db, err = sql.Open("postgres", m.cfg.Storage.Endpoint)
	//if err != nil {
	//	m.Stop()
	//	return err
	//}

	bch, err := blockchain.NewAPI(m.ctx, blockchain.WithConfig(m.cfg.Blockchain))
	if err != nil {
		m.Stop()
		return fmt.Errorf("failed to create NewAPI: %v", err)
	}
	m.blockchain = bch

	//m.storage = newPostgresStorage(numBenchmarks)

	wg := errgroup.Group{}
	wg.Go(m.serveGRPC)
	wg.Go(m.serveHTTP)

	return wg.Wait()
}

func (m *Server) serveGRPC() error {
	lis, err := func() (net.Listener, error) {
		m.mu.Lock()
		defer m.mu.Unlock()

		certRotator, TLSConfig, err := util.NewHitlessCertRotator(m.ctx, m.key)
		if err != nil {
			return nil, err
		}

		m.certRotator = certRotator
		m.creds = util.NewTLS(TLSConfig)
		m.grpc = xgrpc.NewServer(
			m.logger,
			xgrpc.Credentials(m.creds),
			xgrpc.DefaultTraceInterceptor(),
		)
		sonm.RegisterTensorServer(m.grpc, m)
		grpc_prometheus.Register(m.grpc)

		lis, err := net.Listen("tcp", m.cfg.GRPCListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen on %s: %v", m.cfg.GRPCListenAddr, err)
		}

		return lis, nil
	}()
	if err != nil {
		return err
	}

	return m.grpc.Serve(lis)
}

func (m *Server) serveHTTP() error {
	lis, err := func() (net.Listener, error) {
		m.mu.Lock()
		defer m.mu.Unlock()

		options := []rest.Option{rest.WithLog(m.logger)}
		lis, err := net.Listen("tcp", m.cfg.HTTPListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create http listener: %v", err)
		}

		srv := rest.NewServer(options...)

		err = srv.RegisterService((*sonm.DWHServer)(nil), m)
		if err != nil {
			return nil, fmt.Errorf("failed to RegisterService: %v", err)
		}
		m.http = srv

		return lis, err
	}()
	if err != nil {
		return err
	}

	return m.http.Serve(lis)
}

func (m *Server) Stop() error {
	return nil
}

func (m *Server) GetJupyterNode(ctx context.Context, req *sonm.JupyterNodeRequest) (*sonm.JupyterNodeResponse, error) {
	wc, err := m.getWorkerClient(req.SupplierID.Unwrap())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to getWorkerClient")
	}

	repl, err := wc.StartTask(m.ctx, &sonm.StartTaskRequest{
		DealID: req.DealID,
		Spec: &sonm.TaskSpec{
			Container: &sonm.Container{
				Image:        JupyterImage,
				CommitOnStop: true,
				Env:          map[string]string{"repo": req.Repository},
			},
		},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to StartTask")
	}

	return &sonm.JupyterNodeResponse{}, nil
}

func (m *Server) getWorkerClient(addr common.Address) (sonm.WorkerClient, error) {
	cc, err := xgrpc.NewClient(m.ctx, addr.String(), m.creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get gRPC client: %v", err)
	}

	return sonm.NewWorkerClient(cc), nil
}
