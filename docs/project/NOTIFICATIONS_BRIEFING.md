# Dockyrd: briefing "Notificatiekanalen" (Slack, Discord, e-mail, webhook)

> Status: bouwbriefing, Feature-track · Datum: 5 november 2026 · Eigenaar: Joey de Kruis
> Aanleiding: een mislukte deploy, volle disk, verlopen certificaat of mislukte backup moet je bereiken zonder dat het dashboard open staat. Briefing `dockyrd_BRIEFING.md` §1.2 (v0.2: "Events-tabel bestaat al; alleen kanalen toevoegen").
> Bronnen: codebase `dockyrd` (branch `development`, v0.0.45), `dockyrd_BRIEFING.md` §3, §8.3, §11.4, §15, `RESOURCE_LINKING_BRIEFING.md` (conventies), externe docs (bijlage D).

---

## 0. Samenvatting

1. **Doel:** vier soorten alerts (deploy failed, disk vol, certificaat verloopt, backup failed) bezorgen via **Slack, Discord, e-mail en een generieke webhook**, betrouwbaar, zonder zware dependencies, met een footprint die binnen de 60 MB-RSS-eis blijft.
2. **Belangrijkste bevinding uit de codebase: "alleen kanalen toevoegen" klopt maar half.** De inbox, de events-tabel en SMTP bestaan, maar:
   - er is **geen centrale dispatch**: de deploy-runtime roept `inbox.Dispatcher` rechtstreeks aan (`deploy/runtime.go:354`), en `events` heeft maar twee schrijvers (`recordDeployEvent`, `admission.go`);
   - **drie van de vier gevraagde events bestaan niet als signaal**: disk (alleen een build-guard + metrics), certificaat (`domains.cert_expires_at` wordt nergens gevuld) en backup (de runner zet alleen `status='failed'` in `database_backup_runs`);
   - de **job-queue is ongeschikt** voor bezorging (serieel, geen retry/backoff, een build van 20 min blokkeert alles);
   - de reconciler **markeert elke 30 s lopende builds als `failed (interrupted)`** (gelezen, nog niet gereproduceerd). Dat moet eerst gefixt, anders wordt `deploy.failed` een bron van valse alerts (§2.1 gat 7);
   - `mail.SendPlain` heeft **geen timeouts** en **geen CRLF-sanitering**, en wordt nu synchronou vanuit het deploy-pad aangeroepen.
3. **Voorstel in drie lagen:** **(a) bronnen** (detectors: disk, certificaat, backup-overdue + hooks op deploy/backup), **(b) kern** (`internal/notify`: getypeerd `Event`, alert-state met dedup/hysterese, transactionele **outbox** met eigen dispatcher, retries, circuit-breaker), **(c) kanalen** (vier kleine `Sender`-implementaties zonder SDK's, alleen `net/http` en `net/smtp`).
4. **Ontwerpkeuzes in één regel:** instance-brede kanalen (admin beheert), typed events als opt-in (ongetypeerde events gaan nooit naar buiten), at-least-once met delivery-ID, secrets versleuteld met de bestaande master-key, SSRF-guard op dial-niveau, `resolved`-berichten voor stateful alerts.
5. **Fasering:** F0 hardening (direct-track) → F1 kern + webhook/Slack/Discord + deploy → F2 e-mail + backup → F3 disk + certificaat + alert-state-UI → F4 optioneel (crash/OOM, updates, heartbeat, digest).
6. **Eerlijke beperking:** een server die plat ligt of 100 % vol is kan zichzelf niet meer melden. Mitigatie: alert **voor** volraken (85 %/95 %) en een optionele **heartbeat** (dead-man's switch) in F4. Zie §9.

---

## 1. Doel, scope en niet-functionele eisen

### 1.1 In scope (v1)

| Onderdeel | Inhoud |
|---|---|
| Kanalen | Slack (incoming webhook), Discord (webhook), e-mail (bestaande SMTP), generieke webhook (JSON, HMAC-ondertekend) |
| Events (P0, gevraagd) | `deploy.failed`, `disk.high`, `cert.expiring`, `backup.failed` |
| Events (P1, goedkoop meeliften) | `deploy.succeeded` (default uit extern), `backup.overdue`, `backup.succeeded` (default uit), `notification.channel_failed`, `test.ping` |
| Routering | Per kanaal: aan/uit, minimaal niveau, lijst van event-types, optioneel beperkt tot één project |
| Betrouwbaarheid | Outbox, retries met backoff, `Retry-After`, circuit-breaker, crash-herstel, dedup, `resolved`-berichten |
| Beheer | Admin-UI (lijst, toevoegen, test, bezorglog, opnieuw proberen), REST-API, audit-log |

### 1.2 Buiten scope (nu)

Per-user kanalen, Slack/Discord-OAuth-apps en interactieve knoppen (ack vanuit Slack), Microsoft Teams, Telegram, PagerDuty/Opsgenie, SMS, escalatiebeleid, stille uren, bericht-templates bewerken in de UI, digest-berichten (F4), inkomende webhooks. Een generieke webhook dekt ntfy/Gotify/n8n/Zapier voor wie ze wil.

### 1.3 Niet-functionele eisen (hard)

| Eis | Meetbare invulling |
|---|---|
| Lightweight | **Geen nieuwe Go-dependencies.** Alleen `net/http`, `net/smtp`, `crypto/hmac`. Idle: dispatcher is één goroutine met een tick van 2 s + wake-channel; dockyrd blijft < 60 MB RSS (`scripts/measure-rss.sh` voor/na) |
| Stabiel | `Emit` blokkeert nooit en faalt nooit een deploy/backup: alleen een korte DB-transactie, geen netwerk-IO, nil-safe |
| Betrouwbaar | Een event dat is gecommit wordt bezorgd of expliciet `dead` gemarkeerd (zichtbaar in UI). Na `kill -9` geen verlies, hooguit een dubbel bericht (at-least-once, delivery-ID als idempotentiesleutel) |
| Veilig | Secrets versleuteld (AES-GCM, bestaande master-key), nooit in API/logs; SSRF-guard; geen redirects; geen secret-waarden in berichten |
| Voorspelbaar | Rate-limit per kanaal, dedup voor stateful alerts, geen alert-stormen bij flapperende condities |
| Eenvoudig | Setup in < 1 minuut: kies kanaal → plak URL → "Send test" → klaar; defaults zijn verstandig |

---

## 2. Analyse: waar staan we nu (codebase)

Alles hieronder is gelezen in de repo (branch `development`, `package.json` v0.0.45), niet gedraaid.

| Onderdeel | Stand | Bron |
|---|---|---|
| Events | Tabel `events` (level, category, title, detail, ids, `acknowledged_at`), `eventlog.Logger.Record` + SSE-topic `events`. **Vrije tekst**: geen stabiel `type`, geen dedup, geen state. Slechts twee aanroepen: `recordDeployEvent` en `build.waiting` (`admission.go:46`, categorie `build`, level `info`) | `eventlog/log.go`, `deploy/runtime.go:368`, `deploy/admission.go` |
| Inbox | Per-user rijen in `inbox_notifications` (category CHECK `deploys\|usage\|system`), prefs in `settings` onder `user_notifications:<id>` (`deployments`, `security_alerts`, `product_updates`). E-mail alleen voor `deploy failed`, **synchroon** via `mail.SendPlain` binnen `insertForUser` | `inbox/dispatch.go`, `migrations/00002` |
| Aanroep | Rechtstreeks in de deploy-runtime: `NotifyDeployFailed` (`runtime.go:354`, in `fail()`), `NotifyDeployLive` (`runtime.go:236`, `static.go:137`) | `deploy/runtime.go`, `deploy/static.go` |
| SMTP | Instance-instelling `smtp` in `settings` (wachtwoord AES-GCM met `crypto.LoadMasterKey(cfg.SecretKeyPath())`), test-endpoint, `email_log` (+API `settings/email/log`). `recordEmailLog` zit in package `server`, niet in `inbox` | `instance/settings.go`, `server/api_instance.go`, `server/api_email_log.go` |
| Mail-versturen | `smtp.SendMail`/eigen TLS-pad. **Geen dial-/totaal-timeout**, headers worden niet gesaneerd (CRLF), subject niet RFC 2047-gecodeerd, geen `Date`/`Message-ID` | `mail/smtp.go` |
| Job-queue | SQLite `jobs`, **één serieel loop** (`RunLoop`, 2 s). `FailJob` zet `failed` zonder retry; `RequeueJob` bestaat maar wordt niet gebruikt; een `running` job na crash wordt niet teruggezet. Deploy-builds mogen 20 min duren | `jobs/worker.go`, `db/queries/jobs.sql` |
| Disk | Host-collector schrijft elke 10 s `metric_host` (`disk_used/total` van `cfg.DataDir`). Enige drempel: build-guard `dockyrd_MIN_FREE_DISK_PERCENT` (10 %). **Geen alert** | `metrics/collector.go`, `build/guard.go` |
| Certificaat | Kolommen `domains.cert_status`/`cert_expires_at` bestaan; **niets vult ze** (alleen het API-antwoord leest ze) | `migrations/00001`, `server/api_domains.go` |
| Backup | `backup.Runner.runDestination` schrijft `database_backup_runs.status='failed'` en geeft de fout terug; `ProcessOnce` logt alleen ("job failed"). Runner heeft geen Events/Inbox-veld | `backup/runner.go:110-156` |
| Container-restarts | `metric_container.restarts` wordt hard op `0` gezet | `metrics/container_collector.go:50` |
| Reconciler | Elke 30 s. `Tick` roept **onvoorwaardelijk** `MarkInterruptedDeployments` aan: een bulk-`UPDATE` die elke deployment met status `building`/`starting`/`healthchecking` op `failed` + `error='interrupted'` zet (geen guard zoals `thawed`, geen lease-check). De deploy-runtime zet precies die statussen tijdens een lopende build (`runtime.go:88/109/211`, `static.go:49`). **Dit lijkt een bug** (de briefing §3.3 bedoelt dit alleen bij herstart); zie §2.1 gat 7. Er gaat geen notificatie uit | `reconcile/reconcile.go:29-37`, `db/queries/deployments.sql:127`, `deploy/runtime.go` |
| Housekeeping | Reconciler ruimt elke 10e tick images op; er is geen prune van `events` of bezorglogs | `reconcile/reconcile.go` |
| API-stijl | **Plain `net/http` `ServeMux` + `writeJSON`/`writeJSONError`, `requireAdmin`/`requireUser`**, tests via `testMuxWithOwnerSession`. Huma staat wel in de briefing en `AGENTS.md`, maar **niet in `go.mod`** | `server/api_*.go`, `go.mod` |
| UI | Settings → Server → Email (`email-settings-page.tsx`) is het patroon (cards, `Modal`, `StatusBadge`, log-modal). Inbox-UI heet `notifications-*` (let op naamconflict) | `web/components/dockyrd/` |
| Wiring | Alle achtergrondtaken starten in `server/runtime.go: startBackground` | `server/runtime.go` |

### 2.1 Gaten die dit oplevert (belangrijk)

1. **Er is geen "event → bezorging"-pad.** We moeten het bouwen, anders is elke bron een eigen mail-aanroep.
2. **Drie bronnen ontbreken** (disk, certificaat, backup). Dat is meer dan de helft van het werk en het echte waardevolle deel.
3. **`build.waiting` en andere info-events mogen nooit extern.** Daarom: alleen events met een expliciet `type` worden bezorgd.
4. **Synchrone SMTP in het deploy-pad** kan een deploy-worker minuten laten hangen. Dit is een bestaand risico en wordt in F0 gerepareerd.
5. **Dubbele e-mail:** een admin krijgt nu al inbox-e-mail bij `deploy failed`. Een e-mailkanaal mag daar niet bovenop komen (zie §6.4).
6. **Docker-root versus data-dir.** `SampleHost(cfg.DataDir)` meet één mountpoint. Wat in de praktijk volloopt is vaak `/var/lib/docker` (images, build-cache, container-logs). Controleer of dat dezelfde mount is; zo niet, meet beide (zie §5.1).
7. **`MarkInterruptedDeployments` draait elke 30 s en niet alleen bij opstart (gelezen, niet gedraaid).** Een build die langer dan één reconcile-tick duurt (builds mogen 20 min) zou tijdelijk `failed (interrupted)` worden en daarna door de runtime weer overschreven worden (`live`). Voor notificaties is dit **een blokkerende voorwaarde**: als we hier `deploy.failed` aan hangen, krijg je valse alerts voor vrijwel elke build. **Fix vóór F1 (Direct-track, F0):** alleen uitvoeren bij opstart vóór de worker-loop start (zoals het `thawed`-patroon), of alleen voor deployments zonder levende job-lease. Daarna pas koppelen aan `Emit`. Check ook of het dashboard nu al flikkerende "failed"-rijen toont.
8. **Een job die tijdens een crash `running` was, wordt nooit teruggezet** (`ClaimNextJob` pakt alleen `pending`, `RequeueJob` heeft geen aanroeper). Los daarvan van notificaties, maar het verklaart waarom we voor bezorging een eigen lease-mechanisme bouwen (§6.6).

---

## 3. Ontwerpbeslissingen

| # | Vraag | Advies | Reden |
|---|---|---|---|
| D1 | Instance-breed of per user? | **Instance-breed**, beheer door admin+ | Kanalen zijn team-infrastructuur ("#ops"). Per-user blijft de inbox + persoonlijke e-mail. Past bij `requireAdmin` en Sevalla/Coolify-gedrag |
| D2 | Hergebruik job-queue of eigen outbox? | **Eigen outbox + eigen dispatcher** | Job-loop is serieel en zonder retry; een build van 20 min zou alerts blokkeren, juist wanneer je ze het hardst nodig hebt |
| D3 | Wat gaat naar buiten? | Alleen **getypeerde** events (`events.type NOT NULL`) | Voorkomt dat interne info-events (`build.waiting`) per ongeluk in Slack belanden |
| D4 | Bezorggarantie | **At-least-once** met `delivery_id` | Exactly-once bestaat niet over HTTP. Webhook-ontvangers krijgen idempotentiesleutel |
| D5 | Conditie-alerts (disk, cert, backup-overdue) | **Stateful** (`firing` → `resolved`) met dedup-sleutel, hysterese en herinnering (default 24 h) | Zonder state spam je elke minuut, zonder `resolved` blijft de ontvanger onzeker |
| D6 | SDK's? | **Nee**, rauwe `net/http` met kleine renderers | Slack/Discord-SDK's trekken tientallen dependencies; payloads zijn simpel JSON |
| D7 | Slack-koppeling | **Incoming webhook-URL** (geen OAuth-app) | Eén URL plakken, geen app-review, past bij "setup < 1 min". Nadeel: geen interactieve knoppen (bewust buiten scope) |
| D8 | Webhook-authenticatie | **HMAC-SHA256 over `timestamp.body`**, optionele eigen headers (bv. `Authorization`) | Timestamp voorkomt replay; consistent met §8.3 "webhooks: HMAC verplicht" |
| D9 | Private/loopback-doelen? | **Standaard geblokkeerd** voor `webhook`, instelbaar door owner | Self-hosters willen soms een LAN-n8n; maar default-deny verkleint SSRF-risico |
| D10 | `deploy.succeeded` extern? | **Default uit** per kanaal, wel in inbox (zoals nu) | Anders is #ops binnen een dag gedempt |
| D11 | E-mailontvangers | Alle admins/owners (default) + vrije extra adressen + specifieke users | Dekt "ops@bureau.nl" en gedeelde postbussen |
| D12 | Eén bericht of één per ontvanger (e-mail) | **Eén delivery per ontvanger** | Retries en fouten per adres, geen gedeelde bounce |
| D13 | Waar leeft de config? | Tabel `notification_channels` + drempels in `settings` onder key `notifications` | Kanalen zijn lijst-objecten met status; drempels zijn één JSON-blob zoals `instance`/`smtp` |

---

## 4. Architectuur

### 4.1 Overzicht

```
 BRONNEN (detectors / hooks)                 KERN  internal/notify                     KANALEN (Sender)
 ───────────────────────────                 ─────────────────────────────             ───────────────────
 deploy.Runtime.fail / live ───┐
 reconcile: interrupted deploy ┤             ┌────────────────────────────┐            ┌─ slack    (net/http)
 backup.Runner (ok/failed)  ───┤  Emit(ev)   │ 1. alert_state (dedup,     │  outbox    ├─ discord  (net/http)
 watch.Disk   (60 s)        ───┼───────────► │    hysterese, reminder)    │ ─────────► ├─ webhook  (net/http+HMAC)
 watch.Certs  (6 u)         ───┤             │ 2. INSERT events           │ deliveries └─ email    (net/smtp)
 watch.BackupOverdue (10 m) ───┘             │ 3. match kanalen + filter  │      ▲
                                             │ 4. INSERT deliveries       │      │ claim + lease
                                             │    (één DB-transactie)     │  ┌───┴──────────┐
                                             └──────┬───────────┬─────────┘  │ Dispatcher   │  retries, backoff,
                                                    │           │            │ (1 goroutine │  Retry-After,
                                          SSE hub ◄─┘           └─► inbox    │  + ≤4 workers│  circuit-breaker
                                          (events, inbox:<user>)   (bestaand)│  per kanaal 1x)│
                                                                             └──────────────┘
```

Eigenschappen:

- **Bronnen** praten alleen met `notify.Emit`. Geen bron kent kanalen.
- `Emit` doet in **één transactie**: alert-state bijwerken, `events`-rij, deliveries per matchend kanaal. Na commit: non-blocking wake van de dispatcher, SSE-publish en de bestaande inbox-fan-out (ongewijzigd intern).
- **Dispatcher** haalt due deliveries op met een lease, rendert per kanaal (`Sender.Render`), verstuurt (`Sender.Send`), en schrijft het resultaat terug. Per kanaal maximaal één in-flight bericht (volgorde + rate-limit).
- Alles draait in het bestaande proces, gestart in `startBackground` (`server/runtime.go`). Geen extra container, geen extra poort.

### 4.2 Waarom een outbox (en niet "gewoon `go send()`")

| Probleem | Fire-and-forget | Outbox |
|---|---|---|
| Panel herstart tijdens bezorging | bericht weg | delivery staat nog `pending`/`sending`, lease verloopt, wordt opnieuw opgepakt |
| Slack geeft 429/500 | verloren of ad-hoc retry | gecontroleerde backoff, zichtbaar in UI |
| Kanaal tijdelijk onbereikbaar (netwerk, DNS) | verloren | retries tot TTL (24 u), dan `dead` + zichtbaar |
| Wat is er verstuurd? | onbekend | bezorglog per kanaal |
| Deploy-pad vertraagd door netwerk | ja | nee (alleen DB-insert) |

### 4.3 Package-indeling

```
internal/
├── notify/
│   ├── event.go          # Event, Level, State, type-catalogus (consts + defaults)
│   ├── notifier.go       # Notifier: Emit(), channel-match, alert_state, tx
│   ├── alertstate.go     # state machine: firing/resolved, hysterese, reminder (klok injecteerbaar)
│   ├── dispatcher.go     # claim, lease-herstel, worker-pool, backoff, circuit-breaker
│   ├── sender.go         # interface Sender + registry
│   ├── slack.go          # renderer + sender
│   ├── discord.go
│   ├── webhook.go        # payload v1 + HMAC
│   ├── email.go          # gebruikt mail.* (F0-versie)
│   ├── render.go         # gedeelde helpers: truncate (rune-safe), escape, links, kleuren
│   ├── netsafe.go        # http.Client-factory: dial-guard, geen redirects, redactie van URL in errors
│   └── *_test.go
├── watch/                # detectors (draaien in startBackground)
│   ├── disk.go
│   ├── certs.go
│   └── backups.go        # overdue-check
├── db/migrations/00006_notifications.sql
├── db/queries/notifications.sql
└── server/api_notifications.go   # registerNotificationRoutes (mux), admin only
web/
├── components/dockyrd/notification-channels-settings-page.tsx   # NIET "notifications-*" (inbox)
├── components/dockyrd/notification-channel-modal.tsx
├── components/dockyrd/notification-delivery-log-modal.tsx
└── app/settings/server/channels/page.tsx
docs/user/notifications.md
scripts/verify-notify-smoke.sh
```

### 4.4 Event-model

```go
package notify

type Level string // zelfde waarden als events.level en inbox: "error","warning","success","info"
type State string // "" (discreet), "firing", "resolved"

type Event struct {
    Type         string         // stabiele sleutel, bv. "backup.failed" (verplicht)
    Level        Level
    State        State          // alleen voor stateful types
    DedupKey     string         // verplicht bij stateful, bv. "disk:/var/lib/dockyrd", "backup:<serviceID>:remote"
    Title        string         // kort, zonder markup
    Detail       string         // ≤ 500 tekens, gescrubd (zie §8.5)
    ProjectID    string
    ServiceID    string
    DeploymentID string
    Href         string         // relatief pad ("/applications/web"); renderer plakt PublicBaseURL ervoor
    Data         map[string]any // type-specifiek, voor webhook-payload en renderers
}

func (n *Notifier) Emit(ctx context.Context, ev Event) error // nil-safe receiver; nooit netwerk-IO
```

Regels voor `Emit`:

- Receiver mag `nil` zijn (tests, installaties zonder DB-ready), net zoals `eventlog.Logger.Record`.
- Callers negeren de fout (`_ =`, bestaand patroon); `Emit` logt zelf via `slog`.
- Duurt < 10 ms: één `BEGIN IMMEDIATE`-transactie op de bestaande schrijfconnectie.

### 4.5 Event-catalogus (v1)

| Type | Level | Stateful | Bron | Default naar extern | Dedup-sleutel |
|---|---|---|---|---|---|
| `deploy.failed` | error | nee | `deploy.Runtime.fail`, reconciler (interrupted) | **aan** | – |
| `deploy.succeeded` | success | nee | `deploy.Runtime` / `static.go` | uit | – |
| `backup.failed` | error | **ja** | `backup.Runner.runDestination` | **aan** | `backup:<service>:<dest>` |
| `backup.overdue` | warning | **ja** | `watch.BackupOverdue` | **aan** | `backup-overdue:<service>` |
| `backup.succeeded` | success | resolvet `backup.failed` | `backup.Runner` | uit (resolved wél, zie §5.2) | `backup:<service>:<dest>` |
| `disk.high` | warning/error | **ja** | `watch.Disk` | **aan** | `disk:<mount>` |
| `cert.expiring` | warning/error | **ja** | `watch.Certs` | **aan** | `cert:<hostname>` |
| `notification.channel_failed` | warning | **ja** | dispatcher | alleen inbox + andere kanalen | `channel:<id>` |
| `test.ping` | info | nee | test-endpoint | n.v.t. (rechtstreeks) | – |

Tier 2 (F4): `service.crash_loop`, `service.oom`, `update.available`, `update.failed`.

---

## 5. Bronnen (detectors)

### 5.1 Disk (`watch/disk.go`)

- **Interval:** elke 60 s. Lees de **laatste rij van `metric_host`** (wordt al elke 10 s gevuld) in plaats van opnieuw `SampleHost` aan te roepen; geen dubbele syscalls.
- **Drempels** (instelbaar in `settings.notifications`): `warn` ≥ 85 % (briefing §11.4: "disk > 85 %"), `critical` ≥ 95 %, **resolve** < 80 % (hysterese van 5 punten tegen flapperen).
- **Niveau-escalatie:** `warning` → `error` bij 95 % is een *update* op dezelfde dedup-sleutel en wordt direct gemeld (niveau steeg), ook binnen de herinneringsperiode.
- **Mountpoints:** `cfg.DataDir` altijd. Vraag ook `DockerRootDir` op (`docker info`) en meet die als het een andere mount is. Als meerdere volumes: aparte dedup-sleutels (`disk:/`, `disk:/var/lib/docker`).
- **Inhoud:** percentage, vrij/totaal (GiB), top-verbruikers die we zonder shell kennen: Docker build-cache, dangling images, oude releases (de cleanup-API `GET /server/cleanup` levert dit al), plus link naar Analytics/Cleanup.
- **Bijzonder:** bij ≥ 95 % kan SQLite-schrijven falen (zie §9.1). Daarom altijd ook een `slog.Error`.

### 5.2 Backups

**Hook (event-gedreven)** in `backup.Runner.runDestination` na het schrijven van `UpdateDatabaseBackupRun`:

```go
// status == "failed"  → Emit(backup.failed,  firing,   key backup:<svc>:<dest>, Data{service, destination, error, run_id})
// status == "succeeded" → Emit(backup.succeeded, resolved, zelfde key)   // alleen als er een open firing is
```

- `Runner` krijgt een `Notify *notify.Notifier`-veld (nil-safe), gezet in `jobs.Worker` (`runner := &backup.Runner{... Notify: w.Notify}`); `Worker` krijgt het veld erbij.
- Let op: `ProcessOnce` markeert de job als `failed` zonder retry. Eén gemiste backup is dus definitief tot de volgende geplande run, wat de **eerste** fout alert-waardig maakt.
- `Data.error` is de foutmelding van de dump/upload: **scrub** (§8.5), want S3-/dump-fouten kunnen endpoints en soms keys bevatten.

**Overdue-detector** (`watch/backups.go`, elke 10 min): voor elke `database_backup_policies` met `enabled=1`: `now - laatste succesvolle run > 2 × interval + 30 min` (of nooit gelukt en policy ouder dan 2 × interval) → `backup.overdue` firing. Dit vangt wat een hook niet ziet: scheduler die niet draait, job die blijft hangen in `running`, remote target die stil faalt. **Dit is de alert met de meeste waarde per regel code**; "backup mislukt" zie je pas als iemand het merkt, "backup is er al 3 dagen niet" ook als er niets faalt.

### 5.3 Certificaten (`watch/certs.go`)

Geen externe calls naar Let's Encrypt; we kijken naar wat de proxy **daadwerkelijk serveert**:

- Elke 6 uur en direct na het aanmaken van een domein (met 30 min grace): voor elk domein (+ het panel-domein uit `instance.Panel.PublicURL`) een **TLS-dial naar `127.0.0.1:443`** met `ServerName = hostname`, handshake-timeout 5 s.
- Lees leaf `NotBefore/NotAfter`, verifieer hostname + keten tegen de systeemroots (handmatig, met `InsecureSkipVerify` alleen om het certificaat te *lezen* en daarna zelf te verifiëren; zo onderscheiden we "verlopen" van "niet uitgegeven / Caddy-fallback").
- **Schrijf terug** naar `domains.cert_status` (`valid|expiring|expired|missing`) en `cert_expires_at`; de bestaande domeinen-UI krijgt dit gratis.
- **Drempels relatief aan de levensduur** (niet hard 30/14 dagen): Let's Encrypt kondigde kortere certificaatlevensduren aan, en een vaste dagdrempel klopt dan niet meer. Regel: `warning` bij resterend ≤ 15 % van `NotAfter-NotBefore` (maximaal 14 dagen), `error` bij ≤ 5 % (maximaal 3 dagen) of verlopen. Resolve zodra resterend > 33 % (= vers vernieuwd).
- **Waarom pas laat alarmeren:** Caddy vernieuwt automatisch ruim voor de deadline en herhaalt bij falen met toenemende backoff (max. een dag tussen pogingen, tot 30 dagen, eerst alle challenge-types, dan de volgende CA). Een certificaat dat desondanks de 15 %-grens haalt betekent dus dat **vernieuwen echt faalt**: precies de actiegerichte melding. (De exacte vernieuwvenster-ratio van CertMagic staat niet op de Caddy-docspagina; verifieer die in de bron vóór je de 33 %-resolve-regel vastlegt.)
- **Typische oorzaak in de melding:** `dns_status != ok` (domein wijst niet meer naar de server), poort 80 geblokkeerd, rate-limit. Voeg `dns_status` uit `domains` toe aan `Data` zodat het bericht al een hint bevat.
- Sla platformdomeinen (`is_platform=1`) over wanneer er geen publieke URL is; skip domeinen van gestopte services.

### 5.4 Deploy

- Vervang de twee directe aanroepen (`NotifyDeployFailed` in `runtime.go:354`, `NotifyDeployLive` in `runtime.go:236` en `static.go:137`) door `Notify.Emit(...)`. De `Notifier` roept **binnen** de bestaande inbox-`Dispatcher` aan, zodat de inbox-logica niet verandert (zie §6.4 voor e-mail).
- `Data`: project/service-slug, `commit_sha` (7 tekens), `commit_message` (eerste regel, **attacker-controlled**, zie §8.4), `branch`, `trigger`, `failure_kind` (kolom uit migratie 00005: oom/timeout/disk), `duration`.
- **Interrupted deploys (pas ná de F0-fix uit §2.1 gat 7):** `MarkInterruptedDeployments` is een bulk-`UPDATE` zonder id's terug. Beperk hem eerst tot opstart (of lease-gebaseerd), pas daarna de query aan (`SELECT` eerst, of `RETURNING id, service_id` na compatibiliteitscheck met sqlc) zodat per deploy `deploy.failed` ("interrupted by restart") geëmit kan worden. Zonder deze volgorde veroorzaakt de koppeling valse alerts.
- `recordDeployEvent` blijft de `events`-rij schrijven; laat `Notify.Emit` die rij schrijven in plaats van beide (één bron van waarheid: migreer de twee aanroepen en laat `eventlog.Record` alleen voor ongetypeerde interne events bestaan).

---

## 6. Kern: alert-state, outbox en dispatcher

### 6.1 Datamodel (goose `00006_notifications.sql`)

```sql
-- +goose Up
CREATE TABLE notification_channels (
  id                   TEXT PRIMARY KEY,                       -- ULID
  kind                 TEXT NOT NULL CHECK (kind IN ('slack','discord','email','webhook')),
  name                 TEXT NOT NULL,
  enabled              INTEGER NOT NULL DEFAULT 1,
  config               TEXT NOT NULL DEFAULT '{}',             -- JSON zonder secrets (zie 6.2)
  secret_enc           BLOB,                                   -- AES-GCM(JSON): url, signing_secret, headers
  min_level            TEXT NOT NULL DEFAULT 'warning' CHECK (min_level IN ('info','warning','error')),
  event_types          TEXT NOT NULL DEFAULT '[]',             -- JSON-array; '[]' = defaults uit catalogus
  project_id           TEXT REFERENCES projects(id) ON DELETE CASCADE,  -- NULL = alles
  status               TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','degraded','disabled')),
  disabled_reason      TEXT,
  consecutive_failures INTEGER NOT NULL DEFAULT 0,
  last_success_at      INTEGER,
  last_failure_at      INTEGER,
  last_error           TEXT,
  created_by           TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at           INTEGER NOT NULL,
  updated_at           INTEGER NOT NULL
);

-- events krijgt type/state/dedup/data; bestaande rijen blijven NULL = nooit extern
ALTER TABLE events ADD COLUMN type      TEXT;
ALTER TABLE events ADD COLUMN state     TEXT;
ALTER TABLE events ADD COLUMN dedup_key TEXT;
ALTER TABLE events ADD COLUMN data      TEXT;                  -- JSON
CREATE INDEX idx_events_type_created ON events (type, created_at DESC);

CREATE TABLE alert_state (
  dedup_key        TEXT PRIMARY KEY,
  type             TEXT NOT NULL,
  state            TEXT NOT NULL CHECK (state IN ('firing','resolved')),
  level            TEXT NOT NULL,
  first_fired_at   INTEGER NOT NULL,
  last_notified_at INTEGER NOT NULL,
  resolved_at      INTEGER,
  occurrences      INTEGER NOT NULL DEFAULT 1,
  event_id         TEXT                                       -- laatste event-rij
);

CREATE TABLE notification_deliveries (
  id              TEXT PRIMARY KEY,                           -- = X-dockyrd-Delivery (idempotentiesleutel)
  event_id        TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  channel_id      TEXT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
  target          TEXT NOT NULL DEFAULT '',                   -- e-mail: ontvanger; anders ''
  dedup_key       TEXT,                                       -- voor coalescing (6.5)
  status          TEXT NOT NULL CHECK (status IN ('pending','sending','sent','dead','skipped')),
  attempts        INTEGER NOT NULL DEFAULT 0,
  next_attempt_at INTEGER NOT NULL,
  lease_until     INTEGER,
  http_status     INTEGER,
  last_error      TEXT,
  created_at      INTEGER NOT NULL,
  sent_at         INTEGER,
  UNIQUE (event_id, channel_id, target)
);
CREATE INDEX idx_deliveries_due     ON notification_deliveries (status, next_attempt_at);
CREATE INDEX idx_deliveries_channel ON notification_deliveries (channel_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_deliveries_channel;
DROP INDEX IF EXISTS idx_deliveries_due;
DROP TABLE IF EXISTS notification_deliveries;
DROP TABLE IF EXISTS alert_state;
DROP INDEX IF EXISTS idx_events_type_created;
ALTER TABLE events DROP COLUMN data;
ALTER TABLE events DROP COLUMN dedup_key;
ALTER TABLE events DROP COLUMN state;
ALTER TABLE events DROP COLUMN type;
DROP TABLE IF EXISTS notification_channels;
```

Notities:

- Briefing §14: migraties zijn **forward-only en backwards-compatible voor één minor-versie**. Alle wijzigingen hier zijn additief (nieuwe tabellen, nullable kolommen), dus rollback naar de vorige binary werkt.
- **`inbox_notifications.category` heeft een CHECK** (`deploys|usage|system`) die in SQLite niet zonder tabel-rebuild te wijzigen is. Mapping zonder migratie: `deploy.*` → `deploys`, `disk.high` → `usage`, `backup.*`, `cert.*`, `notification.*` → `system`.
- `events.level`/`category` zijn vrije tekst; categorieën voor nieuwe types: `deploys`, `backups`, `system`, `certificates` (controleer dat `eventNeedsAttention` en de dashboard-filter `deploys|issues|system` ze goed behandelen).
- Bezorglog-retentie: prune deliveries ouder dan 30 dagen en getypeerde events ouder dan 90 dagen in de bestaande prune-tick van de reconciler (`tickCount%10`), **batchgewijs** (`LIMIT 500` per tick) zodat de schrijfconnectie niet blokkeert.
- Tijd: unix-seconden (zoals `events.created_at`). Houd sqlc-typen (`sql.NullInt64`) consistent met de rest.

### 6.2 `config` en `secret_enc` per kanaaltype

| Kind | `config` (plain JSON) | `secret_enc` (AES-GCM JSON) |
|---|---|---|
| `slack` | `{ "channel_label": "#ops" }` (alleen cosmetisch) | `{ "url": "https://hooks.slack.com/services/…" }` |
| `discord` | `{ "username": "Dockyrd" }` | `{ "url": "https://discord.com/api/webhooks/<id>/<token>" }` |
| `webhook` | `{ "method": "POST", "include_data": true }` | `{ "url": "…", "signing_secret": "…", "headers": {"Authorization": "Bearer …"} }` |
| `email` | `{ "recipients": {"admins": true, "user_ids": [...], "extra": ["ops@x.nl"]} }` | niet nodig (SMTP is instance-breed) |

Slack-/Discord-URLs bevatten het token en zijn dus **secrets** en versleuteld (net als het SMTP-wachtwoord in `instance.SaveSMTP`). Gebruik `crypto.Encrypt(key, …)` met `crypto.LoadMasterKey(cfg.SecretKeyPath())`. In API-antwoorden alleen een gemaskeerde vorm: `https://hooks.slack.com/services/T0…/B0…/••••` en `secret_set: true`.

### 6.3 `Emit`-algoritme (één transactie)

```
Emit(ev):
  if ev.Type == "" → return error (ongetypeerd nooit extern)
  BEGIN IMMEDIATE
    if stateful(ev.Type):
      row = alert_state[ev.DedupKey]
      switch (row?.state, ev.State):
        (nil|resolved, firing)   → upsert firing; notify=true; occurrences=1
        (firing, firing)         → occurrences++; notify = (ev.Level > row.level)            // escalatie
                                                    || (now - row.last_notified_at ≥ reminder) // herinnering
        (firing, resolved)       → upsert resolved; notify=true; (geen firing bekend → drop stilletjes)
        (resolved, resolved)     → notify=false
      if !notify → COMMIT; return nil                                                       // dedup: niets naar buiten, niets in events
    INSERT events (id, level, category, title, detail, ..., type, state, dedup_key, data)
    for ch in matchingChannels(ev):                  // enabled ∧ status≠disabled ∧ type ∈ filter ∧ level ≥ min_level ∧ project-scope
      for target in ch.targets(ev):                  // e-mail: ontvangers; anders {""}
        coalesce(ch, ev.DedupKey)                    // zie 6.5
        INSERT notification_deliveries (..., status='pending', next_attempt_at=now)
  COMMIT
  wake dispatcher (non-blocking send op chan struct{} met buffer 1)
  hub.Publish("events", …); inbox fan-out (bestaande Dispatcher)
```

`matchingChannels`: lees kanalen in één query (≤ tientallen rijen); cache niet (SQLite lokaal is snel, en een cache is een extra foutbron bij wijzigingen).

### 6.4 E-mail en de bestaande inbox-mail: geen dubbele berichten

Nu mailt `inbox.Dispatcher` bij `deploy.failed` elke actieve user die `prefs.Deployments` aan heeft. Voor het e-mailkanaal geldt:

- `UNIQUE (event_id, channel_id, target)` voorkomt dubbele deliveries binnen het kanaal.
- Tussen inbox-mail en e-mailkanaal: **één regel**: het e-mailkanaal slaat ontvangers over die dit event al via de inbox per mail kregen (zelfde lowercase adres). Implementatie: `Emit` verzamelt eerst de inbox-ontvangers (bestaande prefs-logica), dan het kanaal-doel minus die set.
- **Voorkeursroute (F2):** verplaats de inbox-mail zelf naar de outbox (delivery op een impliciet systeemkanaal `inbox_email`). Dan verdwijnt de synchrone SMTP-aanroep uit het deploy-pad (risico #4 in §2.1) en is er één retry-mechanisme. Tot die tijd: F0 zet timeouts in `mail`, zodat het worst-case zeker begrensd is.

### 6.5 Coalescing (supersede) en volgorde

Een `resolved` mag nooit vóór een nog niet bezorgde `firing` aankomen (ontvanger denkt dan dat het probleem open staat).

- Bij het inserten van een delivery voor `(channel, dedup_key)`: zet bestaande `pending` deliveries met dezelfde sleutel op `skipped` ("superseded by <event>"). Een niet-bezorgde `firing` gevolgd door `resolved` levert dan één bericht op (of geen, als je ook de `resolved` skipt wanneer de `firing` nooit is verstuurd; keuze: **verstuur alleen de laatste staat**).
- Dispatcher claimt per kanaal op `created_at`, één in-flight tegelijk.

### 6.6 Dispatcher

```sql
-- name: ClaimDueDeliveries :many
UPDATE notification_deliveries
SET status = 'sending', lease_until = @now + 60, attempts = attempts + 1
WHERE id IN (
  SELECT d.id FROM notification_deliveries d
  JOIN notification_channels c ON c.id = d.channel_id
  WHERE d.status = 'pending' AND d.next_attempt_at <= @now AND c.status <> 'disabled'
  ORDER BY d.next_attempt_at, d.created_at
  LIMIT @limit)
RETURNING id, event_id, channel_id, target, attempts;
```

(`UPDATE … RETURNING` werkt in SQLite ≥ 3.35; controleer de SQLite-versie in `modernc.org/sqlite v1.36.0`, dit is ruim recent. Als sqlc het niet leuk vindt: twee statements in één transactie.)

Lus:

1. **Opstart:** `UPDATE … SET status='pending' WHERE status='sending' AND lease_until < now` (herstel na crash). Herhaal elke tick (vangt ook een vastgelopen worker).
2. **Tick:** elke 2 s of bij wake. Claim ≤ 8 deliveries; start per kanaal maximaal één goroutine (in-process `map[channelID]*sync.Mutex`/semaphore); globaal ≤ 4 gelijktijdig.
3. **Per delivery:** laad event + kanaal, decrypt secret, `Sender.Render`, `Sender.Send(ctx met timeout 10 s)`.
4. **Resultaat → status:**

| Resultaat | Actie |
|---|---|
| 2xx / SMTP accepted | `sent`, `consecutive_failures=0`, `last_success_at`; herstel `degraded` → `active` |
| 429 | `pending`, `next_attempt_at = now + Retry-After` (Discord: `retry_after` uit JSON-body, seconden, mogelijk fractioneel; cap 1 u). Telt **niet** als kanaalfout |
| 408, 5xx, netwerkfout, timeout, SMTP 4xx | `pending` met backoff (hieronder), `consecutive_failures++` |
| 400 (payload ongeldig) | `dead` direct, `last_error` bewaren; dit is een bug bij ons, log `slog.Error` |
| 401/403 | `dead` + kanaal naar `disabled` ("authenticatie geweigerd") |
| 404/410 (Slack `no_service`/`no_active_hooks`; Discord "Unknown Webhook") | `dead` + kanaal `disabled` ("webhook bestaat niet meer") |
| 3xx | `dead` ("redirect niet gevolgd"), geen redirect-follow (SSRF) |
| SMTP 5xx (ongeldig adres) | `dead` voor dat target; kanaal blijft actief |

5. **Backoff:** poging 1→2: 30 s, 2→3: 2 min, 3→4: 10 min, 4→5: 1 u, 5→6: 6 u, daarna `dead`; ±20 % jitter. **TTL 24 u** vanaf `created_at`: daarna `dead` ("verlopen"); een alert van gisteren is ruis.
6. **Circuit-breaker:** `consecutive_failures ≥ 5` → kanaal `degraded` (UI toont waarschuwing, bezorging gaat door); ≥ 24 u onafgebroken falen of een permanente fout (401/403/404/410) → `disabled` + `disabled_reason`. In beide gevallen **één** `notification.channel_failed`-event (stateful, sleutel `channel:<id>`), dat **nooit** naar het falende kanaal zelf gaat (inbox + andere kanalen). Dit voorkomt een lus en zorgt dat je het wél hoort.
7. **Pacing:** token bucket per kanaal: Slack/Discord/webhook **1 bericht/s, burst 3**. Slack documenteert 1 request/s voor incoming webhooks met korte bursts toegestaan en antwoordt anders met 429 + `Retry-After`.
8. **Graceful shutdown:** `ctx` geannuleerd → in-flight sends krijgen max. 5 s; rest blijft `sending` en wordt door lease-herstel opgepakt.

### 6.7 Waarom één schrijver en geen aparte database

De dispatcher gebruikt dezelfde `*sqlc.Queries` als de rest. Alle writes zijn korte statements; HTTP-IO gebeurt **buiten** transacties. Dit past bij WAL + één schrijver + `busy_timeout=5000` (briefing §4.1). Een belasting van tientallen deliveries per minuut is verwaarloosbaar.

---

## 7. Kanalen

### 7.1 Gemeenschappelijk contract

```go
type Rendered struct { /* kanaalspecifiek: body []byte, contentType, headers, of mail{Subject, Text, HTML} */ }

type Sender interface {
    Kind() string
    Validate(cfg Config, secret Secret) error            // bij opslaan: formaat, host-allowlist, SSRF-pre-check
    Render(ev EventView, cfg Config, base string) (Rendered, error)
    Send(ctx context.Context, cfg Config, secret Secret, r Rendered, meta DeliveryMeta) Result
}
type Result struct { OK bool; HTTPStatus int; RetryAfter time.Duration; Permanent bool; DisableChannel bool; Err string }
```

Registry: `map[string]Sender`, vast ingebouwd. Geen plugin-systeem.

**Gedeelde HTTP-client** (`netsafe.go`): één `http.Client` per proces met `Timeout: 10s`, `CheckRedirect: return http.ErrUseLastResponse`, een `Transport` met `DialContext` die **na DNS-resolutie** het IP checkt (zie §8.1), `MaxIdleConns` klein, `ForceAttemptHTTP2: true`, en een expliciete `User-Agent: dockyrd/<version> (+https://dockyrd.io)`. **Fouten worden omgezet naar een redactie-veilige string** (Go's `*url.Error` bevat de volledige URL, dus het Slack-/Discord-token, in `err.Error()`; nooit ruw loggen of opslaan, §8.2).

### 7.2 Slack (incoming webhook)

- **URL-validatie:** `https`, host exact `hooks.slack.com` (of `hooks.slack-gov.com`), pad begint met `/services/`. Anders weigeren bij opslaan (voorkomt SSRF via "Slack"-kanaal en typo's).
- **Payload:** `attachments[0]` met `color` (kleurbalk) + `blocks`, plus top-level `text` als fallback voor notificaties/screenreaders:

```json
{
  "text": "Deploy failed · webshop / web",
  "attachments": [{
    "color": "#B3261E",
    "blocks": [
      { "type": "header",  "text": { "type": "plain_text", "text": "Deploy failed" } },
      { "type": "section", "text": { "type": "mrkdwn", "text": "*webshop / web*\n`a1b2c3d` Fix checkout rounding\nBuild failed: out of memory (exit 137)" } },
      { "type": "context", "elements": [{ "type": "mrkdwn", "text": "Dockyrd · <instance> · 5 Nov 20:14 UTC" }] },
      { "type": "actions", "elements": [{ "type": "button", "text": { "type": "plain_text", "text": "View build log" }, "url": "https://panel.example.com/applications/web" }] }
    ]
  }]
}
```

- **Link-knop** (alleen URL, geen interactie → werkt met een gewone webhook). Weglaten als er geen publieke URL is ingesteld (nooit `localhost`-links sturen).
- **Escaping:** in mrkdwn `&`→`&amp;`, `<`→`&lt;`, `>`→`&gt;` voor **alle** vrije tekst (commit-berichten, foutmeldingen, service-namen). Voorkomt mention-/link-injectie (`<!channel>`, `<https://evil|klik>`).
- **Limieten:** houd `section`-tekst ruim onder 3000 tekens (afkappen op ~2500, rune-safe, met "…") en `header` onder 150. De Slack-docs-pagina over rate limits noemt geen tekstlimiet voor webhook-berichten; deze waarden komen uit de Block Kit-conventies, **controleer ze tegen de actuele Block Kit-referentie** voordat je ze als harde grens vastlegt.
- **Responses:** `200 ok`. Fouten zijn korte tekstbodies (`invalid_payload`, `no_service`, `channel_not_found`, …); sla de eerste 200 tekens op in `last_error`.

### 7.3 Discord (webhook)

- **URL-validatie:** `https`, host ∈ {`discord.com`, `discordapp.com`, `ptb.discord.com`, `canary.discord.com`}, pad `/api/webhooks/<id>/<token>`.
- **Payload:** één `embed` met kleur (decimaal uit de status-tokens), `timestamp` (RFC 3339), `fields`, `footer`; **altijd** `allowed_mentions: {"parse": []}` zodat vrije tekst nooit `@everyone`/rollen pingt.

```json
{
  "username": "Dockyrd",
  "allowed_mentions": { "parse": [] },
  "embeds": [{
    "title": "Backup failed · uploads-db",
    "description": "Remote backup to S3 failed: connection refused",
    "color": 11740702,
    "timestamp": "2026-11-05T19:14:03Z",
    "url": "https://panel.example.com/databases/uploads-db",
    "fields": [
      { "name": "Project", "value": "client-x", "inline": true },
      { "name": "Destination", "value": "remote", "inline": true }
    ],
    "footer": { "text": "Dockyrd · <instance>" }
  }]
}
```

- **Limieten** (geverifieerd, zie bijlage D): title 256, description 4096, ≤ 25 fields, field name 256, field value 1024, footer 2048, author 256, **totaal ≤ 6000 tekens over alle embeds**, ≤ 10 embeds. Eén embed per bericht, afkappen met marge (bv. description ≤ 1800) en rune-safe.
- **Escaping:** Markdown-tekens in vrije tekst escapen (`` ` * _ ~ | > ``); URL-tekst niet laten renderen als maskeerde link.
- **Succes:** 204 No Content (zonder `?wait=true`). **Rate-limit:** headers `X-RateLimit-*`; bij 429 JSON met `retry_after` (seconden, float): respecteren (cap 1 u). 404 = webhook verwijderd → kanaal `disabled`.
- Zet een expliciete `User-Agent` (zie 7.1) zodat de verzoeken herkenbaar zijn en niet als anonieme default-client worden geweerd.

### 7.4 Generieke webhook

- **Request:** `POST` JSON, `Content-Type: application/json`. Alleen `https` (of `http` wanneer owner "private targets" heeft toegestaan, §8.1). Antwoord `2xx` = succes; **geen redirects**.
- **Headers:**

| Header | Inhoud |
|---|---|
| `User-Agent` | `dockyrd/<versie> (+https://dockyrd.io)` |
| `X-dockyrd-Event` | bv. `backup.failed` |
| `X-dockyrd-Delivery` | delivery-ID (ULID) = idempotentiesleutel |
| `X-dockyrd-Timestamp` | unix-seconden |
| `X-dockyrd-Signature` | `sha256=<hex>` = HMAC-SHA256(`signing_secret`, `timestamp + "." + body`) |
| eigen headers | uit `secret_enc.headers` (bv. `Authorization`); nooit gelogd |

- **Payload `dockyrd.notification.v1`:**

```json
{
  "schema": "dockyrd.notification.v1",
  "id": "01JC9K2M7Q…",
  "type": "disk.high",
  "level": "warning",
  "state": "firing",
  "occurred_at": "2026-11-05T19:14:03Z",
  "title": "Disk usage is high",
  "message": "/var/lib/docker is 87 % full (13.1 GiB free of 100 GiB).",
  "url": "https://panel.example.com/server",
  "instance": { "name": "prod-01", "url": "https://panel.example.com", "version": "0.0.46" },
  "project": null,
  "service": null,
  "data": { "mount": "/var/lib/docker", "used_percent": 87.2, "free_bytes": 14065000000, "threshold": 85 }
}
```

- **Versiebeleid:** `schema` is een contract. Velden toevoegen mag; verwijderen/hernoemen = `v2`. Beschrijf dit in `docs/user/notifications.md`.
- **Verificatie-voorbeeld voor ontvangers** (in de docs opnemen, ook Node/Python):

```go
mac := hmac.New(sha256.New, []byte(secret))
mac.Write([]byte(ts + "." + string(body)))
ok := hmac.Equal(mac.Sum(nil), sigBytes) && abs(now-ts) < 300
```

- **Test:** de test-knop stuurt `type: "test.ping"` met dezelfde handtekening.

### 7.5 E-mail

Hergebruikt de instance-SMTP (`instance.LoadSMTPConfig`); geen eigen SMTP per kanaal in v1. Het kanaal is **uitgeschakeld/waarschuwend** zolang SMTP niet geconfigureerd is (UI: "Configure SMTP first" met link naar `/settings/server/email`).

- **Ontvangers:** `admins` (alle actieve owner/admin-users, `disabled_at IS NULL`), `user_ids`, `extra` (vrije adressen, gevalideerd met `net/mail.ParseAddress`, max. 20). Eén delivery per ontvanger.
- **Inhoud:** `Subject: [dockyrd] <niveau-label> <titel>`; bijv. `[dockyrd] Backup failed · uploads-db`. Body `text/plain` (v1; HTML optioneel later) met: wat/waar, detail, deep-link, "Sent by dockyrd <instance>". Voor `firing`/`resolved` in de subject: `[RESOLVED]`-prefix.
- **Headers:** `Date`, `Message-ID` (`<delivery-id@<host>>`), `Auto-Submitted: auto-generated`, `X-Auto-Response-Suppress: All`, `MIME-Version`, `Content-Type: text/plain; charset=UTF-8`, `Content-Transfer-Encoding: 8bit` (of quoted-printable). Subject met `mime.QEncoding.Encode("utf-8", s)`.
- **Hardening in `mail` (F0):** zie §8.3. De e-mailsender gebruikt dezelfde `mail.Send…` met timeouts en logt elk verzonden/mislukt bericht in **`email_log`** met `kind = "notification"` (nu staat de log-aanroep in `server`; verplaats `recordEmailLog` naar een pakket dat zowel `server` als `notify` kan gebruiken, of geef `notify` een kleine `MailLog`-interface).
- **Deliverability-beperking (eerlijk melden in docs):** alerts van een vers VPS-IP via eigen SMTP belanden makkelijk in spam. Advies: gebruik een transactionele SMTP-provider met eigen domein (SPF/DKIM); `From` hoort bij een domein dat die provider mag gebruiken.

---

## 8. Beveiliging en robuustheid

### 8.1 SSRF-guard (outbound HTTP vanuit het control plane)

De control plane draait als root op de host en kan Docker-netwerken, de metadata-endpoint van de cloud (169.254.169.254) en localhost-services bereiken. Een door een admin ingevulde URL is dus een SSRF-oppervlak (ook al is de admin vertrouwd: gestolen sessie, te ruime rol).

- **Dial-time check:** `Transport.DialContext` resolvet zelf, loopt alle IP's langs en weigert verbinden met: loopback, link-local (incl. `169.254.0.0/16`, `fe80::/10`), RFC 1918, `100.64.0.0/10` (CGNAT), ULA `fc00::/7`, `0.0.0.0/8`, multicast, en IPv4-mapped IPv6 daarvan. Daarna wordt **het gecheckte IP gedialed** (geen tweede DNS-lookup → geen DNS-rebinding).
- **Slack/Discord:** vaste host-allowlist (§7.2/§7.3); de IP-check geldt ook hier.
- **Webhook:** standaard dezelfde blokkade. Instance-instelling `allow_private_targets` (alleen **owner** mag die aanzetten, audit-log) laat RFC 1918/ULA toe; **169.254.0.0/16 blijft altijd geblokkeerd.**
- **Geen redirects.** Geen `http://` tenzij private targets zijn toegestaan.
- **Validatie bij opslaan én bij verzenden** (URL kan later naar een ander IP wijzen).

### 8.2 Secrets en logging

- URLs/headers/signing-secrets **alleen versleuteld** in `secret_enc`; masker in alle API-responses; PATCH zonder `secret` laat de oude staan (zelfde patroon als `SaveSMTP`: wachtwoord alleen vervangen als het niet leeg is).
- `*url.Error` en `http`-fouten bevatten de volledige URL. Alle fouten gaan door `redact(err, secretURL)` voor ze naar `slog`, `last_error` of API gaan. Voeg een test toe die een Slack-token-URL gebruikt en controleert dat het token **nergens** in logs/DB/API voorkomt.
- Bezorglog bewaart `http_status` en een korte, geredigeerde fouttekst, nooit request-bodies met headers.

### 8.3 SMTP-hardening (F0, `internal/mail`)

| Probleem | Fix |
|---|---|
| Geen timeouts (`smtp.SendMail`, `tls.Dial`) | `net.Dialer{Timeout: 10s}`, `conn.SetDeadline(now+30s)`, ctx-annulering; STARTTLS afdwingen wanneer `UseTLS` (nu alleen impliciete TLS op 465 expliciet afgehandeld; controleer dat 587 niet stil naar plaintext zakt) |
| CRLF-injectie in `Subject`/`To` (service-namen, foutmeldingen zijn user-/attacker-beïnvloed) | strip `\r`/`\n` uit alle header-waarden; valideer `To` met `mail.ParseAddress` |
| Subject niet gecodeerd | `mime.QEncoding` |
| Geen `Date`/`Message-ID` | toevoegen (spamfilters straffen dit) |
| Alleen `PlainAuth` | oké voor nu; bij "no auth"-relays (poort 25 intern) `auth=nil` toestaan als `Username==""` |

Dit is een **Direct-track** fix en verbetert direct password-reset/invite-mail.

### 8.4 Inhoud is deels aanvaller-gecontroleerd

Commit-berichten, branchnamen en foutmeldingen kunnen door iedereen met push-rechten (of door een kwaadaardige dependency in de build) worden gevuld. Daarom: per kanaal escapen (§7), eerste regel van commit-berichten, afkappen op ~200 tekens, nooit HTML renderen, nooit mentions toestaan (Discord `allowed_mentions`, Slack-escaping).

### 8.5 Geen secrets in berichten

Build-/foutmeldingen kunnen env-waarden bevatten. Hergebruik de **bestaande scrubber** in `internal/build/scrub.go` (met tests in `scrub_test.go`; briefing §8.3: "secret-waarden vervangen door `••••`") op `Detail`/`error` vóór `Emit`. Controleer of de API ervan los van een build-context bruikbaar is; zo niet, trek hem uit naar een klein gedeeld pakket in plaats van een tweede te bouwen. Verder: nooit env-variabelen, connection-strings of volledige logs meesturen, alleen een korte samenvatting + deep-link naar het panel.

### 8.6 RBAC, audit, misbruik

- Alle `/settings/notifications/*`-routes: `requireAdmin`. Owner-only: `allow_private_targets`.
- Audit-acties: `notification_channel.create|update|delete|test`, `notification_settings.update`.
- Test-endpoint rate-limit: 5/min per user (voorkomt gebruik als HTTP-kanon).
- Max. 20 kanalen per instance; max. 20 extra e-mailadressen per kanaal.

### 8.7 Falen van de melder zelf

| Scenario | Gedrag |
|---|---|
| Kanaal onbereikbaar | retries → `dead`/`degraded`/`disabled` + melding via inbox en andere kanalen |
| SMTP niet geconfigureerd | e-mailkanaal `degraded` ("SMTP not configured"), geen retries die nooit slagen |
| Panel herstart tijdens verzenden | lease verloopt, delivery opnieuw (hooguit dubbel) |
| Disk 100 % → SQLite-write faalt | zie §9.1 |
| Notifier `nil` / DB niet klaar | `Emit` is no-op; geen paniek in deploy-pad |

---

## 9. Beperkingen en wat dit **niet** oplost

### 9.1 Een volle of dode server

- **Disk 100 %:** `Emit` doet een DB-write; bij 100 % kan die falen. Daarom alarmeren we bij 85 %/95 % (ruim vóór volraken) en loggen we altijd ook naar journald. Overweeg bij `critical` een **geheugen-only fallback**: de watcher probeert direct (zonder outbox) één send naar alle kanalen als de DB-insert faalt (best-effort, eenmalig).
- **Server of proces onbereikbaar:** wat niet draait kan niets melden. Dit is het echte gat in "bereik me zonder dashboard".
- **Mitigatie (F4, klein):** **heartbeat-kanaal**: elke N minuten (default 5) een `GET`/`POST` naar een externe ping-URL (healthchecks.io, Uptime Kuma push, Better Stack). De externe dienst alarmeert wanneer de pings **uitblijven** (dead-man's switch). Dat is ~60 regels (een tweede `kind` zonder outbox, met zijn eigen ticker) en dekt precies het scenario dat in-process alerts niet kunnen. Sterk aanbevolen follow-up; apart issue.

### 9.2 Andere beperkingen

- **At-least-once:** zeldzame dubbele berichten bij crash tussen verzenden en commit; webhooks kunnen dedupliceren op `X-dockyrd-Delivery`.
- **Slack/Discord-rate-limits** kunnen bij een alert-storm berichten vertragen; dedup + pacing maken dit zeldzaam. Een digest (F4) is de echte oplossing.
- **E-mail-deliverability** hangt af van de SMTP-provider (§7.5).
- **Tijdzones:** berichten in UTC met ISO-8601 (webhook) of leesbare UTC (Slack/Discord/e-mail); geen locale-gedoe in v1.
- **Meerdere dockyrd-servers naar één Slack-kanaal:** daarom `instance.name` (setting, default hostname) in elke footer/subject/payload.

---

## 10. API (`/api/v1/settings/notifications/…`, plain `net/http`, admin-only)

Past bij het bestaande patroon (`/api/v1/settings/email/log`, `registerEmailLogRoutes`). Nieuwe `registerNotificationRoutes(mux, deps)` aanroepen waar `registerEmailLogRoutes` wordt geregistreerd (`server/api_auth.go:82`).

```
GET    /api/v1/settings/notifications                       → { settings, event_types[], smtp_configured }
PATCH  /api/v1/settings/notifications                       → {disk_warn_pct, disk_crit_pct, cert_warn_pct, reminder_hours, instance_name, allow_private_targets*}   (*owner)

GET    /api/v1/settings/notifications/channels              → { channels: [Channel] }
POST   /api/v1/settings/notifications/channels              → 201 Channel    (validatie: 422 met veldfouten)
GET    /api/v1/settings/notifications/channels/{id}
PATCH  /api/v1/settings/notifications/channels/{id}         (secret weglaten = behouden)
DELETE /api/v1/settings/notifications/channels/{id}

POST   /api/v1/settings/notifications/channels/{id}/test    → { ok, http_status, latency_ms, error? }   (synchroon, 10 s timeout, rate-limited)
GET    /api/v1/settings/notifications/channels/{id}/deliveries?status=&limit=&before=
POST   /api/v1/settings/notifications/deliveries/{id}/retry  (dead → pending)
```

`Channel`-respons (geen secrets):

```json
{ "id": "01J…", "kind": "slack", "name": "#ops", "enabled": true, "status": "active",
  "config": { "channel_label": "#ops" }, "secret_set": true, "secret_hint": "https://hooks.slack.com/services/T0…/B0…/••••",
  "min_level": "warning", "event_types": ["deploy.failed","backup.failed","disk.high","cert.expiring"],
  "project_id": null, "last_success_at": 1762370043, "last_failure_at": null, "last_error": null,
  "consecutive_failures": 0, "created_at": 1762300000 }
```

- Fouten: bestaand `writeJSONError`. Validatie-fouten als `{"error":"…","fields":{"url":"…"}}` zodat de UI per veld kan tonen.
- De **test** gaat **niet** via de outbox (directe feedback in de modal), maar wel door dezelfde `Sender.Render/Send` en dezelfde `netsafe`-client. Resultaat wordt in de bezorglog getoond als `test`.
- `GET …/deliveries` levert per rij: tijd, event-type + titel, status, pogingen, `http_status`, fout.
- SSE: bestaand topic `events` blijft; nieuw topic `notify:channels` (status-wijziging van een kanaal) zodat de UI-badge live bijwerkt.
- OpenAPI/TS-client: net als de rest handmatig in `web/lib/api.ts` (er is geen Huma-generatie in de huidige code). Als je alsnog naar Huma gaat: apart besluit, niet in dit issue meenemen.

---

## 11. UI

Principes uit het design-contract blijven: Base UI (geen Radix), tokens, control-hoogte 44 px, mono-uppercase knoppen, UI-copy **Dockyrd**, primitives uit `web/components/dockyrd/ui/`.

### 11.1 Route en navigatie

`/settings/server/channels` ("Notification channels"), naast **Email** in de server-settings-zijnav. Reden voor deze naam: `/notifications` en `notifications-*` zijn al de inbox. Verwijs vanuit de e-mailinstellingen ("Send alerts by email → Notification channels") en vanuit het account-notificatiescherm ("Server-wide channels are managed by admins").

### 11.2 Scherm "Notification channels"

```
NOTIFICATION CHANNELS                                           [ + ADD CHANNEL ]
Get alerted when a deploy fails, a backup breaks, the disk fills up or a certificate is about to expire.

┌──────────────────────────────────────────────────────────────────────────────┐
│ [▣] #ops               Slack      ● Active     Last sent 12 min ago      ⋯  │
│ [▣] Team Discord       Discord    ▲ Degraded   Failing since 19:02       ⋯  │
│ [▣] ops@bureau.nl +2   Email      ● Active     Last sent yesterday       ⋯  │
│ [▣] n8n                Webhook    ○ Disabled   Webhook no longer exists  ⋯  │
└──────────────────────────────────────────────────────────────────────────────┘
```

- `ResourceRow`-patroon (icon tile `brand-soft`, naam 600, meta 13 px `muted`, `StatusBadge`). ⋯ `PopoverMenu`: *Send test*, *Edit*, *Delivery log*, *Enable/Disable*, *Delete* (`ConfirmDialog`).
- **Lege staat** (`EmptyState`) met vier kaarten (Slack, Discord, Email, Webhook) en één zin per kaart.
- **Statusbadges:** Active (ok), Degraded (warn, met reden), Disabled (idle, met `disabled_reason`).
- **Waarschuwingsbanner** (`Alert`) bovenaan wanneer er **geen enkel actief kanaal** is én er productie-services bestaan ("No alert channels. You will only see problems in the dashboard.") (niet-blokkerend).

### 11.3 Modal "Add channel"

Eén modal, stappen via `WizardModalShell`/`Modal`:

1. **Kind** kiezen (vier kaarten).
2. **Instellingen per kind:** Slack: *Webhook URL* (+ hulplink "How to create a Slack webhook"); Discord: *Webhook URL*; Webhook: *URL*, *Signing secret* (knop "Generate"), *Custom headers* (key/value, secret-masker); E-mail: ontvangers (checkbox "All admins", user-multiselect, vrije adressen als chips).
3. **Wat melden:** checklist gegroepeerd (Deploys, Backups, Server, Certificates) met de defaults aangevinkt; *Minimum level*; optioneel *Only for project* (`SearchSelect`).
4. **Naam** (prefill, bv. "Slack #ops") en **SEND TEST** (knop in de modal, toont resultaat inline: ✓ "Delivered in 210 ms" / foutmelding per veld). **Opslaan pas actief na een geslaagde test**, met een bewuste "Save without testing"-uitweg (geen harde blokkade).

Secrets: veld toont na opslaan alleen `secret_hint` + "Replace"; leeg laten = behouden.

### 11.4 Bezorglog-modal

Hergebruik het patroon van `EmailLogDetailModal` en `email-settings-page.tsx`: tabel (tijd, event, status-chip, pogingen, HTTP-status), klik opent detail (geredigeerde fout, volgende poging), knop **Retry** op `dead`. Filter: All / Failed.

### 11.5 Overige

- **Server-instellingen → drempels:** kaart "Alert thresholds" (disk warn/critical, certificaat-waarschuwing, herinneringsinterval, instance-naam) onderaan dezelfde pagina; `allow_private_targets` met uitleg, alleen zichtbaar voor owner.
- **Dashboard "Needs attention":** `backup.overdue`/`cert.expiring`/`disk.high` verschijnen daar al via `events` (open firing). Zorg dat `resolved` het item automatisch afhandelt (ack). Controleer `eventNeedsAttention`.
- **Inbox:** nieuwe types verschijnen in de bestaande bel (categorie-mapping §6.1).
- **Design-gat:** dit scherm staat niet in het canvas. Volgens `AGENTS.md` (design-when) is een designer-ronde **niet nodig** zolang we de bestaande settings-patronen hergebruiken; plan alleen een wireframe voor de kanaal-modal als je een afwijkend patroon wilt.
- **Toegankelijkheid:** echte `<button>`/`<a>`, `aria-label` op icon-only, status nooit alleen kleur (badge heeft label), foutmeldingen gekoppeld aan velden (`aria-describedby`).

---

## 12. Berichtteksten (voorbeelden)

| Event | Titel | Body (kern) | Link |
|---|---|---|---|
| `deploy.failed` | Deploy failed · `webshop / web` | `a1b2c3d` commit-regel · reden (bv. "out of memory", "build timeout", "healthcheck failed") | build-log |
| `backup.failed` | Backup failed · `uploads-db` | bestemming (local/remote), fout (geredigeerd), "Last successful backup: 2 days ago" | database-detail |
| `backup.overdue` | No recent backup · `uploads-db` | "Expected every 24 h, last successful 3 days ago." | database-detail |
| `disk.high` | Disk usage is high (87 %) | "/var/lib/docker: 13.1 GiB free of 100 GiB. Build cache 6.2 GiB, old images 3.4 GiB." | server/cleanup |
| `disk.high` (critical) | Disk almost full (96 %) | idem + "Builds are blocked below 10 % free." | server/cleanup |
| `cert.expiring` | Certificate expires in 6 days · `shop.example.com` | "Automatic renewal failed. DNS points to 203.0.113.9, expected 198.51.100.4." | domeinen-tab |
| `*.resolved` | ✅ Resolved · `<titel>` | "Disk usage back to 71 %." / "Backup succeeded." / "Certificate renewed." | – |
| `notification.channel_failed` | Alert channel failing · `#ops` | "Last 5 deliveries failed: HTTP 404. Channel disabled." | channels |

Toon (ook in alle kanalen) wélke instance het is en tijd in UTC.

---

## 13. Fasering en acceptatiecriteria

**Aanpak volgens `AGENTS.md`:** Feature-track, GitHub issue + `docs/project/issues/<n>/spec.md` + `rubric.md` (planner), generator met TDD, reviewers: **security-reviewer** (SSRF, secrets) en **database-reviewer** (migratie, claim-query); `api-reviewer` optioneel. Grootte: **L**. Aanbevolen splitsing in 4 issues (F1–F4), F0 als Direct-track.

### Fase 0 · Hardening (Direct-track, klein)

- `mail`: timeouts, STARTTLS-afdwinging, CRLF-sanitering, RFC 2047-subject, `Date`/`Message-ID`; unit-tests (CRLF-injectie, timeout tegen een hangende fake-SMTP).
- `netsafe`: http-client-factory met dial-guard en URL-redactie (+ tests). Nog niet gebruikt, wel klaar.
- **`MarkInterruptedDeployments` beperken tot opstart** (of lease-gebaseerd), met regressietest "een build in `building` blijft `building` over meerdere reconcile-ticks" (§2.1 gat 7). **Voorwaarde voor F1.** Eerst bevestigen met een reproductie (lange build + `Tick` aanroepen) omdat dit uit codelezen komt.
- **Klaar als:** bestaande reset/invite-mail ongewijzigd werkt; een hangende SMTP-server houdt `SendPlain` niet langer dan de timeout vast; een lopende build wordt niet meer als `interrupted` gemarkeerd.

### Fase 1 · Kern + webhook/Slack/Discord + deploy (Feature)

- Migratie `00006`, sqlc-queries, `internal/notify` (Emit, alert-state, outbox, dispatcher, backoff, circuit-breaker), `Sender`-registry, Slack/Discord/webhook, `deploy.failed/succeeded` via `Emit`, API + UI (lijst, toevoegen, test, bezorglog, retry), `docs/user/notifications.md`.
- **Klaar als:**
  - mislukte deploy → bericht in Slack, Discord en webhook **binnen 10 s** (met echte test-kanalen), met correcte kleur/escaping/link;
  - kanaal geeft 500 → retries volgens schema; 429 + `Retry-After` wordt gerespecteerd; 404 → kanaal `disabled` + `channel_failed`-melding via inbox;
  - `kill -9 dockyrd` tijdens bezorging → na herstart wordt de delivery alsnog verstuurd (max. één dubbel);
  - secrets komen nergens in API/logs/DB-plaintext voor (test met token-URL);
  - SSRF-tests: `http://127.0.0.1`, `169.254.169.254`, `10.0.0.1`, DNS-naam die naar loopback wijst, redirect naar intern: **alle geweigerd**;
  - Footprint: idle RSS binnen +2 MB van de baseline.

### Fase 2 · E-mailkanaal + backups

- E-mailsender (+ `email_log`-integratie `kind=notification`), ontvanger-logica (admins/users/extra), dedup met inbox-mail, optioneel inbox-mail naar outbox verplaatsen; `backup.failed` (+ `resolved`) hook; `watch.BackupOverdue`.
- **Klaar als:** mislukte backup (gesimuleerd: onbereikbare S3) → alert; volgende succesvolle backup → `resolved`; policy zonder run > 2× interval → `backup.overdue`; admin krijgt voor één `deploy.failed` **niet twee** mails; ontbrekende SMTP geeft duidelijke UI-staat in plaats van stille retries.

### Fase 3 · Disk + certificaten + drempels

- `watch.Disk` (+ docker-root-mount), `watch.Certs` (+ vullen van `domains.cert_status/cert_expires_at`), alert-state-UI/drempels, herinneringen, escalatie warning → error.
- **Klaar als:** disk op 86 % (gesimuleerd met een gevulde loop-mount of een testhook) → één `warning`; 96 % → één `error`; daling < 80 % → één `resolved`; geen herhaling binnen de herinneringsperiode; een domein met geforceerd verlopen test-certificaat → `cert.expiring`; na vernieuwing → `resolved`.

### Fase 4 · Optioneel (apart plannen)

`service.crash_loop`/`service.oom` (Docker events of `RestartCount` uit inspect; de collector zet `restarts` nu hard op 0), `update.available`/`update.failed`, **heartbeat** (§9.1), digest bij storm, ntfy/Telegram-presets bovenop de webhook-sender, "Test all channels".

---

## 14. Teststrategie

| Laag | Wat | Hoe |
|---|---|---|
| Unit | Renderers (golden files per kanaal), afkappen op limieten (rune-safe, grenswaarden 255/256/257 etc.), escaping (Slack `<!channel>`, Discord `@everyone`), HMAC-testvector, backoff-schema, filter-matching (level/type/project), alert-state-transities met nep-klok (incl. hysterese en escalatie), CRLF-strip, URL-redactie | `go test`, golden files naast de code |
| SSRF | Tabel van IP's/hosts (v4, v6, mapped, CGNAT, link-local), DNS-rebinding met eigen resolver, redirect-weigering | `go test` met custom `Resolver`/`Dialer` |
| Integratie (zonder Docker) | `httptest`-servers die 200/204/400/401/404/410/429(+`Retry-After`)/500/timeout/redirect teruggeven; assert statusovergangen, `degraded`/`disabled`, `channel_failed`-event, geen lus | `go test ./internal/notify/...` |
| Crash-herstel | delivery met verlopen lease → opnieuw geclaimd; `kill`-simulatie tussen send en commit → ten hoogste dubbel | `go test` |
| SMTP | kleine in-process fake SMTP-server (geen extra dependency): 250, 4xx, 5xx, hang | `go test` |
| API | RBAC (member → 403), secrets nooit in GET, PATCH zonder secret behoudt, 422-validatie (Slack-host-allowlist), test-endpoint, rate-limit; volg `testMuxWithOwnerSession` | `go test ./internal/server/...` |
| Migratie | up/down op een kopie van een gevulde DB; oude binary leest nieuwe DB | `go test` + handmatige check |
| UI | kanaal toevoegen + test + opslaan, lege staat, degraded-badge, secret-masker | Playwright (`e2e/tests`) |
| Smoke | `scripts/verify-notify-smoke.sh`: lokale listener, webhook-kanaal, falende deploy-fixture, assert payload+signature (volgt `verify-m4-smoke.sh`) | shell |
| Handmatig | Echte Slack- en Discord-testkanalen: bericht, kleur, link, escaping van een kwaadaardige commit-tekst | runbook in `docs/user/notifications.md` |
| Footprint | `scripts/measure-rss.sh` voor/na; 1.000 gesimuleerde deliveries zonder geheugengroei | script |

Gates volgens `docs/project/setup.md`: `make test`, `make lint`, `make build`, `pnpm --dir web exec tsc --noEmit`, `./scripts/check-no-radix.sh`.

---

## 15. Risico's

| Risico | Mitigatie |
|---|---|
| Alert-storm (flapperende disk, crash-lus) | alert-state, hysterese, herinneringsinterval, coalescing, pacing per kanaal |
| Lekken van webhook-token via foutmeldingen | `redact()` overal, test op token-aanwezigheid, nooit `err.Error()` van `*url.Error` ruw loggen |
| SSRF via webhook | dial-time IP-guard, geen redirects, default-deny private, owner-only opt-in |
| Mail in spam | docs-advies (transactionele SMTP + SPF/DKIM), `email_log` voor diagnose |
| Dubbele e-mail (inbox + kanaal) | ontvanger-dedup (§6.4), uiteindelijk één pad via outbox |
| Meldingen verdwijnen bij lege kanalen onopgemerkt | banner "No alert channels", `channel_failed`-melding via inbox |
| Schema-uitbreiding van `events` breekt bestaande queries | alleen nullable kolommen toevoegen; sqlc-queries en `eventFromRow` uitbreiden; bestaande rijen blijven `type=NULL` |
| `UPDATE … RETURNING` + sqlc-compatibiliteit | fallback met twee statements in één transactie |
| Certificaat-check hit eigen proxy en vertraagt | handshake-timeout 5 s, sequentieel met kleine pauze, elke 6 u |
| Docker-root op andere mount dan data-dir | detecteren via Docker-info, aparte dedup-sleutel per mount |
| Scope-creep (Teams, PagerDuty, SMS, interactieve Slack) | expliciet buiten scope (§1.2); generieke webhook als uitweg |

---

## 16. Open beslissingen voor Joey

1. **Instance-brede kanalen** (D1) of toch per user? Voorstel: instance-breed, admin-beheer.
2. **Defaults per kanaal:** `deploy.failed`, `backup.failed`, `backup.overdue`, `disk.high`, `cert.expiring` aan; `deploy.succeeded`/`backup.succeeded` uit. Akkoord?
3. **Drempels:** disk 85 %/95 % (resolve < 80 %), certificaat 15 %/5 % van de levensduur (max. 14/3 dagen), herinnering 24 u. Akkoord?
4. **`resolved`-berichten** altijd sturen (ook als `backup.succeeded` extern uit staat)? Voorstel: ja, alleen als er een open `firing` was.
5. **Private webhook-doelen** standaard geblokkeerd met owner-opt-in (D9)? Voorstel: ja.
6. **Inbox-mail verplaatsen naar de outbox** in F2 (voorkeur) of laten staan na F0-timeouts?
7. **Heartbeat/dead-man's switch** meenemen in F4 of als apart issue vóór v0.2-release? Mijn advies: apart, direct na F3; het dekt het scenario dat deze feature principieel niet kan.
8. **Huma versus stdlib-mux:** `AGENTS.md`/briefing noemen Huma, de code gebruikt `net/http` + `writeJSON` en heeft Huma niet in `go.mod`. Voorstel: nieuwe routes volgen de code (stdlib) en de docs corrigeren; geen Huma-migratie in dit issue.
9. **Instance-naam** (`instance_name`) verplicht tonen in alle berichten? Voorstel: default hostname, aanpasbaar.
10. **`MarkInterruptedDeployments` (§2.1 gat 7):** akkoord dat dit eerst als Direct-track bugfix wordt rechtgezet (alleen bij opstart)? Daarna: interrupted deploys ook als `deploy.failed` melden? Voorstel: ja en ja. Kun jij bevestigen dat je nu al flikkerende "interrupted"-deploys in de UI ziet, of werkt het in de praktijk doordat de runtime de status daarna overschrijft?

---

## 17. Definition of done

- [ ] Mislukte deploy, disk ≥ 85 %, certificaat bijna verlopen en mislukte backup bereiken Slack, Discord, e-mail én webhook, **zonder dashboard**.
- [ ] Per kanaal: test-knop, bezorglog, retries, status (Active/Degraded/Disabled), `channel_failed`-melding.
- [ ] Stateful alerts: één `firing`, geen spam, één `resolved`, herinnering na 24 u, escalatie bij niveau-stijging.
- [ ] `kill -9` tijdens bezorging: niets kwijt (max. één dubbel); deploy-pad wordt nooit vertraagd door netwerk.
- [ ] Geen secrets in API/logs/berichten; SSRF-tests groen; geen redirects; CRLF-tests groen.
- [ ] Geen nieuwe Go-dependencies; idle RSS binnen +2 MB; `check-no-radix.sh` groen.
- [ ] Migratie additief; rollback naar vorige binary werkt.
- [ ] `docs/user/notifications.md` (kanalen opzetten, webhook-schema + verificatie, troubleshooting), `CHANGELOG.md`-entry (patch +0.0.1), issue-handoff met testoutput.

---

## Bijlage A: statusmachine alert-state

```
                 firing (nieuw / na resolved)
   ┌─────────┐ ───────────────────────────────► ┌─────────┐
   │resolved │                                  │ firing  │ ◄─┐ firing, zelfde level, < reminder: onderdrukken
   └─────────┘ ◄─────────────────────────────── └─────────┘ ──┘ firing, level ↑ of ≥ reminder: melden (occurrences++)
                 resolved (conditie weg, met hysterese)
```

## Bijlage B: sequentie "mislukte backup"

```
backup.Runner ──UpdateDatabaseBackupRun(failed)──► SQLite
      │
      └─ Notify.Emit(backup.failed, firing, key backup:<svc>:remote)
              │  BEGIN IMMEDIATE
              │   alert_state: nil → firing                 (melden)
              │   events: INSERT (type=backup.failed)
              │   channels: Slack ✓  Discord ✓  Email(3 ontvangers) ✓  Webhook ✗ (filter)
              │   deliveries: INSERT ×5 (pending)
              │  COMMIT → wake dispatcher
              ▼
Dispatcher ──claim(lease 60 s)──► render ──► POST Slack ──200──► sent
                                         └─► POST Discord ─429 Retry-After 2──► pending (+2 s) ──204──► sent
                                         └─► SMTP ×3 ──250──► sent  (+ email_log kind=notification)
volgende succesvolle backup ──► Emit(backup.succeeded, resolved) ──► "✅ Resolved" naar dezelfde kanalen
```

## Bijlage C: aanraakpunten in de code (verwacht)

| Bestand | Wijziging |
|---|---|
| `internal/db/migrations/00006_notifications.sql`, `db/queries/notifications.sql`, `sqlc` regenereren | nieuw |
| `internal/notify/*`, `internal/watch/*` | nieuw |
| `internal/mail/smtp.go` | F0-hardening |
| `internal/server/runtime.go` | `Notifier`, dispatcher, watchers starten in `startBackground`; doorgeven aan `deploy.Runtime`, `jobs.Worker`, `reconcile.Loop` |
| `internal/deploy/runtime.go` (`fail`, live-pad ~236, `recordDeployEvent`), `deploy/static.go:137`, `deploy/admission.go` | `Inbox.Notify…` → `Notify.Emit`; `build.waiting` blijft ongetypeerd |
| `internal/backup/runner.go` (`runDestination`), `internal/jobs/worker.go` (`Worker`, `runner :=`) | `Notify`-veld, `Emit` bij failed/succeeded |
| `internal/reconcile/reconcile.go` + `queries` (`MarkInterruptedDeployments`) | ids teruggeven, `deploy.failed` (interrupted); prune-batches voor deliveries/events |
| `internal/inbox/dispatch.go` | aanroepbaar vanuit `Notifier`; ontvanger-set teruggeven voor e-mail-dedup; later mail via outbox |
| `internal/server/api_notifications.go`, `server/api_auth.go:82` | routes registreren |
| `internal/server/api_events.go` (`eventNeedsAttention`, categorie-filter) | nieuwe types/categorieën meenemen |
| `web/lib/api.ts`, `web/components/dockyrd/notification-channels-*`, settings-zijnav, `web/app/settings/server/channels` | UI |
| `docs/user/notifications.md`, `docs/BRIEFING.md` §1.2/§15, `CHANGELOG.md`, `docs/project/roadmap.md` | docs |

## Bijlage D: bronnen en wat is geverifieerd

Geverifieerd tijdens het schrijven:

- **Discord-embedlimieten** (title 256, description 4096, ≤ 25 fields, field name 256 / value 1024, footer 2048, author 256, totaal 6000 over alle embeds, ≤ 10 embeds): https://www.pythondiscord.com/pages/guides/python-guides/discord-embed-limits/ (community-gids; controleer bij twijfel tegen de officiële Discord-docs).
- **Slack-ratelimit** voor incoming webhooks (1 request/s, korte bursts toegestaan, 429 met `Retry-After`): https://docs.slack.dev/apis/web-api/rate-limits
- **Caddy bij mislukte uitgifte/vernieuwing** (opnieuw proberen, andere challenge-types, volgende CA, exponentiële backoff met max. een dag tussen pogingen tot 30 dagen): https://caddyserver.com/docs/automatic-https

**Niet bevestigd** (als zodanig in de tekst gemarkeerd, vóór implementatie checken):

- Block Kit-tekstlimieten voor Slack (`section` 3000, `header` 150) en of `attachments` + `blocks` voor de kleurbalk nog de aanbevolen vorm is.
- Het exacte vernieuwvenster van CertMagic/Caddy (niet op de docspagina) en de actuele plannen van Let's Encrypt voor kortere levensduren; daarom zijn de certificaatdrempels **relatief** aan de levensduur.
- Discord-gedrag rond `User-Agent` voor webhooks (we zetten er altijd een expliciete).
- SQLite-versie-eigenschappen (`UPDATE … RETURNING`) voor `modernc.org/sqlite v1.36.0`; controleren met een unit-test.

Intern: `internal/eventlog`, `internal/inbox`, `internal/mail`, `internal/jobs`, `internal/backup`, `internal/metrics`, `internal/instance`, `internal/reconcile`, `internal/server/api_*`, migraties `00001`–`00005`, `dockyrd_BRIEFING.md` §1.2, §3, §8.3, §11.4, §14, §15, `AGENTS.md`, `CLAUDE.md`.
