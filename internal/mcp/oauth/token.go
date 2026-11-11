package oauth

import (
	"encoding/json"
	"net/http"
)

func writeOAuthJSONError(w http.ResponseWriter, status int, errCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errCode,
		"error_description": description,
	})
}

func redirectURIMatchesAuth(stored, requested string) bool {
	if stored == requested {
		return true
	}
	return redirectURIMatches(stored, requested)
}
