package sonm

import (
	"github.com/ccding/go-stun/stun"
)

func NewNATType(nat stun.NATType) NATType {
	switch nat {
	case stun.NATNone:
		return NATType_NONE
	case stun.NATError, stun.NATUnknown:
		return NATType_UNKNOWN
	case stun.NATBlocked:
		return NATType_BLOCKED
	case stun.NATFull:
		return NATType_FULL
	case stun.NATSymetric:
		return NATType_SYMMETRIC
	case stun.NATRestricted:
		return NATType_RESTRICTED
	case stun.NATPortRestricted:
		return NATType_PORT_RESTRICTED
	case stun.NATSymetricUDPFirewall:
		return NATType_SYMMETRIC_UDP_FIREWALL
	}

	return NATType_UNKNOWN
}
