package ctxlog

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerTagType struct{}
type sugaredloggerTagType struct{}

var (
	// G is a shorcut for GetLogger
	G = GetLogger
	// S is a shortcut for GetSugaredLogger
	S = GetSugaredLogger

	loggerTag        = loggerTagType{}
	sugaredloggerTag = sugaredloggerTagType{}

	// TraceBitLevelEnabler controls enabled log level for requests with TraceBit
	// It's set to DebugLevel by default
	TraceBitLevelEnabler = zap.NewAtomicLevel()
)

func init() {
	TraceBitLevelEnabler.SetLevel(zapcore.DebugLevel)
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
}

// WithLogger attaches logger to a given context. Later the logger can be
// obtained by GetLogger
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	// At first we attach SugaredLogger as it's not used so often
	ctx = context.WithValue(ctx, sugaredloggerTag, logger.Sugar())
	return context.WithValue(ctx, loggerTag, logger)
}

// WithTraceBitLogger attaches TraceBitLogger based on a logger from context
func WithTraceBitLogger(ctx context.Context) context.Context {
	traceBitLogger := withTraceBitCore(GetLogger(ctx))
	return WithLogger(ctx, traceBitLogger)
}

// GetLogger either returns an attached Logger from the context
// or global logger if nothing is attached
func GetLogger(ctx context.Context) *zap.Logger {
	l := ctx.Value(loggerTag)
	if l == nil {
		return zap.L()
	}
	return l.(*zap.Logger)
}

// GetSugaredLogger either returns an attached SugaredLogger from the context
// or global sugared logger if nothing is attached
func GetSugaredLogger(ctx context.Context) *zap.SugaredLogger {
	l := ctx.Value(sugaredloggerTag)
	if l == nil {
		return zap.S()
	}
	return l.(*zap.SugaredLogger)
}

var traceBitCoreOption = zap.WrapCore(
	func(core zapcore.Core) zapcore.Core {
		return traceBitCore{
			Core: core,
		}
	})

func withTraceBitCore(l *zap.Logger) *zap.Logger {
	return l.WithOptions(traceBitCoreOption)
}

type traceBitCore struct {
	zapcore.Core
}

func (t traceBitCore) Enabled(lvl zapcore.Level) bool {
	return TraceBitLevelEnabler.Enabled(lvl)
}

func (t traceBitCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if TraceBitLevelEnabler.Enabled(ent.Level) {
		return ce.AddCore(ent, t)
	}
	return ce
}

func (t traceBitCore) With(f []zapcore.Field) zapcore.Core {
	t.Core = t.Core.With(f)
	return t
}
