package logging

import (
	"context"
	"fmt"

	"github.com/opentracing/basictracer-go"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const (
	traceIdFieldKey = "span.trace_id"
	spanIdFieldKey  = "span.span_id"
)

func WithTrace(ctx context.Context, log *zap.Logger) *zap.Logger {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return log
	}

	spanContext, ok := span.Context().(basictracer.SpanContext)
	if ok && spanContext.Sampled {
		log = log.With(
			zap.String(traceIdFieldKey, fmt.Sprintf("%x", spanContext.TraceID)),
			zap.String(spanIdFieldKey, fmt.Sprintf("%x", spanContext.SpanID)),
		)
	}

	return log
}
