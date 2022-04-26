// Package presence provides a high-level store of information about user presence on pages
package presence

import (
	"sort"
	"sync"
)

type User struct {
	ID         string
	Identifier string
}

type Store struct {
	users     map[string]User
	umu       sync.RWMutex
	presences map[string]map[string]bool
	pmu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		users:     make(map[string]User),
		presences: make(map[string]map[string]bool),
	}
}

func (s *Store) Add(topic, sessionID string) {
	s.pmu.Lock()
	defer s.pmu.Unlock()

	presence, ok := s.presences[topic]
	if !ok {
		presence = make(map[string]bool)
		s.presences[topic] = presence
	}
	presence[sessionID] = true
}

func (s *Store) Remove(topic, sessionID string) {
	s.pmu.Lock()
	defer s.pmu.Unlock()

	presence, ok := s.presences[topic]
	if !ok {
		return
	}
	delete(presence, sessionID)
	if len(presence) == 0 {
		delete(s.presences, topic)
	}
}

func (s *Store) List(topic string) (users []User, anonymousSessions []string) {
	s.pmu.RLock()
	defer s.pmu.RUnlock()

	presence, ok := s.presences[topic]
	if !ok {
		return []User{}, []string{}
	}

	s.umu.RLock()
	defer s.umu.RUnlock()

	anonymousSessions = make([]string, 0, len(presence))
	knownUsers := make(map[string]User)
	for sessionID := range presence {
		if user, ok := s.users[sessionID]; ok {
			knownUsers[user.ID] = user // multiple sessions may refer to the same user, so we use a map
		} else {
			anonymousSessions = append(anonymousSessions, sessionID)
		}
	}
	users = make([]User, 0, len(knownUsers))
	for _, user := range knownUsers {
		users = append(users, user)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].Identifier < users[j].Identifier
	})
	// We should never display the list of anonymous session IDs, so we don't need to sort it here.
	return users, anonymousSessions
}

func (s *Store) IsKnown(sessionID string) bool {
	s.umu.RLock()
	defer s.umu.RUnlock()

	_, ok := s.users[sessionID]
	return ok
}

func (s *Store) Remember(sessionID, userID, userIdentifier string) {
	s.umu.Lock()
	defer s.umu.Unlock()

	s.users[sessionID] = User{
		ID:         userID,
		Identifier: userIdentifier,
	}
}

func (s *Store) Forget(sessionID string) {
	s.umu.Lock()
	defer s.umu.Unlock()

	delete(s.users, sessionID)
}
