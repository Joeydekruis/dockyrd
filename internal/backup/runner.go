package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

// JobPayload is stored on database_backup jobs.
type JobPayload struct {
	ServiceID string `json:"service_id"`
	PolicyID  string `json:"policy_id"`
}

// Runner executes database backup jobs.
type Runner struct {
	Cfg    config.Config
	Q      *sqlc.Queries
	Docker *docker.Client
	Log    *slog.Logger
	Notify *notify.Notifier
	Events *eventlog.Logger
}

func scheduleFromPolicy(p sqlc.GetDatabaseBackupPolicyRow) ScheduleParams {
	return ScheduleParams{
		Interval:    p.IntervalKind,
		RunAtMinute: p.RunAtMinute,
		RunWeekday:  p.RunWeekday,
	}
}


// RunJob performs local and/or remote backup per policy flags.
func (r *Runner) RunJob(ctx context.Context, payload JobPayload) error {
	if r.Q == nil {
		return fmt.Errorf("database not ready")
	}
	policy, err := r.Q.GetDatabaseBackupPolicy(ctx, payload.ServiceID)
	if err != nil {
		return fmt.Errorf("policy: %w", err)
	}
	if policy.Enabled != 1 {
		return fmt.Errorf("backup policy disabled")
	}
	svc, err := r.Q.GetServiceByID(ctx, payload.ServiceID)
	if err != nil {
		return err
	}
	containerID, err := serviceops.FindDatabaseContainerID(ctx, r.Docker, svc.ID)
	if err != nil {
		return err
	}
	key, err := crypto.LoadMasterKey(r.Cfg.SecretKeyPath())
	if err != nil {
		return err
	}
	raw, err := serviceConfigJSON(svc.Config)
	if err != nil {
		return err
	}
	now := time.Now()
	stamp := now.UTC().Format("20060102T150405Z")
	var runErr error
	if policy.LocalEnabled == 1 {
		runErr = r.runDestination(ctx, runInput{
			svc: svc, policy: policy, containerID: containerID, raw: raw, key: key,
			destination: "local", stamp: stamp, now: now,
		})
	}
	if runErr == nil && policy.RemoteEnabled == 1 {
		if err := r.runDestination(ctx, runInput{
			svc: svc, policy: policy, containerID: containerID, raw: raw, key: key,
			destination: "remote", stamp: stamp, now: now,
		}); err != nil {
			runErr = err
		}
	}
	finished := time.Now().Unix()
	if runErr == nil {
		_ = r.Q.UpdateDatabaseBackupPolicySchedule(ctx, sqlc.UpdateDatabaseBackupPolicyScheduleParams{
			LastRunAt: sql.NullInt64{Int64: finished, Valid: true},
			NextRunAt: sql.NullInt64{
				Int64: NextRunAfter(time.Unix(finished, 0), scheduleFromPolicy(policy)).Unix(),
				Valid: true,
			},
			UpdatedAt: sql.NullInt64{Int64: finished, Valid: true},
			ID:        policy.ID,
		})
		if err := r.pruneLocal(ctx, svc.ID, policy.RetentionDays); err != nil && r.Log != nil {
			r.Log.Warn("backup local prune", "service", svc.ID, "err", err)
		}
	}
	return runErr
}

// RunLocalSnapshot writes a one-off local backup run (pre-restore snapshot).
func (r *Runner) RunLocalSnapshot(ctx context.Context, svc sqlc.Service, policy sqlc.GetDatabaseBackupPolicyRow) (runID string, err error) {
	containerID, err := serviceops.FindDatabaseContainerID(ctx, r.Docker, svc.ID)
	if err != nil {
		return "", err
	}
	key, err := crypto.LoadMasterKey(r.Cfg.SecretKeyPath())
	if err != nil {
		return "", err
	}
	raw, err := serviceConfigJSON(svc.Config)
	if err != nil {
		return "", err
	}
	now := time.Now()
	stamp := now.UTC().Format("20060102T150405Z") + "-presnap"
	runID = ulid.Make().String()
	in := runInput{
		svc: svc, policy: policy, containerID: containerID, raw: raw, key: key,
		destination: "local", stamp: stamp, now: now, forcedRunID: runID,
	}
	if err := r.runDestination(ctx, in); err != nil {
		return runID, err
	}
	return runID, nil
}

type runInput struct {
	svc         sqlc.Service
	policy      sqlc.GetDatabaseBackupPolicyRow
	containerID string
	raw         json.RawMessage
	key         []byte
	destination string
	stamp       string
	now         time.Time
	forcedRunID string
}

func (r *Runner) runDestination(ctx context.Context, in runInput) error {
	runID := in.forcedRunID
	if runID == "" {
		runID = ulid.Make().String()
	}
	started := in.now.Unix()
	_ = r.Q.InsertDatabaseBackupRun(ctx, sqlc.InsertDatabaseBackupRunParams{
		ID:          runID,
		ServiceID:   in.svc.ID,
		PolicyID:    sql.NullString{String: in.policy.ID, Valid: in.policy.ID != ""},
		Status:      "running",
		Destination: in.destination,
		StartedAt:   sql.NullInt64{Int64: started, Valid: true},
	})

	if err := CheckDiskGuard(r.Cfg.DataDir, r.Cfg.MinFreeDiskPercent()); err != nil {
		return r.failRun(ctx, in, runID, started, "", 0, "", err)
	}

	tmpFile, objectKey, size, dumpErr := r.writeDump(ctx, in)
	finished := time.Now().Unix()
	errMsg := ""
	sha := ""
	var verifiedAt sql.NullInt64
	if dumpErr != nil {
		errMsg = dumpErr.Error()
	}
	if dumpErr == nil && in.svc.Kind == "postgres" {
		var verifyErr error
		sha, verifyErr = DigestFile(tmpFile)
		if verifyErr != nil {
			dumpErr = verifyErr
			errMsg = verifyErr.Error()
		} else if verifyErr = VerifyPostgresDump(ctx, r.Docker, in.containerID, tmpFile); verifyErr != nil {
			dumpErr = verifyErr
			errMsg = verifyErr.Error()
		} else {
			verifiedAt = sql.NullInt64{Int64: finished, Valid: true}
		}
	}
	if dumpErr == nil && in.svc.Kind == "redis" {
		var verifyErr error
		sha, verifyErr = DigestFile(tmpFile)
		if verifyErr != nil {
			dumpErr = verifyErr
			errMsg = verifyErr.Error()
		}
	}
	if dumpErr == nil && in.destination == "remote" {
		if err := r.uploadRemote(ctx, objectKey, tmpFile); err != nil {
			dumpErr = err
			errMsg = err.Error()
		} else if err := RemotePruneAfterUpload(ctx, r.Cfg, r.Q, in.svc.Slug, in.policy.RetentionDays); err != nil && r.Log != nil {
			r.Log.Warn("remote backup prune", "service", in.svc.ID, "err", err)
		}
	}
	if dumpErr == nil && in.destination == "local" && tmpFile != "" {
		objectKey = tmpFile
	}
	status := "succeeded"
	if dumpErr != nil {
		status = "failed"
	}
	_ = r.Q.UpdateDatabaseBackupRun(ctx, sqlc.UpdateDatabaseBackupRunParams{
		Status:     status,
		ObjectKey:  objectKey,
		SizeBytes:  size,
		Sha256:     sha,
		VerifiedAt: verifiedAt,
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		Error:      errMsg,
		ID:         runID,
	})
	if in.destination == "remote" && in.policy.LocalEnabled != 1 && tmpFile != "" {
		_ = os.Remove(tmpFile)
	}
	r.emitBackupNotify(ctx, in, status, errMsg, runID)
	if status == "failed" {
		r.recordBackupEvent(ctx, in.svc, errMsg)
	}
	return dumpErr
}

func (r *Runner) failRun(ctx context.Context, in runInput, runID string, started int64, objectKey string, size int64, sha string, err error) error {
	finished := time.Now().Unix()
	errMsg := err.Error()
	_ = r.Q.UpdateDatabaseBackupRun(ctx, sqlc.UpdateDatabaseBackupRunParams{
		Status:     "failed",
		ObjectKey:  objectKey,
		SizeBytes:  size,
		Sha256:     sha,
		FinishedAt: sql.NullInt64{Int64: finished, Valid: true},
		Error:      errMsg,
		ID:         runID,
	})
	r.emitBackupNotify(ctx, in, "failed", errMsg, runID)
	r.recordBackupEvent(ctx, in.svc, errMsg)
	return err
}

func (r *Runner) recordBackupEvent(ctx context.Context, svc sqlc.Service, detail string) {
	if r.Events == nil {
		return
	}
	_ = r.Events.Record(ctx, eventlog.RecordInput{
		Level:     "error",
		Category:  "backups",
		Title:     "Backup failed · " + svc.Name,
		Detail:    detail,
		ServiceID: svc.ID,
	})
}

func (r *Runner) emitBackupNotify(ctx context.Context, in runInput, status, errMsg, runID string) {
	if r.Notify == nil {
		return
	}
	dedup := fmt.Sprintf("backup:%s:%s", in.svc.ID, in.destination)
	if status == "failed" {
		scrubbed := notify.ScrubDetailForEmit(errMsg)
		_ = r.Notify.Emit(ctx, notify.Event{
			Type: notify.TypeBackupFailed, Level: notify.LevelError, State: notify.StateFiring,
			DedupKey: dedup, Title: "Backup failed · " + in.svc.Name, Detail: scrubbed,
			ServiceID: in.svc.ID, Href: "/databases/" + in.svc.Slug,
			Data: map[string]any{"service_id": in.svc.ID, "destination": in.destination, "run_id": runID, "error": scrubbed},
		})
		return
	}
	_ = r.Notify.Emit(ctx, notify.Event{
		Type: notify.TypeBackupSucceeded, Level: notify.LevelSuccess, State: notify.StateResolved,
		DedupKey: dedup, Title: "Backup succeeded · " + in.svc.Name, ServiceID: in.svc.ID,
	})
}

func (r *Runner) writeDump(ctx context.Context, in runInput) (localPath, objectKey string, size int64, err error) {
	dir := filepath.Join(r.Cfg.DatabaseBackupsDir(), in.svc.ID)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", "", 0, err
	}
	ext := ".dump"
	if in.svc.Kind == "redis" {
		ext = ".rdb"
	}
	localPath = filepath.Join(dir, in.stamp+ext)
	objectKey = filepath.Join(in.svc.Slug, in.stamp+ext)

	switch in.svc.Kind {
	case "postgres":
		pg, _, err := dbkind.ParsePostgres(in.raw, in.key)
		if err != nil {
			return "", "", 0, err
		}
		f, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
		if err != nil {
			return "", "", 0, err
		}
		defer f.Close()
		cmd := []string{"pg_dump", "-Fc", "-U", pg.User, pg.Database}
		if err := r.Docker.ExecStream(ctx, in.containerID, cmd, f); err != nil {
			_ = os.Remove(localPath)
			return "", "", 0, err
		}
		st, err := f.Stat()
		if err != nil {
			return localPath, objectKey, 0, nil
		}
		return localPath, objectKey, st.Size(), nil
	case "redis":
		_, pw, err := dbkind.ParseRedis(in.raw, in.key)
		if err != nil {
			return "", "", 0, err
		}
		tmpInContainer := "/tmp/dy-backup.rdb"
		cmd := []string{"valkey-cli", "-a", pw, "--rdb", tmpInContainer}
		if err := r.Docker.ExecStream(ctx, in.containerID, cmd, io.Discard); err != nil {
			return "", "", 0, err
		}
		reader, err := r.Docker.CopyFromContainerPath(ctx, in.containerID, tmpInContainer)
		if err != nil {
			return "", "", 0, err
		}
		defer reader.Close()
		f, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
		if err != nil {
			return "", "", 0, err
		}
		defer f.Close()
		n, err := io.Copy(f, reader)
		if err != nil {
			_ = os.Remove(localPath)
			return "", "", 0, err
		}
		return localPath, objectKey, n, nil
	default:
		return "", "", 0, fmt.Errorf("unsupported kind %q", in.svc.Kind)
	}
}

func (r *Runner) pruneLocal(ctx context.Context, serviceID string, retentionDays int64) error {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour).Unix()
	rows, err := r.Q.ListDatabaseBackupRunsForPrune(ctx, sqlc.ListDatabaseBackupRunsForPruneParams{
		ServiceID:   serviceID,
		Destination: "local",
		StartedAt:   sql.NullInt64{Int64: cutoff, Valid: true},
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.ObjectKey != "" {
			_ = os.Remove(row.ObjectKey)
		}
	}
	return nil
}

func (r *Runner) uploadRemote(ctx context.Context, objectKey, filePath string) error {
	server, err := r.Q.GetLocalServer(ctx)
	if err != nil {
		return err
	}
	target, err := r.Q.GetBackupTargetForServer(ctx, server.ID)
	if err != nil {
		return fmt.Errorf("remote backup target not configured")
	}
	key, err := crypto.LoadMasterKey(r.Cfg.SecretKeyPath())
	if err != nil {
		return err
	}
	access, err := crypto.Decrypt(key, target.AccessKeyEnc)
	if err != nil {
		return err
	}
	secret, err := crypto.Decrypt(key, target.SecretKeyEnc)
	if err != nil {
		return err
	}
	prefix := strings.Trim(strings.TrimSpace(target.Prefix), "/")
	keyPath := objectKey
	if prefix != "" {
		keyPath = prefix + "/" + objectKey
	}
	return uploadFile(ctx, s3Options{
		Endpoint:  target.Endpoint,
		Region:    target.Region,
		Bucket:    target.Bucket,
		AccessKey: string(access),
		SecretKey: string(secret),
	}, keyPath, filePath)
}

func serviceConfigJSON(v any) (json.RawMessage, error) {
	switch x := v.(type) {
	case json.RawMessage:
		return x, nil
	case []byte:
		return json.RawMessage(x), nil
	case string:
		return json.RawMessage(x), nil
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
}
