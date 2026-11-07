package addons

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
)

var ErrNotInstallable = errors.New("addon install not available in this release")

// Row is catalog merged with DB state for APIs.
type Row struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Kind         string `json:"kind"`
	Installable  bool   `json:"installable"`
	State        string `json:"state"`
	Version      string `json:"version,omitempty"`
	InstalledAt  int64  `json:"installed_at,omitempty"`
	LastError    string `json:"last_error,omitempty"`
	HostAction   string `json:"host_action,omitempty"`
	ComingSoon   string `json:"coming_soon,omitempty"`
}

// ListMerged returns catalog rows with sqlite state.
func ListMerged(ctx context.Context, q *sqlc.Queries) ([]Row, error) {
	rows, err := q.ListAddons(ctx)
	if err != nil {
		return nil, err
	}
	byID := map[string]sqlc.Addon{}
	for _, r := range rows {
		byID[r.ID] = r
	}
	out := make([]Row, 0, len(catalog))
	for _, c := range catalog {
		st := "not_installed"
		var ver string
		var installed int64
		var lastErr string
		if db, ok := byID[c.ID]; ok {
			st = db.State
			if db.Version.Valid {
				ver = db.Version.String
			}
			if db.InstalledAt.Valid {
				installed = db.InstalledAt.Int64
			}
			if db.LastError.Valid {
				lastErr = db.LastError.String
			}
		}
		row := Row{
			ID: c.ID, Title: c.Title, Description: c.Description, Kind: string(c.Kind),
			Installable: c.Installable, State: st, Version: ver, InstalledAt: installed, LastError: lastErr,
			HostAction: c.HostAction,
		}
		if !c.Installable {
			row.ComingSoon = "Available in a later release"
		}
		out = append(out, row)
	}
	return out, nil
}

// PlanInstall aggregates host action Plan previews without mutations.
func PlanInstall(ctx context.Context, reg *hostops.Registry, id string) (hostops.Plan, error) {
	entry, err := ByID(id)
	if err != nil {
		return nil, err
	}
	if !entry.Installable || entry.HostAction == "" {
		return nil, ErrNotInstallable
	}
	act, err := reg.New(entry.HostAction)
	if err != nil {
		return nil, err
	}
	return act.Plan(ctx)
}

// MarkInstalling sets addon state before enqueueing a job.
func MarkInstalling(ctx context.Context, q *sqlc.Queries, id string) error {
	row, err := q.GetAddon(ctx, id)
	if err != nil {
		return err
	}
	if row.State != "not_installed" && row.State != "failed" {
		return fmt.Errorf("addon %s is %s", id, row.State)
	}
	return q.UpsertAddon(ctx, sqlc.UpsertAddonParams{
		ID: id, State: "installing",
		LastError: sql.NullString{},
	})
}

// CompleteInstall marks addon installed after host action success.
func CompleteInstall(ctx context.Context, q *sqlc.Queries, id, version, hostActionID string) error {
	cfg, _ := json.Marshal(map[string]string{"last_host_action_id": hostActionID})
	now := time.Now().Unix()
	return q.UpsertAddon(ctx, sqlc.UpsertAddonParams{
		ID: id, State: "installed",
		Version:     sql.NullString{String: version, Valid: version != ""},
		InstalledAt: sql.NullInt64{Int64: now, Valid: true},
		ConfigJson:  sql.NullString{String: string(cfg), Valid: true},
		LastError:   sql.NullString{},
	})
}

// FailInstall records failure and optional error text.
func FailInstall(ctx context.Context, q *sqlc.Queries, id, msg string) error {
	return q.UpsertAddon(ctx, sqlc.UpsertAddonParams{
		ID: id, State: "failed",
		LastError: sql.NullString{String: msg, Valid: msg != ""},
	})
}

// MarkUninstalling prepares uninstall.
func MarkUninstalling(ctx context.Context, q *sqlc.Queries, id string) error {
	row, err := q.GetAddon(ctx, id)
	if err != nil {
		return err
	}
	if row.State != "installed" {
		return fmt.Errorf("addon %s is %s", id, row.State)
	}
	return q.UpsertAddon(ctx, sqlc.UpsertAddonParams{ID: id, State: "uninstalling"})
}

// LastHostActionID reads the install host action from addon config_json.
func LastHostActionID(row sqlc.Addon) string {
	if !row.ConfigJson.Valid || row.ConfigJson.String == "" {
		return ""
	}
	var cfg map[string]string
	if err := json.Unmarshal([]byte(row.ConfigJson.String), &cfg); err != nil {
		return ""
	}
	return cfg["last_host_action_id"]
}

// CompleteUninstall resets addon to not_installed.
func CompleteUninstall(ctx context.Context, q *sqlc.Queries, id string) error {
	return q.UpsertAddon(ctx, sqlc.UpsertAddonParams{
		ID: id, State: "not_installed",
	})
}

// Health returns ok/warn for installed baseline addons.
func Health(ctx context.Context, reg *hostops.Registry, id string) (string, error) {
	entry, err := ByID(id)
	if err != nil {
		return "", err
	}
	if entry.HostAction == "" {
		return "n/a", nil
	}
	act, err := reg.New(entry.HostAction)
	if err != nil {
		return "", err
	}
	if err := act.Verify(ctx); err != nil {
		return "warn", nil
	}
	return "ok", nil
}

// HostActionFor returns linked host action kind.
func HostActionFor(id string) (string, error) {
	entry, err := ByID(id)
	if err != nil {
		return "", err
	}
	if entry.HostAction == "" {
		return "", fmt.Errorf("no host action for %s", id)
	}
	return entry.HostAction, nil
}
