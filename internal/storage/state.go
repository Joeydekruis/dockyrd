package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

const SettingKeyGarage = "garage"

// GarageState is persisted in settings.value for object storage.
type GarageState struct {
	Installed     bool   `json:"installed"`
	Provider      string `json:"provider"`
	Version       string `json:"version"`
	AdminTokenEnc string `json:"admin_token_enc"`
	RPCSecretEnc  string `json:"rpc_secret_enc"`
}

// LoadGarage reads garage install state from settings.
func LoadGarage(ctx context.Context, q *sqlc.Queries) (GarageState, error) {
	row, err := q.GetSetting(ctx, SettingKeyGarage)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GarageState{}, nil
		}
		return GarageState{}, err
	}
	var st GarageState
	data, err := settingValueBytes(row.Value)
	if err != nil {
		return GarageState{}, err
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return GarageState{}, err
	}
	return st, nil
}

func settingValueBytes(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}

// SaveGarage persists garage install state.
func SaveGarage(ctx context.Context, q *sqlc.Queries, st GarageState) error {
	raw, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return q.UpsertSetting(ctx, sqlc.UpsertSettingParams{
		Key:   SettingKeyGarage,
		Value: string(raw),
	})
}

// IsGarageInstalled reports whether Garage was installed successfully.
func IsGarageInstalled(ctx context.Context, q *sqlc.Queries) (bool, error) {
	st, err := LoadGarage(ctx, q)
	if err != nil {
		return false, nil
	}
	return st.Installed && st.Provider == "garage", nil
}

// InternalEndpoint is the S3 API URL apps use on Docker networks.
const InternalEndpoint = "http://dy-garage:3900"

// ValidateInstallRequest checks POST /storage/install body.
func ValidateInstallRequest(provider string) error {
	if provider != "garage" {
		return fmt.Errorf("unsupported storage provider %q (only garage in MVP)", provider)
	}
	return nil
}
