package xgrpc

import (
	"context"
	"crypto/tls"

	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/lucas-clemente/quic-go"
	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util/xnet"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newTracer() opentracing.Tracer {
	return basictracer.NewWithOptions(basictracer.Options{
		ShouldSample:   func(traceID uint64) bool { return true },
		MaxLogsPerSpan: 100,
		Recorder:       newNoOpRecorder(),
	})
}

// NewClient creates new gRPC client connection on given addr and wraps it
// with given credentials (if provided).
//
// The address argument can be optionally used as other peer's verification
// using ETH authentication. To enable this the argument should be in
// format "ethAddr@Endpoint".
func NewClient(ctx context.Context, addr string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	authEndpoint, err := auth.ParseAddr(addr)
	if err != nil {
		return nil, err
	}

	netAddr, err := authEndpoint.Addr()
	if err != nil {
		return nil, err
	}

	ethAddr, err := authEndpoint.ETH()
	if err != nil {
		return newClient(ctx, addr, credentials, opts...)
	}

	return newClient(ctx, netAddr, auth.NewWalletAuthenticator(credentials, ethAddr), opts...)
}

func newClient(ctx context.Context, addr string, credentials credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var secureOpt = grpc.WithInsecure()
	if credentials != nil {
		secureOpt = grpc.WithTransportCredentials(credentials)
	}

	var extraOpts = append(opts, secureOpt,
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(newTracer()))),
		grpc.WithStreamInterceptor(grpc_opentracing.StreamClientInterceptor()),
	)
	cc, err := grpc.DialContext(ctx, addr, extraOpts...)
	if err != nil {
		return nil, err
	}
	return cc, err
}

func NewQUICClient(ctx context.Context, addr string, tlsConfig *tls.Config, credentials credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	authEndpoint, err := auth.ParseAddr(addr)
	if err != nil {
		return nil, err
	}

	netAddr, err := authEndpoint.Addr()
	if err != nil {
		return nil, err
	}

	session, err := quic.DialAddr(netAddr, tlsConfig, xnet.DefaultQUICConfig())
	if err != nil {
		return nil, err
	}

	conn, err := xnet.NewQUICConn(session)
	if err != nil {
		return nil, err
	}

	opts = append(opts, WithConn(conn))

	ethAddr, err := authEndpoint.ETH()
	if err != nil {
		return newClient(ctx, addr, credentials, opts...)
	}

	return newClient(ctx, netAddr, auth.NewWalletAuthenticator(credentials, ethAddr), opts...)
}
