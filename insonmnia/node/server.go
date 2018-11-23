package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/lucas-clemente/quic-go"
	"github.com/sonm-io/core/util/defergroup"
	"github.com/sonm-io/core/util/rest"
	"github.com/sonm-io/core/util/xgrpc"
	"github.com/sonm-io/core/util/xnet"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type LocalEndpoints struct {
	GRPC []net.Addr
	QRPC []net.Addr
	REST []net.Addr
}

type serverNetwork struct {
	mu            sync.Mutex
	ListenersGRPC []net.Listener
	ListenersQRPC []net.PacketConn
	ListenersREST []net.Listener
}

func newServerNetwork(listenersGRPC []net.Listener, listenersQRPC []net.PacketConn, listenersREST []net.Listener) *serverNetwork {
	m := &serverNetwork{
		ListenersGRPC: listenersGRPC,
		ListenersQRPC: listenersQRPC,
		ListenersREST: listenersREST,
	}

	return m
}

func (m *serverNetwork) LocalEndpoints() LocalEndpoints {
	return LocalEndpoints{
		GRPC: toLocalAddrs(m.ListenersGRPC),
		QRPC: m.localQRPCAddrs(),
		REST: toLocalAddrs(m.ListenersREST),
	}
}

func (m *serverNetwork) localQRPCAddrs() []net.Addr {
	var addrs []net.Addr
	for id := range m.ListenersQRPC {
		addrs = append(addrs, m.ListenersQRPC[id].LocalAddr())
	}

	return addrs
}

func (m *serverNetwork) Pop() *serverNetwork {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ListenersGRPC == nil && m.ListenersQRPC == nil && m.ListenersREST == nil {
		return nil
	}

	network := &serverNetwork{
		ListenersGRPC: m.ListenersGRPC,
		ListenersQRPC: m.ListenersQRPC,
		ListenersREST: m.ListenersREST,
	}

	m.ListenersGRPC = nil
	m.ListenersQRPC = nil
	m.ListenersREST = nil

	return network
}

type Services interface {
	RegisterGRPC(server *grpc.Server) error
	RegisterREST(server *rest.Server) error
	Interceptor() grpc.UnaryServerInterceptor
	StreamInterceptor() grpc.StreamServerInterceptor
	Run(ctx context.Context) error
}

type SSHServer interface {
	Serve(ctx context.Context) error
}

// Server is a server for LocalNode instance.
//
// Its responsibility is to manage network part, i.e. creating TCP servers and
// exposing API provided outside.
type Server struct {
	network   *serverNetwork
	endpoints LocalEndpoints

	// Node API.
	services Services

	// Servers for processing requests.
	serverGRPC *grpc.Server
	serverREST *rest.Server
	serverSSH  SSHServer

	tlsConfig *tls.Config

	log *zap.SugaredLogger
}

// NewServer creates new Local Node server instance.
//
// The provided "services" describes Node API, while "options" specifies how
// those API should be exposed.
//
// Note, that you MUST call "Serve", otherwise allocated resources will leak.
func newServer(cfg nodeConfig, services Services, options ...ServerOption) (*Server, error) {
	opts := newServerOptions()
	for _, o := range options {
		if err := o(opts); err != nil {
			return nil, err
		}
	}

	dg := defergroup.DeferGroup{}
	defer dg.Exec()

	listenersGRPC, err := xnet.ListenLoopback("tcp", cfg.BindPort)
	if err != nil {
		return nil, err
	}
	dg.Defer(func() { closeListeners(listenersGRPC) })

	listenersQRPC, err := xnet.ListenPacketLoopback("udp", cfg.BindPort)
	if err != nil {
		return nil, err
	}
	dg.Defer(func() { closePacketConns(listenersQRPC) })

	listenersREST, err := xnet.ListenLoopback("tcp", cfg.HttpBindPort)
	if err != nil {
		return nil, err
	}
	dg.Defer(func() { closeListeners(listenersREST) })

	serverNetwork := newServerNetwork(listenersGRPC, listenersQRPC, listenersREST)
	m := &Server{
		network:   serverNetwork,
		endpoints: serverNetwork.LocalEndpoints(),
		services:  services,
		serverSSH: opts.sshProxy,

		tlsConfig: opts.TLSConfig,

		log: opts.log.Sugar(),
	}

	if opts.allowGRPC {
		options := append(opts.optionsGRPC, []xgrpc.ServerOption{
			xgrpc.DefaultTraceInterceptor(),
			xgrpc.RequestLogInterceptor([]string{"PushTask", "PullTask"}),
			xgrpc.VerifyInterceptor(),
			xgrpc.UnaryServerInterceptor(services.Interceptor()),
			xgrpc.StreamServerInterceptor(services.StreamInterceptor()),
		}...)

		m.serverGRPC = xgrpc.NewServer(m.log.Desugar(), options...)
		if err := services.RegisterGRPC(m.serverGRPC); err != nil {
			return nil, err
		}

		m.log.Infow("registered gRPC services", zap.Any("services", xgrpc.Services(m.serverGRPC)))
	}

	if opts.allowREST {
		options := append([]rest.Option{
			rest.WithLog(opts.log),
			rest.WithInterceptors(
				xgrpc.OpenTracingZapUnaryInterceptor(),
				xgrpc.RequestLogUnaryInterceptor(map[string]bool{}),
				xgrpc.VerifyUnaryInterceptor(),
				services.Interceptor(),
			),
		}, opts.optionsREST...)

		m.serverREST = rest.NewServer(options...)
		if err := services.RegisterREST(m.serverREST); err != nil {
			return nil, err
		}

		m.log.Infow("registered REST services", zap.Any("services", m.serverREST.Services()))
	}

	if opts.exposeGRPCMetrics {
		grpc_prometheus.Register(m.serverGRPC)
		m.log.Info("registered gRPC metrics collector")
	}

	dg.CancelExec()

	return m, nil
}

func (m *Server) LocalEndpoints() LocalEndpoints {
	return m.endpoints
}

// Serve starts serving current Node instance until either critical error
// occurs or the given context is canceled.
//
// Warning: after this call the next callings of "Serve" will have no effect.
func (m *Server) Serve(ctx context.Context) error {
	network := m.network.Pop()
	if network == nil {
		return nil
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return m.serveGRPC(ctx, network.ListenersGRPC...)
	})
	wg.Go(func() error {
		return m.serveQUIC(ctx, network.ListenersQRPC...)
	})
	wg.Go(func() error {
		return m.serveHTTP(ctx, network.ListenersREST...)
	})
	wg.Go(func() error {
		return m.services.Run(ctx)
	})
	wg.Go(func() error {
		return m.serverSSH.Serve(ctx)
	})

	<-ctx.Done()

	m.close()

	return wg.Wait()
}

func (m *Server) serveGRPC(ctx context.Context, listeners ...net.Listener) error {
	if m.serverGRPC == nil {
		return nil
	}

	wg := errgroup.Group{}

	for id := range listeners {
		listener := listeners[id]

		wg.Go(func() error {
			m.log.Infof("exposing gRPC server on %s", listener.Addr().String())
			return m.serverGRPC.Serve(listener)
		})
	}

	defer m.log.Infof("stopped gRPC server on %s", formatListeners(listeners))

	return wg.Wait()
}

func (m *Server) serveQUIC(ctx context.Context, conns ...net.PacketConn) error {
	if m.tlsConfig == nil {
		return nil
	}

	wg := errgroup.Group{}

	for id := range conns {
		listener, err := quic.Listen(conns[id], m.tlsConfig, xnet.DefaultQUICConfig())
		if err != nil {
			return err
		}

		wg.Go(func() error {
			quicListener := &xnet.QUICListener{Listener: listener}
			m.log.Infof("exposing QUIC gRPC server on %s", quicListener.Addr().String())
			defer m.log.Infof("stopped QUIC gRPC server on %s", quicListener.Addr().String())

			return m.serverGRPC.Serve(quicListener)
		})
	}

	return wg.Wait()
}

func (m *Server) serveHTTP(ctx context.Context, listeners ...net.Listener) error {
	if m.serverREST == nil {
		return nil
	}

	defer m.log.Infof("stopped REST server on %s", formatListeners(listeners))

	go func() {
		<-ctx.Done()
		for _, listener := range listeners {
			listener.Close()
		}
	}()

	return m.serverREST.Serve(listeners...)
}

func (m *Server) close() {
	if m.serverGRPC != nil {
		m.serverGRPC.Stop()
	}
	if m.serverREST != nil {
		m.serverREST.Close()
	}
}

func toLocalAddrs(listeners []net.Listener) []net.Addr {
	var addrs []net.Addr
	for id := range listeners {
		addrs = append(addrs, listeners[id].Addr())
	}

	return addrs
}

func formatListeners(listeners []net.Listener) string {
	var addrs []string
	for _, addr := range toLocalAddrs(listeners) {
		addrs = append(addrs, addr.String())
	}

	return fmt.Sprintf("[%s]", strings.Join(addrs, ", "))
}

func closeListeners(listeners []net.Listener) {
	for id := range listeners {
		listeners[id].Close()
	}
}

func closePacketConns(connections []net.PacketConn) {
	for id := range connections {
		connections[id].Close()
	}
}
