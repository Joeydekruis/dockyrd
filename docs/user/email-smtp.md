# Email (SMTP)

Password reset and team invites can send email when SMTP is configured.

## Configure in the dashboard

1. Sign in as an **admin**.
2. Open **Server settings → Email** (`/settings/server/email`).
3. Set **Public URL** on **Server settings → General** to the HTTPS origin users open in the browser (used in email links).
4. Fill in **SMTP** (host, port, username, password, From address) and save.
5. Use **Test SMTP** to verify delivery. Attempts appear in **Email log** on the same page.

Passwords are encrypted with the same master key as other secrets (`secret.key` under your config directory).

## Environment fallback

If **Public URL** is empty in settings, the panel uses `dockyrd_PUBLIC_URL` from the environment when building links.

## Without SMTP

- **Forgot password** shows instructions to ask an admin or run `sudo dockyrd admin reset-password`.
- **Invites** still work: the dashboard shows a copyable invite link after you create an invite.
