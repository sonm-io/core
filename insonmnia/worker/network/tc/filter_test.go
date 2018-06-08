package tc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestU32Cmd(t *testing.T) {
	filter := &U32{
		FilterAttrs: FilterAttrs{
			Link:     nil,
			Parent:   NewHandle(0x8001, 0),
			Priority: 7,
			Protocol: ProtoAll,
		},
		FlowID: NewHandle(0x8001, 1),
		Selector: U32Key{
			Val:     0x0,
			Mask:    0x0,
			Off:     0,
			OffMask: 0,
		},
		Actions: []Action{},
	}
	assert.Equal(t, []string{"parent", "8001:0", "protocol", "all", "prio", "7", "u32", "match", "u32", "0", "0", "flowid", "8001:1"}, filter.Cmd())
}

func TestU32CmdIpSelector(t *testing.T) {
	filter := &U32{
		FilterAttrs: FilterAttrs{
			Link:     nil,
			Parent:   NewHandle(0x8001, 0),
			Priority: 7,
			Protocol: ProtoIP,
		},
		FlowID: NewHandle(0x8001, 1),
		Selector: U32Key{
			Val:     0xac000000,
			Mask:    0xff000000,
			Off:     0,
			OffMask: 0,
		},
		Actions: []Action{},
	}
	assert.Equal(t, []string{"parent", "8001:0", "protocol", "ip", "prio", "7", "u32", "match", "u32", "ac000000", "ff000000", "flowid", "8001:1"}, filter.Cmd())
}
