---
github: https://github.com/Joeydekruis/dockyrd/issues/14
reviewers: none
---

## Done when

- Spec AC met tests
- `go test ./internal/docker/...` green without Docker
- Integration test skipped in default CI unit pass

## Gates

```bash
go test ./internal/docker/...
```
