package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/apitoken"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerDeployBreakRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/deploy-break/leaderboard", deps.handleDeployBreakLeaderboard)
	mux.HandleFunc("GET /api/v1/deployments/{id}/break-score", deps.handleGetDeploymentBreakScore)
	mux.HandleFunc("PUT /api/v1/deployments/{id}/break-score", deps.handlePutDeploymentBreakScore)
}

func (d apiDeps) handleGetDeploymentBreakScore(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetDeploymentByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	projectID, err := d.deploymentProjectID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if _, ok := d.requireProjectScope(w, r, apitoken.ScopeRead, projectID); !ok {
		return
	}

	var yourScore *int64
	if row, err := d.q.GetDeploymentBreakScoreForUser(ctx, sqlc.GetDeploymentBreakScoreForUserParams{
		DeploymentID: id,
		UserID:       user.ID,
	}); err == nil {
		yourScore = &row.Score
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	var deploymentBest map[string]any
	if bestRow, err := d.q.GetDeploymentBreakBest(ctx, id); err == nil {
		deploymentBest = map[string]any{
			"score":      bestRow.Score,
			"user_id":    bestRow.UserID,
			"user_name":  bestRow.UserName,
			"updated_at": bestRow.UpdatedAt,
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}

	allTimeRow, err := d.q.GetDeploymentBreakAllTimeBest(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	var allTimeBest map[string]any
	if err == nil {
		allTimeBest = map[string]any{
			"score":         allTimeRow.Score,
			"user_id":       allTimeRow.UserID,
			"user_name":     allTimeRow.UserName,
			"deployment_id": allTimeRow.DeploymentID,
			"service_name":  allTimeRow.ServiceName,
			"service_slug":  allTimeRow.ServiceSlug,
			"service_kind":  allTimeRow.ServiceKind,
			"project_slug":  allTimeRow.ProjectSlug,
			"updated_at":    allTimeRow.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"your_score":        yourScore,
		"deployment_best":   deploymentBest,
		"all_time_best":     allTimeBest,
	})
}

type putBreakScoreBody struct {
	Score int64 `json:"score"`
}

func (d apiDeps) handlePutDeploymentBreakScore(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetDeploymentByID(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	projectID, err := d.deploymentProjectID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if _, ok := d.requireProjectScope(w, r, apitoken.ScopeRead, projectID); !ok {
		return
	}

	var body putBreakScoreBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Score < 0 || body.Score > 9_999_999 {
		writeJSONError(w, http.StatusBadRequest, "invalid score")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpsertDeploymentBreakScore(ctx, sqlc.UpsertDeploymentBreakScoreParams{
		ID:           ulid.Make().String(),
		DeploymentID: id,
		UserID:       user.ID,
		Score:        body.Score,
		UpdatedAt:    now,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleDeployBreakLeaderboard(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	row, err := d.q.GetDeploymentBreakAllTimeBest(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"best": nil})
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"best": map[string]any{
			"score":         row.Score,
			"user_id":       row.UserID,
			"user_name":     row.UserName,
			"deployment_id": row.DeploymentID,
			"service_name":  row.ServiceName,
			"service_slug":  row.ServiceSlug,
			"service_kind":  row.ServiceKind,
			"project_slug":  row.ProjectSlug,
			"updated_at":    row.UpdatedAt,
		},
	})
}

func (d apiDeps) deploymentProjectID(ctx context.Context, deploymentID string) (string, error) {
	row, err := d.q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		return "", err
	}
	return d.q.GetServiceProjectID(ctx, row.ServiceID)
}
