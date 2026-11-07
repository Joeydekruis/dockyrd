package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/build"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/github"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
	"github.com/Joeydekruis/dockyrd/internal/resources"
)

func registerBuildResilienceRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/server/build-budget", deps.handleBuildBudget)
	mux.HandleFunc("GET /api/v1/services/{id}/github-workflow", deps.handleGitHubWorkflow)
	mux.HandleFunc("GET /api/v1/services/{id}/deploy-token", deps.handleGetDeployTokenStatus)
	mux.HandleFunc("POST /api/v1/services/{id}/deploy-token", deps.handleCreateDeployToken)
	mux.HandleFunc("POST /api/v1/hooks/deploy/{service_id}", deps.handleDeployHook)
	mux.HandleFunc("GET /api/v1/builders", deps.handleListBuilders)
	mux.HandleFunc("POST /api/v1/builders", deps.handleCreateBuilder)
	mux.HandleFunc("DELETE /api/v1/builders/{id}", deps.handleDeleteBuilder)
	mux.HandleFunc("POST /api/v1/builders/{id}/test", deps.handleTestBuilder)
	mux.HandleFunc("GET /api/v1/builders/{id}/bundle", deps.handleBuilderBundle)
	mux.HandleFunc("PATCH /api/v1/services/{id}/build-location", deps.handlePatchBuildLocation)
}

func (d apiDeps) handleBuildBudget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	res := resources.New(d.cfg.BuildIsolationEnabled())
	snap, err := res.Snapshot(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "snapshot failed")
		return
	}
	b := build.ComputeBudget(snap, build.BudgetParams{})
	writeJSON(w, http.StatusOK, map[string]any{
		"ram_budget":      b.RAMBytes,
		"swap_budget":     b.SwapBytes,
		"apps_ws":         snap.AppsWS,
		"system_ws":       snap.SystemWS,
		"kernel_reserve":  snap.KernelReserve,
		"mem_total":       snap.MemTotal,
		"isolation_active": res.Available(),
	})
}

func (d apiDeps) handleGitHubWorkflow(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	_ = user
	serviceID := r.PathValue("id")
	raw, err := github.LoadServiceConfig(r.Context(), d.q, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "service not found")
		return
	}
	cfg, err := appconfig.Parse(raw)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid app config")
		return
	}
	repo := cfg.Source.Repo
	if repo == "" {
		writeJSONError(w, http.StatusBadRequest, "service has no GitHub repo")
		return
	}
	base := d.githubPublicBaseURL(r)
	yaml, err := build.RenderGitHubWorkflow(build.GitHubWorkflowInput{
		Branch:      cfg.Source.BranchOrDefault(),
		Context:     cfg.BuildContextDir("."),
		Image:       build.ExpectedGHCRImage(repo),
		DockyrdURL: base,
		ServiceID:   serviceID,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "workflow render failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"workflow": ".github/workflows/dockyrd.yml",
		"yaml":     yaml,
		"secrets": []map[string]string{
			{"name": "dockyrd_URL", "value": base},
			{"name": "dockyrd_DEPLOY_TOKEN", "description": "Create via POST /api/v1/services/{id}/deploy-token"},
		},
	})
}

func (d apiDeps) handleGetDeployTokenStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if _, err := d.q.GetServiceByID(r.Context(), serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	scope := apitoken.DeployScope(serviceID)
	n, err := d.q.CountAPITokensByScope(r.Context(), scope)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"configured": n > 0,
		"service_id": serviceID,
	})
}

func (d apiDeps) handleCreateDeployToken(w http.ResponseWriter, r *http.Request) {
	admin, ok := d.requireAdmin(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	if _, err := d.q.GetServiceByID(r.Context(), serviceID); err != nil {
		writeJSONError(w, http.StatusNotFound, "service not found")
		return
	}
	plain, hash, err := apitoken.Generate()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "token generation failed")
		return
	}
	scope := apitoken.DeployScope(serviceID)
	_ = d.q.DeleteAPITokensByScope(r.Context(), scope)
	now := time.Now().Unix()
	id := ulid.Make().String()
	if err := d.q.InsertAPIToken(r.Context(), sqlc.InsertAPITokenParams{
		ID:        id,
		Name:      sql.NullString{String: "deploy hook", Valid: true},
		TokenHash: hash,
		Scope:     scope,
		CreatedBy: sql.NullString{String: admin.ID, Valid: true},
		ProjectID: sql.NullString{},
		CreatedAt: now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "store token failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      plain,
		"scope":      scope,
		"service_id": serviceID,
	})
}

func (d apiDeps) handleDeployHook(w http.ResponseWriter, r *http.Request) {
	if d.q == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not ready")
		return
	}
	serviceID := r.PathValue("service_id")
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	hash, err := apitoken.Hash(auth)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "token required")
		return
	}
	row, err := d.q.GetAPITokenByHash(r.Context(), hash)
	if err != nil || row.Scope != apitoken.DeployScope(serviceID) {
		writeJSONError(w, http.StatusForbidden, "invalid deploy token")
		return
	}
	if row.RevokedAt.Valid {
		writeJSONError(w, http.StatusForbidden, "invalid deploy token")
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	var payload struct {
		Image     string `json:"image"`
		CommitSHA string `json:"commit_sha"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || strings.TrimSpace(payload.Image) == "" {
		writeJSONError(w, http.StatusBadRequest, "image required")
		return
	}
	raw, err := github.LoadServiceConfig(r.Context(), d.q, serviceID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "service not found")
		return
	}
	repo := github.RepoFromServiceConfig(raw)
	if repo == "" || !build.ImageMatchesRepo(repo, payload.Image) {
		writeJSONError(w, http.StatusForbidden, "image does not match service repository")
		return
	}
	now := time.Now().Unix()
	_ = d.q.TouchAPITokenUsed(r.Context(), sqlc.TouchAPITokenUsedParams{
		LastUsedAt: sql.NullInt64{Int64: now, Valid: true},
		LastUsedIp: sql.NullString{String: clientIP(r), Valid: true},
		ID:         row.ID,
	})
	deploymentID, err := jobs.EnqueueHookDeploy(r.Context(), d.q, serviceID, strings.TrimSpace(payload.Image), jobs.GitMeta{CommitSHA: payload.CommitSHA})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "enqueue failed")
		return
	}
	if d.events != nil {
		d.events.Publish("deployments:"+deploymentID, map[string]string{"status": "queued"})
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"deployment_id": deploymentID})
}
