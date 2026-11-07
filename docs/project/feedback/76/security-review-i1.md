# Security review - 76 i1
## Verdict: BLOCK
## Boundaries
- `/api/v1/security/*` writes use `requireAdmin`; unauth → 401 covered in tests
- Check/fix maps only allowlisted check IDs → host action kinds (no arbitrary `Registry.New`)
- Addon IDs from static catalog; non-installable add-ons rejected at API
- Host commands via `exec.CommandContext` (no shell); lockout timer uses discrete `systemd-run` args
- Confirm/revert host-actions: admin-only; audit on check run / fix / addon install
## Blockers
- `internal/hostops/ssh_passwordauth.go:46-64` - Pre-check only tests panel UID `authorized_keys` when `passwordauthentication yes`; spec/rubric require blocking password-only sessions (not merely root keys) → enforce session/key preconditions before `Plan` or return 422
- `internal/server/api_security_addons.go:95-106` - `uninstall` updates SQLite only; spec requires host revert for firewall/unattended → enqueue revert job / `Runner.Revert` before `CompleteUninstall`
- `internal/hostops/runner.go:98-100` - `Apply` errors return without reverting; `ssh_passwordauth.go` / `firewall_ensure.go` may leave drop-in/rules on disk while row stays `planned` → revert using captured payload or delete staged files on `Apply` failure
## Notes
- Addon install marks `installed` after `Verify` while lockout actions sit `awaiting_confirm`; aligns with spec §install flow—UI should surface pending confirm
