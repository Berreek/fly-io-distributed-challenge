package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct {
	ID          int
	NeighbourID string
	Time        time.Time
}

type Store struct {
	messages                       map[int]int
	wasMessagePublishedToNeighbour map[string]*Message
	messagesLock                   sync.RWMutex
	publishedLock                  sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		messages:                       map[int]int{},
		wasMessagePublishedToNeighbour: map[string]*Message{},
		messagesLock:                   sync.RWMutex{},
		publishedLock:                  sync.RWMutex{},
	}
}

func (s *Store) ReadMessages() []int {
	s.messagesLock.RLock()
	defer s.messagesLock.RUnlock()

	res := make([]int, 0, len(s.messages))
	for _, v := range s.messages {
		res = append(res, v)
	}
	return res
}

func (s *Store) AddMessage(msg int) bool {
	s.messagesLock.Lock()
	defer s.messagesLock.Unlock()

	if _, ok := s.messages[msg]; ok {
		return true
	}

	s.messages[msg] = msg
	return false
}

func (s *Store) AddMessageToNodePublishInformation(msg int, neighbour string) {
	s.publishedLock.Lock()
	defer s.publishedLock.Unlock()
	s.wasMessagePublishedToNeighbour[s.publishInfoKey(msg, neighbour)] = &Message{
		ID:          msg,
		NeighbourID: neighbour,
		Time:        time.Now(),
	}
}

func (s *Store) MessagePublishedToNode(msg int, neighbour string) {
	s.publishedLock.Lock()
	defer s.publishedLock.Unlock()
	delete(s.wasMessagePublishedToNeighbour, s.publishInfoKey(msg, neighbour))
}

func (s *Store) UnpublishedMessages() []*Message {
	s.publishedLock.RLock()
	defer s.publishedLock.RUnlock()

	var res []*Message
	for _, v := range s.wasMessagePublishedToNeighbour {
		res = append(res, v)
	}

	return res
}

func (s *Store) publishInfoKey(msg int, neighbour string) string {
	return fmt.Sprintf("%v-%s", msg, neighbour)
}
