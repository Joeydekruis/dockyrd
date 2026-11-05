# Notification channels

Dockyard can deliver platform alerts to Slack, Discord, email, and signed HTTPS webhooks without keeping the dashboard open.

## Setup (admin)

1. Open **Settings → Server → Channels** in Dockyrd.
2. Choose a channel type and paste credentials (Slack/Discord incoming webhook URL, or webhook URL + signing secret).
3. Select which events to deliver (defaults include deploy failures, backup issues, disk pressure, and certificate warnings).
4. Use **Test** to verify delivery before relying on the channel.

SMTP must be configured under **Settings → Server → Email** before email channels can send.

## Webhook verification

Outbound webhooks use schema `dockyard.notification.v1` with headers:

- `X-Dockyard-Event` — event type (e.g. `deploy.failed`)
- `X-Dockyard-Delivery` — unique delivery id (idempotency key)
- `X-Dockyard-Timestamp` — unix seconds
- `X-Dockyard-Signature` — `sha256=` HMAC over `timestamp + "." + body`

Reject replays older than five minutes. Verify with the signing secret configured on the channel.

## Troubleshooting

- **Degraded channel** — recent delivery failures; check the delivery log in the UI and fix credentials or network paths.
- **Disabled channel** — permanent auth or 404 from Slack/Discord; recreate the webhook.
- **No email for deploy** — inbox email and notification email dedupe by address; use a dedicated ops alias on the email channel if needed.
- **Alerts while building** — interrupted-deploy marking runs once at panel startup only; long builds should stay `building`.

## Limits

The panel cannot alert when the host is entirely down or disk is 100% full (SQLite may not accept writes). Disk alerts fire at configurable thresholds (default 85% / 95%) before that point. Optional external heartbeat (issue #66) covers dead-man scenarios.
