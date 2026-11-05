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

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/backup"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
)

func registerBackupRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/settings/server/backup-target", deps.handleGetBackupTarget)
	mux.HandleFunc("PATCH /api/v1/settings/server/backup-target", deps.handlePatchBackupTarget)
	mux.HandleFunc("POST /api/v1/settings/server/backup-target/garage", deps.handlePostGarageBackupTarget)
	mux.HandleFunc("POST /api/v1/settings/server/backup-target/test", deps.handlePostTestBackupTarget)
	mux.HandleFunc("GET /api/v1/services/{id}/backup-policy", deps.handleGetDatabaseBackupPolicy)
	mux.HandleFunc("PATCH /api/v1/services/{id}/backup-policy", deps.handlePatchDatabaseBackupPolicy)
	mux.HandleFunc("GET /api/v1/services/{id}/backup-runs", deps.handleListDatabaseBackupRuns)
	mux.HandleFunc("POST /api/v1/services/{id}/backup-runs", deps.handleCreateDatabaseBackupRun)
	mux.HandleFunc("POST /api/v1/services/{id}/backup-runs/{runId}/restore", deps.handleRestoreDatabaseBackup)
	mux.HandleFunc("GET /api/v1/services/{id}/restore-runs", deps.handleListDatabaseRestoreRuns)
	mux.HandleFunc("GET /api/v1/services/{id}/restore-runs/{restoreId}", deps.handleGetDatabaseRestoreRun)
}

type backupTargetResponse struct {
	Configured    bool   `json:"configured"`
	Provider      string `json:"provider"`
	Endpoint      string `json:"endpoint"`
	Bucket        string `json:"bucket"`
	Prefix        string `json:"prefix"`
	Region        string `json:"region"`
	HasKeys       bool   `json:"has_keys"`
	ManagedGarage bool   `json:"managed_garage"`
	LastTestAt    *int64 `json:"last_test_at,omitempty"`
	LastTestError string `json:"last_test_error,omitempty"`
}

type databaseBackupPolicyResponse struct {
	Enabled       bool   `json:"enabled"`
	Interval      string `json:"interval"`
	RetentionDays int64  `json:"retention_days"`
	LocalEnabled  bool   `json:"local_enabled"`
	RemoteEnabled bool   `json:"remote_enabled"`
	RunAtMinute   int64  `json:"run_at_minute"`
	RunWeekday    int64  `json:"run_weekday"`
	LastRunAt     *int64 `json:"last_run_at,omitempty"`
	NextRunAt     *int64 `json:"next_run_at,omitempty"`
}

func backupTargetFromRow(row sqlc.GetBackupTargetForServerRow) backupTargetResponse {
	resp := backupTargetResponse{
		Configured:    row.Bucket != "",
		Provider:      row.Provider,
		Endpoint:      row.Endpoint,
		Bucket:        row.Bucket,
		Prefix:        row.Prefix,
		Region:        row.Region,
		HasKeys:       len(row.AccessKeyEnc) > 0 && len(row.SecretKeyEnc) > 0,
		ManagedGarage: row.ManagedGarage == 1,
		LastTestError: row.LastTestError,
	}
	if row.LastTestAt.Valid {
		v := row.LastTestAt.Int64
		resp.LastTestAt = &v
	}
	return resp
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
	writeJSON(w, http.StatusOK, backupTargetFromRow(row))
}

func (d apiDeps) handlePatchBackupTarget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		Endpoint  string `json:"endpoint"`
		Bucket    string `json:"bucket"`
		Prefix    string `json:"prefix"`
		Region    string `json:"region"`
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
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
	var lastTest sql.NullInt64
	lastTestErr := ""
	if existingErr == nil {
		targetID = existing.ID
		accessEnc = existing.AccessKeyEnc
		secretEnc = existing.SecretKeyEnc
		lastTest = existing.LastTestAt
		lastTestErr = existing.LastTestError
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
		ID:            targetID,
		ServerID:      server.ID,
		Provider:      "s3",
		Endpoint:      strings.TrimSpace(body.Endpoint),
		Bucket:        body.Bucket,
		Prefix:        strings.TrimSpace(body.Prefix),
		Region:        strings.TrimSpace(body.Region),
		AccessKeyEnc:  accessEnc,
		SecretKeyEnc:  secretEnc,
		ManagedGarage: 0,
		LastTestAt:    lastTest,
		LastTestError: lastTestErr,
		CreatedAt:     sql.NullInt64{Int64: now, Valid: true},
		UpdatedAt:     sql.NullInt64{Int64: now, Valid: true},
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "save error")
		return
	}
	d.handleGetBackupTarget(w, r)
}

func (d apiDeps) handlePostGarageBackupTarget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	if d.docker == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "docker unavailable")
		return
	}
	if err := backup.ConfigureGarageBackupTarget(r.Context(), d.cfg, d.q, d.docker); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	d.handleGetBackupTarget(w, r)
}

func (d apiDeps) handlePostTestBackupTarget(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireAdmin(w, r); !ok {
		return
	}
	ctx := r.Context()
	server, err := d.q.GetLocalServer(ctx)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if err := backup.TestBackupTarget(ctx, d.cfg, d.q); err != nil {
		backup.RecordBackupTargetTestFailure(ctx, d.q, server.ID, err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
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
			RunAtMinute:   180,
			RunWeekday:    0,
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
		RunAtMinute   *int64 `json:"run_at_minute"`
		RunWeekday    *int64 `json:"run_weekday"`
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
	runAtMinute := int64(180)
	runWeekday := int64(0)
	existing, err := d.q.GetDatabaseBackupPolicy(ctx, serviceID)
	policyID := ulid.Make().String()
	var lastRun, nextRun sql.NullInt64
	if err == nil {
		policyID = existing.ID
		lastRun = existing.LastRunAt
		nextRun = existing.NextRunAt
		runAtMinute = existing.RunAtMinute
		runWeekday = existing.RunWeekday
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if body.RunAtMinute != nil {
		runAtMinute = *body.RunAtMinute
	}
	if runAtMinute < 0 || runAtMinute > 1439 {
		writeJSONError(w, http.StatusBadRequest, "run_at_minute must be 0–1439")
		return
	}
	if body.RunWeekday != nil {
		runWeekday = *body.RunWeekday
	}
	if runWeekday < 0 || runWeekday > 6 {
		writeJSONError(w, http.StatusBadRequest, "run_weekday must be 0–6")
		return
	}
	now := time.Now().Unix()
	sched := backup.ScheduleParams{Interval: interval, RunAtMinute: runAtMinute, RunWeekday: runWeekday}
	if enabled {
		nextRun = sql.NullInt64{Int64: backup.NextRunAfter(time.Unix(now, 0), sched).Unix(), Valid: true}
	} else {
		nextRun = sql.NullInt64{}
	}
	if err := d.q.UpsertDatabaseBackupPolicy(ctx, sqlc.UpsertDatabaseBackupPolicyParams{
		ID:            policyID,
		ServiceID:     serviceID,
		Enabled:       boolToInt64(enabled),
		IntervalKind:  interval,
		RetentionDays: retention,
		LocalEnabled:  boolToInt64(local),
		RemoteEnabled: boolToInt64(remote),
		RunAtMinute:   runAtMinute,
		RunWeekday:    runWeekday,
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
			"sha256":      row.Sha256,
			"error":       row.Error,
		}
		if row.VerifiedAt.Valid {
			item["verified_at"] = row.VerifiedAt.Int64
			item["verified"] = true
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

func (d apiDeps) handleCreateDatabaseBackupRun(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if err := d.requireDatabaseService(ctx, serviceID); err != nil {
		writeServiceKindError(w, err)
		return
	}
	policy, err := d.q.GetDatabaseBackupPolicy(ctx, serviceID)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, http.StatusBadRequest, "configure a backup policy first")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if policy.LocalEnabled != 1 && policy.RemoteEnabled != 1 {
		writeJSONError(w, http.StatusBadRequest, "enable local or remote backup on the policy")
		return
	}
	if err := backup.EnqueueManualBackup(ctx, d.q, serviceID, policy.ID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "enqueue error")
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"queued": true})
}

func (d apiDeps) handleRestoreDatabaseBackup(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	runID := r.PathValue("runId")
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
	if svc.Kind == "redis" {
		writeJSONError(w, http.StatusUnprocessableEntity, restoreUnsupportedRedisMessage)
		return
	}
	if svc.Kind != "postgres" {
		writeJSONError(w, http.StatusBadRequest, "backup policy applies to postgres and redis only")
		return
	}
	var body struct {
		Mode    string `json:"mode"`
		Confirm string `json:"confirm"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	body.Mode = strings.TrimSpace(body.Mode)
	if body.Mode == "" {
		body.Mode = "in_place"
	}
	if body.Mode == "new_database" {
		writeJSONError(w, http.StatusUnprocessableEntity, "Restore to a new database is not available yet")
		return
	}
	if body.Mode != "in_place" {
		writeJSONError(w, http.StatusBadRequest, "mode must be in_place")
		return
	}
	dbName, err := postgresDatabaseName(ctx, d.cfg, svc)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if strings.TrimSpace(body.Confirm) != dbName {
		writeJSONError(w, http.StatusBadRequest, "type the database name to confirm restore")
		return
	}
	run, err := d.q.GetDatabaseBackupRun(ctx, runID)
	if err != nil || run.ServiceID != serviceID || run.Status != "succeeded" {
		writeJSONError(w, http.StatusBadRequest, "backup run not available for restore")
		return
	}
	restoreID, err := backup.CreateRestoreRun(ctx, d.q, serviceID, runID, body.Mode, user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "restore error")
		return
	}
	if err := backup.EnqueueRestore(ctx, d.q, restoreID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "enqueue error")
		return
	}
	_ = d.audit.Record(ctx, audit.Entry{
		UserID:     user.ID,
		Action:     "database.restore",
		TargetKind: "service",
		TargetID:   serviceID,
		Meta:       map[string]any{"run_id": runID, "restore_id": restoreID},
	})
	writeJSON(w, http.StatusAccepted, map[string]any{"restore_id": restoreID})
}

func postgresDatabaseName(ctx context.Context, cfg config.Config, svc sqlc.Service) (string, error) {
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return "", err
	}
	pg, _, err := dbkind.ParsePostgres(serviceConfigJSON(svc.Config), key)
	if err != nil {
		return "", err
	}
	return pg.Database, nil
}

const restoreUnsupportedRedisMessage = "Restore is not supported for Redis databases. Copy the RDB snapshot manually after stopping Valkey."

func (d apiDeps) handleListDatabaseRestoreRuns(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	if err := d.requireDatabaseService(r.Context(), serviceID); err != nil {
		writeServiceKindError(w, err)
		return
	}
	rows, err := d.q.ListDatabaseRestoreRuns(r.Context(), sqlc.ListDatabaseRestoreRunsParams{
		ServiceID: serviceID,
		Limit:     20,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		out = append(out, restoreRunJSON(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": out})
}

func (d apiDeps) handleGetDatabaseRestoreRun(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	serviceID := r.PathValue("id")
	restoreID := r.PathValue("restoreId")
	row, err := d.q.GetDatabaseRestoreRun(r.Context(), restoreID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "restore run not found")
		return
	}
	if row.ServiceID != serviceID {
		writeJSONError(w, http.StatusNotFound, "restore run not found")
		return
	}
	writeJSON(w, http.StatusOK, restoreRunJSON(row))
}

func restoreRunJSON(row sqlc.DatabaseRestoreRun) map[string]any {
	item := map[string]any{
		"id":            row.ID,
		"backup_run_id": row.BackupRunID,
		"mode":          row.Mode,
		"status":        row.Status,
		"log_tail":      row.LogTail,
		"error":         row.Error,
		"created_at":    row.CreatedAt,
	}
	if row.PreSnapshotRunID.Valid {
		item["pre_snapshot_run_id"] = row.PreSnapshotRunID.String
	}
	if row.StartedAt.Valid {
		item["started_at"] = row.StartedAt.Int64
	}
	if row.FinishedAt.Valid {
		item["finished_at"] = row.FinishedAt.Int64
	}
	return item
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
	errServiceNotFound    = errors.New("service not found")
	errNotDatabaseService = errors.New("not a database service")
	errDatabase           = errors.New("database error")
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

func policyFromRow(row sqlc.GetDatabaseBackupPolicyRow) databaseBackupPolicyResponse {
	resp := databaseBackupPolicyResponse{
		Enabled:       row.Enabled == 1,
		Interval:      row.IntervalKind,
		RetentionDays: row.RetentionDays,
		LocalEnabled:  row.LocalEnabled == 1,
		RemoteEnabled: row.RemoteEnabled == 1,
		RunAtMinute:   row.RunAtMinute,
		RunWeekday:    row.RunWeekday,
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
