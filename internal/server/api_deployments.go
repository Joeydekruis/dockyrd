package server

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/deploy"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/jobs"
)

type deploymentResponse struct {
	ID         string  `json:"id"`
	ServiceID  string  `json:"service_id"`
	Status     string  `json:"status"`
	Trigger    string  `json:"trigger"`
	ImageRef    *string `json:"image_ref,omitempty"`
	ReleasePath *string `json:"release_path,omitempty"`
	BuildLogPath *string `json:"build_log_path,omitempty"`
	CommitSHA    *string `json:"commit_sha,omitempty"`
	CommitMessage *string `json:"commit_message,omitempty"`
	CommitAuthor  *string `json:"commit_author,omitempty"`
	Branch       *string `json:"branch,omitempty"`
	Error        *string `json:"error,omitempty"`
	MemPeakBytes  *int64   `json:"mem_peak_bytes,omitempty"`
	SwapPeakBytes *int64   `json:"swap_peak_bytes,omitempty"`
	CpuSeconds    *float64 `json:"cpu_seconds,omitempty"`
	WaitedMs      *int64   `json:"waited_ms,omitempty"`
	FrozenMs      *int64   `json:"frozen_ms,omitempty"`
	BudgetBytes   *int64   `json:"budget_bytes,omitempty"`
	Builder       *string  `json:"builder,omitempty"`
	FailureKind   *string  `json:"failure_kind,omitempty"`
	QueuedAt   int64   `json:"queued_at"`
	StartedAt  *int64  `json:"started_at,omitempty"`
	FinishedAt *int64  `json:"finished_at,omitempty"`
}

func registerDeploymentRoutes(mux *http.ServeMux, deps apiDeps) {
	mux.HandleFunc("GET /api/v1/services/{id}/deployments", deps.handleListDeployments)
	mux.HandleFunc("POST /api/v1/services/{id}/deployments", deps.handleCreateDeployment)
	mux.HandleFunc("POST /api/v1/deployments/{id}/cancel", deps.handleCancelDeployment)
	mux.HandleFunc("POST /api/v1/deployments/{id}/promote", deps.handlePromoteDeployment)
	mux.HandleFunc("POST /api/v1/deployments/{id}/redeploy", deps.handleRedeployDeployment)
	mux.HandleFunc("GET /api/v1/deployments/{id}/log", deps.handleDeploymentLog)
}

func (d apiDeps) handleListDeployments(w http.ResponseWriter, r *http.Request) {
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
	rows, err := d.q.ListDeploymentsForService(r.Context(), serviceID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	out := make([]deploymentResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, deploymentFromRow(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"deployments": out})
}

func (d apiDeps) handleCreateDeployment(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	serviceID := r.PathValue("id")
	ctx := r.Context()
	if _, err := d.q.GetServiceByID(ctx, serviceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "service not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	deploymentID, err := jobs.EnqueueServiceDeploy(ctx, d.q, serviceID, "manual", user.ID, jobs.GitMeta{})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "create deployment error")
		return
	}
	if d.events != nil {
		d.events.Publish("deployments:"+deploymentID, map[string]string{"status": "queued"})
	}
	row, err := d.q.GetDeploymentByID(ctx, deploymentID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, deploymentFromFullRow(row))
}

func (d apiDeps) handleCancelDeployment(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	row, err := d.q.GetDeploymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if row.Status != "queued" {
		writeJSONError(w, http.StatusConflict, "deployment not cancellable")
		return
	}
	finished := time.Now().Unix()
	if err := d.q.UpdateDeploymentStatus(ctx, sqlc.UpdateDeploymentStatusParams{
		Status:    "cancelled",
		Error:     sql.NullString{},
		StartedAt: sql.NullInt64{},
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		ID:        id,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "update error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (d apiDeps) handleRedeployDeployment(w http.ResponseWriter, r *http.Request) {
	user, ok := d.requireUser(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	row, err := d.q.GetDeploymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	newID, err := jobs.EnqueueServiceDeploy(ctx, d.q, row.ServiceID, "redeploy", user.ID, jobs.GitMeta{})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "enqueue error")
		return
	}
	rec, err := d.q.GetDeploymentByID(ctx, newID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, deploymentFromFullRow(rec))
}

func (d apiDeps) handlePromoteDeployment(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	ctx := r.Context()
	row, err := d.q.GetDeploymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	rt := &deploy.Runtime{Cfg: d.cfg, Q: d.q}
	if cli, err := docker.NewClient(); err == nil {
		if err := cli.Ping(r.Context()); err == nil {
			rt.Docker = cli
		} else {
			_ = cli.Close()
		}
	}
	if err := rt.PromoteRelease(r.Context(), row.ServiceID, id); err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	updated, err := d.q.GetDeploymentByID(ctx, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, deploymentFromFullRow(updated))
}

func (d apiDeps) handleDeploymentLog(w http.ResponseWriter, r *http.Request) {
	if _, ok := d.requireUser(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	row, err := d.q.GetDeploymentByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "database error")
		return
	}
	if !row.BuildLogPath.Valid || row.BuildLogPath.String == "" {
		writeJSONError(w, http.StatusNotFound, "build log not available")
		return
	}
	if r.URL.Query().Get("stream") == "1" {
		d.streamBuildLog(w, r, row.BuildLogPath.String)
		return
	}
	data, err := os.ReadFile(row.BuildLogPath.String)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "build log not found")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(data)
}

func (d apiDeps) streamBuildLog(w http.ResponseWriter, r *http.Request, path string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	offset := int64(0)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			f, err := os.Open(path)
			if err != nil {
				continue
			}
			info, _ := f.Stat()
			if info.Size() > offset {
				buf := make([]byte, info.Size()-offset)
				_, _ = f.Seek(offset, 0)
				n, _ := f.Read(buf)
				offset += int64(n)
				if n > 0 {
					_, _ = w.Write([]byte("data: "))
					_, _ = w.Write(buf[:n])
					_, _ = w.Write([]byte("\n\n"))
					flusher.Flush()
				}
			}
			_ = f.Close()
		}
	}
}

func deploymentFromRow(row sqlc.ListDeploymentsForServiceRow) deploymentResponse {
	return deploymentResponse{
		ID:         row.ID,
		ServiceID:  row.ServiceID,
		Status:     row.Status,
		Trigger:    row.Trigger,
		ImageRef:     nullStrPtr(row.ImageRef),
		ReleasePath:  nullStrPtr(row.ReleasePath),
		CommitSHA:    nullStrPtr(row.CommitSha),
		CommitMessage: nullStrPtr(row.CommitMessage),
		CommitAuthor:  nullStrPtr(row.CommitAuthor),
		Branch:       nullStrPtr(row.Branch),
		BuildLogPath: nullStrPtr(row.BuildLogPath),
		Error:        nullStrPtr(row.Error),
		MemPeakBytes:  nullInt64Ptr(row.MemPeakBytes),
		SwapPeakBytes: nullInt64Ptr(row.SwapPeakBytes),
		CpuSeconds:    nullFloat64Ptr(row.CpuSeconds),
		WaitedMs:      int64PtrIfNonZero(row.WaitedMs),
		FrozenMs:      int64PtrIfNonZero(row.FrozenMs),
		BudgetBytes:   nullInt64Ptr(row.BudgetBytes),
		Builder:       stringPtrIfNonEmpty(row.Builder),
		FailureKind:   nullStrPtr(row.FailureKind),
		QueuedAt:   nullInt64(row.QueuedAt),
		StartedAt:  nullInt64Ptr(row.StartedAt),
		FinishedAt: nullInt64Ptr(row.FinishedAt),
	}
}

func deploymentFromFullRow(row sqlc.GetDeploymentByIDRow) deploymentResponse {
	return deploymentResponse{
		ID:         row.ID,
		ServiceID:  row.ServiceID,
		Status:     row.Status,
		Trigger:    row.Trigger,
		ImageRef:     nullStrPtr(row.ImageRef),
		ReleasePath:  nullStrPtr(row.ReleasePath),
		CommitSHA:    nullStrPtr(row.CommitSha),
		CommitMessage: nullStrPtr(row.CommitMessage),
		CommitAuthor:  nullStrPtr(row.CommitAuthor),
		Branch:       nullStrPtr(row.Branch),
		BuildLogPath: nullStrPtr(row.BuildLogPath),
		Error:        nullStrPtr(row.Error),
		MemPeakBytes:  nullInt64Ptr(row.MemPeakBytes),
		SwapPeakBytes: nullInt64Ptr(row.SwapPeakBytes),
		CpuSeconds:    nullFloat64Ptr(row.CpuSeconds),
		WaitedMs:      int64PtrIfNonZero(row.WaitedMs),
		FrozenMs:      int64PtrIfNonZero(row.FrozenMs),
		BudgetBytes:   nullInt64Ptr(row.BudgetBytes),
		Builder:       stringPtrIfNonEmpty(row.Builder),
		FailureKind:   nullStrPtr(row.FailureKind),
		QueuedAt:   nullInt64(row.QueuedAt),
		StartedAt:  nullInt64Ptr(row.StartedAt),
		FinishedAt: nullInt64Ptr(row.FinishedAt),
	}
}

func nullStrPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func nullInt64(v sql.NullInt64) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}

func nullFloat64(v sql.NullFloat64) float64 {
	if !v.Valid {
		return 0
	}
	return v.Float64
}

func nullFloat64Ptr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	n := v.Float64
	return &n
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}

func int64PtrIfNonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}

func stringPtrIfNonEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
