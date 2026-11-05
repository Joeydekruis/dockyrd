package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/compose"
	"github.com/Joeydekruis/dockyrd/internal/deploy"
)

func registerComposeRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("POST /api/v1/services/{id}/compose/validate", deps.handleComposeValidate)
	mux.HandleFunc("GET /api/v1/services/{id}/compose/containers", deps.handleComposeContainers)
}

type composeValidateRequest struct {
	YAML string `json:"yaml"`
}

type composeValidateResponse struct {
	OK             bool               `json:"ok"`
	Errors         []compose.Issue    `json:"errors"`
	Warnings       []compose.Issue    `json:"warnings"`
	Services       []string           `json:"services"`
	ExposablePorts []compose.Exposure `json:"exposable_ports"`
}

func (d apiDeps) composeDeployRT() *deploy.Runtime {
	return &deploy.Runtime{Cfg: d.cfg, Q: d.q, Docker: d.docker}
}

func (d apiDeps) handleComposeValidate(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	var body composeValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.YAML = strings.TrimSpace(body.YAML)
	if body.YAML == "" {
		writeJSONError(w, http.StatusBadRequest, "yaml required")
		return
	}
	ctx := r.Context()
	svc, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if svc.Kind != "compose" {
		writeJSONError(w, http.StatusBadRequest, "service is not compose kind")
		return
	}
	cfg, err := compose.Parse(serviceConfigJSON(svc.Config))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctxRow, err := d.q.GetServiceDeployContext(ctx, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	dir, err := os.MkdirTemp("", "dy-compose-validate-*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "temp dir error")
		return
	}
	defer os.RemoveAll(dir)
	inputPath := filepath.Join(dir, "compose.input.yaml")
	if err := os.WriteFile(inputPath, []byte(body.YAML), 0o600); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "write error")
		return
	}
	envLines, _ := d.composeDeployRT().RuntimeEnv(ctx, serviceID)
	mutIn := compose.MutateInput{
		ServiceID:    serviceID,
		ServerID:     ctxRow.ServerID,
		ProjectID:    ctxRow.ProjectID,
		StackDataDir: compose.StackDataDir(d.cfg.DataDir, serviceID),
	}
	runner := compose.CLIRunner{}
	doc, issues, err := compose.NormalizePipeline(ctx, runner, cfg, mutIn, dir, inputPath, envLines, false)
	resp := composeValidateResponse{
		OK:       err == nil && !compose.HasBlockingErrors(issues),
		Errors:   []compose.Issue{},
		Warnings: []compose.Issue{},
	}
	for _, i := range issues {
		if i.Warning {
			resp.Warnings = append(resp.Warnings, i)
		} else {
			resp.Errors = append(resp.Errors, i)
		}
	}
	if err != nil && resp.OK {
		resp.OK = false
		resp.Errors = append(resp.Errors, compose.Issue{Rule: "compose", Message: err.Error()})
	}
	if doc != nil {
		resp.Services = compose.ListServiceNames(doc)
		resp.ExposablePorts = compose.ExposablePorts(doc)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (d apiDeps) handleComposeContainers(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	ctx := r.Context()
	svc, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if svc.Kind != "compose" {
		writeJSONError(w, http.StatusBadRequest, "service is not compose kind")
		return
	}
	if !svc.LiveDeploymentID.Valid {
		writeJSON(w, http.StatusOK, map[string]any{"containers": []any{}})
		return
	}
	cfg, err := compose.Parse(serviceConfigJSON(svc.Config))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	releaseDir := filepath.Join(compose.StackDataDir(d.cfg.DataDir, serviceID), "releases", svc.LiveDeploymentID.String)
	composeFile := filepath.Join(releaseDir, "compose.dockyard.json")
	runner := compose.CLIRunner{}
	statuses, err := runner.PS(ctx, cfg.ProjectName, releaseDir, composeFile)
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, "compose ps failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"containers": statuses})
}
