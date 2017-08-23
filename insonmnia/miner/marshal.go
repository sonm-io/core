package miner

import (
	"github.com/ccding/go-stun/stun"

	pb "github.com/sonm-io/core/proto"
)

func marshalNATType(nat stun.NATType) pb.NATType {
	switch nat {
	case stun.NATNone:
		return pb.NATType_NONE
	case stun.NATError, stun.NATUnknown:
		return pb.NATType_UNKNOWN
	case stun.NATBlocked:
		return pb.NATType_BLOCKED
	case stun.NATFull:
		return pb.NATType_FULL
	case stun.NATSymetric:
		return pb.NATType_SYMMETRIC
	case stun.NATRestricted:
		return pb.NATType_RESTRICTED
	case stun.NATPortRestricted:
		return pb.NATType_PORT_RESTRICTED
	case stun.NATSymetricUDPFirewall:
		return pb.NATType_SYMMETRIC_UDP_FIREWALL
	}

	return pb.NATType_UNKNOWN
}
