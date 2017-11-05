package hub

import (
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
	Remove(credentials string)
	// Has checks whether the given worker credentials contains in the
	// storage.
	Has(credentials string) bool
}

type workerACLStorage struct {
	storage *set.Set
}

func NewACLStorage() ACLStorage {
	return &workerACLStorage{
		storage: set.New(),
	}
}

func (s *workerACLStorage) Insert(credentials string) {
	s.storage.Add(credentials)
}

func (s *workerACLStorage) Remove(credentials string) {
	s.storage.Remove(credentials)
}

func (s *workerACLStorage) Has(credentials string) bool {
	return s.storage.Has(credentials)
}
