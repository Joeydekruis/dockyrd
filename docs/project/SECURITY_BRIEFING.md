# Dockyrd: Security- en performance-briefing

Status: ontwerp en design klaar (canvas "Dockyrd Dashboard", pagina **Security**). Nog geen code. Gebaseerd op de repo (`development` @ `2a00c7c`), gelezen en doorzocht, niet gedraaid. Wat niet bevestigd is, staat gemarkeerd met **(niet bevestigd)** en moet vóór implementatie gecheckt worden.

## 0. Samenvatting

1. **Eerst de gaten dichten die er al zijn** (§2): `no-new-privileges` ontbreekt op gewone apps en databases terwijl de docs het beloven, er is geen compressie, geen cache-control en geen security headers in de Caddy-config.
2. **Geen eigen WAF.** We bieden een **curated add-on-catalogus** aan die Dockyrd zelf installeert, test, kan terugdraaien en logt: firewall, automatische security updates, CrowdSec, en een Dockyrd-proxybuild met Coraza (WAF) en rate limiting.
3. **Drie lagen met een vloer en een plafond:** server (instance policy), project (isolatie, bestaat al) en site (security-profiel). De instance zet defaults en minimums, een site kan strenger worden, en losser alleen als de admin dat toestaat.
4. **Eén concept per site: het security-profiel** (Baseline, Strict, Custom). De solo dev hoeft niets te doen, de freelancer krijgt isolatie en limieten per klantsite, het bedrijf krijgt een policy-lock.
5. **Alles wat de host aanraakt gaat door getypeerde host actions** (`internal/hostops`) met preview, verificatie, revert en audit log. Lockout-bescherming (SSH, firewall) draait buiten het panel-proces zodat een crash niet uitsluit.

## 1. Doel, gebruikers en scope

| Gebruiker | Wat ze nodig hebben | Wat Dockyrd doet |
|---|---|---|
| **Solo dev** | Veilig zonder nadenken | Baseline-profiel overal, security-checklist met knoppen, geen configuratie nodig |
| **Freelancer met klantsites** | Klantsites mogen elkaar en de server niet raken | Netwerkisolatie per project (bestaat, M6), CPU/geheugen/pids-limieten per site, WAF-modus per site, klanten als Viewer, per-site security events |
| **Bedrijf met eigen sites** | Beleid afdwingen, aantoonbaar | Instance policy met lock, verplichte 2FA (bestaat), audit log (bestaat), vaste add-on-set |

**Uitgangspunten**

- Veilig by default, zonder configuratie.
- Geen vrije package-installatie. Alleen items uit de catalogus, elk met preview, verificatie, uninstall en audit.
- Elke wijziging aan de host toont vooraf precies wat er verandert.
- Lean: geen extra containers of processen tenzij de gebruiker een add-on aanzet.
- Detect-only vóór block. Nieuwe blokkerende functies starten in simulatie.

**Niet in scope (nu):** een eigen WAF-engine, ClamAV (bewust overgeslagen: ~1 GB RAM past niet bij het footprint-doel), vrije package-installatie, WordPress-specifieke regels, multi-server (v0.3), SSO.

## 2. Stand van de code: gaten

Gevonden met zoeken en lezen. Elk punt is klein en hoort in de Direct-track (geen issue, wel regressietest).

| # | Bevinding | Waar | Gevolg | Fase |
|---|---|---|---|---|
| 1 | Apps en databases draaien zonder `no-new-privileges`. Alleen Compose-services krijgen het. De docs (`site/docs/security.html`) zeggen dat alle user containers het hebben. | `internal/docker/run.go`, `database.go`; wel in `internal/compose/mutate.go:94` | Docs kloppen niet. Privilege escalation via setuid-binaries in een gecompromitteerde container is niet geblokkeerd. | P0 |
| 2 | Geen `cap_drop`, geen pids-limiet, geen CPU-limiet. Alleen geheugen. | `internal/docker/run.go` (HostConfig), `database.go` | Fork bomb of CPU-hog van één site raakt alle andere sites. Dit is het noisy-neighbour-risico voor freelancers. | P0 |
| 3 | De gerenderde Caddy-config heeft geen `encode` (zstd/gzip). De briefing (§3.5) beschrijft het wel. | `internal/proxy/render.go`, `static_handles.go` | Geen compressie: trager en duurder in bandbreedte. | P0 |
| 4 | Geen `Cache-Control` voor statische assets en geen security headers. | idem | Onnodige herhaalde downloads, geen HSTS/nosniff/referrer-policy. | P0 |
| 5 | Geen server-timeouts of request-limieten in de proxyconfig. | idem | Slowloris en grote bodies zijn niet begrensd. **(niet bevestigd)** dat er niets elders staat. | P0 |
| 6 | Geen rate limiting op sites. Login-brute-force op het panel is wel afgedekt. | `internal/auth`, `internal/proxy` | Sites zijn onbeschermd tegen floods en credential stuffing op eigen loginpaden. | P2 |
| 7 | **`dockyrd.service` heeft geen `User=`, dus de service draait als root** (met `NoNewPrivileges=true`, `ProtectSystem=full`). Het panel heeft bovendien de Docker-socket. | `packaging/dockyrd.service` | Zie §3.3: een aparte "helper zonder root" levert nu geen privilege-scheiding op. | ADR |
| 8 | De host-firewall ziet gepubliceerde Docker-poorten niet: Docker DNAT't in `PREROUTING`, die verkeer komt nooit door de `INPUT`-keten waar UFW zit. | Docker-gedrag, zie bijlage A | Een "firewall aan"-vinkje geeft schijnveiligheid. Dockyrd moet gepubliceerde poorten zelf controleren. | P1 |

## 3. Architectuur

### 3.1 Lagen en de vloer/plafond-regel

| Laag | Beheerd door | Inhoud |
|---|---|---|
| **Server (instance)** | Owner/Admin | Firewall, SSH, automatische updates, CrowdSec, proxybuild, standaardprofiel, headers, rate-limit-defaults, CDN/trusted proxies, standaardlimieten, lock |
| **Project** | Owner/Admin | Netwerkisolatie en gerichte links (bestaat, M6), toegang via rollen (bestaat) |
| **Site (app/static)** | Admin, Member | Profiel, CSP, Access (wachtwoord, IP-allowlist, onderhoud), rate limit, WAF-modus, resources |

**Resolutie.** Per instelling geldt: `effectief = strengste(instance.vloer, site.keuze)`. Staat de **lock** uit, dan is de instance-waarde alleen een default. Staat de lock aan, dan wordt een zwakkere sitewaarde bij opslaan afgewezen (HTTP 422 met de naam van de instelling). Een admin kan de lock voor één site opheffen via een expliciete uitzondering die in de audit log komt.

### 3.2 Security-profielen

| Instelling | Baseline | Strict | Custom |
|---|---|---|---|
| Headers | HSTS (6 maanden), `nosniff`, `Referrer-Policy: strict-origin-when-cross-origin`, `X-Frame-Options: SAMEORIGIN`, geen `Server`-header | Baseline + `Permissions-Policy` (minimaal), HSTS 1 jaar | Per header aan/uit, nooit zwakker dan de lock |
| Compressie | zstd + gzip | idem | idem |
| Rate limit (indien proxybuild) | 600 req/min per IP, 30 req/min op login- en admin-paden | 300 en 20 | Zelf in te vullen |
| WAF (indien proxybuild) | Uit | **Detect-only** | Uit, Detect-only of Block |
| Container-hardening | Altijd aan, los van het profiel | idem | idem |

Strict zet de WAF bewust op **detect-only**, niet op block. Een blokkerende WAF veroorzaakt false positives en dus uitval. Block is een expliciete keuze per site, nadat je de events hebt gezien.

### 3.3 Host actions (`internal/hostops`)

Alles wat de host wijzigt, loopt door getypeerde acties. Geen shell-strings, geen vrije invoer in commando's.

Elke actie implementeert:

| Methode | Doel |
|---|---|
| `Plan()` | Geeft een preview terug: bestanden, diff, pakketten, regels. Wordt in de UI getoond vóór bevestiging. |
| `Apply()` | Voert uit. Schrijft een journaalregel (`host_actions`). |
| `Verify()` | Controleert het resultaat (bijv. `sshd -T`, `nft list`, `systemctl is-active`). Een mislukte verify draait automatisch terug. |
| `Revert()` | Zet de vorige toestand terug. Elke actie bewaart wat nodig is om terug te draaien. |

**Acties v1:** `firewall.ensure`, `ssh.passwordauth.disable`, `unattended.enable`, `crowdsec.bouncer.install` (route: zie §11), `proxy.swap`. Geen generieke "voer commando uit".

**Lockout-bescherming** (SSH en firewall). Na `Apply()` start Dockyrd een **tijdelijke systemd-timer** (`systemd-run --on-active=120 dockyrd host revert <id>`). De gebruiker bevestigt via de UI ("Keep change"). Zonder bevestiging draait de timer de actie terug. Omdat de timer in systemd leeft, werkt het ook als het panel crasht of herstart. Bij opstart reconcilieert Dockyrd openstaande acties met hun deadline.

**SSH-details.**

- De drop-in heet `/etc/ssh/sshd_config.d/00-dockyrd.conf`. De naam is belangrijk: bij sshd wint de **eerste** waarde, en cloud-init zet vaak `50-cloud-init.conf` met `PasswordAuthentication yes`. Onze file moet dus ervoor sorteren.
- Vóór het herladen: `sshd -t`. Na het herladen: `sshd -T` moet `passwordauthentication no` tonen. Zo niet, dan revert.
- Pre-checks: er bestaat een `authorized_keys` voor root of een andere sudo-gebruiker, en de huidige sessie gebruikte een sleutel. Anders is de actie geblokkeerd.

**Waarom (nog) geen aparte helper-daemon.** Het panel draait al als root en heeft de Docker-socket, wat root-equivalent is. Een helper naast een root-panel voegt geen beveiligingsgrens toe. De waarde van `hostops` zit in vaste acties, preview, revert en audit, niet in privilege-scheiding. Echte scheiding vraagt een grotere stap: panel als eigen gebruiker, Docker via een beperkte socket-proxy, en dan pas een helper voor de weinige root-acties. Dat is een eigen ADR (§11, punt 1). De Add-ons-pagina noemt dit nu "Host actions" en belooft geen scheiding.

### 3.4 Add-on-framework

Een add-on is een Go-struct in de Dockyrd-binary (geen door gebruikers aangeleverd manifest):

| Veld | Betekenis |
|---|---|
| `ID`, `Title`, `Kind` | `baseline` of `optional` |
| `Requires` | Bijv. distro-ondersteuning, ander add-on |
| `Plan()` | Lijst host actions en containerstappen voor de preview |
| `Install()`, `Uninstall()` | Idempotent, met rollback bij falen |
| `Health()` | Voor `dockyrd doctor` en de status op de pagina |

**Toestanden:** `not_installed` → `installing` → `installed` | `failed` → `uninstalling` → `not_installed`. Elke overgang schrijft een event en een audit-regel. Mislukte installaties laten de host terug in de oude toestand en tonen de logregels in de UI.

## 4. Features

### 4.1 Baseline-checks (Server → Security → Overview)

Elke check heeft een id, een detectiemethode, een ernst en optioneel een fix.

| Check | Detectie | Fix |
|---|---|---|
| `firewall.active` | nft-tabel `dockyrd` aanwezig, default policy | `firewall.ensure` |
| `docker.published_ports` | Containers met HostPort op `0.0.0.0` of `::` behalve de proxy op 80/443 | Geen automatische fix. Lijst met link naar de service. Dockyrd sluit nooit zelf poorten. |
| `ssh.password_auth` | `sshd -T` | `ssh.passwordauth.disable` (ook `PermitRootLogin prohibit-password`) |
| `updates.unattended` | Pakket en apt-config aanwezig | `unattended.enable` (alleen Debian/Ubuntu, anders "handmatig") |
| `reboot.pending` | `/var/run/reboot-required` | Notice, geen automatische reboot |
| `panel.2fa` | Instelling "2FA verplicht voor owners en admins" | Link naar Settings |
| `panel.https` | Dashboard-URL is https | Link naar Instance |
| `proxy.defaults` | Gerenderde config bevat headers en encode | Automatisch na P0 |

De pagina draait de checks bij openen, via "Run checks again" en elke 24 uur. Een afwijking maakt een event "Needs attention" (bestaat al).

### 4.2 Add-on-catalogus

| Add-on | Soort | Doet | Nodig |
|---|---|---|---|
| **Firewall** | Baseline | nftables: alleen 22, 80, 443 binnen. Aparte controle op gepubliceerde Docker-poorten. | Host action |
| **Automatische security updates** | Baseline | `unattended-upgrades`, alleen security, nooit zelf rebooten | Host action, Debian/Ubuntu |
| **CrowdSec** | Aanbevolen | Leest proxy- en SSH-logs, blokkeert IP's aan de firewall, community-blocklists | Container, firewall-bouncer |
| **WAF (Coraza)** | Opt-in per site | OWASP core rules | Dockyrd-proxybuild |
| **Rate limiting** | Opt-in per site | Limieten per IP per site | Dockyrd-proxybuild |

WAF en rate limiting delen één proxybuild. De tweede installeren vraagt geen proxywissel meer.

### 4.3 CrowdSec

- **Runtime:** container `dy-crowdsec` (gepinde digest) op een systeemnetwerk. Collections: `crowdsecurity/sshd`, `crowdsecurity/caddy`, `crowdsecurity/linux`. **(niet bevestigd)** exacte collectienamen en of de Caddy-collectie onze JSON access log (`/var/lib/dockyrd/caddy/access.log`) zonder aanpassing parsed.
- **SSH-logs:** CrowdSec heeft toegang nodig tot de auth-log of journald. Op Debian 12 kan alleen journald beschikbaar zijn. **(niet bevestigd)** welke datasource per distro werkt.
- **Simulatiemodus:** CrowdSec kent per scenario `cscli simulation enable <scenario>`. Gesimuleerde beslissingen verschijnen in `cscli decisions list` als `(simul)ban` en "won't be effective". **(niet bevestigd)** of bouncers gesimuleerde beslissingen ook echt niet krijgen. Dat moet in een test staan voordat we "Start in simulation mode" beloven. Vallen we hierop terug, dan draait de bouncer pas bij "Start enforcing".
- **Altijd toegestaan (allowlist):** het huidige IP van de installerende gebruiker, de IP's van de server, en recente admin-sign-in-IP's (30 dagen). Dockyrd vult dit automatisch en toont het onder "Never blocked". **(niet bevestigd)** welk allowlist-mechanisme de gebruikte CrowdSec-versie heeft (`cscli allowlists` of een parser-whitelist).
- **Bouncer:** zie §11, punt 4.
- **UI:** de tabel "Decisions" komt uit de lokale API van CrowdSec. "Allow IP" voegt toe aan de allowlist en heft het blok op.
- **Uninstall:** verwijdert container, bouncer en alle firewallregels. Audit-regel blijft.

### 4.4 Dockyrd-proxybuild (WAF en rate limiting)

- Eigen Caddy-image uit onze release-pipeline met `xcaddy build --with github.com/corazawaf/coraza-caddy/v2 --with github.com/mholt/caddy-ratelimit`. Gepind op digest, gesigneerd (cosign) en gepubliceerd naast de Dockyrd-release.
- **Wissel (`proxy.swap`):** nieuwe image pullen, de volledige config valideren met de **nieuwe** build, nieuwe container starten, schakelen, oude container bewaren tot de nieuwe antwoordt. Mislukt een stap, dan blijft de oude proxy actief en komt er een event. Dit sluit aan op het bestaande patroon (volledige config via `POST /load`, atomisch).
- **Alles uit na installatie.** Geen enkele site verandert van gedrag zolang WAF en rate limit uit staan.
- **Coraza:** de directive heet `coraza_waf` en vraagt `order coraza_waf first`, `load_owasp_crs` en een `directives`-blok met `SecRuleEngine`. De bronpagina toont alleen `On`. `DetectionOnly` is een standaardwaarde van Coraza/ModSecurity maar is niet op de gebruikte pagina bevestigd, **(niet bevestigd)** dus testen. De JSON-handlernaam is ook niet bevestigd. Valideren via een golden test met `caddy adapt`.
- **Rate limit:** JSON-handler `rate_limit` met zones `{key, window, max_events}`. Antwoord is 429 met `Retry-After`. Let op: de module is geen onderdeel van het officiële Caddy-project, en de volgorde ten opzichte van `basic_auth` telt. **(niet bevestigd)** de exacte placeholder-vorm voor het IP in JSON (`{http.request.remote.host}`).
- **Eerlijke kanttekening:** Coraza kost CPU en geeft false positives. Daarom detect-only als standaard, een events-lijst per site, en per-pad uitzonderingen voordat iemand Block kiest.

### 4.5 Instance policy (Server → Security → Policy)

- **Standaardprofiel** (Baseline of Strict) en de **lock**.
- **Headers** per stuk aan/uit, met waarde zichtbaar. CSP staat hier niet: dat is per site, want Dockyrd raadt geen policy.
- **Rate limits:** default per minuut en voor login/admin-paden. Staat uitgegrijsd met "Needs the proxy build" zolang die niet geïnstalleerd is.
- **CDN of proxy:** Direct, Cloudflare of Overig. Zonder dit zien rate limits, CrowdSec en analytics alleen het IP van de CDN.
  - Cloudflare: Dockyrd haalt de gepubliceerde IP-ranges dagelijks op en rendert ze als statische `trusted_proxies`. Header `CF-Connecting-IP` via `client_ip_headers`. Zo is er geen plugin nodig. **(niet bevestigd)** exacte JSON-veldnamen (`trusted_proxies.source = static`, `ranges`, `client_ip_headers`, `trusted_proxies_strict`); controleren tegen de Caddy-docs van de gepinde versie.
  - Overig: zelf ranges en headernaam invullen.
  - Detectie: als er verzoeken binnenkomen vanaf bekende Cloudflare-ranges, toont de UI een hint.
- **Standaardlimieten voor nieuwe services:** geheugen (512 MB), CPU (1.0 vCPU), processen (512).

### 4.6 Site-niveau (Application en Static site → Security)

Nieuwe tab "Security" (vóór Settings). Zie design `ApplicationSecurity`.

| Blok | Inhoud | Toepassing |
|---|---|---|
| **Profiel** | Baseline, Strict, Custom. Toont "Instance default" en de lock. | Direct, via proxyconfig |
| **Content security** | CSP-tekstveld, schakelaar "Report only" (`Content-Security-Policy-Report-Only`) | Direct |
| **Access** | Wachtwoordbeveiliging (basic auth), "Only allow these IPs" (403 voor de rest), onderhoudspagina (503) | Direct |
| **Traffic protection** | Rate limit: instance default, custom, uit. WAF: uit, detect, block. | Direct, alleen als proxybuild er is |
| **Resources** | Geheugen, CPU, processen | **Herstart** met blue/green, dus zonder downtime |
| **Security events** | Verzoeken die een regel of limiet raakten | Zie §5 |

- **Wachtwoord:** wordt als hash opgeslagen (bcrypt, via Caddy's `basic_auth`) en één keer getoond. Dockyrd kan het niet teruggeven.
- **IP-allowlist en je eigen IP:** de UI toont het huidige IP van de gebruiker en waarschuwt als dat er niet in staat.
- **Static sites:** zelfde blok, zonder Resources en WAF-modus per container. De `Access`-tab uit briefing §11.2 (v0.2) valt hier samen met dit blok.
- **Zonder proxybuild:** "Traffic protection" toont een uitleg met "Install" in plaats van bedieningselementen.

### 4.7 Valkey-presets (databases)

Valkey blijft de Redis-compatibele keuze (BSD). Het label in de UI blijft "Redis". We voegen geen echte Redis toe.

| | **Cache** | **Queue / Sessions** |
|---|---|---|
| `maxmemory` | 75% van de containerlimiet | 75% van de containerlimiet |
| `maxmemory-policy` | `allkeys-lru` | `noeviction` |
| Persistence | Uit | AOF aan (`appendonly yes`, `everysec`) |
| Bedoeld voor | Next.js-cache, ISR, tijdelijke data | Jobqueues, sessies, data die niet weg mag |
| Waarschuwing | geen | Event bij 80% geheugen |

Altijd: wachtwoord (bestaat), geen host-poort, `protected-mode yes`. De preset is een keuze bij aanmaken en daarna aan te passen met een gecontroleerde herstart (bestaat al voor config-wijzigingen).

### 4.8 Container-hardening

| Maatregel | Apps (Railpack/Dockerfile/image) | Databases | Opmerking |
|---|---|---|---|
| `no-new-privileges` | Altijd | Altijd | Moet getest worden met de officiële images, **(niet bevestigd)** per image dat de privilege-drop in de entrypoint (gosu/su-exec) blijft werken. |
| `cap_drop: ALL` | Altijd | `ALL` plus een minimale set per image | De databaseset (bijv. `CHOWN`, `DAC_OVERRIDE`, `FOWNER`, `SETGID`, `SETUID`) is **(niet bevestigd)** en per image te testen in e2e. |
| Pids-limiet | 512 (instelbaar) | 512 | Voorkomt fork bombs |
| CPU-limiet | Standaard 1.0 vCPU (instelbaar) | Geen standaard | Builds lopen apart (bestaat) |
| Read-only rootfs | Optie per app, standaard uit | Nee | Breekt veel apps, dus alleen opt-in |
| Docker-socket | Nooit (bestaat, ook in Compose-policy) | Nooit | |

### 4.9 Performance

| Onderdeel | Wijziging |
|---|---|
| Compressie | `encode` met zstd en gzip op elke route (apps en static). |
| Statische assets | `Cache-Control: public, max-age=31536000, immutable` voor bestanden met een hash in de naam (bijv. `/_next/static/*`); `no-cache` voor HTML. `ETag` en `Last-Modified` van `file_server`. |
| Static precompress | Optioneel in de build: `.zst` en `.gz` genereren en `precompressed` gebruiken. Later. |
| HTTP/3 | Caddy-standaard, niets te doen. |
| Timeouts | `read_header_timeout`, `idle_timeout`, maximale headergrootte in de serverconfig. **(niet bevestigd)** de JSON-eenheden (string of nanoseconden). |
| Request body | Standaardlimiet met per-site override. |
| Valkey | Presets (§4.7). |
| Proxy zonder downtime | Bestaat (blue/green). |
| Micro-cache voor dynamische apps | Niet nu. Mogelijk v0.3 met een cache-module en Valkey als opslag. |

### 4.10 `dockyrd doctor`

Voegt toe: de baseline-checks (§4.1), distro-ondersteuning per add-on, proxybuild-versie, CrowdSec-gezondheid, en gepubliceerde poorten. Uitvoer in dezelfde vorm als de bestaande checks, met een `FixHint` en de bijbehorende host action.

## 5. Datamodel (SQLite, goose)

Volgende vrije migratienummers. Alleen parameterized SQL (sqlc).

| Tabel | Kolommen (kern) |
|---|---|
| `security_policy` (één rij) | `default_profile`, `lock`, `headers_json`, `rl_default_rpm`, `rl_login_rpm`, `cdn_mode`, `cdn_ranges_json`, `cdn_header`, `default_mem_mb`, `default_cpu`, `default_pids` |
| `service_security` | `service_id` (pk), `profile`, `csp`, `csp_report_only`, `basic_user`, `basic_hash`, `ip_allowlist_json`, `maintenance`, `rl_mode`, `rl_rpm`, `rl_login_rpm`, `waf_mode`, `mem_mb`, `cpu`, `pids` |
| `addons` | `id`, `state`, `version`, `config_json`, `installed_at`, `last_error` |
| `host_actions` | `id`, `kind`, `state`, `plan_json`, `revert_json`, `actor_user_id`, `applied_at`, `confirm_deadline`, `confirmed_at`, `reverted_at` |
| `security_events` | `ts`, `service_id`, `source` (`ratelimit`, `waf`, `crowdsec`), `rule`, `path`, `action`, `ip`, `detail` |
| `crowdsec_allow` | `cidr`, `reason`, `source` (`auto`, `user`) |

`security_events` en privacy: de analytics-briefing zegt "geen IP-opslag". Voor security events is een IP nodig. Besluit nodig (§11, punt 2). Voorstel: bewaren 7 dagen, instelbaar tot 30, apart van analytics en gedocumenteerd.

## 6. API (Huma, OpenAPI → TS-client)

Alle schrijvende routes vragen rol Admin of hoger, tenzij anders vermeld, en schrijven een audit-regel.

| Route | Doel |
|---|---|
| `GET /api/security/checks` · `POST …/run` | Baseline-checks lezen en opnieuw draaien |
| `GET /api/security/addons` | Catalogus met toestand |
| `POST /api/security/addons/{id}/plan` | Preview van een installatie (geen wijziging) |
| `POST /api/security/addons/{id}/install` · `…/uninstall` | Uitvoeren (async job, status via SSE) |
| `POST /api/security/host-actions/{id}/confirm` · `…/revert` | Lockout-bevestiging |
| `GET/PUT /api/security/policy` | Instance policy |
| `GET /api/security/crowdsec/decisions` · `POST …/allow` · `PUT …/mode` | CrowdSec-beheer |
| `GET/PUT /api/services/{id}/security` | Site-instellingen (valideert tegen vloer en lock) |
| `GET /api/services/{id}/security/events` | Events per site |

## 7. Wijzigingen in de Caddy-render

Per route, in vaste volgorde: **headers → encode → (rate_limit) → (waf) → (access: basic auth, IP-matcher, onderhoud) → reverse_proxy of file_server**. Per site wordt één zone-naam gebruikt (`site_<id>`). Op serverniveau komen `trusted_proxies`, `client_ip_headers` en timeouts.

Voorbeeld voor een app met Strict-profiel en de proxybuild. **(niet bevestigd)**: de exacte JSON-vormen van `headers`, `deferred` en `delete`, en de handlernaam van Coraza. Controleer met `caddy validate` in een golden test.

```json
{ "match": [{ "host": ["api.clientx.nl"] }],
  "handle": [
    { "handler": "headers", "response": {
        "set": { "Strict-Transport-Security": ["max-age=31536000"],
                 "X-Content-Type-Options": ["nosniff"],
                 "Referrer-Policy": ["strict-origin-when-cross-origin"],
                 "X-Frame-Options": ["SAMEORIGIN"] },
        "deferred": true, "delete": ["Server"] } },
    { "handler": "encode", "encodings": { "zstd": {}, "gzip": {} }, "prefer": ["zstd", "gzip"] },
    { "handler": "rate_limit", "rate_limits": {
        "site_01jc9k": { "key": "{http.request.remote.host}", "window": "1m", "max_events": 300 } } },
    { "handler": "reverse_proxy", "upstreams": [{ "dial": "dy-api-01jc9k:3000" }],
      "load_balancing": { "try_duration": "5s" } }
  ],
  "terminal": true }
```

## 8. UI en design

Canvas-pagina **Security** in "Dockyrd Dashboard":

| Bord | Inhoud |
|---|---|
| `ServerSecurity` | Overzicht: baseline-checks, protection add-ons, SSH-fix met preview, diff en lockout-bevestiging (interactief) |
| `ServerSecurityAddons` | Catalogus, host actions, installatieflow (preview, bezig, klaar) voor updates, CrowdSec en proxybuild |
| `ServerSecurityCrowdSec` | Simulatie- en enforcing-modus, decisions met filter en "Allow IP", collections, allowlist, engine, uninstall |
| `ServerSecurityPolicy` | Standaardprofiel en lock, headers, rate limits (uitgegrijsd zonder proxybuild), CDN, standaardlimieten |
| `ApplicationSecurity` | Profiel, CSP, Access, Traffic protection, Resources, Security events |

**Navigatie:** de sidebar-groep "Server" krijgt "Security" tussen Analytics en Settings. Dat staat nu alleen op de nieuwe borden. De overige borden en de tab "Security" op de overige applicatieborden en op de static-site-borden moeten nog volgen.

**Nieuwe UI-primitives** (Base UI, in `web/components/dockyrd/ui/`, ook in Storybook): `Switch`, `RadioCard`, `Segmented`, `CodeBlock` met diff-kleuring, `StatusRow` (check of waarschuwing), `StepList` (voortgang). Geen `@radix-ui/*`.

## 9. Fases

De volgorde is bewust: eerst gaten, dan het raamwerk, dan de opt-in extra's.

| Fase | Inhoud | Track |
|---|---|---|
| **P0** | Container-hardening (§4.8), `encode`, cache-control, baseline headers en timeouts in de proxy (§4.9), Valkey-presets (§4.7), docs corrigeren zolang `no-new-privileges` nog niet overal staat | Direct, met regressietests en e2e |
| **P1a** | `internal/hostops` met journaal, revert en lockout-timer | Feature |
| **P1b** | Baseline-checks API en UI, SSH-fix, unattended-upgrades, firewall | Feature |
| **P1c** | `security_policy`, profielen, `service_security`, site-tab (headers, CSP, Access) | Feature |
| **P1d** | CDN en trusted proxies, Cloudflare-ranges | Feature |
| **P2a** | Release-pipeline voor de proxybuild (xcaddy, digest, cosign) en `proxy.swap` | Feature |
| **P2b** | Rate limiting (instance, site) | Feature |
| **P2c** | Coraza in detect-only, events per site, per-pad uitzonderingen, Block | Feature |
| **P3** | CrowdSec: container, bouncer, simulatie, allowlist, UI | Feature |

## 10. Tests en verificatie

- **Unit:** profielresolutie (vloer, lock, uitzondering), Caddy-render (golden files per profiel), `host_actions`-toestandsmachine, validatie van allowlist en wachtwoord.
- **Regressie P0:** een app- en een databasecontainer starten met de nieuwe `HostConfig` en gezond blijven voor Postgres, Valkey, een Next.js-app en een Dockerfile-app. Fail als `no-new-privileges` of `cap_drop` een entrypoint breekt.
- **E2E (Docker-host):** SSH-drop-in op Debian 12 en Ubuntu 24.04 in een VM of systemd-container, inclusief `50-cloud-init.conf` met wachtwoord aan. Revert-timer: panel stoppen na `Apply()` en controleren dat de actie terugdraait.
- **Chaos:** proxywissel met een kapotte nieuwe build (oude moet blijven), CrowdSec uninstall met actieve regels, installatie die halverwege faalt.
- **WAF:** fixtures met bekende false positives (Next.js API-routes, JSON-bodies) om te controleren dat detect-only niets blokkeert.
- **Performance:** vóór en na P0 `curl -H "Accept-Encoding: zstd,gzip"` en responsegroottes op een Next.js-fixture en een static site.
- **Bevestigen vóór bouwen:** alles wat in dit document **(niet bevestigd)** is.

## 11. Risico's en open punten

1. **Panel als root.** Het panel draait al als root met Docker-toegang. Een helper met privilege-scheiding vraagt een eigen ADR: panel als eigen gebruiker, Docker via een socket-proxy, een kleine root-helper voor de weinige host-acties. Besluit nodig of dat vóór of na P1 komt. Tot die tijd noemt de UI het "Host actions" en belooft geen scheiding.
2. **IP's in security events.** Strijdig met "geen IP-opslag" uit de analytics-briefing. Voorstel: 7 dagen bewaren, instelbaar tot 30, in de privacy-documentatie vermelden. AVG-impact beoordelen voor de betaalde EU-cloud.
3. **Distro-ondersteuning.** `unattended-upgrades` en apt-acties alleen op Debian en Ubuntu (officieel ondersteund). Op andere distro's toont de check "handmatig".
4. **CrowdSec-bouncer.** Twee routes: host-package (vraagt een apt-bron en dus een extra supply-chain-stap) of een eigen container met host-netwerk en `NET_ADMIN`. Voorkeur: container, zodat Dockyrd geen externe apt-bron toevoegt. Uitzoeken of daar een betrouwbaar image voor bestaat.
5. **Eigen proxybuild.** Onderhoud en supply chain: wie bouwt, signeert en publiceert, hoe snel volgen we Caddy-patches, en wat is de beeldgrootte? Voorstel: één extra image per release, gesigneerd, met een CI-test die de volledige gerenderde config valideert.
6. **Firewall en Docker.** `firewall.ensure` mag Docker niet breken. We zetten regels in de `DOCKER-USER`-keten of een eigen tabel met eigen prioriteit. Gepubliceerde poorten blijven een aparte check (§2, punt 8).
7. **IPv6.** Firewall, allowlist en trusted proxies moeten v6 aankunnen.
8. **Detectie van Cloudflare** is een hint, geen zekerheid. De gebruiker bevestigt altijd.
9. **ClamAV** is bewust overgeslagen. Sites met uploads kunnen later een optionele add-on krijgen.
10. **Docs en markt.** Zodra P0 staat, kloppen de docs weer. Een "Security"-pagina op de marketingsite (en de betaalde cloud met Baseline als standaard) is een logische vervolgstap.

## Bijlage A: Referenties

- CrowdSec, simulatiemodus: https://docs.crowdsec.net/docs/log_processor/scenarios/simulation
- Coraza voor Caddy: https://github.com/corazawaf/coraza-caddy/wiki
- Rate limiting voor Caddy: https://www.github.com/mholt/caddy-ratelimit
- Docker en UFW, gepubliceerde poorten: https://dev.to/alanwest/why-docker-bypasses-ufw-and-how-to-actually-lock-it-down-26ep
- **Niet gecontroleerd in deze sessie:** Caddy-documentatie voor `trusted_proxies`, `client_ip_headers`, timeouts en de JSON-vorm van `headers`. Controleren tegen de gepinde Caddy-versie.

## Bijlage B: Host action, toestanden

```
planned ──► applied ──► verified ──► awaiting_confirm ──► confirmed
   │           │            │               │
   │           │            └─► reverted    └─► reverted (timer of "Revert now")
   │           └─► failed (verify mislukt) ──► reverted
   └──► cancelled
```

Alleen `internal/hostops` mag overgangen maken, in een DB-transactie. Elke overgang schrijft een audit-regel. Bij opstart worden `awaiting_confirm`-acties met een verstreken deadline alsnog teruggedraaid.
