package antifraud

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func mklog(n int) string {
	var s string
	for i := 0; i < n; i++ {
		s += fmt.Sprintf("ETH - Total Speed: %d.000 Mh/s, Total Shares: 127, Rejected: 0, Time: 00:02\n", i)
	}
	return s
}

func TestClaymoreLogParser(t *testing.T) {
	rd := strings.NewReader(`ETH - Total Speed: 100.000 Mh/s, Total Shares: 127, Rejected: 0, Time: 00:02`)
	p := &commonLogProcessor{log: zap.NewNop()}

	claymoreLogParser(context.Background(), p, rd)
	assert.Equal(t, float64(100e6), p.hashrate, "new value should be parsed and set")
}

func TestClaymoreLogParser_InvalidLine(t *testing.T) {
	rd := strings.NewReader(`Oops! Claymore failed`)
	p := &commonLogProcessor{log: zap.NewNop(), hashrate: 100500}

	claymoreLogParser(context.Background(), p, rd)
	assert.Equal(t, float64(100500), p.hashrate, "previous value should be kept")
}

func TestClaymoreLogParser_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	rd := strings.NewReader(mklog(1000))
	p := &commonLogProcessor{log: zap.NewNop(), hashrate: 1.2345}
	cancel()

	claymoreLogParser(ctx, p, rd)
	assert.Equal(t, float64(1.2345), p.hashrate, "new value should be parsed and set")
}
