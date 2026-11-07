---
github: https://github.com/Joeydekruis/dockyrd/issues/35
---

# #35 · dockyrd doctor

## AC

- [x] `dockyrd doctor` prints checks for data/config paths, master key, SQLite, disk guard, Docker ping, Caddy admin socket, listen address, clock
- [x] Exit code 1 when any check is `fail`; warnings do not fail
- [x] Unit test covers missing layout path
