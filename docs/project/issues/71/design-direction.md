---
github: https://github.com/Joeydekruis/dockyrd/issues/71
wireframe: assets/domains-urls-wireframe.html
---

# Dockyrd: Apps domain and default URLs, design direction

Answers to the open questions in the issue 71 brief, plus the principles, state machine and copy for the boards on the canvas page **Domains & URLs**.

## Principle

Wildcard DNS, verified, then pretty default URLs is the happy path. sslip.io is the zero-DNS fallback and never feels like an error. The UI shows one small block with one DNS row, not a certificate or DNS textbook.

## Naming (user-facing)

| Concept | Name in the UI | Note |
|---|---|---|
| Panel URL | **Dashboard URL** | Where Dockyrd lives. |
| Apps wildcard | **Apps domain** | Pattern like `*.apps.clientx.nl`. |
| Platform hostname | **Default URL** | Per service, badge "Platform". |
| Bring your own hostname | **Your domains** | Badge "Custom". Always a separate section. |

Status pills: **Using sslip.io**, **Not set up**, **Unverified**, **Checking…**, **Verified**, **Check failed**. Row badges: **On your domain** (wildcard URL), **sslip.io**, **Your domain** (custom).

## Decisions where the brief said "assume if silent"

1. **Wildcard convention:** always `*.apps.<apex>`, where apex is the registrable domain of the Dashboard URL host. `panel.clientx.nl` gives `*.apps.clientx.nl`. It is a suggestion chip ("Use *.apps.clientx.nl"), always editable. Engineering needs the public suffix list for the apex.
2. **Hostname shape:** keep `{slug}-{4hex}.{wildcard-suffix}`. Not guessable, collision safe. A clean `{slug}.apps…` would need an explicit product call.
3. **Verify flow:** inline in the settings block (DNS row, probe hostname, Verify). "Show full DNS help" is a link, not a modal.
4. **Primary CTA:** Verify is the prominent button while the apps domain is unverified. Save URL is primary otherwise.
5. **Setup wizard:** a full step "Default URLs for apps", skippable, two choices (Use my domain / Skip for now). Not a collapsed advanced section. Exits: Verified, or Skipped with a pointer to Server settings › Instance.
6. **Platform rows:** with a verified wildcard, users can still add an sslip.io URL as a secondary platform row. The wildcard URL is primary by default. A custom primary domain wins over both, and both platform rows then redirect to it.
7. **Admin only:** nudges and the Instance page are for instance admins. Members see nothing.
8. **Polling:** "Checking DNS…" lives on the button and the pill. No page loader. Suggested: re-check automatically every 30 s while the settings page is open.
9. **Databases:** no Domains section in this release. Per-service wildcards are out of scope.

## State machine: Instance › Apps on your domain

| State | Trigger | What the user sees | Primary action |
|---|---|---|---|
| No panel URL | IP only | Info: apps already get sslip.io URLs. | Set Dashboard URL |
| Panel URL, no wildcard | Dashboard URL saved | Suggestion chip + one DNS line. | Set apps domain |
| Unverified | Wildcard saved | DNS helper (A, `*.apps`, server IPv4), probe hostname, "You can deploy now". | Verify |
| Checking | Verify clicked | Button "Checking DNS…", pill "Checking…". | none |
| Verified | Probe resolves to server IP | Success, verified-at, example URL. | Re-check, Change apps domain |
| Failed | Wrong target or NXDOMAIN | What we saw, what it should be, "up to an hour", docs link. | Verify again |

Changing a verified apps domain asks for confirmation and resets to Unverified. Assumed behavior, to confirm with engineering: existing apps keep their current URLs, new URLs on the new domain appear after verification.

## Nudges

One slim banner, never a modal. Dashboard home (missing, unverified), Service › Domains (service-scoped), Instance settings (inline block, strongest). Dismiss is per user and returns only if the state gets worse. Verifying removes it for everyone.

## Domains tab (applications, static sites, compose)

Two sections, same on all three kinds: **Default URL** (Platform badge, blue accent) and **Your domains** (Custom badge, white card). Frames cover wildcard primary, sslip primary with nudge, both platform rows, custom domain in the same zone (with a "your domain vs the default URL" callout), empty app and compose. Compose keeps its routing line in the header. Routing itself is not redesigned.

## Copy pack

- sslip pill: "Using sslip.io". Wildcard badge: "On your domain".
- Unverified: "You can deploy now. Default URLs use sslip.io until this is verified."
- Failed: "We could not see the record yet. Looking up dockyrd-check.apps.clientx.nl returned 198.51.100.9, but this server is 203.0.113.7. DNS changes can take up to an hour."
- Primary redirect hint: "Redirects to api.clientx.nl".
- Delete last platform URL: "Delete default URL? web-api will only be reachable on your own domains. You can add a default URL again any time."
- Change verified wildcard: "Change apps domain? You will need to verify the new domain. Existing apps keep their current URLs until then."
- Custom vs default: "api.clientx.nl is a name you own and verify with DNS. web-api-9f2c.apps.clientx.nl is generated by Dockyrd and stays as a fallback."

## Responsive (1360 and 390)

Cards stack, the DNS table keeps Type / Name / Value on one line and wraps the value. Buttons wrap under the text, the Verify button stays full width on 390. Status pills never truncate. Banners put the actions under the text.

## Open points for product

1. Can the last platform URL be deleted, or is one always required? The copy above assumes it can.
2. What happens to existing default URLs when the apps domain changes (assumed: kept until re-verified).
3. IPv6: the helper shows one A record. Add an AAAA row if the server has an IPv6 address.
4. Settings navigation: the three settings pages (Profile & security, Git providers, Instance) now share a tab row on the Instance board only. The other two boards need it too.
