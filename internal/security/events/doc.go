// Package events persists WAF (and later rate-limit / CrowdSec) hits in security_events.
//
// Log path: host `{DataDir}/caddy/coraza-audit.log` (container `/var/lib/dockyrd/caddy/coraza-audit.log`),
// written when SecAuditLog is set in coraza_waf directives. Client IPs here are security-only
// and purged after RetentionDays (default 7) — separate from analytics.
//
// Mapping strategy: Coraza audit JSON Host header → domains.hostname → service_id.
package events
