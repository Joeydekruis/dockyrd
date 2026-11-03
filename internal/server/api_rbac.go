package server

import (
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func (d apiDeps) requireAdmin(w http.ResponseWriter, r *http.Request) (sqlc.GetUserByIDRow, bool) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return user, false
	}
	if user.Role != "owner" && user.Role != "admin" {
		writeJSONError(w, http.StatusForbidden, "admin required")
		return user, false
	}
	return user, true
}

func validAssignRole(role string) bool {
	switch role {
	case "owner", "admin", "member", "viewer":
		return true
	default:
		return false
	}
}
