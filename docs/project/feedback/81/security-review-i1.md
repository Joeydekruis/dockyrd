# Security review - 81 i1
## Verdict: BLOCK
## Boundaries
- IP retention/purge (`RetentionDays=7`, worker tick + `PurgeSecurityEventsBefore`)
- Events GET authz (`requireProjectScope` + `service_id` list; token project match)
- Path exceptions → Coraza `directives` (`ValidatePathExceptions` + `sanitizeDirectivePath`)
- `coraza_waf` emit gate (`ProxyWAFInstalled` + `AllowWAF` + mode) in render, not UI-only
- Ingest `detail` truncation / audit parts (no body secrets)

## Blockers
- `internal/proxy/waf_handles.go:65` / `internal/securitypolicy/waf.go:76` - path exceptions reject `"`/newlines/`..` but **not `\`**; a stored prefix like `/x\" …` can break out of `"@beginsWith …"` into Coraza directive actions (Member/Admin PUT). → Reject `\` (and keep rejecting `"`) in both `ValidatePathExceptions` and `sanitizeDirectivePath`; add unit/golden coverage for escape payloads.

## Notes
- 7-day purge is wired (`store.RetentionDays`, `Worker.TickOnce`, `runtime.go` loop); IPs stay in `security_events` only.
- Events list is parameterized by `service_id` after project resolve; API tokens cannot cross projects.
- Addon gate enforced in `corazaWAFHandler` / sync `proxy-waf` installed; static `AllowWAF=false`.
