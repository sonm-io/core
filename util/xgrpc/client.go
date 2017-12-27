package xgrpc

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/basictracer-go"
	"github.com/sonm-io/core/insonmnia/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewClient creates new gRPC client connection on given addr and wraps it
// with given credentials.
func NewClient(ctx context.Context, addr string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var secureOpt = grpc.WithInsecure()
	if creds != nil {
		secureOpt = grpc.WithTransportCredentials(creds)
	}

	tracer := basictracer.NewWithOptions(basictracer.Options{
		ShouldSample:   func(traceID uint64) bool { return true },
		MaxLogsPerSpan: 100,
		Recorder:       basictracer.NewInMemoryRecorder(),
	})

	var extraOpts = append(opts, secureOpt,
		grpc.WithCompressor(grpc.NewGZIPCompressor()),
		grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
		grpc.WithUnaryInterceptor(grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(tracer))),
		grpc.WithStreamInterceptor(grpc_opentracing.StreamClientInterceptor()),
	)
	cc, err := grpc.DialContext(ctx, addr, extraOpts...)
	if err != nil {
		return nil, err
	}
	return cc, err
}

func NewWalletAuthenticatedClient(ctx context.Context, creds credentials.TransportCredentials, endpoint string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	socketAddr, ethAddr, err := auth.ParseEndpoint(endpoint)
	if err != nil {
		conn, err := NewClient(ctx, endpoint, creds, opts...)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	locatorCreds := auth.NewWalletAuthenticator(creds, ethAddr)

	conn, err := NewClient(ctx, socketAddr, locatorCreds)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
