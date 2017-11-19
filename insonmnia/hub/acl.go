package hub

import (
	"encoding/json"
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

func (s *workerACLStorage) MarshalJSON() ([]byte, error) {
	if s == nil {
		return json.Marshal(nil)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	set := make([]string, 0)
	s.storage.Each(func(item interface{}) bool {
		set = append(set, item.(string))
		return true
	})
	return json.Marshal(set)
}

func (s *workerACLStorage) UnmarshalJSON(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unmarshalled := make([]string, 0)
	err := json.Unmarshal(data, &unmarshalled)
	if err != nil {
		return err
	}
	s.storage = set.NewNonTS()

	for _, val := range unmarshalled {
		s.storage.Add(val)
	}
	return nil
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
