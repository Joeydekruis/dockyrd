# Security review - 81 i2
## Verdict: PASS
## Boundaries
- Path exceptions → Coraza `@beginsWith` (`ValidatePathExceptions` + `sanitizeDirectivePath`) after `\`/`"` fix
- PUT still validates via `ValidatePathExceptions` → 422 before persist
- Prior i1 OK: 7-day purge, events authz, `coraza_waf` emit gate, audit parts / no body secrets
## Blockers
- None
## Notes
- Escape payloads (`/x\"`, `/x\`, `/api/\..\…`) rejected in both validators; `buildCorazaDirectives` omits sanitized-empty entries (`waf_test.go`, `waf_render_test.go`).
- Defense-in-depth retained: API reject + render-time sanitize.
