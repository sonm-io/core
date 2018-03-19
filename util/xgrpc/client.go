package xgrpc

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/sonm-io/core/insonmnia/auth"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newTracer() opentracing.Tracer {
	return basictracer.NewWithOptions(basictracer.Options{
		ShouldSample:   func(traceID uint64) bool { return true },
		MaxLogsPerSpan: 100,
		Recorder:       basictracer.NewInMemoryRecorder(),
	})
}

// NewClient creates new gRPC client connection on given addr and wraps it
// with given credentials (if provided).
func NewClient(ctx context.Context, addr string, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	var secureOpt = grpc.WithInsecure()
	if creds != nil {
		secureOpt = grpc.WithTransportCredentials(creds)
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

func NewWalletAuthenticatedClient(ctx context.Context, creds credentials.TransportCredentials, endpoint string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	authEndpoint, err := auth.NewEndpoint(endpoint)
	if err != nil {
		return NewClient(ctx, endpoint, creds, opts...)
	}

	return NewClient(ctx, authEndpoint.Endpoint, auth.NewWalletAuthenticator(creds, authEndpoint.EthAddress), opts...)
}
