package tc

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

const (
	Continue ConformExceed = iota
	Drop
	Shot
	Ok
	Pass
	Pipe

	// These values are stolen from linux kernel.
	EgressRedir   MirredActionType = 1
	EgressMirror                   = 2
	IngressRedir                   = 3
	IngressMirror                  = 4
)

type ConformExceed int32

func (m ConformExceed) String() string {
	switch m {
	case Continue:
		return "continue"
	case Drop:
		return "drop"
	case Shot:
		return "shot"
	case Ok:
		return "ok"
	case Pass:
		return "pass"
	case Pipe:
		return "pipe"
	default:
		return fmt.Sprintf("0x%x", int32(m))
	}
}

type Action interface {
	// Kind returns this action's type.
	Kind() string
	// Cmd creates and returns "tc" sub-command using this filter action.
	Cmd() []string
}

type PoliceAction struct {
	// The maximum traffic rate of packets passing this action.
	Rate uint64
	// Set the maximum allowed burst in bytes.
	Burst uint64
	// This is the maximum packet size handled by the policer.
	MTU uint32
	// PeakRate sets the maximum bucket depletion rate, exceeding rate.
	PeakRate uint64
	// Overhead accounts for protocol overhead of encapsulating output devices
	// when computing rate and peakrate.
	Overhead uint64
	// Define how to handle packets which exceed or conform the configured
	// bandwidth limit.
	ConformExceed ConformExceed
}

func (m *PoliceAction) Kind() string {
	return "police"
}

func (m *PoliceAction) Cmd() []string {
	args := []string{
		"action", m.Kind(),
		"rate", fmt.Sprintf("%d", m.Rate),
		"burst", fmt.Sprintf("%d", m.Burst),
	}

	if m.MTU != 0 {
		args = append(args, "mtu", fmt.Sprintf("%d", m.MTU))
	}
	if m.PeakRate != 0 {
		args = append(args, "peakrate", fmt.Sprintf("%d", m.PeakRate))
	}
	args = append(args, "conform-exceed", m.ConformExceed.String())

	return args
}

type MirredActionType int

func (m MirredActionType) String() string {
	switch m {
	case EgressRedir:
		return "egress redirect"
	case EgressMirror:
		return "egress mirror"
	case IngressRedir:
		return "ingress redir"
	case IngressMirror:
		return "ingress mirror"
	default:
		return fmt.Sprintf("unknown mirred action type: %d", int(m))
	}
}

// MirredAction is the action that allows packet mirroring or redirecting
// the packet it receives.
type MirredAction struct {
	Dev    netlink.Link
	Action MirredActionType
}

func (m *MirredAction) Kind() string {
	return "mirred"
}

func (m *MirredAction) Cmd() []string {
	return []string{
		"action", m.Kind(),
		m.Action.String(),
		"dev", m.Dev.Attrs().Name,
	}
}

// Note that lower Priority means a higher priority.
type FilterAttrs struct {
	Link     netlink.Link
	Parent   Handle
	Priority uint16
	Protocol Protocol
}

type Filter interface {
	// Kind returns this filter's type.
	Kind() string
	// Attrs returns basic attributes of this filter.
	Attrs() FilterAttrs
	// Cmd creates and returns "tc" sub-command using this filter.
	Cmd() []string
}

// U32 represents the Universal/Ugly 32bit filter classifier.
type U32 struct {
	FilterAttrs

	FlowID   Handle
	Selector U32Key
	Actions  []Action
}

func (m *U32) Kind() string {
	return "u32"
}

func (m *U32) Attrs() FilterAttrs {
	return m.FilterAttrs
}

func (m *U32) Cmd() []string {
	args := []string{
		"parent", m.Parent.String(),
		"protocol", m.Protocol.String(),
		"prio", fmt.Sprintf("%d", m.Priority),
		m.Kind(),
		"match", "u32", fmt.Sprintf("%x", m.Selector.Val), fmt.Sprintf("%x", m.Selector.Mask),
	}

	for _, action := range m.Actions {
		args = append(args, action.Cmd()...)
	}

	if m.FlowID.UInt32() != 0 {
		args = append(args, "flowid", m.FlowID.String())
	}

	return args
}

type U32Key struct {
	Val     uint32
	Mask    uint32
	Off     int
	OffMask int
}
