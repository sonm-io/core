package logging

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	atom = zap.NewAtomicLevel()
)

// BuildLogger return new zap.Logger instance with given severity and debug settings
func BuildLogger(level Level) *zap.Logger {
	atom.SetLevel(level.Zap())
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
	// LogLevel return log verbosity.
	LogLevel() Level
}

// Level represents a shifted zap logging level that is able to being
// constructed from YAML.
type Level struct {
	level zapcore.Level
}

func NewLevel(level zapcore.Level) *Level {
	return &Level{level}
}

// Zap returns the underlying zap logging level.
func (m Level) Zap() zapcore.Level {
	return m.level
}

func (m *Level) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var level string
	if err := unmarshal(&level); err != nil {
		return err
	}

	v, err := parseLogLevel(level)
	if err != nil {
		return err
	}

	m.level = v

	return nil
}

// ParseLogLevel returns zap logger level by it's name
func parseLogLevel(s string) (zapcore.Level, error) {
	s = strings.ToLower(s)

	lvl := zapcore.DebugLevel
	if err := lvl.Set(s); err != nil {
		return zapcore.DebugLevel, fmt.Errorf("cannot parse config file: \"%s\" is invalid log level", s)
	}

	return lvl, nil
}
