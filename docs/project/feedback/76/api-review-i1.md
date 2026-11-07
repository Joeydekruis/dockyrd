# API review - 76 i1
## Verdict: BLOCK
## Scope
- `GET|POST /api/v1/security/checks`, `POST …/checks/{id}/fix`, `GET|POST /api/v1/security/addons/*` (+ existing host-actions)
- OpenAPI: unchanged (no generated spec in tree yet)
## Smoke
- `go test ./internal/server/... -run Security`: pass (401 unauth GET checks, owner 200, crowdsec install 422)
## Blockers
- `api_security_addons.go:99` — uninstall calls `CompleteUninstall` synchronously after `MarkUninstalling`; spec requires host revert (firewall/unattended) via job pattern (`EnqueueSecurityAddonUninstall` exists but unused; worker has no uninstall kind). API returns `200 not_installed` while host protection may remain active.
## Notes
- All scoped routes use `requireAdmin` before work (aligned with spec default Admin+).
- Install: `202` + `job_id`; plan/install preconditions map to `422`; fix maps `ErrPrecondition` → `422`.
