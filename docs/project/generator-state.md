# Generator state

**Updated:** 2026-11-25

## In progress

None (ready for next pickup).

## Last ship

**#81 P2c — Coraza WAF detect-only & security events** @ **v0.7.22** on `development`.

- Migration `00017`, `coraza_waf` render, events ingest/purge, Application Security WAF UI
- Reviewers: database/api PASS i1; security PASS i2 (backslash escape fix); docs PASS
- Soft-dep #79: fixture ingest; gate on `proxy-waf` installed
- Soft-dep #80: no `rate_limit`
