# Privacy

## Analytics vs security events

**Analytics** aggregates request rates, latency, and status codes. It does **not** store client IP addresses.

**Security events** (Application Security → events table) record hits from traffic protection such as the WAF. Those rows may include the client IP so you can investigate abuse. Security-event IPs are retained for **7 days**, then purged automatically. They are stored separately from analytics and are never mixed into analytics rollups.
