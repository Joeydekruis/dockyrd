package server

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerAuthSecurityRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/auth/mfa", deps.handleAuthMFA)
	mux.HandleFunc("GET /api/v1/auth/security", deps.handleAuthSecurity)
	mux.HandleFunc("POST /api/v1/auth/totp/setup", deps.handleTOTPSetup)
	mux.HandleFunc("POST /api/v1/auth/totp/enable", deps.handleTOTPEnable)
	mux.HandleFunc("GET /api/v1/auth/sessions", deps.handleListSessions)
	mux.HandleFunc("DELETE /api/v1/auth/sessions/{id}", deps.handleRevokeSession)
}

type securityStatusResponse struct {
	TOTPEnabled    bool `json:"totp_enabled"`
	RecoveryRemain int  `json:"recovery_remaining"`
}

type sessionResponse struct {
	ID        string `json:"id"`
	Current   bool   `json:"current"`
	UserAgent string `json:"user_agent,omitempty"`
	IP        string `json:"ip,omitempty"`
	MFAPassed bool   `json:"mfa_passed"`
	CreatedAt int64  `json:"created_at"`
	LastSeen  int64  `json:"last_seen_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func (d apiDeps) handleAuthMFA(w http.ResponseWriter, r *http.Request) {
	sessHash, err := d.sessionHashFromRequest(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	ctx := r.Context()
	sess, err := d.q.GetSession(ctx, sessHash)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	if sess.MfaPassed != 0 {
		writeJSONError(w, http.StatusBadRequest, "mfa already completed")
		return
	}

	var body struct {
		Code         string `json:"code"`
		RecoveryCode string `json:"recovery_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Code = strings.TrimSpace(body.Code)
	body.RecoveryCode = strings.TrimSpace(body.RecoveryCode)

	ok := false
	if body.RecoveryCode != "" {
		ok, err = d.tryRecoveryCode(ctx, sess.UserID, body.RecoveryCode)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "database error")
			return
		}
	} else if body.Code != "" {
		secret, err := d.decryptUserTOTPSecret(ctx, sess.UserID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "2fa not configured")
			return
		}
		ok = auth.ValidateTOTP(secret, body.Code)
	}
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "invalid code")
		return
	}
	if err := d.q.SetSessionMFAPassed(ctx, sessHash); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "session error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (d apiDeps) handleAuthSecurity(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	sec, err := d.q.GetUserSecurity(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	remaining, _ := d.countRecoveryCodes(r.Context(), user.ID)
	writeJSON(w, http.StatusOK, securityStatusResponse{
		TOTPEnabled:    totpEnabled(sec.TotpEnabledAt),
		RecoveryRemain: remaining,
	})
}

func (d apiDeps) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	sec, err := d.q.GetUserSecurity(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if totpEnabled(sec.TotpEnabledAt) {
		writeJSONError(w, http.StatusConflict, "2fa already enabled")
		return
	}
	secret, url, err := auth.NewTOTPSecret(user.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "totp setup error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"secret":      secret,
		"otpauth_url": url,
	})
}

func (d apiDeps) handleTOTPEnable(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Secret = strings.TrimSpace(body.Secret)
	body.Code = strings.TrimSpace(body.Code)
	if body.Secret == "" || body.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "secret and code required")
		return
	}
	if !auth.ValidateTOTP(body.Secret, body.Code) {
		writeJSONError(w, http.StatusBadRequest, "invalid authenticator code")
		return
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encryption key error")
		return
	}
	enc, err := crypto.Encrypt(key, []byte(body.Secret))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "encrypt error")
		return
	}
	ctx := r.Context()
	now := time.Now().Unix()
	if err := d.q.EnableUserTOTP(ctx, sqlc.EnableUserTOTPParams{
		TotpSecretEnc: enc,
		TotpEnabledAt: sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:     now,
		ID:            user.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	plain, hashes, err := auth.NewRecoveryCodes(10)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "recovery codes error")
		return
	}
	_ = d.q.DeleteRecoveryCodesForUser(ctx, sqlNullUserID(user.ID))
	for _, h := range hashes {
		if err := d.q.InsertRecoveryCode(ctx, sqlc.InsertRecoveryCodeParams{
			UserID:   sqlNullUserID(user.ID),
			CodeHash: h,
		}); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "recovery codes error")
			return
		}
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID: user.ID,
		Action: "totp.enable",
		IP:     clientIP(r),
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"recovery_codes": plain,
	})
}

func (d apiDeps) handleListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	currentHash, _ := d.sessionHashFromRequest(r)
	rows, err := d.q.ListSessionsForUser(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]sessionResponse, 0, len(rows))
	for _, row := range rows {
		id := hex.EncodeToString(row.IDHash)
		item := sessionResponse{
			ID:        id,
			Current:   currentHash != nil && hex.EncodeToString(currentHash) == id,
			MFAPassed: row.MfaPassed != 0,
			CreatedAt: row.CreatedAt,
			LastSeen:  row.LastSeenAt,
			ExpiresAt: row.ExpiresAt,
		}
		if row.UserAgent.Valid {
			item.UserAgent = row.UserAgent.String
		}
		if row.Ip.Valid {
			item.IP = row.Ip.String
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": out})
}

func (d apiDeps) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	idHex := strings.TrimSpace(r.PathValue("id"))
	hash, err := hex.DecodeString(idHex)
	if err != nil || len(hash) == 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid session id")
		return
	}
	if err := d.q.DeleteSessionForUser(r.Context(), sqlc.DeleteSessionForUserParams{
		IDHash: hash,
		UserID: user.ID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) decryptUserTOTPSecret(ctx context.Context, userID string) (string, error) {
	sec, err := d.q.GetUserSecurity(ctx, userID)
	if err != nil {
		return "", err
	}
	if !totpEnabled(sec.TotpEnabledAt) || len(sec.TotpSecretEnc) == 0 {
		return "", errors.New("totp not enabled")
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		return "", err
	}
	plain, err := crypto.Decrypt(key, sec.TotpSecretEnc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (d apiDeps) tryRecoveryCode(ctx context.Context, userID, code string) (bool, error) {
	hash := auth.HashRecoveryCode(code)
	_, err := d.q.RecoveryCodeValid(ctx, sqlc.RecoveryCodeValidParams{
		UserID:   sqlNullUserID(userID),
		CodeHash: hash,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	now := time.Now().Unix()
	if err := d.q.ConsumeRecoveryCode(ctx, sqlc.ConsumeRecoveryCodeParams{
		UsedAt:   sql.NullInt64{Int64: now, Valid: true},
		UserID:   sqlNullUserID(userID),
		CodeHash: hash,
	}); err != nil {
		return false, err
	}
	return true, nil
}

func (d apiDeps) countRecoveryCodes(ctx context.Context, userID string) (int, error) {
	// No list query yet — return 10 when enabled as placeholder count
	sec, err := d.q.GetUserSecurity(ctx, userID)
	if err != nil || !totpEnabled(sec.TotpEnabledAt) {
		return 0, err
	}
	return 10, nil
}

func (d apiDeps) sessionHashFromRequest(r *http.Request) ([]byte, error) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil, err
	}
	return auth.HashSessionToken(c.Value), nil
}

func totpEnabled(at sql.NullInt64) bool {
	return at.Valid && at.Int64 > 0
}

func sqlNullUserID(id string) sql.NullString {
	return sql.NullString{String: id, Valid: id != ""}
}
