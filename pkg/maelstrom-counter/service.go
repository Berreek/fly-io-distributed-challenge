package main

import (
	"context"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	saveLocalNodeCounterIntervalDuration     = 2 * time.Second
	updateDistributedCounterIntervalDuration = 5 * time.Second
)

type CounterService struct {
	store                          *Store
	cluster                        *Cluster
	saveLocalNodeCounterTicker     *time.Ticker
	updateDistributedCounterTicker *time.Ticker
	done                           chan bool
}

func NewCounterService(kv *maelstrom.KV) *CounterService {
	return &CounterService{
		store:   NewStore(kv),
		cluster: NewCluster(),

		saveLocalNodeCounterTicker:     time.NewTicker(saveLocalNodeCounterIntervalDuration),
		updateDistributedCounterTicker: time.NewTicker(updateDistributedCounterIntervalDuration),
		done:                           make(chan bool),
	}
}

func (s *CounterService) Start() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.updateDistributedCounterTicker.C:
				s.updateDistributedCounter()
			}
		}
	}()

	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.saveLocalNodeCounterTicker.C:
				s.saveLocalNodeCounter()
			}
		}
	}()
}

func (s *CounterService) InitialiseCluster(neighbourIDs []string, currentNodeID string) {
	s.cluster.SetNeighbourIDsAndNodeID(neighbourIDs, currentNodeID)
}

func (s *CounterService) saveLocalNodeCounter() {
	ctx, cancel := context.WithTimeout(context.Background(), saveLocalNodeCounterIntervalDuration)
	defer cancel()
	err := s.store.SaveLocalCounter(ctx, s.cluster.NodeID())
	if err != nil {
		log.Printf("%e", err)
	}
}

func (s *CounterService) updateDistributedCounter() {
	possibleNewDistributedCounter := 0
	ctx, cancel := context.WithTimeout(context.Background(), saveLocalNodeCounterIntervalDuration)
	defer cancel()

	for _, nodeID := range s.cluster.NeighbourIDs() {
		v, err := s.store.ReadCounterForNeighbour(ctx, nodeID)
		if err != nil {
			log.Printf("%e", err)
		}
		possibleNewDistributedCounter = possibleNewDistributedCounter + v
	}

	s.store.TryToUpdateDistributedCounter(possibleNewDistributedCounter)
}

func (s *CounterService) IncreaseLocalNodeCounter(delta int) {
	s.store.IncreaseLocalNodeCounter(delta)
}

func (s *CounterService) DistributedCounter() int {
	return s.store.DistributedCounter()
}
