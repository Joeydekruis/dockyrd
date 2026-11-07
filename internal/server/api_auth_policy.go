package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/instance"
)

// Owners must enable 2FA before using the panel (briefing §19 / §8).
func ownerMustHave2FA(user sqlc.GetUserByIDRow) bool {
	if os.Getenv("dockyrd_REQUIRE_OWNER_2FA") == "0" {
		return false
	}
	return user.Role == "owner" && !totpEnabled(user.TotpEnabledAt)
}

func pathAllows2FASetup(path string) bool {
	switch path {
	case "/api/v1/auth/me", "/api/v1/auth/logout", "/api/v1/auth/mfa",
		"/api/v1/auth/account", "/api/v1/auth/notifications", "/api/v1/auth/locale",
		"/api/v1/auth/password/change":
		return true
	default:
		if strings.HasPrefix(path, "/api/v1/auth/sessions") {
			return true
		}
		return strings.HasPrefix(path, "/api/v1/auth/totp/")
	}
}

func (d apiDeps) enforce2FASetupPolicy(w http.ResponseWriter, r *http.Request, user sqlc.GetUserByIDRow) bool {
	if totpEnabled(user.TotpEnabledAt) {
		return true
	}
	panel, err := instance.LoadPanel(r.Context(), d.q)
	if err != nil {
		return true
	}
	mustSetup := ownerMustHave2FA(user) || panel.Require2FAAll
	if !mustSetup {
		return true
	}
	if pathAllows2FASetup(r.URL.Path) {
		return true
	}
	writeJSON(w, http.StatusForbidden, map[string]any{
		"error": "two-factor authentication required",
		"code":  "2fa_setup_required",
	})
	return false
}
