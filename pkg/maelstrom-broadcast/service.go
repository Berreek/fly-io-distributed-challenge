package main

import (
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastingService struct {
	node           *maelstrom.Node
	store          *Store
	cluster        *Cluster
	retryingTicker *time.Ticker
	done           chan bool
}

func NewBroadcastingService() *BroadcastingService {
	return &BroadcastingService{
		node:           maelstrom.NewNode(),
		store:          NewStore(),
		cluster:        NewCluster(),
		retryingTicker: time.NewTicker(1500 * time.Millisecond),
		done:           make(chan bool),
	}
}

func (s *BroadcastingService) Start() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.retryingTicker.C:
				messagesPerNeighbourID := make(map[string][]int)
				messages := s.store.UnpublishedMessages()
				for _, msg := range messages {
					messagesPerNeighbourID[msg.NeighbourID] = append(messagesPerNeighbourID[msg.NeighbourID], msg.ID)
				}

				for neighbourID, ids := range messagesPerNeighbourID {
					nIDcp := neighbourID
					err := s.node.RPC(neighbourID, &PropagateRequest{Type: "propagate", Messages: ids},
						func(msg maelstrom.Message) error {
							for _, id := range ids {
								s.store.MessagePublishedToNode(id, nIDcp)
							}
							return nil
						})
					if err != nil {
						log.Print(err)
					}
				}
			}
		}
	}()
}

func (s *BroadcastingService) Stop() {
	s.retryingTicker.Stop()
	close(s.done)
}

func (s *BroadcastingService) Node() *maelstrom.Node {
	return s.node
}

func (s *BroadcastingService) InitialiseCluster(neighbourIDs []string, currentNodeID string) {
	s.cluster.SetNeighbourIDs(neighbourIDs, currentNodeID)
}

func (s *BroadcastingService) ReadMessages() []int {
	return s.store.ReadMessages()
}

func (s *BroadcastingService) SaveMessageAsLeader(message int, msg maelstrom.Message) {
	dup := s.store.AddMessage(message)
	if dup {
		return
	}

	for _, neighbourID := range s.cluster.NeighbourIDs() {
		if neighbourID == msg.Src {
			continue
		}

		s.store.AddMessageToNodePublishInformation(message, neighbourID)
	}

	return
}

func (s *BroadcastingService) SaveMessagesAsReplica(messages []int) {
	for _, m := range messages {
		s.store.AddMessage(m)
	}
}
