package watch

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

type notifySettings struct {
	DiskWarnPct   int    `json:"disk_warn_pct"`
	DiskCritPct   int    `json:"disk_crit_pct"`
	CertWarnPct   int    `json:"cert_warn_pct"`
	HeartbeatURL  string `json:"heartbeat_url"`
}

func loadNotifySettings(ctx context.Context, q *sqlc.Queries) notifySettings {
	out := notifySettings{DiskWarnPct: 85, DiskCritPct: 95, CertWarnPct: 15}
	row, err := q.GetSetting(ctx, "notifications")
	if err != nil {
		return out
	}
	raw, err := notifySettingBytes(row.Value)
	if err != nil {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	if out.DiskWarnPct == 0 {
		out.DiskWarnPct = 85
	}
	if out.DiskCritPct == 0 {
		out.DiskCritPct = 95
	}
	if out.CertWarnPct == 0 {
		out.CertWarnPct = 15
	}
	return out
}

func notifySettingBytes(v any) ([]byte, error) {
	switch t := v.(type) {
	case []byte:
		return t, nil
	case string:
		return []byte(t), nil
	default:
		return json.Marshal(v)
	}
}

func sqlNullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}
