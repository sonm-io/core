package state

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	"github.com/mohae/deepcopy"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/hardware"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type stateJSON struct {
	AskPlans   map[string]*pb.AskPlan `json:"ask_plans"`
	Benchmarks map[uint64]bool        `json:"benchmarks"`
	Hardware   *hardware.Hardware     `json:"hardware"`
	HwHash     string                 `json:"hw_hash"`
}

type Storage struct {
	mu  sync.Mutex
	ctx context.Context

	store store.Store
	data  *stateJSON
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
		AskPlans:   make(map[string]*pb.AskPlan, 0),
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

func (s *Storage) AskPlans() map[string]*pb.AskPlan {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]*pb.AskPlan)
	for id, plan := range s.data.AskPlans {
		result[id] = plan
	}

	return result
}

func (s *Storage) SaveAskPlan(askPlan *pb.AskPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(askPlan.GetID()) == 0 {
		return errors.New("could not create ask plan  - missing id")
	}
	s.data.AskPlans[askPlan.ID] = askPlan
	if err := s.dump(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) AskPlan(planID string) (*pb.AskPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	askPlan, ok := s.data.AskPlans[planID]
	if !ok {
		return nil, errors.New("specified ask-plan does not exist")
	}
	copy := deepcopy.Copy(askPlan).(*pb.AskPlan)
	return copy, nil
}

func (s *Storage) RemoveAskPlan(planID string) (*pb.AskPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	askPlan, ok := s.data.AskPlans[planID]
	if !ok {
		return nil, errors.New("specified ask-plan does not exist")
	}

	delete(s.data.AskPlans, planID)
	err := s.dump()
	return askPlan, err
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
