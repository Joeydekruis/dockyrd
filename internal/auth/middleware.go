package auth

import (
	"context"
	"net/http"
)

type contextKey int

const sessionKey contextKey = 1

// Session is a minimal session view (full login in later sprints).
type Session struct {
	UserID string
}

// WithSession stores a session on the request context.
func WithSession(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

// SessionFromRequest returns the session if present.
func SessionFromRequest(r *http.Request) (*Session, bool) {
	s, ok := r.Context().Value(sessionKey).(*Session)
	return s, ok && s != nil
}

// RequireAuth passes through until session validation is wired (Sprint 4).
func RequireAuth(next http.Handler) http.Handler {
	return next
}
