package terminal

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrTicketNotFound = errors.New("terminal ticket not found")
	ErrTicketUsed     = errors.New("terminal ticket already used")
	ErrTicketExpired  = errors.New("terminal ticket expired")
)

// SessionSpec describes a container exec target for one WebSocket session.
type SessionSpec struct {
	UserID         string
	ServiceID      string
	ContainerID    string
	ComposeService string
	Mode           string // shell | psql | valkey-cli
	ExecUser       string // empty = container default
	Cmd            []string
	Env            []string
}

type ticketRecord struct {
	spec      SessionSpec
	expiresAt time.Time
	used      bool
}

// TicketStore holds short-lived single-use WebSocket tickets.
type TicketStore struct {
	mu      sync.Mutex
	records map[string]ticketRecord
}

func NewTicketStore() *TicketStore {
	return &TicketStore{records: make(map[string]ticketRecord)}
}

func (s *TicketStore) Issue(spec SessionSpec) (string, time.Time, error) {
	token, err := randomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	exp := time.Now().Add(TicketTTL)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[token] = ticketRecord{spec: spec, expiresAt: exp}
	return token, exp, nil
}

// Peek validates a ticket without consuming it (for use before WebSocket upgrade).
func (s *TicketStore) Peek(token string) (SessionSpec, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[token]
	if !ok {
		return SessionSpec{}, ErrTicketNotFound
	}
	if rec.used {
		return SessionSpec{}, ErrTicketUsed
	}
	if time.Now().After(rec.expiresAt) {
		delete(s.records, token)
		return SessionSpec{}, ErrTicketExpired
	}
	return rec.spec, nil
}

// Redeem validates and consumes a ticket exactly once.
func (s *TicketStore) Redeem(token string) (SessionSpec, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.records[token]
	if !ok {
		return SessionSpec{}, ErrTicketNotFound
	}
	if rec.used {
		return SessionSpec{}, ErrTicketUsed
	}
	if time.Now().After(rec.expiresAt) {
		delete(s.records, token)
		return SessionSpec{}, ErrTicketExpired
	}
	rec.used = true
	delete(s.records, token)
	return rec.spec, nil
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
