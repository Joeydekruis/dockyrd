package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func registerBackupRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/settings/server/backup-target", deps.handleGetBackupTarget)
	mux.HandleFunc("PATCH /api/v1/settings/server/backup-target", deps.handlePatchBackupTarget)
	mux.HandleFunc("GET /api/v1/services/{id}/backup-policy", deps.handleGetDatabaseBackupPolicy)
	mux.HandleFunc("PATCH /api/v1/services/{id}/backup-policy", deps.handlePatchDatabaseBackupPolicy)
	mux.HandleFunc("GET /api/v1/services/{id}/backup-runs", deps.handleListDatabaseBackupRuns)
}

type backupTargetResponse struct {
	Configured bool   `json:"configured"`
	Provider   string `json:"provider"`
	Endpoint   string `json:"endpoint"`
	Bucket     string `json:"bucket"`
	Prefix     string `json:"prefix"`
	Region     string `json:"region"`
	HasKeys    bool   `json:"has_keys"`
}

type databaseBackupPolicyResponse struct {
	Enabled       bool   `json:"enabled"`
	Interval      string `json:"interval"`
	RetentionDays int64  `json:"retention_days"`
	LocalEnabled  bool   `json:"local_enabled"`
	RemoteEnabled bool   `json:"remote_enabled"`
	LastRunAt     *int64 `json:"last_run_at,omitempty"`
	NextRunAt     *int64 `json:"next_run_at,omitempty"`
}

func (d apiDeps) handleGetBackupTarget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	server, err := d.q.GetLocalServer(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	row, err := d.q.GetBackupTargetForServer(r.Context(), server.ID)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, backupTargetResponse{Provider: "s3"})
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, backupTargetResponse{
		Configured: row.Bucket != "",
		Provider:   row.Provider,
		Endpoint:   row.Endpoint,
		Bucket:     row.Bucket,
		Prefix:     row.Prefix,
		Region:     row.Region,
		HasKeys:    len(row.AccessKeyEnc) > 0 && len(row.SecretKeyEnc) > 0,
	})
}

func (d apiDeps) handlePatchBackupTarget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		Endpoint   string `json:"endpoint"`
		Bucket     string `json:"bucket"`
		Prefix     string `json:"prefix"`
		Region     string `json:"region"`
		AccessKey  string `json:"access_key"`
		SecretKey  string `json:"secret_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	body.Bucket = strings.TrimSpace(body.Bucket)
	if body.Bucket == "" {
		writeJSONError(w, http.StatusBadRequest, "bucket is required")
		return
	}
	ctx := r.Context()
	server, err := d.q.GetLocalServer(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	key, err := crypto.LoadMasterKey(d.cfg.SecretKeyPath())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "secrets not configured")
		return
	}
	existing, existingErr := d.q.GetBackupTargetForServer(ctx, server.ID)
	var accessEnc, secretEnc []byte
	targetID := ulid.Make().String()
	if existingErr == nil {
		targetID = existing.ID
		accessEnc = existing.AccessKeyEnc
		secretEnc = existing.SecretKeyEnc
	} else if !errors.Is(existingErr, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if strings.TrimSpace(body.AccessKey) != "" {
		accessEnc, err = crypto.Encrypt(key, []byte(strings.TrimSpace(body.AccessKey)))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "encrypt error")
			return
		}
	}
	if strings.TrimSpace(body.SecretKey) != "" {
		secretEnc, err = crypto.Encrypt(key, []byte(strings.TrimSpace(body.SecretKey)))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "encrypt error")
			return
		}
	}
	now := time.Now().Unix()
	if err := d.q.UpsertBackupTarget(ctx, sqlc.UpsertBackupTargetParams{
		ID:           targetID,
		ServerID:     server.ID,
		Provider:     "s3",
		Endpoint:     strings.TrimSpace(body.Endpoint),
		Bucket:       body.Bucket,
		Prefix:       strings.TrimSpace(body.Prefix),
		Region:       strings.TrimSpace(body.Region),
		AccessKeyEnc: accessEnc,
		SecretKeyEnc: secretEnc,
		CreatedAt:    sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:    sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save error")
		return
	}
	d.handleGetBackupTarget(w, r)
}

func (d apiDeps) handleGetDatabaseBackupPolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if err := d.requireDatabaseService(r.Context(), serviceID); err != nil {
		writeServiceKindError(w, err)
		return
	}
	row, err := d.q.GetDatabaseBackupPolicy(r.Context(), serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, databaseBackupPolicyResponse{
			Enabled:       false,
			Interval:      "daily",
			RetentionDays: 30,
			LocalEnabled:  true,
			RemoteEnabled: false,
		})
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, policyFromRow(row))
}

func (d apiDeps) handlePatchDatabaseBackupPolicy(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if err := d.requireDatabaseService(ctx, serviceID); err != nil {
		writeServiceKindError(w, err)
		return
	}
	var body struct {
		Enabled       *bool  `json:"enabled"`
		Interval      string `json:"interval"`
		RetentionDays *int64 `json:"retention_days"`
		LocalEnabled  *bool  `json:"local_enabled"`
		RemoteEnabled *bool  `json:"remote_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}
	interval := strings.TrimSpace(body.Interval)
	if interval == "" {
		interval = "daily"
	}
	if interval != "hourly" && interval != "daily" && interval != "weekly" {
		writeJSONError(w, http.StatusBadRequest, "interval must be hourly, daily, or weekly")
		return
	}
	retention := int64(30)
	if body.RetentionDays != nil && *body.RetentionDays > 0 {
		retention = *body.RetentionDays
	}
	enabled := false
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	local := true
	if body.LocalEnabled != nil {
		local = *body.LocalEnabled
	}
	remote := false
	if body.RemoteEnabled != nil {
		remote = *body.RemoteEnabled
	}
	if !local && !remote {
		writeJSONError(w, http.StatusBadRequest, "enable at least one of local or remote backup")
		return
	}
	existing, err := d.q.GetDatabaseBackupPolicy(ctx, serviceID)
	policyID := ulid.Make().String()
	var lastRun, nextRun sql.NullInt64
	if err == nil {
		policyID = existing.ID
		lastRun = existing.LastRunAt
		nextRun = existing.NextRunAt
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	now := time.Now().Unix()
	if err := d.q.UpsertDatabaseBackupPolicy(ctx, sqlc.UpsertDatabaseBackupPolicyParams{
		ID:            policyID,
		ServiceID:     serviceID,
		Enabled:       boolToInt64(enabled),
		IntervalKind:  interval,
		RetentionDays: retention,
		LocalEnabled:  boolToInt64(local),
		RemoteEnabled: boolToInt64(remote),
		LastRunAt:     lastRun,
		NextRunAt:     nextRun,
		CreatedAt:     sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:     sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save error")
		return
	}
	d.handleGetDatabaseBackupPolicy(w, r)
}

func (d apiDeps) handleListDatabaseBackupRuns(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if err := d.requireDatabaseService(r.Context(), serviceID); err != nil {
		writeServiceKindError(w, err)
		return
	}
	rows, err := d.q.ListDatabaseBackupRuns(r.Context(), sqlc.ListDatabaseBackupRunsParams{
		ServiceID: serviceID,
		Limit:     20,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		item := map[string]any{
			"id":          row.ID,
			"status":      row.Status,
			"destination": row.Destination,
			"object_key":  row.ObjectKey,
			"size_bytes":  row.SizeBytes,
			"error":       row.Error,
		}
		if row.StartedAt.Valid {
			item["started_at"] = row.StartedAt.Int64
		}
		if row.FinishedAt.Valid {
			item["finished_at"] = row.FinishedAt.Int64
		}
		out = append(out, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": out})
}

func (d apiDeps) requireDatabaseService(ctx context.Context, serviceID string) error {
	row, err := d.q.GetServiceByID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errServiceNotFound
		}
		return errDatabase
	}
	if row.Kind != "postgres" && row.Kind != "redis" {
		return errNotDatabaseService
	}
	return nil
}

var (
	errServiceNotFound   = errors.New("service not found")
	errNotDatabaseService = errors.New("not a database service")
	errDatabase          = errors.New("database error")
)

func writeServiceKindError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errServiceNotFound):
		writeJSONError(w, http.StatusNotFound, "service not found")
	case errors.Is(err, errNotDatabaseService):
		writeJSONError(w, http.StatusBadRequest, "backup policy applies to postgres and redis only")
	default:
		writeJSONError(w, http.StatusInternalServerError, "database error")
	}
}

func policyFromRow(row sqlc.DatabaseBackupPolicy) databaseBackupPolicyResponse {
	resp := databaseBackupPolicyResponse{
		Enabled:       row.Enabled == 1,
		Interval:      row.IntervalKind,
		RetentionDays: row.RetentionDays,
		LocalEnabled:  row.LocalEnabled == 1,
		RemoteEnabled: row.RemoteEnabled == 1,
	}
	if row.LastRunAt.Valid {
		v := row.LastRunAt.Int64
		resp.LastRunAt = &v
	}
	if row.NextRunAt.Valid {
		v := row.NextRunAt.Int64
		resp.NextRunAt = &v
	}
	return resp
}

func boolToInt64(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
