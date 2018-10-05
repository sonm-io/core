package node

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
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
	REST []net.Addr
}

type serverNetwork struct {
	mu            sync.Mutex
	ListenersGRPC []net.Listener
	ListenersREST []net.Listener
}

func (m *serverNetwork) Pop() *serverNetwork {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ListenersGRPC == nil && m.ListenersREST == nil {
		return nil
	}

	network := &serverNetwork{
		ListenersGRPC: m.ListenersGRPC,
		ListenersREST: m.ListenersREST,
	}

	m.ListenersGRPC = nil
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
	network   serverNetwork
	endpoints LocalEndpoints

	// Node API.
	services Services

	// Servers for processing requests.
	serverGRPC *grpc.Server
	serverREST *rest.Server
	serverSSH  SSHServer

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

	listenersREST, err := xnet.ListenLoopback("tcp", cfg.HttpBindPort)
	if err != nil {
		return nil, err
	}
	dg.Defer(func() { closeListeners(listenersREST) })

	m := &Server{
		log: opts.log.Sugar(),
		network: serverNetwork{
			ListenersGRPC: listenersGRPC,
			ListenersREST: listenersREST,
		},
		endpoints: LocalEndpoints{
			GRPC: toLocalAddrs(listenersGRPC),
			REST: toLocalAddrs(listenersREST),
		},
		services:  services,
		serverSSH: opts.sshProxy,
	}

	if opts.allowGRPC {
		options := append([]xgrpc.ServerOption{
			xgrpc.DefaultTraceInterceptor(),
			xgrpc.RequestLogInterceptor(m.log.Desugar(), []string{"PushTask", "PullTask"}),
			xgrpc.VerifyInterceptor(),
			xgrpc.UnaryServerInterceptor(services.Interceptor()),
			xgrpc.StreamServerInterceptor(services.StreamInterceptor()),
		}, opts.optionsGRPC...)

		m.serverGRPC = xgrpc.NewServer(m.log.Desugar(), options...)
		if err := services.RegisterGRPC(m.serverGRPC); err != nil {
			return nil, err
		}

		m.log.Infow("registered gRPC services", zap.Any("services", xgrpc.Services(m.serverGRPC)))
	}

	if opts.allowREST {
		options := append([]rest.Option{
			rest.WithLog(opts.log),
			rest.WithInterceptor(services.Interceptor()),
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

// TODO: Compose those three functions into a separate struct.
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
