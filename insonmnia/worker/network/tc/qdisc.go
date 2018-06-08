package tc

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

type QDiscAttrs struct {
	Link   netlink.Link
	Handle Handle
	Parent Handle
}

// QDisc describes traffic control queueing discipline.
type QDisc interface {
	// Type returns this qdisc name, for example "tbf", "htb", etc.
	Type() string
	// Attrs returns basic attributes of this qdisc.
	Attrs() QDiscAttrs
	// Cmd creates and returns "tc" sub-command using this qdisc.
	Cmd() []string
}

// PfifoQDisc represents packet limited First In, First Out queue.
type PfifoQDisc struct {
	QDiscAttrs
	// Limit constrains the queue size as measured in packets.
	Limit uint32
}

func (m *PfifoQDisc) Type() string {
	return "pfifo"
}

func (m *PfifoQDisc) Attrs() QDiscAttrs {
	return m.QDiscAttrs
}

func (m *PfifoQDisc) Cmd() []string {
	return []string{
		"parent", m.Parent.String(),
		"handle", m.Handle.String(),
		m.Type(),
		"limit", fmt.Sprintf("%d", m.Limit),
	}
}

type Ingress struct {
	QDiscAttrs
}

func (m *Ingress) Type() string {
	return "ingress"
}

func (m *Ingress) Attrs() QDiscAttrs {
	return m.QDiscAttrs
}

func (m *Ingress) Cmd() []string {
	return []string{
		m.Parent.String(),
		"handle", m.Handle.String(),
	}
}

type TFBQDisc struct {
	QDiscAttrs
	// Rate in bits per second.
	Rate uint64
	// Size of bucket in bytes.
	Burst uint64
	// Latency in microseconds.
	Latency uint32
}

func (m *TFBQDisc) Type() string {
	return "tbf"
}

func (m *TFBQDisc) Attrs() QDiscAttrs {
	return m.QDiscAttrs
}

func (m *TFBQDisc) Cmd() []string {
	return []string{
		"parent", m.Parent.String(),
		"handle", m.Handle.String(),
		m.Type(),
		"rate", fmt.Sprintf("%d", m.Rate),
		"burst", fmt.Sprintf("%d", m.Burst),
		"latency", fmt.Sprintf("%d", m.Latency),
	}
}

type HTBQDisc struct {
	QDiscAttrs
}

func (m *HTBQDisc) Type() string {
	return "htb"
}

func (m *HTBQDisc) Attrs() QDiscAttrs {
	return m.QDiscAttrs
}

func (m *HTBQDisc) Cmd() []string {
	return []string{
		"parent", m.Parent.String(),
		"handle", m.Handle.String(),
		m.Type(),
	}
}
