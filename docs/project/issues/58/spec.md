---
github: https://github.com/Joeydekruis/dockyrd/issues/58
title: Resource links Fase 0 — network prerequisites
size: S
track: Direct
---

## Goal

Prepare Docker networking so per-link isolation (Fase 2) is meaningful.

## AC

1. Apps connect to `dy-edge` only when they have at least one domain row.
2. Regression test documents cross-project reachability expectation on edge.
3. Document/merge `default-address-pools` for install path; doctor warns if missing.

## References

- `docs/project/briefings/resource-linking.md` §6 Fase 0, §2.1 item 3
