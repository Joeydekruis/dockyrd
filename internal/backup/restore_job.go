package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/eventlog"
	"github.com/Joeydekruis/dockyrd/internal/notify"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

// RestoreJobPayload is stored on database_restore jobs.
type RestoreJobPayload struct {
	RestoreRunID string `json:"restore_run_id"`
}

// RestoreRunner executes async database restore jobs.
type RestoreRunner struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Docker *docker.Client
	Log    *slog.Logger
	Notify *notify.Notifier
	Events *eventlog.Logger
}

// EnqueueRestore schedules a restore job after creating the restore run row.
func EnqueueRestore(ctx context.Context, q *sqlc.Queries, restoreRunID string) error {
	payload, _ := json.Marshal(RestoreJobPayload{RestoreRunID: restoreRunID})
	now := time.Now().Unix()
	return q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID:        ulid.Make().String(),
		Kind:      sql.NullString{String: "database_restore", Valid: true},
		Payload:   payload,
		RunAfter:  sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
}

// RunJob performs in-place Postgres restore with pre-snapshot.
func (r *RestoreRunner) RunJob(ctx context.Context, payload RestoreJobPayload) error {
	if r.Q == nil || r.Docker == nil {
		return fmt.Errorf("restore not available")
	}
	restoreRow, err := r.Q.GetDatabaseRestoreRun(ctx, payload.RestoreRunID)
	if err != nil {
		return fmt.Errorf("restore run not found")
	}
	if restoreRow.Status != "queued" && restoreRow.Status != "running" {
		return nil
	}
	started := time.Now().Unix()
	_ = r.Q.UpdateDatabaseRestoreRun(ctx, sqlc.UpdateDatabaseRestoreRunParams{
		Status:    "running",
		StartedAt: sql.NullInt64{Int64: started, Valid: true},
		LogTail:   restoreRow.LogTail,
		ID:        restoreRow.ID,
	})

	preSnapID, logTail, runErr := r.runInPlace(ctx, restoreRow)
	finished := time.Now().Unix()
	status := "succeeded"
	errMsg := ""
	if runErr != nil {
		status = "failed"
		errMsg = runErr.Error()
	}
	if logTail == "" && errMsg != "" {
		logTail = TailBytes(errMsg, maxLogTailBytes)
	}
	_ = r.Q.UpdateDatabaseRestoreRun(ctx, sqlc.UpdateDatabaseRestoreRunParams{
		Status:           status,
		PreSnapshotRunID: sql.NullString{String: preSnapID, Valid: preSnapID != ""},
		LogTail:          logTail,
		Error:            errMsg,
		FinishedAt:       sql.NullInt64{Int64: finished, Valid: true},
		StartedAt:        sql.NullInt64{Int64: started, Valid: true},
		ID:               restoreRow.ID,
	})
	if runErr != nil {
		svc, _ := r.Q.GetServiceByID(ctx, restoreRow.ServiceID)
		r.recordRestoreFailure(ctx, svc, errMsg)
	}
	return runErr
}

func (r *RestoreRunner) runInPlace(ctx context.Context, restoreRow sqlc.DatabaseRestoreRun) (preSnapID, logTail string, err error) {
	run, err := r.Q.GetDatabaseBackupRun(ctx, restoreRow.BackupRunID)
	if err != nil {
		return "", "", fmt.Errorf("backup run not found")
	}
	if run.ServiceID != restoreRow.ServiceID {
		return "", "", fmt.Errorf("backup run mismatch")
	}
	if run.Status != "succeeded" {
		return "", "", fmt.Errorf("backup run did not succeed")
	}
	svc, err := r.Q.GetServiceByID(ctx, restoreRow.ServiceID)
	if err != nil {
		return "", "", err
	}
	if svc.Kind != "postgres" {
		return "", "", fmt.Errorf("restore supports postgres only")
	}
	policy, err := r.Q.GetDatabaseBackupPolicy(ctx, restoreRow.ServiceID)
	if err != nil {
		return "", "", fmt.Errorf("backup policy required for pre-snapshot")
	}
	snapPolicy := policy
	snapPolicy.LocalEnabled = 1
	snapPolicy.RemoteEnabled = 0
	backupRunner := &Runner{Cfg: r.Cfg, Q: r.Q, Docker: r.Docker, Log: r.Log, Notify: r.Notify, Events: r.Events}
	preSnapID, err = backupRunner.RunLocalSnapshot(ctx, svc, snapPolicy)
	if err != nil {
		return preSnapID, "", fmt.Errorf("pre-restore snapshot: %w", err)
	}

	dumpPath, cleanup, err := resolveDumpPath(ctx, r.Cfg, r.Q, run)
	if err != nil {
		return preSnapID, "", err
	}
	if cleanup != nil {
		defer cleanup()
	}
	if run.Sha256 != "" {
		sum, err := DigestFile(dumpPath)
		if err != nil {
			return preSnapID, "", err
		}
		if !strings.EqualFold(sum, run.Sha256) {
			return preSnapID, "", fmt.Errorf("backup checksum mismatch")
		}
	}
	logBuf, err := restorePostgresDump(ctx, r.Cfg, r.Q, r.Docker, restoreRow.ServiceID, dumpPath)
	logTail = TailBytes(logBuf, maxLogTailBytes)
	return preSnapID, logTail, err
}

func restorePostgresDump(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client, serviceID, dumpPath string) (string, error) {
	st, err := os.Stat(dumpPath)
	if err != nil {
		return "", err
	}
	containerID, err := serviceops.FindDatabaseContainerID(ctx, dc, serviceID)
	if err != nil {
		return "", err
	}
	svc, err := q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return "", err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return "", err
	}
	raw, err := serviceConfigJSON(svc.Config)
	if err != nil {
		return "", err
	}
	pg, _, err := dbkind.ParsePostgres(raw, key)
	if err != nil {
		return "", err
	}
	f, err := os.Open(dumpPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	const destFile = "dy-restore.dump"
	if err := dc.CopyFileToContainer(ctx, containerID, "/tmp", destFile, f, st.Size()); err != nil {
		return "", fmt.Errorf("copy dump: %w", err)
	}
	var logBuf strings.Builder
	cmd := []string{
		"pg_restore",
		"-U", pg.User,
		"-d", pg.Database,
		"--clean",
		"--if-exists",
		"/tmp/" + destFile,
	}
	if err := dc.ExecStream(ctx, containerID, cmd, &logBuf); err != nil {
		return logBuf.String(), fmt.Errorf("pg_restore: %w", err)
	}
	return logBuf.String(), nil
}

func (r *RestoreRunner) recordRestoreFailure(ctx context.Context, svc sqlc.Service, detail string) {
	if r.Events != nil {
		_ = r.Events.Record(ctx, eventlog.RecordInput{
			Level:     "error",
			Category:  "backups",
			Title:     "Restore failed · " + svc.Name,
			Detail:    detail,
			ServiceID: svc.ID,
		})
	}
	if r.Notify != nil {
		_ = r.Notify.Emit(ctx, notify.Event{
			Type: notify.TypeBackupFailed, Level: notify.LevelError, State: notify.StateFiring,
			DedupKey: "restore:" + svc.ID, Title: "Restore failed · " + svc.Name, Detail: detail,
			ServiceID: svc.ID, Href: "/databases/" + svc.Slug,
		})
	}
}

// CreateRestoreRun inserts a queued restore run row.
func CreateRestoreRun(ctx context.Context, q *sqlc.Queries, serviceID, backupRunID, mode, requestedBy string) (string, error) {
	id := ulid.Make().String()
	now := time.Now().Unix()
	if mode == "" {
		mode = "in_place"
	}
	if err := q.InsertDatabaseRestoreRun(ctx, sqlc.InsertDatabaseRestoreRunParams{
		ID:          id,
		ServiceID:   serviceID,
		BackupRunID: backupRunID,
		Mode:        mode,
		Status:      "queued",
		RequestedBy: sql.NullString{String: requestedBy, Valid: requestedBy != ""},
		CreatedAt:   now,
	}); err != nil {
		return "", err
	}
	return id, nil
}
