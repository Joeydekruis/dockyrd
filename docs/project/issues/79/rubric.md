---
github: https://github.com/Joeydekruis/dockyrd/issues/79
---

## Done when

- [ ] `go test ./internal/proxybuild/... ./internal/hostops/... ./internal/docker/...` pass
- [ ] `proxy.swap` registered (not stub); `RequiresLockoutConfirm` false
- [ ] Proxybuild workflow + Dockerfile present; cosign on tag push
- [ ] Addon `proxybuild` installable; migration seeds row
- [ ] `proxybuild_installed` exposed in security policy + service security GET

## Reviewers

`none` (Go/docker pipeline; skip frontend-eval)