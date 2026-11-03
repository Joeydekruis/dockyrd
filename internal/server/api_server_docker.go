package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Joeydekruis/dockyrd/internal/docker"
)

func (d apiDeps) requireDocker(w http.ResponseWriter) (*docker.Client, bool) {
	if d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return nil, false
	}
	return d.docker, true
}

func (d apiDeps) handleServerContainers(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	cli, ok := d.requireDocker(w)
	if !ok {
		return
	}
	rows, err := cli.ListContainerRows(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "docker error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"containers": rows})
}

func (d apiDeps) handleServerCleanupGet(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	cli, ok := d.requireDocker(w)
	if !ok {
		return
	}
	sum, err := cli.CleanupSummary(r.Context())
	if err != nil {
		if errors.Is(err, docker.ErrUnavailable) {
			writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "docker error")
		return
	}
	writeJSON(w, http.StatusOK, sum)
}

func (d apiDeps) handleServerCleanupPost(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	cli, ok := d.requireDocker(w)
	if !ok {
		return
	}
	var body struct {
		Targets []string `json:"targets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(body.Targets) == 0 {
		writeJSONError(w, http.StatusBadRequest, "targets required")
		return
	}
	results, err := cli.RunCleanup(r.Context(), body.Targets)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reclaimed": results})
}
