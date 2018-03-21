package miner

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/sonm-io/core/insonmnia/hardware"
)

const stateKey = "state"

type stateJSON struct {
	Benchmarks map[string]bool    `json:"benchmarks"`
	HwHash     string             `json:"hw_hash"`
	Hardware   *hardware.Hardware `json:"hardware"`
}

func newEmptyState() *stateJSON {
	return &stateJSON{
		Benchmarks: map[string]bool{},
		HwHash:     "",
	}
}

type state struct {
	mu   sync.Mutex
	ctx  context.Context
	db   store.Store
	data *stateJSON
}

func initStorage(path, bucket string) (store.Store, error) {
	boltdb.Register()
	config := store.Config{
		Bucket: bucket,
	}

	return libkv.NewStore(store.BOLTDB, []string{path}, &config)
}

// NewState returns state storage that uses boltdb as backend
func NewState(ctx context.Context, config Config) (*state, error) {
	stor, err := initStorage(config.StorePath(), config.StoreBucket())
	if err != nil {
		return nil, err
	}

	s := &state{
		ctx: ctx,
		db:  stor,
		data: &stateJSON{
			Benchmarks: map[string]bool{},
		},
	}

	err = s.loadInitial()
	if err != nil {
		return nil, err
	}

	return s, err
}

// loadInitial loads state from boltdb
func (s *state) loadInitial() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	kv, err := s.db.Get(stateKey)
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}

	if kv != nil {
		// unmarshal exiting state
		err = json.Unmarshal(kv.Value, &s.data)
		if err != nil {
			return err
		}
	} else {
		// create new state (clean start)
		s.data = newEmptyState()
	}

	err = s.save()
	if err != nil {
		return fmt.Errorf("cannot save state into storage: %v", err)
	}

	return nil
}

// save dumps current state on disk.
//
// Warn: need no be protected by `s.mu` mutex
func (s *state) save() error {
	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return s.db.Put(stateKey, b, &store.WriteOptions{})
}

func (s *state) getPassedBenchmarks() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Benchmarks
}

func (s *state) setPassedBenchmarks(v map[string]bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Benchmarks = v
	return s.save()
}

func (s *state) getHardwareHash() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.HwHash
}

func (s *state) setHardwareHash(v string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.HwHash = v
	return s.save()
}

func (s *state) getHardwareWithBenchmarks() *hardware.Hardware {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Hardware
}

func (s *state) setHardwareWithBenchmarks(hw *hardware.Hardware) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Hardware = hw
	return s.save()
}
