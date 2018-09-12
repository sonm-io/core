package state

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	"go.uber.org/zap"
)

type stateJSON struct {
	Benchmarks map[uint64]bool    `json:"benchmarks"`
	Hardware   *hardware.Hardware `json:"hardware"`
	HwHash     string             `json:"hw_hash"`
}

type Storage struct {
	mu  sync.Mutex
	ctx context.Context

	store store.Store
	data  *stateJSON
}

type KeyedStorage struct {
	key     string
	storage SimpleStorage
}

type StorageConfig struct {
	Endpoint string `yaml:"endpoint" required:"true" default:"/var/lib/sonm/worker.boltdb"`
	Bucket   string `yaml:"bucket" required:"true" default:"sonm"`
}

func makeStore(ctx context.Context, cfg *StorageConfig) (store.Store, error) {
	boltdb.Register()
	log.G(ctx).Info("creating store", zap.Any("store", cfg))

	config := store.Config{
		Bucket: cfg.Bucket,
	}

	return libkv.NewStore(store.BOLTDB, []string{cfg.Endpoint}, &config)
}

func newEmptyState() *stateJSON {
	return &stateJSON{
		Benchmarks: make(map[uint64]bool),
		Hardware:   new(hardware.Hardware),
	}
}

func NewState(ctx context.Context, cfg *StorageConfig) (*Storage, error) {
	clusterStore, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, err
	}

	out := &Storage{
		ctx:   ctx,
		store: clusterStore,
		data:  newEmptyState(),
	}

	if err := out.loadInitial(); err != nil {
		return nil, err
	}

	return out, nil
}

type SimpleStorage interface {
	Save(key string, value interface{}) error
	Load(key string, value interface{}) (bool, error)
}

func NewKeyedStorage(key string, storage SimpleStorage) *KeyedStorage {
	return &KeyedStorage{key, storage}
}

func (s *Storage) dump() error {
	bin, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return s.store.Put("state", bin, &store.WriteOptions{})
}

func (s *Storage) loadInitial() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	kv, err := s.store.Get("state")
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}

	if kv != nil {
		// unmarshal exiting state
		if err := json.Unmarshal(kv.Value, &s.data); err != nil {
			return err
		}
	}

	return s.dump()
}

func (s *Storage) PassedBenchmarks() map[uint64]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Benchmarks
}

func (s *Storage) SetPassedBenchmarks(v map[uint64]bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Benchmarks = v
	return s.dump()
}

func (s *Storage) HardwareHash() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.HwHash
}

func (s *Storage) SetHardwareHash(v string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.HwHash = v
	return s.dump()
}

func (s *Storage) HardwareWithBenchmarks() *hardware.Hardware {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Hardware
}

func (s *Storage) SetHardwareWithBenchmarks(hw *hardware.Hardware) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Hardware = hw
	return s.dump()
}

func (s *Storage) Save(key string, value interface{}) error {
	bin, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.store.Put(key, bin, &store.WriteOptions{})
}

func (s *Storage) Load(key string, value interface{}) (bool, error) {
	kv, err := s.store.Get(key)
	if err != nil && err != store.ErrKeyNotFound {
		return false, err
	}

	if kv != nil {
		// unmarshal existing state
		if err := json.Unmarshal(kv.Value, value); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (s *Storage) Remove(key string) (bool, error) {
	err := s.store.Delete(key)
	if err == store.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *KeyedStorage) Save(value interface{}) error {
	return m.storage.Save(m.key, value)
}

func (m *KeyedStorage) Load(value interface{}) (bool, error) {
	return m.storage.Load(m.key, value)
}
