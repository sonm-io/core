package tc

import (
	"github.com/vishvananda/netlink"
)

const (
	HandleNone    Handle = netlink.HANDLE_NONE
	HandleRoot           = netlink.HANDLE_ROOT
	HandleIngress        = netlink.HANDLE_INGRESS
)

// Handle specifies both qdisc, class and filter ID.
type Handle uint32

// NewHandle computes tc handle based on major and minor parts.
func NewHandle(major uint16, minor uint16) Handle {
	return Handle(netlink.MakeHandle(major, minor))
}

// UInt32 returns internal handle representation as an uint32.
func (m Handle) UInt32() uint32 {
	return uint32(m)
}

func (m Handle) String() string {
	return netlink.HandleStr(m.UInt32())
}

func (m Handle) WithMinor(minor uint16) Handle {
	return NewHandle(uint16((m&0xFFFF0000)>>16), minor)
}
