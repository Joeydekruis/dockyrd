package auth

import (
	"encoding/json"
	"net/http"
)

type setupStatusResponse struct {
	Complete bool `json:"complete"`
}

// HandleSetupStatus is a stub until setup-token flow lands (Sprint 3).
func HandleSetupStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(setupStatusResponse{Complete: false})
}
