package hub

import (
	"sync"

	"gopkg.in/fatih/set.v0"
)

// ACLStorage describes an ACL storage for workers.
//
// A worker connection can be accepted only and the only if its credentials
// provided with the certificate contains in this storage.
type ACLStorage interface {
	// Insert inserts the given worker credentials to the storage.
	Insert(credentials string)
	// Remove removes the given worker credentials from the storage.
	// Returns true if it was actually removed.
	Remove(credentials string) bool
	// Has checks whether the given worker credentials contains in the
	// storage.
	Has(credentials string) bool
}

type workerACLStorage struct {
	storage *set.SetNonTS
	mu      sync.RWMutex
}

func NewACLStorage() ACLStorage {
	return &workerACLStorage{
		storage: set.NewNonTS(),
	}
}

func (s *workerACLStorage) Insert(credentials string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.storage.Add(credentials)
}

func (s *workerACLStorage) Remove(credentials string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exists := s.storage.Has(credentials)
	if exists {
		s.storage.Remove(credentials)
	}
	return exists
}

func (s *workerACLStorage) Has(credentials string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.storage.Has(credentials)
}
