package backup

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/storage"
)

const garageBackupBucket = "dy-backups"

// ConfigureGarageBackupTarget ensures dy-backups exists and upserts a managed target.
func ConfigureGarageBackupTarget(ctx context.Context, cfg config.Config, q *sqlc.Queries, dc *docker.Client) error {
	if q == nil || dc == nil {
		return fmt.Errorf("backup target unavailable")
	}
	installed, err := storage.IsGarageInstalled(ctx, q)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("install Garage from Storage before using it as a backup target")
	}
	if err := storage.EnsureGarageRunning(ctx, cfg, dc); err != nil {
		return err
	}
	if err := ensureGarageBucket(ctx, dc, garageBackupBucket); err != nil {
		return err
	}
	accessKey, secretKey, err := ensureGarageBackupKey(ctx, dc, garageBackupBucket)
	if err != nil {
		return err
	}
	server, err := q.GetLocalServer(ctx)
	if err != nil {
		return err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return err
	}
	accessEnc, err := crypto.Encrypt(key, []byte(accessKey))
	if err != nil {
		return err
	}
	secretEnc, err := crypto.Encrypt(key, []byte(secretKey))
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	targetID := ulid.Make().String()
	var lastTest sql.NullInt64
	var lastTestErr string
	existing, existingErr := q.GetBackupTargetForServer(ctx, server.ID)
	if existingErr == nil {
		targetID = existing.ID
		lastTest = existing.LastTestAt
		lastTestErr = existing.LastTestError
	}
	return q.UpsertBackupTarget(ctx, sqlc.UpsertBackupTargetParams{
		ID:             targetID,
		ServerID:       server.ID,
		Provider:       "s3",
		Endpoint:       storage.InternalEndpoint,
		Bucket:         garageBackupBucket,
		Prefix:         "databases",
		Region:         "garage",
		AccessKeyEnc:   accessEnc,
		SecretKeyEnc:   secretEnc,
		ManagedGarage:  1,
		LastTestAt:     lastTest,
		LastTestError:  lastTestErr,
		CreatedAt:      sqlNullInt64Now(now),
		UpdatedAt:      sqlNullInt64Now(now),
	})
}

func ensureGarageBucket(ctx context.Context, dc *docker.Client, name string) error {
	garageID, err := dc.FindContainerByName(ctx, "dy-garage")
	if err != nil || garageID == "" {
		return fmt.Errorf("dy-garage is not running")
	}
	check := []string{"garage", "bucket", "info", name}
	if err := dc.ExecStream(ctx, garageID, check, io.Discard); err == nil {
		return nil
	}
	create := []string{"garage", "bucket", "create", name}
	return dc.ExecStream(ctx, garageID, create, io.Discard)
}

func ensureGarageBackupKey(ctx context.Context, dc *docker.Client, bucket string) (access, secret string, err error) {
	garageID, err := dc.FindContainerByName(ctx, "dy-garage")
	if err != nil || garageID == "" {
		return "", "", fmt.Errorf("dy-garage is not running")
	}
	keyName := "dy-backups-dockyard"
	var buf strings.Builder
	show := []string{"garage", "key", "info", keyName}
	if err := dc.ExecStream(ctx, garageID, show, &buf); err == nil {
		return parseGarageKeyInfo(buf.String())
	}
	var createOut strings.Builder
	create := []string{"garage", "key", "create", keyName, "--bucket", bucket, "--read", "--write"}
	if err := dc.ExecStream(ctx, garageID, create, &createOut); err != nil {
		return "", "", fmt.Errorf("garage key create: %w", err)
	}
	return parseGarageKeyInfo(createOut.String())
}

func parseGarageKeyInfo(raw string) (access, secret string, err error) {
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Key ID:") {
			access = strings.TrimSpace(strings.TrimPrefix(line, "Key ID:"))
		}
		if strings.HasPrefix(line, "Secret key:") {
			secret = strings.TrimSpace(strings.TrimPrefix(line, "Secret key:"))
		}
	}
	if access == "" || secret == "" {
		return "", "", fmt.Errorf("could not parse garage key output")
	}
	return access, secret, nil
}

// TestBackupTarget writes, reads, and deletes a probe object on the configured target.
func TestBackupTarget(ctx context.Context, cfg config.Config, q *sqlc.Queries) (err error) {
	server, err := q.GetLocalServer(ctx)
	if err != nil {
		return err
	}
	target, err := q.GetBackupTargetForServer(ctx, server.ID)
	if err != nil {
		return fmt.Errorf("backup target not configured")
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
	probeID, err := randomHex(8)
	if err != nil {
		return err
	}
	objectKey := fullObjectKey(target.Prefix, ".dy-probe/"+probeID)
	tmp, err := os.CreateTemp("", "dy-probe-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	payload := []byte("dockyard-backup-probe")
	if _, err := tmp.Write(payload); err != nil {
		_ = tmp.Close()
		return err
	}
	_ = tmp.Close()
	opt := s3Options{
		Endpoint:  target.Endpoint,
		Region:    target.Region,
		Bucket:    target.Bucket,
		AccessKey: string(access),
		SecretKey: string(secret),
	}
	if err := uploadFile(ctx, opt, objectKey, tmpPath); err != nil {
		return err
	}
	downloadPath := tmpPath + ".get"
	defer os.Remove(downloadPath)
	if err := downloadObject(ctx, opt, objectKey, downloadPath); err != nil {
		return err
	}
	got, err := os.ReadFile(downloadPath)
	if err != nil {
		return err
	}
	if string(got) != string(payload) {
		return fmt.Errorf("probe read mismatch")
	}
	if err := removeObject(ctx, opt, objectKey); err != nil {
		return err
	}
	now := time.Now().Unix()
	return q.UpdateBackupTargetTestResult(ctx, sqlc.UpdateBackupTargetTestResultParams{
		LastTestAt:    sqlNullInt64Now(now),
		LastTestError: "",
		UpdatedAt:     sqlNullInt64Now(now),
		ServerID:      server.ID,
	})
}

// RecordBackupTargetTestFailure stores the probe error on the target row.
func RecordBackupTargetTestFailure(ctx context.Context, q *sqlc.Queries, serverID string, probeErr error) {
	if q == nil || probeErr == nil {
		return
	}
	now := time.Now().Unix()
	_ = q.UpdateBackupTargetTestResult(ctx, sqlc.UpdateBackupTargetTestResultParams{
		LastTestAt:    sqlNullInt64Now(now),
		LastTestError: probeErr.Error(),
		UpdatedAt:     sqlNullInt64Now(now),
		ServerID:      serverID,
	})
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sqlNullInt64Now(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}
