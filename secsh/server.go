package secsh

import (
	"context"
	"crypto/ecdsa"
	"crypto/tls"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/insonmnia/npp"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/secsh/secshc"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/xgrpc"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type execStream struct {
	Server sonm.RemotePTY_ExecServer
}

func (m *execStream) Write(b []byte) (int, error) {
	chunk := &sonm.RemotePTYExecResponseChunk{
		Out:  b,
		Done: false,
	}

	if err := m.Server.Send(chunk); err != nil {
		return 0, err
	}

	return len(b), nil
}

type RemotePTYServer struct {
	cfg        *Config
	privateKey *ecdsa.PrivateKey
	log        *zap.SugaredLogger
}

func NewRemotePTYServer(cfg *Config, log *zap.SugaredLogger) (*RemotePTYServer, error) {
	key, err := cfg.Eth.LoadKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load Ethereum keys: %v", err)
	}

	m := &RemotePTYServer{
		cfg:        cfg,
		privateKey: key,
		log:        log,
	}

	return m, nil
}

func (m *RemotePTYServer) Run(ctx context.Context) error {
	m.log.Infof("ETH address: %s", crypto.PubkeyToAddress(m.privateKey.PublicKey).Hex())

	certRotator, tlsConfig, err := util.NewHitlessCertRotator(ctx, m.privateKey)
	if err != nil {
		return err
	}

	defer certRotator.Close()

	server := m.makeServer(tlsConfig)
	service := &RemotePTYService{
		execPath:   m.cfg.SecExecPath,
		policyPath: m.cfg.SeccompPolicyDir,
		log:        m.log,
	}

	sonm.RegisterRemotePTYServer(server, service)

	listener, err := npp.NewListener(ctx, "127.0.0.1:0",
		npp.WithProtocol(secshc.Protocol),
		npp.WithNPPBacklog(m.cfg.NPP.Backlog),
		npp.WithNPPBackoff(m.cfg.NPP.MinBackoffInterval, m.cfg.NPP.MaxBackoffInterval),
		npp.WithRendezvous(m.cfg.NPP.Rendezvous, xgrpc.NewTransportCredentials(tlsConfig)),
		npp.WithLogger(m.log.Desugar()),
	)
	if err != nil {
		return fmt.Errorf("failed to create NPP listener: %v", err)
	}

	m.log.Infof("exposed remote PTY server on %s", listener.Addr().String())
	defer m.log.Infof("stopped remote PTY server on %s", listener.Addr().String())

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return server.Serve(listener)
	})

	<-ctx.Done()
	if err := listener.Close(); err != nil {
		m.log.Warnw("failed to close TCP listener", zap.Error(err))
	}

	return wg.Wait()
}

func (m *RemotePTYServer) makeServer(tlsConfig *tls.Config) *grpc.Server {
	options := []xgrpc.ServerOption{
		xgrpc.DefaultTraceInterceptor(),
		xgrpc.RequestLogInterceptor([]string{}),
		xgrpc.VerifyInterceptor(),
		xgrpc.Credentials(auth.NewWalletAuthenticator(xgrpc.NewTransportCredentials(tlsConfig), crypto.PubkeyToAddress(m.privateKey.PublicKey))),
	}

	return xgrpc.NewServer(m.log.Desugar(), options...)
}
