package antifraud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

func mklog(n int) string {
	var s string
	for i := 0; i < n; i++ {
		s += fmt.Sprintf("ETH - Total Speed: %d.000 Mh/s, Total Shares: 127, Rejected: 0, Time: 00:02\n", i)
	}
	return s
}

func newTestProcessor() *logProcessor {
	return &logProcessor{
		log:      zap.NewNop(),
		hashrate: atomic.NewFloat64(0),
		cfg: &LogProcessorConfig{
			Pattern:    "Total Speed:",
			Field:      4,
			Multiplier: 1000000,
		},
	}
}

func TestClaymoreLogParser(t *testing.T) {
	rd := strings.NewReader(`ETH - Total Speed: 100.000 Mh/s, Total Shares: 127, Rejected: 0, Time: 00:02`)
	p := newTestProcessor()

	p.logParser(context.Background(), rd)
	assert.Equal(t, float64(100e6), p.hashrate.Load(), "new value should be parsed and set")
}

func TestClaymoreLogParser_InvalidLine(t *testing.T) {
	rd := strings.NewReader(`Oops! Claymore failed`)
	p := newTestProcessor()
	p.hashrate = atomic.NewFloat64(100500)

	p.logParser(context.Background(), rd)
	assert.Equal(t, float64(100500), p.hashrate.Load(), "previous value should be kept")
}

func TestClaymoreLogParser_ShortLine(t *testing.T) {
	rd := strings.NewReader(`Total Speed:`)
	p := newTestProcessor()

	p.logParser(context.Background(), rd)
	assert.Equal(t, float64(0), p.hashrate.Load(), "previous value should be kept")
}

func TestClaymoreLogParser_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	rd := strings.NewReader(mklog(1000))
	p := &logProcessor{log: zap.NewNop(), hashrate: atomic.NewFloat64(1.2345)}
	cancel()

	p.logParser(ctx, rd)
	assert.Equal(t, float64(1.2345), p.hashrate.Load(), "previous value should be kept")
}
