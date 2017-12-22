package libkv

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/docker/libkv/store"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	ds "github.com/sonm-io/core/insonmnia/locator/datastruct"
)

type Storage struct {
	nodeTTL time.Duration

	logger *zap.Logger

	store     store.Store
	keyPrefix string
}

func NewStorage(nodeTTL time.Duration, store store.Store) (*Storage, error) {
	if store == nil {
		return nil, errors.New("store param cannot be nil")
	}

	s := &Storage{
		nodeTTL:   nodeTTL,
		store:     store,
		keyPrefix: "node/",
	}

	return s, nil
}

func (s *Storage) Put(node *ds.Node) error {
	node.TS = time.Now()

	value, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("cannot put node into storage: %v", err)
	}

	key := s.keyPrefix + node.EthAddr.Hex()
	if err := s.store.Put(key, value, nil); err != nil {
		return fmt.Errorf("cannot put value at key: %v", key)
	}

	return nil
}

func (s *Storage) ByEthAddr(ethAddr common.Address) (*ds.Node, error) {
	key := s.keyPrefix + ethAddr.Hex()
	pair, err := s.store.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot get value at key: %s", key)
	}

	value := &ds.Node{}
	if err := json.Unmarshal(pair.Value, value); err != nil {
		return nil, fmt.Errorf("cannot get value from storage: %v", err)
	}

	deadline := time.Now().Add(-1 * s.nodeTTL)

	if value.TS.Before(deadline) {
		s.store.Delete(key)
		return nil, fmt.Errorf("value timed out at key: %s", key)
	}

	return value, nil
}
