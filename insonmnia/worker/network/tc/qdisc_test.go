package tc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPfifoCmd(t *testing.T) {
	qdisc := &PfifoQDisc{
		QDiscAttrs: QDiscAttrs{
			Link:   nil,
			Handle: NewHandle(0x8001, 0),
			Parent: HandleRoot,
		},
		Limit: 42,
	}
	assert.Equal(t, []string{"parent", "root", "handle", "8001:0", "pfifo", "limit", "42"}, qdisc.Cmd())
}
