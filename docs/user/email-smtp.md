# Email (SMTP)

Password reset and team invites can send email when SMTP is configured.

## Configure in the dashboard

1. Sign in as an **admin**.
2. Open **Settings → Instance**.
3. Set **Public URL** to the HTTPS origin users open in the browser (used in email links).
4. Fill in **SMTP** (host, port, username, password, From address) and save.
5. Use **Test SMTP** to verify delivery.

Passwords are encrypted with the same master key as other secrets (`secret.key` under your config directory).

## Environment fallback

If **Public URL** is empty in settings, the panel uses `DOCKYARD_PUBLIC_URL` from the environment when building links.

## Without SMTP

- **Forgot password** shows instructions to ask an admin or run `sudo dockyard admin reset-password`.
- **Invites** still work: the dashboard shows a copyable invite link after you create an invite.
