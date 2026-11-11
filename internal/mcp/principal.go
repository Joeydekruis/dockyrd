package mcp

import (
	"context"
	"errors"
	"slices"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type ctxKeyPrincipal struct{}

// Principal is the authenticated MCP actor.
type Principal struct {
	User       sqlc.GetUserByIDRow
	Scopes     []string
	APIToken   *sqlc.ApiToken
	ProjectID  string // set when authenticated via project-scoped dyk_* token
}

// WithPrincipal stores p on ctx.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKeyPrincipal{}, p)
}

// PrincipalFromContext returns the MCP principal or an error.
func PrincipalFromContext(ctx context.Context) (Principal, error) {
	v := ctx.Value(ctxKeyPrincipal{})
	if v == nil {
		return Principal{}, errors.New("mcp: no principal")
	}
	p, ok := v.(Principal)
	if !ok {
		return Principal{}, errors.New("mcp: bad principal")
	}
	return p, nil
}

// HasScope returns true when the principal holds scope s (or a superset via API token rank).
func (p Principal) HasScope(s string) bool {
	if slices.Contains(p.Scopes, s) {
		return true
	}
	if p.APIToken != nil {
		return apiTokenCoversMCScope(p.APIToken.Scope, s)
	}
	return false
}

func apiTokenCoversMCScope(tokenScope, need string) bool {
	switch need {
	case ScopeRead:
		return apitoken.ScopeAllows(tokenScope, apitoken.ScopeRead)
	case ScopeDeploy:
		return apitoken.ScopeAllows(tokenScope, apitoken.ScopeDeploy)
	case ScopeWrite:
		return apitoken.ScopeAllows(tokenScope, apitoken.ScopeWrite)
	default:
		return false
	}
}
