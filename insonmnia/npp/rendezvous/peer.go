package rendezvous

import (
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/proto"
	"google.golang.org/grpc/peer"
)

type Peer struct {
	peer.Peer
	ID           PeerID
	privateAddrs []*sonm.Addr
}

func NewPeer(peerInfo peer.Peer, privateAddrs []*sonm.Addr) Peer {
	return Peer{peerInfo, NewPeerID(), privateAddrs}
}

// PeerID represents an unique peer id generated at the time of either
// publishing or resolving another peer.
//
// This id is used for cleaning resources when a request is finished. Note,
// that we cannot use peer's Ethereum address as an id, because it can be
// shared across multiple servers/clients.
// It's also intentionally that
type PeerID string

// NewPeerID constructs and returns a new unique id used for internal
// identifying peers connected for publishing and resolving each other.
func NewPeerID() PeerID {
	return PeerID(uuid.New())
}

func (m PeerID) String() string {
	return string(m)
}
