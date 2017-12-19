package inmemory

import (
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	ds "github.com/sonm-io/core/insonmnia/locator/datastruct"
)

var errNodeNotFound = errors.New("a Node with the given Eth address is not found")

type Storage struct {
	cleanupPeriod time.Duration
	nodeTTL       time.Duration

	logger *zap.Logger

	mx sync.RWMutex
	db map[common.Address]*ds.Node
}

func NewStorage(cleanUpPeriod, nodeTTL time.Duration, logger *zap.Logger) *Storage {
	if logger == nil {
		logger = zap.NewNop()
	}

	s := &Storage{
		cleanupPeriod: cleanUpPeriod,
		nodeTTL:       nodeTTL,

		logger: logger,
		db:     make(map[common.Address]*ds.Node),
	}
	go s.cleanExpiredNodes()

	return s
}

func (s *Storage) Put(n *ds.Node) {
	s.mx.Lock()
	defer s.mx.Unlock()

	n.TS = time.Now()
	s.db[n.EthAddr] = n
}

func (s *Storage) ByEthAddr(ethAddr common.Address) (*ds.Node, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	n, ok := s.db[ethAddr]
	if !ok {
		return nil, errNodeNotFound
	}

	return n, nil
}

func (s *Storage) cleanExpiredNodes() {
	t := time.NewTicker(s.cleanupPeriod)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			s.traverseAndClean()
		}
	}
}

func (s *Storage) traverseAndClean() {
	deadline := time.Now().Add(-1 * s.nodeTTL)

	s.mx.Lock()
	defer s.mx.Unlock()

	var (
		total = len(s.db)
		del   uint64
		keep  uint64
	)
	for addr, node := range s.db {
		if node.TS.Before(deadline) {
			delete(s.db, addr)
			del++
		} else {
			keep++
		}
	}

	s.logger.Debug("expired Nodes cleaned",
		zap.Int("total", total), zap.Uint64("keep", keep), zap.Uint64("del", del))
}
