package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
	"github.com/Joeydekruis/dockyrd/internal/security/addons"
)

const (
	KindSecurityAddonInstall   = "security.addon.install"
	KindSecurityAddonUninstall = "security.addon.uninstall"
)

type SecurityAddonPayload struct {
	AddonID     string `json:"addon_id"`
	ActorUserID string `json:"actor_user_id"`
}

// EnqueueSecurityAddonInstall schedules install work.
func EnqueueSecurityAddonInstall(ctx context.Context, q *sqlc.Queries, addonID, actor string) (string, error) {
	payload, _ := json.Marshal(SecurityAddonPayload{AddonID: addonID, ActorUserID: actor})
	now := time.Now().Unix()
	id := ulid.Make().String()
	err := q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID: id, Kind: sql.NullString{String: KindSecurityAddonInstall, Valid: true},
		Payload: payload, RunAfter: sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
	return id, err
}

// EnqueueSecurityAddonUninstall schedules uninstall work.
func EnqueueSecurityAddonUninstall(ctx context.Context, q *sqlc.Queries, addonID, actor string) (string, error) {
	payload, _ := json.Marshal(SecurityAddonPayload{AddonID: addonID, ActorUserID: actor})
	now := time.Now().Unix()
	id := ulid.Make().String()
	err := q.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		ID: id, Kind: sql.NullString{String: KindSecurityAddonUninstall, Valid: true},
		Payload: payload, RunAfter: sql.NullInt64{Int64: now, Valid: true},
		CreatedAt: sql.NullInt64{Int64: now, Valid: true},
	})
	return id, err
}

// RunSecurityAddonInstall executes host action chain for a baseline addon.
func RunSecurityAddonInstall(ctx context.Context, cfg config.Config, db *sql.DB, q *sqlc.Queries, payload SecurityAddonPayload) error {
	entry, err := addons.ByID(payload.AddonID)
	if err != nil {
		return err
	}
	if !entry.Installable {
		return addons.ErrNotInstallable
	}
	reg := hostops.DefaultRegistry(cfg)
	runner := hostops.NewRunner(db, q, reg, audit.Logger{Q: q}, hostops.NewSystemdLockoutTimer())
	actionID, err := runner.Plan(ctx, entry.HostAction, payload.ActorUserID)
	if err != nil {
		_ = addons.FailInstall(ctx, q, payload.AddonID, err.Error())
		return err
	}
	if err := runner.Apply(ctx, actionID, payload.ActorUserID); err != nil {
		_ = addons.FailInstall(ctx, q, payload.AddonID, err.Error())
		return err
	}
	if err := runner.Verify(ctx, actionID); err != nil {
		_ = addons.FailInstall(ctx, q, payload.AddonID, err.Error())
		return err
	}
	return addons.CompleteInstall(ctx, q, payload.AddonID, entry.Version, actionID)
}

// RunSecurityAddonUninstall reverts the last host action and marks addon not_installed.
func RunSecurityAddonUninstall(ctx context.Context, cfg config.Config, db *sql.DB, q *sqlc.Queries, payload SecurityAddonPayload) error {
	row, err := q.GetAddon(ctx, payload.AddonID)
	if err != nil {
		return err
	}
	actionID := addons.LastHostActionID(row)
	reg := hostops.DefaultRegistry(cfg)
	runner := hostops.NewRunner(db, q, reg, audit.Logger{Q: q}, hostops.NewSystemdLockoutTimer())
	if actionID != "" {
		if err := runner.Revert(ctx, actionID, hostops.RevertMeta{Source: "addon.uninstall", Actor: payload.ActorUserID}); err != nil {
			return err
		}
	}
	return addons.CompleteUninstall(ctx, q, payload.AddonID)
}
