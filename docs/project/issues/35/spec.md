---
github: https://github.com/Joeydekruis/dockyrd/issues/35
---

# #35 · dockyard doctor

## AC

- [ ] `dockyard doctor` prints checks for data/config paths, master key, SQLite, disk guard, Docker ping, Caddy admin socket, listen address, clock
- [ ] Exit code 1 when any check is `fail`; warnings do not fail
- [ ] Unit test covers missing layout path
