package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type requestPrincipal struct {
	User  sqlc.GetUserByIDRow
	Token *sqlc.ApiToken
}

// Action is the minimum API token scope (default-deny unless explicitly allowed).
type Action string

const (
	ActionRead   Action = apitoken.ScopeRead
	ActionDeploy Action = apitoken.ScopeDeploy
	ActionWrite  Action = apitoken.ScopeWrite
)

func (d apiDeps) requireSessionUser(w http.ResponseWriter, r *http.Request) (sqlc.GetUserByIDRow, bool) {
	return d.requireUser(w, r)
}

func sessionAllowsAction(user sqlc.GetUserByIDRow, min Action) bool {
	switch min {
	case ActionRead:
		return true
	case ActionDeploy:
		return user.Role != "viewer"
	case ActionWrite:
		return user.Role == "owner" || user.Role == "admin" || user.Role == "member"
	default:
		return false
	}
}

// requireAction authenticates the request and enforces token scope or session role (default deny).
func (d apiDeps) requireAction(w http.ResponseWriter, r *http.Request, min Action) (requestPrincipal, bool) {
	p, err := d.resolvePrincipal(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return requestPrincipal{}, false
	}
	if p.Token != nil {
		if !apitoken.ValidateScope(p.Token.Scope, string(min)) {
			writeJSONError(w, http.StatusForbidden, "insufficient token scope")
			return requestPrincipal{}, false
		}
		d.touchAPIToken(r, *p.Token)
		return p, true
	}
	if !d.enforce2FASetupPolicy(w, r, p.User) {
		return requestPrincipal{}, false
	}
	if !sessionAllowsAction(p.User, min) {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return requestPrincipal{}, false
	}
	return p, true
}

func (d apiDeps) sessionUserFromRequest(r *http.Request) (sqlc.GetUserByIDRow, error) {
	if d.q == nil {
		return sqlc.GetUserByIDRow{}, errors.New("no db")
	}
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return sqlc.GetUserByIDRow{}, err
	}
	sess, err := d.q.GetSession(r.Context(), auth.HashSessionToken(c.Value))
	if err != nil {
		return sqlc.GetUserByIDRow{}, err
	}
	if time.Now().Unix() > sess.ExpiresAt {
		return sqlc.GetUserByIDRow{}, errors.New("expired")
	}
	user, err := d.q.GetUserByID(r.Context(), sess.UserID)
	if err != nil {
		return sqlc.GetUserByIDRow{}, err
	}
	if totpEnabled(user.TotpEnabledAt) && sess.MfaPassed == 0 {
		return sqlc.GetUserByIDRow{}, errors.New("mfa required")
	}
	return user, nil
}

// resolvePrincipal tries session cookies first, then Bearer user API tokens.
func (d apiDeps) resolvePrincipal(r *http.Request) (requestPrincipal, error) {
	if d.q == nil {
		return requestPrincipal{}, errors.New("no db")
	}
	if user, err := d.sessionUserFromRequest(r); err == nil {
		return requestPrincipal{User: user, Token: nil}, nil
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(authHeader, "Bearer ") {
		plain := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		hash, err := apitoken.Hash(plain)
		if err != nil {
			return requestPrincipal{}, err
		}
		row, err := d.q.GetAPITokenByHash(r.Context(), hash)
		if err != nil {
			return requestPrincipal{}, errors.New("invalid token")
		}
		if !apitoken.IsUserScope(row.Scope) {
			return requestPrincipal{}, errors.New("invalid token")
		}
		if !row.CreatedBy.Valid || row.CreatedBy.String == "" {
			return requestPrincipal{}, errors.New("invalid token")
		}
		user, err := d.q.GetUserByID(r.Context(), row.CreatedBy.String)
		if err != nil {
			return requestPrincipal{}, errors.New("invalid token")
		}
		if row.RevokedAt.Valid {
			return requestPrincipal{}, errors.New("token revoked")
		}
		if row.ExpiresAt.Valid && time.Now().Unix() > row.ExpiresAt.Int64 {
			return requestPrincipal{}, errors.New("token expired")
		}
		return requestPrincipal{User: user, Token: &row}, nil
	}
	return requestPrincipal{}, errors.New("not authenticated")
}

func (d apiDeps) touchAPIToken(r *http.Request, row sqlc.ApiToken) {
	now := time.Now().Unix()
	ip := clientIP(r)
	_ = d.q.TouchAPITokenUsed(r.Context(), sqlc.TouchAPITokenUsedParams{
		LastUsedAt: sql.NullInt64{Int64: now, Valid: true},
		LastUsedIp: sql.NullString{String: ip, Valid: ip != ""},
		ID:         row.ID,
	})
}

func tokenProjectID(row sqlc.ApiToken) string {
	if row.ProjectID.Valid {
		return row.ProjectID.String
	}
	return ""
}

func (d apiDeps) requireProjectScope(
	w http.ResponseWriter,
	r *http.Request,
	need string,
	projectID string,
) (requestPrincipal, bool) {
	p, err := d.resolvePrincipal(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return requestPrincipal{}, false
	}
	if p.Token != nil {
		if !apitoken.ValidateScope(p.Token.Scope, need) {
			writeJSONError(w, http.StatusForbidden, "insufficient token scope")
			return requestPrincipal{}, false
		}
		if projectID != "" && tokenProjectID(*p.Token) != projectID {
			writeJSONError(w, http.StatusForbidden, "token not valid for project")
			return requestPrincipal{}, false
		}
		d.touchAPIToken(r, *p.Token)
		return p, true
	}
	if !d.enforce2FASetupPolicy(w, r, p.User) {
		return requestPrincipal{}, false
	}
	if !sessionAllowsAction(p.User, Action(need)) {
		writeJSONError(w, http.StatusForbidden, "forbidden")
		return requestPrincipal{}, false
	}
	return p, true
}
