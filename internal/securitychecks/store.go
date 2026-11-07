package securitychecks

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/hostops"
)

const cacheFreshHours = 24

// LoadCached returns a report when ran_at is within 24h.
func LoadCached(ctx context.Context, q *sqlc.Queries, now time.Time) (Report, bool, error) {
	row, err := q.GetSecurityCheckCache(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Report{}, false, nil
		}
		return Report{}, false, err
	}
	var rep Report
	if err := json.Unmarshal([]byte(row.ResultsJson), &rep); err != nil {
		return Report{}, false, err
	}
	age := now.Unix() - row.RanAt
	if age > cacheFreshHours*3600 {
		rep.Stale = true
		return rep, false, nil
	}
	rep.Cached = true
	rep.RanAt = row.RanAt
	return rep, true, nil
}

// Save persists the latest report.
func Save(ctx context.Context, q *sqlc.Queries, rep Report) error {
	b, err := json.Marshal(rep)
	if err != nil {
		return err
	}
	return q.UpsertSecurityCheckCache(ctx, sqlc.UpsertSecurityCheckCacheParams{
		RanAt:       rep.RanAt,
		ResultsJson: string(b),
	})
}

// FixKindForCheck maps check id to host action kind.
func FixKindForCheck(id string) (string, bool) {
	switch id {
	case "firewall.active":
		return hostops.KindFirewallEnsure, true
	case "ssh.password_auth":
		return hostops.KindSSHPasswordAuthDisable, true
	case "updates.unattended":
		return hostops.KindUnattendedEnable, true
	default:
		return "", false
	}
}
