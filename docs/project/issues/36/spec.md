---
github: https://github.com/Joeydekruis/dockyrd/issues/36
---

# #36 · Admin CLI

## AC

- [ ] `dockyard admin reset-password <email> [password]` updates argon2id hash (min 12 chars)
- [ ] `dockyard admin disable-2fa <email>` clears TOTP fields
- [ ] Uses control-plane DB via `config.Load()` paths
