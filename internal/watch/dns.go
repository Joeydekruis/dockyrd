package watch

import (
	"context"
	"log/slog"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dns"
)

// RunDNS periodically resolves custom domain A records against the server IP.
func RunDNS(ctx context.Context, cfg config.Config, q *sqlc.Queries, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Minute
	}
	tickDNS(ctx, cfg, q, log)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tickDNS(ctx, cfg, q, log)
		}
	}
}

func tickDNS(ctx context.Context, cfg config.Config, q *sqlc.Queries, log *slog.Logger) {
	if q == nil {
		return
	}
	expected := cfg.EffectivePublicIPv4()
	rows, err := q.ListDomainsForDNSWatch(ctx)
	if err != nil {
		return
	}
	now := time.Now().Unix()
	for _, row := range rows {
		if row.IsPlatform != 0 {
			continue
		}
		status := dns.EvaluateARecord(ctx, row.Hostname, expected)
		if err := q.UpdateDomainDNSStatus(ctx, sqlc.UpdateDomainDNSStatusParams{
			DnsStatus:    sqlNullString(status),
			DnsCheckedAt: sqlNullInt64(now),
			ID:           row.ID,
		}); err != nil && log != nil {
			log.Warn("update domain dns status", "hostname", row.Hostname, "err", err)
		}
	}
}

// CheckDomainDNS resolves one custom domain and persists dns_status.
func CheckDomainDNS(ctx context.Context, cfg config.Config, q *sqlc.Queries, domainID string) (string, error) {
	if q == nil {
		return "pending", nil
	}
	row, err := q.GetDomainByID(ctx, domainID)
	if err != nil {
		return "", err
	}
	if row.IsPlatform != 0 {
		if row.DnsStatus.Valid {
			return row.DnsStatus.String, nil
		}
		return "active", nil
	}
	status := dns.EvaluateARecord(ctx, row.Hostname, cfg.EffectivePublicIPv4())
	now := time.Now().Unix()
	if err := q.UpdateDomainDNSStatus(ctx, sqlc.UpdateDomainDNSStatusParams{
		DnsStatus:    sqlNullString(status),
		DnsCheckedAt: sqlNullInt64(now),
		ID:           domainID,
	}); err != nil {
		return "", err
	}
	return status, nil
}
