// Package chat provides a high-level store of chat messages
package chat

import (
	"sync"
	"time"
)

type Message struct {
	Time             time.Time
	SenderID         string
	SenderIdentifier string
	Text             string
}

const maxHistory = 100 // max number of messages to store per topic

type Store struct {
	messages map[string][]Message
	mmu      sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		messages: make(map[string][]Message),
	}
}

func (s *Store) Add(topic string, m Message) {
	// We could have less lock contention if we had more granular locks, but we don't care about such
	// scalability yet.
	s.mmu.Lock()
	defer s.mmu.Unlock()

	// It's extremely inefficient to use slices for a queue due to allocations/deallocations, but
	// we aren't at a scale to care about this performance yet.
	s.messages[topic] = append(s.messages[topic], m)
	if len(s.messages[topic]) > maxHistory {
		s.messages[topic] = s.messages[topic][len(s.messages[topic])-maxHistory:]
	}
}

func (s *Store) List(topic string) []Message {
	s.mmu.RLock()
	defer s.mmu.RUnlock()

	messages, ok := s.messages[topic]
	if !ok {
		return []Message{}
	}

	return messages
}
