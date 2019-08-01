// Following strange gRPC tradition client options are constructed using `With`
// prefix, while server options - are not.

package xgrpc

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
	"github.com/rcrowley/go-metrics"
	"github.com/sonm-io/core/insonmnia/auth"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func GRPCServerOptions(opts ...grpc.ServerOption) ServerOption {
	return func(o *options) {
		o.options = append(o.options, opts...)
	}
}

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
			OpenTracingZapUnaryInterceptor(),
		)
		o.interceptors.s = append(o.interceptors.s,
			grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			openTracingZapStreamInterceptor(),
		)
	}
}

type noopRecorder struct{}

func newNoOpRecorder() *noopRecorder {
	return &noopRecorder{}
}

func (noopRecorder) RecordSpan(span basictracer.RawSpan) {}

func DefaultTraceInterceptor() ServerOption {
	return TraceInterceptor(basictracer.New(newNoOpRecorder()))
}

// UnaryServerInterceptor adds an unary interceptor to the chain.
func UnaryServerInterceptor(u grpc.UnaryServerInterceptor) ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, u)
	}
}

// UnaryServerInterceptor adds an unary interceptor to the chain.
func StreamServerInterceptor(s grpc.StreamServerInterceptor) ServerOption {
	return func(o *options) {
		o.interceptors.s = append(o.interceptors.s, s)
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
		logger = zap.New(logger.Core(), zap.AddStacktrace(zapcore.FatalLevel))

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

func OpenTracingZapUnaryInterceptor() grpc.UnaryServerInterceptor {
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

func authStreamInterceptor(router *auth.AuthRouter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ss = &wrappedAuthStream{
			ServerStream: ss,
			router:       router,
			info:         info,
			processed:    atomic.NewBool(false),
		}

		return handler(srv, ss)
	}
}

type wrappedAuthStream struct {
	grpc.ServerStream
	router    *auth.AuthRouter
	info      *grpc.StreamServerInfo
	processed *atomic.Bool
}

func (m *wrappedAuthStream) RecvMsg(msg interface{}) error {
	if err := m.ServerStream.RecvMsg(msg); err != nil {
		return err
	}

	if m.processed.CAS(false, true) {
		if err := m.router.Authorize(m.ServerStream.Context(), auth.Event(m.info.FullMethod), msg); err != nil {
			return err
		}
	}

	return nil
}

func AuthorizationInterceptor(router *auth.AuthRouter) ServerOption {
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, authUnaryInterceptor(router))
		o.interceptors.s = append(o.interceptors.s, authStreamInterceptor(router))
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
		o.interceptors.u = append(o.interceptors.u, VerifyUnaryInterceptor())
	}
}

func VerifyUnaryInterceptor() grpc.UnaryServerInterceptor {
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
func RequestLogInterceptor(truncatedMethods []string) ServerOption {
	var truncatedMethodsSet = map[string]bool{}
	for _, method := range truncatedMethods {
		truncatedMethodsSet[method] = true
	}
	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, RequestLogUnaryInterceptor(truncatedMethodsSet))
		o.interceptors.s = append(o.interceptors.s, requestLogStreamInterceptor(truncatedMethodsSet))
	}
}

func RequestLogUnaryInterceptor(truncatedMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		truncateRequestField := false
		if methodInfo := ParseMethodInfo(info.FullMethod); methodInfo != nil {
			truncateRequestField = truncatedMethods[methodInfo.Method]
		}

		executeRequestLogging(ctx, req, info.FullMethod, truncateRequestField)
		return handler(ctx, req)
	}
}

func requestLogStreamInterceptor(truncatedMethods map[string]bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ss = &requestLoggingWrappedStream{
			ServerStream: ss,
			fullMethod:   info.FullMethod,
			truncate:     truncatedMethods[ParseMethodInfo(info.FullMethod).Method],
		}

		return handler(srv, ss)
	}
}

func executeRequestLogging(ctx context.Context, req interface{}, method string, truncateRequestField bool) {
	log := ctx_zap.Extract(ctx).Sugar()
	if !truncateRequestField {
		log = log.With(zap.Any("request", reflect.Indirect(reflect.ValueOf(req)).Interface()))
	}

	log.Infof("handling %s request", method)
}

type requestLoggingWrappedStream struct {
	grpc.ServerStream
	fullMethod string
	log        *zap.SugaredLogger
	truncate   bool
}

func (m *requestLoggingWrappedStream) RecvMsg(msg interface{}) error {
	err := m.ServerStream.RecvMsg(msg)

	executeRequestLogging(m.Context(), msg, m.fullMethod, m.truncate)

	return err
}

type rateLimiter struct {
	// Constants.
	defaultLimit  float64
	preciseLimits map[string]float64

	mu           sync.Mutex
	currentRates map[common.Address]metrics.Meter
}

func newRateLimiter(ctx context.Context, defaultLimit float64, preciseLimits map[string]float64) *rateLimiter {
	m := &rateLimiter{
		defaultLimit:  defaultLimit,
		preciseLimits: preciseLimits,
		currentRates:  map[common.Address]metrics.Meter{},
	}

	go m.runGC(ctx)

	return m
}

func (m *rateLimiter) runGC(ctx context.Context) {
	timer := time.NewTicker(5 * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			m.collect()
		}
	}
}

func (m *rateLimiter) collect() {
	const RPH = 1.0 / 3600

	m.mu.Lock()
	defer m.mu.Unlock()

	for addr, meter := range m.currentRates {
		if meter.Count() == 0 {
			continue
		}

		if meter.Rate1() < RPH {
			delete(m.currentRates, addr)
		}
	}
}

func (m *rateLimiter) LimitFor(method string) float64 {
	limit, ok := m.preciseLimits[method]
	if !ok {
		return m.defaultLimit
	}
	return limit
}

func (m *rateLimiter) MeterFor(addr common.Address) metrics.Meter {
	m.mu.Lock()
	defer m.mu.Unlock()

	meter, ok := m.currentRates[addr]
	if !ok {
		meter = metrics.NewMeter()
		m.currentRates[addr] = meter
	}

	return meter
}

func (m *rateLimiter) CheckFor(ctx context.Context, method string) error {
	addr, err := auth.ExtractWalletFromContext(ctx)
	if err != nil {
		return err
	}

	meter := m.MeterFor(*addr)
	currentRate := meter.Rate1()
	thresholdRate := m.LimitFor(method)
	if currentRate >= thresholdRate {
		return status.Errorf(codes.Unavailable, "rate limit reached for %s: %f RPS, while the threshold is %f", addr.Hex(), currentRate, thresholdRate)
	}

	meter.Mark(1)
	return nil
}

// RateLimitInterceptor is an option that enables requests rate limit for
// specified methods.
//
// The context is required, because internally a garbage collector is run to
// periodically collect records for addresses that are inactive for a long
// time.
//
// Rates are counted as exponentially-weighted averaged at one-minutes.
//
// Methods that are not listed in the "preciseLimits" argument will have no rate
// limitations.
func RateLimitInterceptor(ctx context.Context, defaultLimit float64, preciseLimits map[string]float64) ServerOption {
	rateLimiter := newRateLimiter(ctx, defaultLimit, preciseLimits)

	return func(o *options) {
		o.interceptors.u = append(o.interceptors.u, rateLimitUnaryInterceptor(rateLimiter))
		o.interceptors.s = append(o.interceptors.s, rateLimitStreamInterceptor(rateLimiter))
	}
}

func rateLimitUnaryInterceptor(rateLimiter *rateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := rateLimiter.CheckFor(ctx, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func rateLimitStreamInterceptor(rateLimiter *rateLimiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := rateLimiter.CheckFor(ss.Context(), info.FullMethod); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}
