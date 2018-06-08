package tc

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

type ClassAttrs struct {
	Link   netlink.Link
	Handle Handle
	Parent Handle
}

type Class interface {
	// Kind returns this class's type.
	Kind() string
	// Attrs returns basic attributes of this filter.
	Attrs() ClassAttrs
	// Cmd creates and returns "tc" sub-command using this qdisc.
	Cmd() []string
}

type HTBClass struct {
	ClassAttrs
	Rate uint64
	Ceil uint64
}

func (m *HTBClass) Kind() string {
	return "htb"
}

func (m *HTBClass) Attrs() ClassAttrs {
	return m.ClassAttrs
}

func (m *HTBClass) Cmd() []string {
	return []string{
		"parent", m.Parent.String(),
		"classid", m.Handle.String(),
		m.Kind(),
		"rate", fmt.Sprintf("%d", m.Rate),
		"ceil", fmt.Sprintf("%d", m.Ceil),
	}
}
