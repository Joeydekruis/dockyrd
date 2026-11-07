package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/events"
	"github.com/Joeydekruis/dockyrd/internal/notify"
	"github.com/Joeydekruis/dockyrd/internal/setup"
)

const sessionCookie = "dy_session"

type apiDeps struct {
	cfg       config.Config
	q         *sqlc.Queries
	events    *events.Hub
	audit     audit.Logger
	docker    *docker.Client
	notifier  *notify.Notifier
	notifyReg notify.Registry
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type setupRequest struct {
	Token    string `json:"token"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	TOTPEnabled  bool   `json:"totp_enabled"`
}

func registerAPIRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/auth/login", deps.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", deps.handleLogout)
	mux.HandleFunc("GET /api/v1/auth/me", deps.handleMe)
	mux.HandleFunc("POST /api/v1/setup", deps.handleSetup)
	registerProjectRoutes(mux, deps)
	registerServiceRoutes(mux, deps)
	registerConnectionRoutes(mux, deps)
	registerDomainRoutes(mux, deps)
	registerDeploymentRoutes(mux, deps)
	registerVariableRoutes(mux, deps)
	registerServiceLinkRoutes(mux, deps)
	registerLogRoutes(mux, deps)
	registerStreamRoutes(mux, deps)
	registerGitHubRoutes(mux, deps)
	registerStorageRoutes(mux, deps)
	registerDatabaseRoutes(mux, deps)
	registerBucketKeyRoutes(mux, deps)
	registerEventRoutes(mux, deps)
	registerServerRoutes(mux, deps)
	registerTeamRoutes(mux, deps)
	registerInvitePublicRoutes(mux, deps)
	registerSearchRoutes(mux, deps)
	registerAuthSecurityRoutes(mux, deps)
	registerAuthAccountRoutes(mux, deps)
	registerApplicationRoutes(mux, deps)
	registerPasswordResetRoutes(mux, deps)
	registerInstanceSettingsRoutes(mux, deps)
	registerBackupRoutes(mux, deps)
	registerEmailLogRoutes(mux, deps)
	registerNotificationRoutes(mux, deps)
	registerUpdatesRoutes(mux, deps)
	registerInboxRoutes(mux, deps)
	registerBuildResilienceRoutes(mux, deps)
	registerComposeRoutes(mux, deps)
	registerTokenRoutes(mux, deps)
	registerTerminalRoutes(mux, deps)
}

func (d apiDeps) handleLogin(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	var body loginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	if body.Email == "" || body.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password required")
		return
	}

	ctx := r.Context()
	user, err := d.q.GetUserByEmail(ctx, body.Email)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if user.DisabledAt.Valid {
		writeJSONError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if !user.PasswordHash.Valid {
		writeJSONError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	ok, err := auth.VerifyPassword(body.Password, user.PasswordHash.String)
	if err != nil || !ok {
		writeJSONError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	if err := d.startSession(w, r, user.ID, !totpEnabled(user.TotpEnabledAt)); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "session error")
		return
	}
	if totpEnabled(user.TotpEnabledAt) {
		writeJSON(w, http.StatusAccepted, map[string]any{"mfa_required": true})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookie)
	if err == nil && d.q != nil {
		_ = d.q.DeleteSession(r.Context(), auth.HashSessionToken(c.Value))
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleMe(w http.ResponseWriter, r *http.Request) {
	p, err := d.resolvePrincipal(r)
	if err != nil {
		if d.q != nil {
			if sess, serr := d.sessionFromRequest(r); serr == nil {
				if time.Now().Unix() <= sess.ExpiresAt {
					u, uerr := d.q.GetUserByID(r.Context(), sess.UserID)
					if uerr == nil && totpEnabled(u.TotpEnabledAt) && sess.MfaPassed == 0 {
						writeJSON(w, http.StatusUnauthorized, map[string]any{
							"error":        "mfa required",
							"mfa_required": true,
						})
						return
					}
				}
			}
		}
		writeJSONError(w, http.StatusUnauthorized, "not signed in")
		return
	}
	if p.Token != nil {
		if !apitoken.ValidateScope(p.Token.Scope, apitoken.ScopeRead) {
			writeJSONError(w, http.StatusForbidden, "insufficient token scope")
			return
		}
		d.touchAPIToken(r, *p.Token)
	} else if !d.enforce2FASetupPolicy(w, r, p.User) {
		return
	}
	user := p.User
	writeJSON(w, http.StatusOK, userResponse{
		ID:          user.ID,
		Email:       user.Email,
		Name:        user.Name,
		Role:        user.Role,
		TOTPEnabled: totpEnabled(user.TotpEnabledAt),
	})
}

func (d apiDeps) sessionFromRequest(r *http.Request) (sqlc.GetSessionRow, error) {
	if d.q == nil {
		return sqlc.GetSessionRow{}, errors.New("no db")
	}
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return sqlc.GetSessionRow{}, err
	}
	return d.q.GetSession(r.Context(), auth.HashSessionToken(c.Value))
}

func (d apiDeps) handleSetup(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	var body setupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Email = strings.TrimSpace(strings.ToLower(body.Email))
	body.Name = strings.TrimSpace(body.Name)
	if len(body.Password) < 12 {
		writeJSONError(w, http.StatusBadRequest, "password must be at least 12 characters")
		return
	}
	if body.Token == "" || body.Email == "" || body.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "token, name, and email required")
		return
	}

	ctx := r.Context()
	count, err := d.q.CountUsers(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if count > 0 {
		writeJSONError(w, http.StatusConflict, "setup already completed")
		return
	}
	if err := setup.ValidateSetupToken(ctx, d.q, body.Token); err != nil {
		writeJSONError(w, http.StatusForbidden, "invalid setup token")
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "password hash error")
		return
	}
	now := time.Now().Unix()
	userID := ulid.Make().String()
	if err := d.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           userID,
		Email:        body.Email,
		Name:         body.Name,
		PasswordHash: sql.NullString{String: hash, Valid: true},
		Role:         "owner",
		CreatedAt:    now,
		UpdatedAt:    now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create user error")
		return
	}
	if err := setup.MarkSetupTokenUsed(ctx, d.q); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "setup token error")
		return
	}
	if _, err := EnsureLocalServer(ctx, d.q); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "server bootstrap error")
		return
	}
	if err := d.startSession(w, r, userID, true); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) startSession(w http.ResponseWriter, r *http.Request, userID string, mfaPassed bool) error {
	plain, hash, err := auth.NewSessionToken()
	if err != nil {
		return err
	}
	mfa := int64(0)
	if mfaPassed {
		mfa = 1
	}
	now := time.Now()
	expires := now.Add(30 * 24 * time.Hour)
	if err := d.q.CreateSession(r.Context(), sqlc.CreateSessionParams{
		IDHash:     hash,
		UserID:     userID,
		UserAgent:  sql.NullString{String: r.UserAgent(), Valid: r.UserAgent() != ""},
		Ip:         sql.NullString{String: clientIP(r), Valid: true},
		MfaPassed:  mfa,
		CreatedAt:  now.Unix(),
		LastSeenAt: now.Unix(),
		ExpiresAt:  expires.Unix(),
	}); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    plain,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
	return nil
}

func (d apiDeps) userFromRequest(r *http.Request) (sqlc.GetUserByIDRow, error) {
	return d.sessionUserFromRequest(r)
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
