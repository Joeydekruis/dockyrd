package setup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// ValidateSetupToken checks the plain token against the stored hash and expiry.
func ValidateSetupToken(ctx context.Context, q *sqlc.Queries, plain string) error {
	row, err := q.GetSetting(ctx, settingKeySetupToken)
	if err != nil {
		return errors.New("invalid setup token")
	}
	setting, err := parseTokenSetting(row.Value)
	if err != nil {
		return errors.New("invalid setup token")
	}
	if setting.UsedAt != nil {
		return errors.New("setup token already used")
	}
	if time.Now().Unix() > setting.ExpiresAt {
		return errors.New("setup token expired")
	}
	sum := sha256.Sum256([]byte(plain))
	if hex.EncodeToString(sum[:]) != setting.TokenHash {
		return errors.New("invalid setup token")
	}
	return nil
}

// MarkSetupTokenUsed sets used_at on the setup token setting.
func MarkSetupTokenUsed(ctx context.Context, q *sqlc.Queries) error {
	row, err := q.GetSetting(ctx, settingKeySetupToken)
	if err != nil {
		return err
	}
	setting, err := parseTokenSetting(row.Value)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	setting.UsedAt = &now
	payload, err := marshalTokenSetting(setting)
	if err != nil {
		return err
	}
	if err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   settingKeySetupToken,
		Value: payload,
	}); err != nil {
		return fmt.Errorf("upsert setup token: %w", err)
	}
	return nil
}
