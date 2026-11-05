package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/preview"
)

type servicePreviewResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	PreviewPR         int64   `json:"preview_pr"`
	PreviewHeadSHA    *string `json:"preview_head_sha,omitempty"`
	PreviewLastPushAt *int64  `json:"preview_last_push_at,omitempty"`
	PrimaryHostname   *string `json:"primary_hostname,omitempty"`
	LiveDeploymentID  *string `json:"live_deployment_id,omitempty"`
}

func registerPreviewRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/previews", deps.handleListServicePreviews)
	mux.HandleFunc("DELETE /api/v1/services/{id}/previews/{previewId}", deps.handleDeleteServicePreview)
	mux.HandleFunc("PATCH /api/v1/services/{id}/preview-config", deps.handlePatchPreviewConfig)
}

func (d apiDeps) handleListServicePreviews(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	parentID := r.PathValue("id")
	if _, err := d.q.GetServiceByID(r.Context(), parentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	rows, err := d.q.ListPreviewServicesByParent(r.Context(), sql.NullString{String: parentID, Valid: true})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]servicePreviewResponse, 0, len(rows))
	for _, row := range rows {
		item := previewFromService(row)
		domains, err := d.q.ListDomainsForService(r.Context(), row.ID)
		if err == nil {
			for _, dom := range domains {
				if dom.IsPrimary != 0 {
					h := dom.Hostname
					item.PrimaryHostname = &h
					break
				}
			}
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"previews": out})
}

func (d apiDeps) handleDeleteServicePreview(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireMember(w, r); !ok {
		return
	}
	parentID := r.PathValue("id")
	previewID := r.PathValue("previewId")
	child, err := d.q.GetServiceByID(r.Context(), previewID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "preview not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !child.ParentServiceID.Valid || child.ParentServiceID.String != parentID {
		writeJSONError(w, http.StatusNotFound, "preview not found")
		return
	}
	if err := d.q.DeleteService(r.Context(), previewID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handlePatchPreviewConfig(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireMember(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	svc, err := d.q.GetServiceByID(r.Context(), serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if svc.ParentServiceID.Valid {
		writeJSONError(w, http.StatusBadRequest, "preview child services cannot configure previews")
		return
	}
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	merged, err := preview.MergeConfig(serviceConfigJSON(svc.Config), preview.Config{Enabled: body.Enabled})
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid service config")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpdateServiceConfig(r.Context(), sqlc.UpdateServiceConfigParams{
		Config:    merged,
		UpdatedAt: sql.NullInt64{Int64: now, Valid: true},
		ID:        serviceID,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"preview": preview.Config{Enabled: body.Enabled},
	})
}

func previewFromService(row sqlc.Service) servicePreviewResponse {
	out := servicePreviewResponse{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}
	if row.PreviewPr.Valid {
		out.PreviewPR = row.PreviewPr.Int64
	}
	if row.PreviewHeadSha.Valid && row.PreviewHeadSha.String != "" {
		s := row.PreviewHeadSha.String
		out.PreviewHeadSHA = &s
	}
	if row.PreviewLastPushAt.Valid {
		v := row.PreviewLastPushAt.Int64
		out.PreviewLastPushAt = &v
	}
	if row.LiveDeploymentID.Valid && row.LiveDeploymentID.String != "" {
		s := row.LiveDeploymentID.String
		out.LiveDeploymentID = &s
	}
	return out
}
