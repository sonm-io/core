package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	atom = zap.NewAtomicLevel()
)

// BuildLogger return new zap.Logger instance with given severity and debug settings
func BuildLogger(level int, development bool) *zap.Logger {
	var encoding string
	var encodingConfig zapcore.EncoderConfig
	if development {
		encoding = "console"
		encodingConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoding = "json"
		encodingConfig = zap.NewProductionEncoderConfig()
	}

	atom.SetLevel(zapcore.Level(level))
	loggerConfig := zap.Config{
		Development:      development,
		Level:            atom,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Encoding:         encoding,
		EncoderConfig:    encodingConfig,
	}

	log, _ := loggerConfig.Build()
	return log
}
