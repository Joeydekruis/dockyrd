package backup

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

// RestoreRun replaces database contents from a succeeded backup run (Postgres only).
func RestoreRun(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client, serviceID, runID string) error {
	if q == nil || dc == nil {
		return fmt.Errorf("restore not available")
	}
	run, err := q.GetDatabaseBackupRun(ctx, runID)
	if err != nil {
		return fmt.Errorf("backup run not found")
	}
	if run.ServiceID != serviceID {
		return fmt.Errorf("backup run mismatch")
	}
	if run.Status != "succeeded" {
		return fmt.Errorf("backup run did not succeed")
	}
	svc, err := q.GetServiceByID(ctx, serviceID)
	if err != nil {
		return err
	}
	if svc.Kind != "postgres" {
		return fmt.Errorf("restore supports postgres only")
	}
	dumpPath, cleanup, err := resolveDumpPath(ctx, cfg, q, run)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}
	st, err := os.Stat(dumpPath)
	if err != nil {
		return err
	}
	containerID, err := serviceops.FindDatabaseContainerID(ctx, dc, serviceID)
	if err != nil {
		return err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return err
	}
	raw, err := serviceConfigJSON(svc.Config)
	if err != nil {
		return err
	}
	pg, _, err := dbkind.ParsePostgres(raw, key)
	if err != nil {
		return err
	}
	f, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer f.Close()
	const destFile = "dy-restore.dump"
	if err := dc.CopyFileToContainer(ctx, containerID, "/tmp", destFile, f, st.Size()); err != nil {
		return fmt.Errorf("copy dump: %w", err)
	}
	cmd := []string{
		"pg_restore",
		"-U", pg.User,
		"-d", pg.Database,
		"--clean",
		"--if-exists",
		"/tmp/" + destFile,
	}
	if err := dc.ExecStream(ctx, containerID, cmd, discardWriter{}); err != nil {
		return fmt.Errorf("pg_restore: %w", err)
	}
	return nil
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func resolveDumpPath(ctx context.Context, cfg config.Config, q *sqlc.Queries, run sqlc.GetDatabaseBackupRunRow) (path string, cleanup func(), err error) {
	if run.Destination == "local" {
		if run.ObjectKey == "" {
			return "", nil, fmt.Errorf("missing backup file path")
		}
		return run.ObjectKey, nil, nil
	}
	if run.Destination != "remote" || run.ObjectKey == "" {
		return "", nil, fmt.Errorf("remote backup key missing")
	}
	tmp, err := os.CreateTemp("", "dy-restore-*.dump")
	if err != nil {
		return "", nil, err
	}
	tmpPath := tmp.Name()
	server, err := q.GetLocalServer(ctx)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	target, err := q.GetBackupTargetForServer(ctx, server.ID)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	access, err := crypto.Decrypt(key, target.AccessKeyEnc)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	secret, err := crypto.Decrypt(key, target.SecretKeyEnc)
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	remoteKey := fullObjectKey(target.Prefix, run.ObjectKey)
	if err := downloadObject(ctx, s3Options{
		Endpoint:  target.Endpoint,
		Region:    target.Region,
		Bucket:    target.Bucket,
		AccessKey: string(access),
		SecretKey: string(secret),
	}, remoteKey, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", nil, err
	}
	_ = tmp.Close()
	return tmpPath, func() { _ = os.Remove(tmpPath) }, nil
}

func fullObjectKey(prefix, objectKey string) string {
	prefix = stringsTrimSlash(prefix)
	objectKey = stringsTrimSlash(objectKey)
	if prefix == "" {
		return objectKey
	}
	if objectKey == "" {
		return prefix
	}
	return prefix + "/" + objectKey
}

func stringsTrimSlash(s string) string {
	s = strings.TrimSpace(s)
	return strings.Trim(s, "/")
}

// RemotePruneAfterUpload deletes remote objects older than retention for a service slug.
func RemotePruneAfterUpload(ctx context.Context, cfg config.Config, q *sqlc.Queries, serviceSlug string, retentionDays int64) error {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	server, err := q.GetLocalServer(ctx)
	if err != nil {
		return err
	}
	target, err := q.GetBackupTargetForServer(ctx, server.ID)
	if err != nil {
		return nil
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
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
	prefix := fullObjectKey(target.Prefix, serviceSlug)
	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	return pruneObjectsOlderThan(ctx, s3Options{
		Endpoint:  target.Endpoint,
		Region:    target.Region,
		Bucket:    target.Bucket,
		AccessKey: string(access),
		SecretKey: string(secret),
	}, prefix, cutoff)
}
