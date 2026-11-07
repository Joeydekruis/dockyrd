package hostops

import (
	"context"
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/audit"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
)

// RevertByID opens the panel DB and reverts a host action (timer callback / CLI).
func RevertByID(ctx context.Context, cfg config.Config, actionID string) error {
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return err
	}
	defer conn.Close()
	reg := DefaultRegistry(cfg)
	r := NewRunner(conn, q, reg, audit.Logger{Q: q}, NoopLockoutTimer{})
	if err := r.Revert(ctx, actionID, RevertMeta{Source: "cli_timer"}); err != nil {
		return fmt.Errorf("revert %s: %w", actionID, err)
	}
	return nil
}
