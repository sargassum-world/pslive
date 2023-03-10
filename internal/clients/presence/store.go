// Package presence provides a high-level store of information about user presence on pages
package presence

import (
	"sort"
	"sync"
)

type (
	UserID         string
	UserIdentifier string
	SessionID      string
	Topic          string
)

type User struct {
	ID         UserID
	Identifier UserIdentifier
}

type Store struct {
	users     map[SessionID]User
	umu       sync.RWMutex
	presences map[Topic]map[SessionID]uint
	pmu       sync.RWMutex // we could use more granular locking, but we don't need that yet
}

func NewStore() *Store {
	return &Store{
		users:     make(map[SessionID]User),
		presences: make(map[Topic]map[SessionID]uint),
	}
}

func (s *Store) Add(topic Topic, sessionID SessionID) (changed bool) {
	s.pmu.Lock()
	defer s.pmu.Unlock()

	counts, ok := s.presences[topic]
	if !ok {
		counts = make(map[SessionID]uint)
		s.presences[topic] = counts
	}
	if counts[sessionID] == 0 {
		changed = true
	}
	counts[sessionID] += 1
	return changed
}

func (s *Store) Remove(topic Topic, sessionID SessionID) (changed bool) {
	s.pmu.Lock()
	defer s.pmu.Unlock()

	counts, ok := s.presences[topic]
	if !ok {
		return false
	}
	if counts[sessionID] > 0 {
		counts[sessionID] -= 1
		if counts[sessionID] == 0 {
			changed = true
		}
	}
	if counts[sessionID] == 0 {
		delete(counts, sessionID)
	}
	if len(counts) == 0 {
		delete(s.presences, topic)
	}
	return changed
}

func (s *Store) List(topic Topic) (users []User, anonymousSessions []SessionID) {
	s.pmu.RLock()
	defer s.pmu.RUnlock()

	counts, ok := s.presences[topic]
	if !ok {
		return []User{}, []SessionID{}
	}

	s.umu.RLock()
	defer s.umu.RUnlock()

	anonymousSessions = make([]SessionID, 0, len(counts))
	knownUsers := make(map[UserID]User)
	for sessionID, count := range counts {
		if count == 0 {
			// We don't attempt to delete the session ID here because we only have a read lock
			continue
		}
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

func (s *Store) Count(topic Topic) (count int) {
	s.pmu.RLock()
	defer s.pmu.RUnlock()

	return len(s.presences[topic])
}

func (s *Store) IsKnown(sessionID SessionID) bool {
	s.umu.RLock()
	defer s.umu.RUnlock()

	_, ok := s.users[sessionID]
	return ok
}

func (s *Store) Remember(sessionID SessionID, userID UserID, userIdentifier UserIdentifier) {
	s.umu.Lock()
	defer s.umu.Unlock()

	// TODO: we should remove the Remember and Forget methods because they cache the user's name,
	// which may change outside the cache; instead, the presence store should only concern itself with
	// storing session IDs, and then the internal/app/pslive/handling/presence.go file should provide
	// an AdaptPresenceList function which looks up session IDs and provides UserPresenceViewData
	// slice containing user identifiers where available (like the AdaptChatMessages function)
	s.users[sessionID] = User{
		ID:         userID,
		Identifier: userIdentifier,
	}
}

func (s *Store) Forget(sessionID SessionID) {
	s.umu.Lock()
	defer s.umu.Unlock()

	delete(s.users, sessionID)
}
