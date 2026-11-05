package watch

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type notifySettings struct {
	DiskWarnPct int `json:"disk_warn_pct"`
	DiskCritPct int `json:"disk_crit_pct"`
}

func loadNotifySettings(ctx context.Context, q *sqlc.Queries) notifySettings {
	out := notifySettings{DiskWarnPct: 85, DiskCritPct: 95}
	row, err := q.GetSetting(ctx, "notifications")
	if err != nil {
		return out
	}
	raw, _ := json.Marshal(row.Value)
	_ = json.Unmarshal(raw, &out)
	if out.DiskWarnPct == 0 {
		out.DiskWarnPct = 85
	}
	if out.DiskCritPct == 0 {
		out.DiskCritPct = 95
	}
	return out
}

func sqlNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}
