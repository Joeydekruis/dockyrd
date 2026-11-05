package watch

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/notify"
)

type certEval struct {
	status string
	level  notify.Level
	emit   bool
	resolve bool
}

// evalCertLifetime derives status and alert intent from leaf validity and now.
func evalCertLifetime(now time.Time, notBefore, notAfter time.Time, warnPct float64) certEval {
	lifetime := notAfter.Sub(notBefore).Seconds()
	if lifetime <= 0 {
		lifetime = 1
	}
	remaining := notAfter.Sub(now)
	remainPct := remaining.Seconds() / lifetime * 100

	if remaining <= 0 {
		return certEval{status: "expired", level: notify.LevelError, emit: true}
	}
	if remainPct <= 5 || remaining <= 72*time.Hour {
		return certEval{status: "expiring", level: notify.LevelError, emit: true}
	}
	warnThreshold := warnPct
	if warnThreshold <= 0 {
		warnThreshold = 15
	}
	if remainPct <= warnThreshold || remaining <= 14*24*time.Hour {
		return certEval{status: "expiring", level: notify.LevelWarning, emit: true}
	}
	if remainPct > 33 {
		return certEval{status: "valid", resolve: true}
	}
	return certEval{status: "valid"}
}

// RunCerts checks TLS certificates served on :443 and updates domain rows.
func RunCerts(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	tickCerts(ctx, cfg, q, n, log)
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			tickCerts(ctx, cfg, q, n, log)
		}
	}
}

func tickCerts(ctx context.Context, cfg config.Config, q *sqlc.Queries, n *notify.Notifier, log *slog.Logger) {
	if q == nil {
		return
	}
	settings := loadNotifySettings(ctx, q)
	rows, err := q.ListDomainsForCertWatch(ctx)
	if err != nil {
		return
	}
	for _, row := range rows {
		if row.DesiredState != "running" {
			continue
		}
		dns := ""
		if row.DnsStatus.Valid {
			dns = row.DnsStatus.String
		}
		checkDomainCert(ctx, q, n, settings.CertWarnPct, row.ID, row.Hostname, dns)
	}
	_ = cfg
	_ = log
}

func checkDomainCert(ctx context.Context, q *sqlc.Queries, n *notify.Notifier, certWarnPct int, domainID, hostname, dnsStatus string) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(&dialer, "tcp", net.JoinHostPort("127.0.0.1", "443"), &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true,
	})
	if err != nil {
		_ = q.UpdateDomainCertStatus(ctx, sqlc.UpdateDomainCertStatusParams{
			CertStatus:    sqlNullString("missing"),
			CertExpiresAt: sql.NullInt64{},
			ID:            domainID,
		})
		return
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return
	}
	leaf := certs[0]
	if !hostnameMatchesCert(hostname, leaf) {
		_ = q.UpdateDomainCertStatus(ctx, sqlc.UpdateDomainCertStatusParams{
			CertStatus:    sqlNullString("missing"),
			CertExpiresAt: sql.NullInt64{},
			ID:            domainID,
		})
		return
	}
	expires := leaf.NotAfter.Unix()
	eval := evalCertLifetime(time.Now(), leaf.NotBefore, leaf.NotAfter, float64(certWarnPct))

	_ = q.UpdateDomainCertStatus(ctx, sqlc.UpdateDomainCertStatusParams{
		CertStatus:    sqlNullString(eval.status),
		CertExpiresAt: sqlNullInt64(expires),
		ID:            domainID,
	})
	if eval.status == "valid" || eval.status == "expiring" {
		now := time.Now().Unix()
		_ = q.UpdateDomainDNSStatus(ctx, sqlc.UpdateDomainDNSStatusParams{
			DnsStatus:    sqlNullString("active"),
			DnsCheckedAt: sqlNullInt64(now),
			ID:           domainID,
		})
	}

	if n == nil {
		return
	}
	if eval.resolve {
		_ = n.Emit(ctx, notify.Event{
			Type: notify.TypeCertExpiring, Level: notify.LevelWarning, State: notify.StateResolved,
			DedupKey: "cert:" + hostname, Title: "Certificate renewed · " + hostname,
		})
		return
	}
	if !eval.emit {
		return
	}
	title := fmt.Sprintf("Certificate expires soon · %s", hostname)
	if eval.status == "expired" {
		title = fmt.Sprintf("Certificate expired · %s", hostname)
	}
	detail := fmt.Sprintf("NotAfter %s (DNS: %s).", leaf.NotAfter.UTC().Format(time.RFC3339), dnsStatus)
	_ = n.Emit(ctx, notify.Event{
		Type: notify.TypeCertExpiring, Level: eval.level, State: notify.StateFiring,
		DedupKey: "cert:" + hostname, Title: title, Detail: detail, Href: "/domains",
		Data: map[string]any{"hostname": hostname, "dns_status": dnsStatus},
	})
}

func hostnameMatchesCert(hostname string, leaf *x509.Certificate) bool {
	if err := leaf.VerifyHostname(hostname); err == nil {
		return true
	}
	h := strings.ToLower(strings.TrimSuffix(hostname, "."))
	for _, name := range leaf.DNSNames {
		if strings.EqualFold(name, h) {
			return true
		}
	}
	if strings.EqualFold(leaf.Subject.CommonName, h) {
		return true
	}
	return false
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
