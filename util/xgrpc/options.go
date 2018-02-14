// Following strange gRPC tradition client options are constructed using `With`
// prefix, while server options - are not.

package xgrpc

import (
	"fmt"
	"net"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/sonm-io/core/insonmnia/auth"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerOption func(options *options)

// Credentials activates credentials for server connections.
func Credentials(creds credentials.TransportCredentials) ServerOption {
	return func(o *options) {
		if creds != nil {
			o.options = append(o.options, grpc.Creds(creds))

			// These interceptors will add peer wallet to the context to be
			// able to audit.
			o.interceptors.u = append(o.interceptors.u, walletAuditUnaryInterceptor())
			o.interceptors.s = append(o.interceptors.s, walletAuditStreamInterceptor())
		}
	}
}

func TraceInterceptor(tracer opentracing.Tracer) ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u,
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			openTracingZapUnaryInterceptor(),
		)
		o.interceptors.s = append(o.interceptors.s,
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			openTracingZapStreamInterceptor(),
		)
	}
}

func DefaultTraceInterceptor() ServerOption {
	return TraceInterceptor(basictracer.New(basictracer.NewInMemoryRecorder()))
}

// UnaryServerInterceptor adds an unary interceptor to the chain.
func UnaryServerInterceptor(u grpc.UnaryServerInterceptor) ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, u)
	}
}

type options struct {
	options      []grpc.ServerOption
	interceptors struct {
		u []grpc.UnaryServerInterceptor
		s []grpc.StreamServerInterceptor
	}
}

func newOptions(logger *zap.Logger, extraOpts ...ServerOption) *options {
	opts := &options{}

	// These must be set before other interceptors to avoid losing logs
	// because of improper interceptor implementation.
	if logger != nil {
		opts.interceptors.u = append(opts.interceptors.u, grpc_zap.UnaryServerInterceptor(logger))
		opts.interceptors.s = append(opts.interceptors.s, grpc_zap.StreamServerInterceptor(logger))
	}

	opts.interceptors.u = append(
		opts.interceptors.u, grpc.UnaryServerInterceptor(grpc_prometheus.UnaryServerInterceptor))
	opts.interceptors.s = append(
		opts.interceptors.s, grpc.StreamServerInterceptor(grpc_prometheus.StreamServerInterceptor))

	for _, o := range extraOpts {
		o(opts)
	}

	opts.options = append(opts.options, grpc.StatsHandler(&Handler{}))

	return opts
}

func contextUnaryInterceptor(fn func(ctx context.Context) context.Context) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := fn(ctx)
		if newCtx != nil {
			ctx = newCtx
		}

		return handler(ctx, req)
	}
}

func contextStreamInterceptor(fn func(ctx context.Context) context.Context) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := fn(ss.Context())
		if ctx == nil {
			return handler(srv, ss)
		}

		wrappedStream := grpc_middleware.WrapServerStream(ss)
		wrappedStream.WrappedContext = ctx
		return handler(srv, wrappedStream)
	}
}

func walletAuditUnaryInterceptor() grpc.UnaryServerInterceptor {
	return contextUnaryInterceptor(wrapZapContextWithWallet)
}

func walletAuditStreamInterceptor() grpc.StreamServerInterceptor {
	return contextStreamInterceptor(wrapZapContextWithWallet)
}

func wrapZapContextWithWallet(ctx context.Context) context.Context {
	wallet, err := auth.ExtractWalletFromContext(ctx)
	if err == nil {
		ctx_zap.AddFields(ctx, zap.String("peer.wallet", wallet.Hex()))
	}
	return ctx
}

func openTracingZapUnaryInterceptor() grpc.UnaryServerInterceptor {
	return contextUnaryInterceptor(wrapZapContextWithTracing)
}

func openTracingZapStreamInterceptor() grpc.StreamServerInterceptor {
	return contextStreamInterceptor(wrapZapContextWithTracing)
}

func wrapZapContextWithTracing(ctx context.Context) context.Context {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	spanContext, ok := span.Context().(basictracer.SpanContext)
	if ok && spanContext.Sampled {
		ctx_zap.AddFields(ctx, zap.String("span.trace_id", hex(spanContext.TraceID)))
		ctx_zap.AddFields(ctx, zap.String("span.span_id", hex(spanContext.SpanID)))
	}
	return ctx
}

func hex(v interface{}) string {
	return fmt.Sprintf("%x", v)
}

func authUnaryInterceptor(router *auth.AuthRouter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := router.Authorize(ctx, auth.Event(info.FullMethod), req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func AuthorizationInterceptor(router *auth.AuthRouter) ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, authUnaryInterceptor(router))
		// TODO: Stream interceptors.
		// o.interceptors.s = append(o.interceptors.s, authStreamInterceptor(router))
	}
}

// WithConn is a client option that specifies a predefined connection used for
// a service.
func WithConn(conn net.Conn) grpc.DialOption {
	return grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
		return conn, nil
	})
}
