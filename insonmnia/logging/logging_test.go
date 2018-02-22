package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in  string
		out zapcore.Level
	}{
		// any case must be parsed
		{in: "warn", out: zapcore.WarnLevel},
		{in: "Warn", out: zapcore.WarnLevel},
		{in: "WARN", out: zapcore.WarnLevel},
		// any other values must be represented as default debug level
		{in: "-1", out: zapcore.DebugLevel},
		{in: "666", out: zapcore.DebugLevel},
	}

	for _, tt := range tests {
		out := ParseLogLevel(tt.in)
		assert.Equal(t, tt.out, out)
	}
}
