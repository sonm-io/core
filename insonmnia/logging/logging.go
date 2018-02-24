package logging

import (
	"strings"

	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	atom = zap.NewAtomicLevel()
)

// BuildLogger return new zap.Logger instance with given severity and debug settings
func BuildLogger(level zapcore.Level) *zap.Logger {
	atom.SetLevel(level)
	loggerConfig := zap.Config{
		Development:      false,
		Level:            atom,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}

	log, _ := loggerConfig.Build()
	return log
}

type Leveler interface {
	// LogLevel return log verbosity
	LogLevel() (zapcore.Level, error)
}

// ParseLogLevel returns zap logger level by it's name
func ParseLogLevel(s string) (zapcore.Level, error) {
	s = strings.ToLower(s)

	var lvl = zapcore.DebugLevel
	err := lvl.Set(s)
	if err != nil {
		return zapcore.DebugLevel, fmt.Errorf("cannot parse config file: \"%s\" is invalid log level", s)
	}

	return lvl, nil
}
