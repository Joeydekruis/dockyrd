---
github: https://github.com/Joeydekruis/dockyrd/issues/36
---

# #36 · Admin CLI

## AC

- [x] `dockyrd admin reset-password <email> [password]` updates argon2id hash (min 12 chars)
- [x] `dockyrd admin disable-2fa <email>` clears TOTP fields
- [x] Uses control-plane DB via `config.Load()` paths
