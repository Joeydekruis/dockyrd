package events

import (
	"context"
	"database/sql"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// RetentionDays is how long client IPs stay in security_events (not analytics).
const RetentionDays = 7

// Inserter persists parsed WAF hits.
type Inserter struct {
	Q *sqlc.Queries
}

// InsertWAFHit resolves hostname → service and inserts source=waf.
func (in Inserter) InsertWAFHit(ctx context.Context, hit ParsedHit) (sqlc.SecurityEvent, error) {
	if in.Q == nil {
		return sqlc.SecurityEvent{}, sql.ErrConnDone
	}
	dom, err := in.Q.GetDomainByHostname(ctx, hit.Hostname)
	if err != nil {
		return sqlc.SecurityEvent{}, err
	}
	return in.Q.InsertSecurityEvent(ctx, sqlc.InsertSecurityEventParams{
		Ts:        hit.TS.UTC().Format(time.RFC3339Nano),
		ServiceID: dom.ServiceID,
		Source:    "waf",
		Rule:      nullStr(hit.Rule),
		Path:      nullStr(hit.Path),
		Action:    nullStr(hit.Action),
		Ip:        nullStr(hit.IP),
		Detail:    nullStr(hit.Detail),
	})
}

// PurgeOlderThan deletes rows with ts before cutoff (RFC3339 comparable lexicographically).
func PurgeOlderThan(ctx context.Context, q *sqlc.Queries, cutoff time.Time) (int64, error) {
	if q == nil {
		return 0, nil
	}
	return q.PurgeSecurityEventsBefore(ctx, cutoff.UTC().Format(time.RFC3339Nano))
}

// PurgeDefaultRetention deletes events older than RetentionDays.
func PurgeDefaultRetention(ctx context.Context, q *sqlc.Queries, now time.Time) (int64, error) {
	return PurgeOlderThan(ctx, q, now.Add(-time.Duration(RetentionDays)*24*time.Hour))
}

func nullStr(s string) sql.NullString {
	s = stringsTrim(s)
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	return s
}
