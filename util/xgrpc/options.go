// Following strange gRPC tradition client options are constructed using `With`
// prefix, while server options - are not.

package xgrpc

import (
	"context"
	"fmt"
	"net"
	"reflect"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	traceIdFieldKey = "span.trace_id"
	spanIdFieldKey  = "span.span_id"
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
		ctx_zap.AddFields(ctx, zap.String(traceIdFieldKey, hex(spanContext.TraceID)))
		ctx_zap.AddFields(ctx, zap.String(spanIdFieldKey, hex(spanContext.SpanID)))
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

// VerifyInterceptor is an interceptor that performs server-side validation
// for both gRPC request and reply if possible.
//
// It automatically checks whether those types has `Validate` method and
// calls it if so.
func VerifyInterceptor() ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, verifyUnaryInterceptor())
	}
}

func verifyUnaryInterceptor() grpc.UnaryServerInterceptor {
	validate := func(v interface{}) error {
		if v, ok := v.(interface {
			Validate() error
		}); ok {
			if err := v.Validate(); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return nil
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(req); err != nil {
			return nil, err
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := validate(resp); err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// RequestLogInterceptor is an options that activates gRPC service call logging before
// the real execution begins.
//
// Note, that to enable tracing logging you should specify this option AFTER
// trace interceptors.
func RequestLogInterceptor(log *zap.Logger, truncatedMethods []string) ServerOption {
	var truncatedMethodsSet = map[string]bool{}
	for _, method := range truncatedMethods {
		truncatedMethodsSet[method] = true
	}
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, requestLogUnaryInterceptor(log.Sugar(), truncatedMethodsSet))
		o.interceptors.s = append(o.interceptors.s, requestLogStreamInterceptor(log.Sugar(), truncatedMethodsSet))
	}
}

func requestLogUnaryInterceptor(log *zap.SugaredLogger, truncatedMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		executeRequestLogging(ctx, req, info.FullMethod, log, truncatedMethods[MethodInfo(info.FullMethod).Method])
		return handler(ctx, req)
	}
}

func requestLogStreamInterceptor(log *zap.SugaredLogger, truncatedMethods map[string]bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ss = &requestLoggingWrappedStream{
			ServerStream: ss,
			fullMethod:   info.FullMethod,
			log:          log,
			truncate:     truncatedMethods[MethodInfo(info.FullMethod).Method],
		}

		return handler(srv, ss)
	}
}

func executeRequestLogging(ctx context.Context, req interface{}, method string, log *zap.SugaredLogger, truncate bool) {
	service, method := MethodInfo(method).IntoTuple()
	attributes := []interface{}{
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}

	if !truncate {
		attributes = append(attributes, zap.Any("request", reflect.Indirect(reflect.ValueOf(req)).Interface()))
	}

	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		spanContext, ok := span.Context().(basictracer.SpanContext)
		if ok && spanContext.Sampled {
			attributes = append(attributes,
				zap.String(traceIdFieldKey, hex(spanContext.TraceID)),
				zap.String(spanIdFieldKey, hex(spanContext.SpanID)),
			)
		}
	}

	log.With(attributes...).Infof("handling %s request", method)
}

type requestLoggingWrappedStream struct {
	grpc.ServerStream
	fullMethod string
	log        *zap.SugaredLogger
	truncate   bool
}

func (m *requestLoggingWrappedStream) RecvMsg(msg interface{}) error {
	err := m.ServerStream.RecvMsg(msg)

	executeRequestLogging(m.Context(), msg, m.fullMethod, m.log, m.truncate)

	return err
}
