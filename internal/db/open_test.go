package db

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

func TestOpenMigratesAndQueries(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "control.db")

	conn, q, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx := context.Background()

	count, err := q.CountUsers(ctx)
	if err != nil {
		t.Fatalf("CountUsers: %v", err)
	}
	if count != 0 {
		t.Fatalf("CountUsers = %d, want 0", count)
	}

	payload, err := json.Marshal(map[string]bool{"complete": false})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := q.UpsertSetting(ctx, sqlc.UpsertSettingParams{Key: "setup", Value: payload}); err != nil {
		t.Fatalf("UpsertSetting: %v", err)
	}

	row, err := q.GetSetting(ctx, "setup")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if row.Key != "setup" {
		t.Fatalf("key = %q", row.Key)
	}
}
