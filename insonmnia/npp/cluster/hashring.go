// This module describes the consistent hash ring used to distribute the relay
// load between multiple discovered servers.
//
// The Continuum evolution.
// 1. [ | | | | | | | | | | | | | | | | ]
// 2. [ | | | |x| | | | | |x| | | | |x| ]
// 3. [ | | | |x| | |K| | |x| | | |U|x| ]
// 4. [v| | | |x| | |K| |v|x| | |v|U|x| ]
// 5. [v| | | | | | |K| |v| | | |v|U| | ]
//
// 1. Initial Continuum.
// 2. After inserting node "x".
// 3. Discovering both key "K" and "U" results in "x".
// 4. After inserting another node "v" discovering result of key "U" left the
// same while discovering key "K" now results in node "v".
// Then K needs to be rediscovered, while U doesn't.
// 5. Removing node "x" results in "U" discovering, but it's done
// automatically, since leaving from the group usually means that the node is
// shutting down.

package cluster

import (
	"sync"

	"github.com/serialx/hashring"
	"github.com/sonm-io/core/insonmnia/npp/nppc"
)

type Continuum struct {
	mu        sync.RWMutex
	continuum *hashring.HashRing
	tracking  map[nppc.ResourceID]string
}

func NewContinuum() *Continuum {
	return &Continuum{
		continuum: hashring.New([]string{}),
	}
}

// Add adds a new weighted node into the Continuum, returning list of ETH
// addresses that are need to be rescheduled.
func (m *Continuum) Add(node string, weight int) []nppc.ResourceID {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.continuum = m.continuum.AddNode(node)

	return m.scanTrackingChanges()
}

// Track starts tracking the given address.
//
// When a new node is inserted into the Continuum the Add method returns the
// list of addresses that must be rescheduled.
func (m *Continuum) Track(addr nppc.ResourceID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.continuum.GetNode(addr.String())
	if ok {
		m.tracking[addr] = node
	} else {
		m.tracking[addr] = ""
	}
}

// StopServerTracking stops tracking the given server connection described by ID.
func (m *Continuum) StopServerTracking(addr nppc.ResourceID) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tracking, addr)
}

// Remove removes the specified node from the continuum
func (m *Continuum) Remove(node string) []nppc.ResourceID {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.continuum = m.continuum.RemoveNode(node)

	return m.scanTrackingChanges()
}

func (m *Continuum) scanTrackingChanges() []nppc.ResourceID {
	addrs := make([]nppc.ResourceID, 0)
	tracking := make(map[nppc.ResourceID]string, 0)

	for addr, trackedNode := range m.tracking {
		node, ok := m.continuum.GetNode(addr.String())
		if ok && node != trackedNode {
			addrs = append(addrs, addr)
		}

		tracking[addr] = node
	}

	m.tracking = tracking
	return addrs
}

// Get returns a node that will serve the specified ETH address.
func (m *Continuum) Get(addr nppc.ResourceID) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.continuum.GetNode(addr.String())
}
