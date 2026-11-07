package terminal

import (
	"errors"
	"sync"
)

var ErrSessionLimit = errors.New("too many concurrent terminal sessions")

type activeSession struct {
	userID string
}

// SessionRegistry tracks live WebSocket terminal sessions per user.
type SessionRegistry struct {
	mu       sync.Mutex
	sessions map[string]activeSession // session id -> user
}

func NewSessionRegistry() *SessionRegistry {
	return &SessionRegistry{sessions: make(map[string]activeSession)}
}

func (r *SessionRegistry) TryStart(userID, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, s := range r.sessions {
		if s.userID == userID {
			count++
		}
	}
	if count >= MaxConcurrentPerUser {
		return ErrSessionLimit
	}
	r.sessions[sessionID] = activeSession{userID: userID}
	return nil
}

func (r *SessionRegistry) End(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
}

func (r *SessionRegistry) CountForUser(userID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, s := range r.sessions {
		if s.userID == userID {
			n++
		}
	}
	return n
}
