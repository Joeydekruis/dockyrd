# Build resilience (free tier)

Product plan for cgroup isolation, build budgets, admission control, pressure freeze, and free “build elsewhere” paths. Source briefing: Nov 2026 (Joey de Kruis).

**Goal:** A build must never take down the panel, proxy, SSH, or live apps. On undersized servers: slower → wait → clear failure (in that order).

## GitHub tracking

| Issue | Layer / scope | Milestone |
|-------|----------------|-----------|
| [#49](https://github.com/Joeydekruis/dockyrd/issues/49) | Docs & BRIEFING alignment | Plan |
| [#50](https://github.com/Joeydekruis/dockyrd/issues/50) | Layer 1 slices + OOM, `internal/resources`, installer, BuildKit placement | M1 |
| [#51](https://github.com/Joeydekruis/dockyrd/issues/51) | Layer 3 admission + `waiting`, peaks, schema migration | M2 |
| [#52](https://github.com/Joeydekruis/dockyrd/issues/52) | Layer 2 static sites on BuildKit + buildkitd.toml sizing | M2 |
| [#53](https://github.com/Joeydekruis/dockyrd/issues/53) | Layer 4 PSI watchdog (freeze/thaw) | M4 |
| [#54](https://github.com/Joeydekruis/dockyrd/issues/54) | Layer 5 env injection + hints | M2 |
| [#55](https://github.com/Joeydekruis/dockyrd/issues/55) | Layer 6a GitHub Actions workflow + deploy hook | M2 |
| [#56](https://github.com/Joeydekruis/dockyrd/issues/56) | API/UI: build budget, resource strip, builders settings | M4 |
| [#57](https://github.com/Joeydekruis/dockyrd/issues/57) | Layer 6b remote BuildKit + doctor + e2e fixtures | M4 / v0.2 |

Local specs: `docs/project/issues/<n>/spec.md` per issue.

## Architecture (six layers)

1. **Slices** — `dy.slice` → `dy-system`, `dy-apps`, `dy-build`; OOM scores; Docker `CgroupParent` per container kind.
2. **BuildKit** — All builds under `dy-buildkit` in `dy-build.slice`; `buildkitd.toml` from server RAM/disk; static sites via generated Dockerfile + `buildx --output type=local`.
3. **Budget & admission** — Formula from meminfo + slice working sets; set `MemoryHigh`/`MemoryMax`/`MemorySwapMax` on `dy-build.slice`; `waiting` deployment state when peak > budget.
4. **Watchdog** — PSI-driven `cgroup.freeze` on build slice only; thaw on startup in reconciler.
5. **Tuning** — `NODE_OPTIONS`, cache mounts, non-blocking hints.
6. **Build elsewhere** — GitHub Actions generator + deploy hook; optional remote BuildKit (mTLS).

## Deployment state machine change

Insert **`waiting`** between `queued` and clone/build work. Non-terminal; reconciler re-queues after restart.

## Go layout

```
internal/resources/   # cgroup, PSI, systemd slice props, meminfo (fake for tests/macOS)
internal/build/
  admission.go
  watchdog.go
  env.go
  staticsite.go       # BuildKit path
  ghactions.go
  remote.go
```

Only build runner + watchdog call `SetBuildLimits` / `FreezeBuild`; reconciler thaws once at start.

## Verification

- Unit: budget table, PSI parse, watchdog FSM with fake `Resources`, workflow YAML golden files.
- Integration: 1 GB VM scenarios in briefing (memhog, cpuhog, spiky-app, etc.) under `e2e/fixtures/` (#57).
- `dockyard doctor`: cgroup v2, slices, freeze=0, BuildKit config, OOM outside build slice.

## Related

- [`docs/BRIEFING.md`](../BRIEFING.md) — § build resilience (cross-ref)
- [`post-m5-backlog.md`](post-m5-backlog.md)
