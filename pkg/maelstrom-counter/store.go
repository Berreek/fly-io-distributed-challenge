package main

import (
	"context"
	"fmt"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Store struct {
	localNodeCounter     int
	localNodeCounterLock sync.RWMutex

	distributedCounter     int
	distributedCounterLock sync.RWMutex

	kv *maelstrom.KV
}

func NewStore(kv *maelstrom.KV) *Store {
	return &Store{
		localNodeCounter:     0,
		localNodeCounterLock: sync.RWMutex{},

		distributedCounter:     0,
		distributedCounterLock: sync.RWMutex{},

		kv: kv,
	}
}

func (s *Store) ReadCounterForNeighbour(ctx context.Context, nodeID string) (int, error) {
	v, err := s.kv.ReadInt(ctx, nodeID)
	if err != nil {
		return 0, fmt.Errorf("reading counter from key value store for node ID %s : %w", nodeID, err)
	}

	return v, nil
}

func (s *Store) TryToUpdateDistributedCounter(newDistributedCounter int) {
	s.distributedCounterLock.Lock()
	defer s.distributedCounterLock.Unlock()
	s.localNodeCounterLock.RLock()
	defer s.localNodeCounterLock.RUnlock()

	newDistributedCounterWithLocalNodeIncluded := newDistributedCounter + s.localNodeCounter
	if newDistributedCounterWithLocalNodeIncluded > s.distributedCounter {
		s.distributedCounter = newDistributedCounterWithLocalNodeIncluded
	}
}

func (s *Store) SaveLocalCounter(ctx context.Context, nodeID string) error {
	s.localNodeCounterLock.RLock()
	defer s.localNodeCounterLock.RUnlock()

	err := s.kv.Write(ctx, nodeID, s.localNodeCounter)
	if err != nil {
		return fmt.Errorf("saving local node counter for node ID %s : %w", nodeID, err)
	}

	return nil
}

func (s *Store) IncreaseLocalNodeCounter(delta int) {
	s.localNodeCounterLock.Lock()
	defer s.localNodeCounterLock.Unlock()
	s.localNodeCounter = s.localNodeCounter + delta
}

func (s *Store) DistributedCounter() int {
	s.distributedCounterLock.RLock()
	defer s.distributedCounterLock.RUnlock()
	return s.distributedCounter
}
