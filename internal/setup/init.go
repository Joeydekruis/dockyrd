package setup

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// Result describes what happened during Init.
type Result struct {
	AlreadyInitialized bool
	Token              string
	TokenExpiresAt     time.Time
	MasterKeyCreated   bool
}

// Init prepares config dirs, master key, SQLite, and a setup token when no users exist.
func Init(cfg config.Config) (Result, error) {
	var res Result
	if err := cfg.EnsureLayout(); err != nil {
		return res, err
	}

	created, err := crypto.EnsureMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return res, err
	}
	res.MasterKeyCreated = created

	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return res, err
	}
	defer conn.Close()

	ctx := context.Background()
	count, err := q.CountUsers(ctx)
	if err != nil {
		return res, err
	}
	if count > 0 {
		res.AlreadyInitialized = true
		return res, nil
	}

	plain, setting, err := loadOrCreateToken(ctx, q)
	if err != nil {
		return res, err
	}
	res.Token = plain
	res.TokenExpiresAt = time.Unix(setting.ExpiresAt, 0).UTC()
	return res, nil
}

func loadOrCreateToken(ctx context.Context, q *sqlc.Queries) (plain string, setting tokenSetting, err error) {
	row, err := q.GetSetting(ctx, settingKeySetupToken)
	if err == nil {
		setting, err = parseTokenSetting(row.Value)
		if err != nil {
			return "", tokenSetting{}, err
		}
		if setting.UsedAt != nil {
			return "", tokenSetting{}, errors.New("setup already completed")
		}
		if time.Now().Unix() <= setting.ExpiresAt {
			return "", setting, nil
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", tokenSetting{}, err
	}

	now := time.Now().UTC()
	plain, setting, err = GenerateToken(now)
	if err != nil {
		return "", tokenSetting{}, err
	}
	payload, err := marshalTokenSetting(setting)
	if err != nil {
		return "", tokenSetting{}, err
	}
	if err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   settingKeySetupToken,
		Value: payload,
	}); err != nil {
		return "", tokenSetting{}, err
	}
	return plain, setting, nil
}

// IssueSetupToken replaces the setup token when no owner exists yet.
func IssueSetupToken(cfg config.Config) (plain string, expiresAt time.Time, err error) {
	if err := cfg.EnsureLayout(); err != nil {
		return "", time.Time{}, err
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return "", time.Time{}, err
	}
	defer conn.Close()

	ctx := context.Background()
	count, err := q.CountUsers(ctx)
	if err != nil {
		return "", time.Time{}, err
	}
	if count > 0 {
		return "", time.Time{}, errors.New("instance already has a user; setup-token not available")
	}

	now := time.Now().UTC()
	plain, setting, err := GenerateToken(now)
	if err != nil {
		return "", time.Time{}, err
	}
	payload, err := marshalTokenSetting(setting)
	if err != nil {
		return "", time.Time{}, err
	}
	if err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   settingKeySetupToken,
		Value: payload,
	}); err != nil {
		return "", time.Time{}, err
	}
	return plain, time.Unix(setting.ExpiresAt, 0).UTC(), nil
}

// IsComplete reports whether onboarding finished (at least one user).
func IsComplete(ctx context.Context, q *sqlc.Queries) (bool, error) {
	count, err := q.CountUsers(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
