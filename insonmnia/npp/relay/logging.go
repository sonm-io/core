package relay

import (
	"io"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

type logAdapter struct {
	log *zap.Logger
	rx  *regexp.Regexp
}

func newLogAdapter(log *zap.Logger) io.Writer {
	return &logAdapter{
		log: log.WithOptions(zap.AddCallerSkip(3)),
		rx:  regexp.MustCompile(`\[(\w+)] \w+:(.*)`),
	}
}

func (m *logAdapter) Write(p []byte) (int, error) {
	matches := m.rx.FindSubmatch(p[20:])
	if len(matches) != 3 {
		return len(p), nil
	}

	level := string(matches[1])
	message := strings.ToLower(strings.TrimSpace(string(matches[2])))
	switch level {
	case "DEBUG":
		m.log.Debug(message)
	case "INFO":
		m.log.Info(message)
	case "WARN":
		m.log.Warn(message)
	case "ERROR", "FATAL":
		m.log.Error(message)
	default:
		m.log.Info(message)
	}

	return len(p), nil
}
