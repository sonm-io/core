package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in       string
		out      zapcore.Level
		mustFail bool
	}{
		// any case must be parsed
		{in: "warn", out: zapcore.WarnLevel},
		{in: "Warn", out: zapcore.WarnLevel},
		{in: "WARN", out: zapcore.WarnLevel},
		// any other values must be represented as default debug level
		{in: "-1", mustFail: true},
		{in: "5", mustFail: true},
		{in: "666", mustFail: true},
	}

	for _, tt := range tests {
		out, err := parseLogLevel(tt.in)
		if tt.mustFail {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.out, out)
		}
	}
}
