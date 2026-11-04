---
name: push-to-main
description: >-
  Full build verification then merge development into main and push.
  Use when the user asks to push to main, merge to main, ship main, or release
  from development.
---

# Push to main (`development` → `main`)

Use **only** when the human explicitly asks to merge/push/ship to **`main`**.
Day-to-day work stays on `development` ([git-branch](../../rules/git-branch.mdc)).

## Preconditions

- [ ] Human requested main ship
- [ ] `development` is clean (or only intentional ship commits) and pushed
- [ ] Latest cycle changelog/version already closed on `development` ([changelog](../../rules/changelog.mdc))

## 1. Sync

```bash
git checkout development
git pull origin development
git fetch origin main
```

Confirm HEAD is `development` and up to date.

## 2. Full build check (must all pass)

Run the **full** matrix from `docs/project/setup.md` — not a subset. Abort merge on any failure.

When Makefile exists (briefing §14.3):

```bash
make lint
make test
make build
./scripts/check-no-radix.sh
pnpm --dir web lint
pnpm --dir web build
go test -tags=integration ./...   # when Docker available
make e2e                          # when e2e/ exists and Docker available
```

If a gate is impossible in this environment (e.g. Docker down), **stop** and report which checks passed/failed — do **not** merge to `main` on partial green unless the human explicitly waives named gates.

## 3. Merge development → main

```bash
git checkout main
git pull origin main
git merge development
```

Resolve conflicts if any; re-run failed gates after conflict resolution before continuing.

## 4. Push main

```bash
git push origin main
git checkout development
```

Never `--force` to `main`.

## 6. GitHub release (only when Joey asked for release / tag / ship binaries)

After section 5, follow [release-ship](../../rules/release-ship.mdc) (skill [release-ship](../release-ship/SKILL.md)):

1. Push tag **`v$(node -p "require('./package.json').version")`** to `origin` (triggers GoReleaser).
2. **Wait** for `.github/workflows/release.yml` **success**.
3. **Verify** both `dockyard_{version}_linux_amd64.tar.gz` and `dockyard_{version}_linux_arm64.tar.gz` (+ `checksums.txt`) on the GitHub Release.

Do **not** tell Joey the release is ready for self-update until step 3 passes.

## 7. Report (merge + optional release)

Return: merge commit SHA (if any), `main` / `development` tips, version from root `package.json` or binary `-version`, which gates ran + pass/fail. If a release was cut: tag, Release workflow status, and asset checklist (amd64 + arm64 tarballs).

## Anti-patterns

- Merging to `main` without the full build check
- Force-pushing `main`
- Staying on `main` to continue feature work
- Skipping changelog/version close on `development` before the ship
- Cutting a GitHub tag or telling users to self-update before [release-ship](../../rules/release-ship.mdc) workflow success and asset verification
