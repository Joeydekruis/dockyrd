# Dockyard: technische briefing v0.1 (MVP)

> Self-hosted, open source deploy-platform op één server. Eén Go-binary, één installatieregel, Next.js control plane.
> Status: bouwbriefing voor de eerste publieke repo · Datum: 2 november 2026 · Eigenaar: Joey de Kruis
> Bron design: Claude Design-canvas "Dockyard Dashboard" (19 schermen)

---

## 0. Samenvatting in tien regels

1. **Eén Go-binary** (`dockyard`) draait als systemd-service op de host: API, scheduler, builds, reconciler, metrics en de **ingebedde Next.js UI** (static export via `go:embed`).
2. **Docker Engine** is de runtime. Alles wat Dockyard maakt, krijgt Docker-labels; die labels zijn de "actual state", SQLite is de "desired state".
3. **Caddy** draait als aparte container (`dy-proxy`) en wordt geconfigureerd via zijn admin-API op een unix-socket. Automatisch HTTPS, zero-downtime switches, JSON access logs voor analytics. **Een herstart of update van Dockyard raakt nooit lopend verkeer.**
4. **SQLite** (CGO-vrij, WAL) als enige database van het control plane. Geen Postgres/Redis voor Dockyard zelf.
5. **Builds**: Railpack (auto-detect, o.a. Next.js) via BuildKit, of Dockerfile, of een kant-en-klaar image.
6. **Static sites** hebben geen container nodig: build-output gaat naar disk, Caddy serveert het, rollback is een atomische symlink-switch.
7. **Databases**: Postgres en Redis-compatible (Valkey) uit officiële images, met gegenereerde credentials en verwijsbare variabelen (`${{db.DATABASE_URL}}`).
8. **Object storage**: Garage (single-node) als ingebouwde S3-server. MinIO valt af (community-editie is gearchiveerd, zie §12).
9. **Gebruikersbeheer** vanaf dag één: owner/admin/member/viewer, sessies, argon2id, TOTP-2FA, invites, optioneel GitHub-login, audit log.
10. **Installatie**: `curl -fsSL https://get.dockyard.dev | sudo sh` → Docker check/installatie, binary met handtekeningcontrole, systemd, eenmalige setup-token. Updates via ondertekende GitHub releases met automatische rollback.

Licentie: **Apache-2.0** (zie §16).

---

## 1. Doel en scope

### 1.1 Einddoel MVP

Een developer installeert Dockyard op een verse VPS (Ubuntu 24.04 / Debian 12+) en kan binnen 10 minuten:

- een **Next.js-app** uit een GitHub-repo live zetten op een eigen domein met HTTPS,
- er een **Postgres-database** naast zetten die de app via een interne verbinding gebruikt,
- optioneel een **S3-bucket** (Garage) koppelen voor uploads,
- een **statische website** (Astro, Vite, Hugo, Next export…) live zetten,
- logs, deploys en servergebruik volgen in het dashboard,
- collega's uitnodigen met beperkte rechten.

### 1.2 Wat expliciet níet in de MVP zit

| Later (v0.2+) | Waarom later |
|---|---|
| Geplande database-backups naar S3 + restore | Design toont ze al; datamodel en jobs-runner zijn erop voorbereid (§15) |
| Multi-server (remote servers via agent) | Schema bevat `server_id` vanaf dag één, maar MVP = lokale server |
| Preview deployments per PR | Vereist GitHub App-events die we al ontvangen; UI later |
| Web-terminal (`docker exec`) | WebSocket + security review |
| Notificaties (e-mail, Slack, Discord, webhooks) | Events-tabel bestaat al; alleen kanalen toevoegen |
| Docker Compose-stacks / one-click templates | Kernmodel eerst stabiel |
| MySQL/MariaDB/MongoDB | Zelfde database-driver-interface, extra implementaties |
| API-tokens voor CI/CLI | Tabel is voorzien; UI later |

### 1.3 Niet-functionele eisen (hard)

| Eis | Meetbare invulling |
|---|---|
| Lightweight | `dockyard` idle < 60 MB RSS, < 1 % CPU op 1 vCPU. Caddy ~30 MB. Totale platform-overhead < 150 MB RAM excl. builds |
| Stabiel | Control plane kan crashen/updaten zonder dat één workload of request faalt. Alle operaties idempotent. Reconciler herstelt drift binnen 60 s |
| Betrouwbaar | Geen deploy laat een half-kapotte situatie achter: nieuwe container pas live na healthcheck, anders blijft de oude draaien |
| Simpel te installeren | Eén commando, < 3 min op een verse VPS, geen vragen in de terminal |
| Minimale requirements | 1 vCPU / 1 GB RAM werkt (installer maakt 2 GB swap aan bij < 2 GB RAM); 2 vCPU / 4 GB aanbevolen voor Next.js-builds |
| Platform | Linux x86_64 en arm64; Ubuntu 22.04/24.04, Debian 12/13 officieel ondersteund |

---

## 2. Lessen uit Coolify en Dokploy

| | Coolify v4 | Dokploy | **Dockyard-keuze** |
|---|---|---|---|
| Stack control plane | PHP/Laravel + Livewire, Postgres, Redis, soketi (meerdere containers) | Next.js-app (UI en backend in één), Postgres, Redis, Traefik | **Eén Go-binary + SQLite**. Geen extra infrastructuur voor het platform zelf |
| Proxy | Traefik (default) of Caddy | Traefik | **Caddy** via admin-API (JSON), los proces |
| Orchestratie | Docker via SSH, ook naar de eigen host | Docker Swarm | **Plain Docker Engine** met labels. Swarm is overkill voor één server en maakt debugging lastiger |
| Builder | Nixpacks, Dockerfile, Compose | Nixpacks, Railpack, Dockerfile, buildpacks | **Railpack + Dockerfile + image**. Nixpacks is bij Railway opgevolgd door Railpack |
| Installatie | `curl … \| bash` | `curl … \| sh` | Zelfde UX, maar met checksum + signatuurcontrole en een setup-token |

**Wat we overnemen:** het mentale model project → omgeving → service; GitHub App-koppeling via het manifest-flow; automatische HTTPS; `sslip.io`-achtige platformdomeinen; een install-script dat Docker zelf installeert.

**Wat we bewust anders doen:**

- **Footprint.** Coolify en Dokploy draaien zelf al 3–5 containers met Postgres en Redis. Dockyard: één proces + Caddy.
- **Geen SSH naar de eigen host.** Coolify beheert ook de lokale server via SSH; dat is een bron van fragiliteit (key-beheer, connection-errors). Dockyard praat met de lokale Docker-socket. Multi-server komt later via een eigen agent, niet via SSH-shellcommando's.
- **Gescheiden data plane.** Het panel kan herstarten zonder dat de proxy of workloads iets merken (ook letterlijk een event in het design: "The panel restarted. No services were interrupted").
- **Setup-token.** Een verse installatie mag niet door de eerste de beste bezoeker van `http://IP:8000` geclaimd worden.
- **Kleine, consistente UI** in plaats van een feature-wand.

---

## 3. Architectuur

### 3.1 Overzicht

```
                         Internet  :80 / :443
                               │
                    ┌──────────▼──────────┐
                    │  dy-proxy (Caddy)    │  container, restart=always
                    │  - TLS / ACME        │  netwerk: dy-edge (+ host poorten 80/443)
                    │  - routes apps/sites │  volumes: caddy-data, /var/lib/dockyard/sites:ro
                    │  - JSON access log   │  admin-API: unix socket
                    └───┬─────────┬────────┘
                        │         │ file_server
          reverse_proxy │         └──────────► /var/lib/dockyard/sites/<site>/current
                        │
     ┌──────────────────▼─────────────────────────────────────────┐
     │  Docker Engine                                              │
     │                                                             │
     │  dy-edge network ─── app containers met publiek domein      │
     │  dy-p-<project>  ─── app + db + redis van één project       │
     │  dy-garage (S3)  ─── op dy-edge + projectnetwerken met buckets│
     │  dy-buildkit     ─── BuildKit daemon voor Railpack-builds   │
     └──────────────────▲─────────────────────────────────────────┘
                        │ /var/run/docker.sock
     ┌──────────────────┴──────────────────────────────┐
     │  dockyard (systemd, Go)                          │
     │  ├─ HTTP: /api/v1 (JSON), /api/v1/stream (SSE),  │
     │  │        /hooks/github, UI (embedded export)    │
     │  ├─ jobs: deploy, build, prune, cert-check …     │
     │  ├─ reconciler (30 s)                            │
     │  ├─ collectors: host, containers, caddy logs     │
     │  ├─ updater (GitHub releases)                    │
     │  └─ SQLite /var/lib/dockyard/dockyard.db         │
     └──────────────────────────────────────────────────┘
```

Het dashboard zelf wordt ook via Caddy geserveerd zodra er een domein is ingesteld (`dockyard.example.com` → `reverse_proxy host.docker.internal:8000` / host-gateway). Vóór de setup is het bereikbaar op `http://<ip>:8000`; na het instellen van een domein kan poort 8000 dicht (optie in Settings).

### 3.2 Waarom Dockyard op de host draait en niet in een container

- Eén binary + systemd = minste bewegende delen, journald-logging, simpele self-update.
- Directe toegang tot `/proc` voor host-metrics en tot `/var/lib/dockyard/sites` zonder bind-mount-gedoe.
- Herstart van Dockyard ≠ herstart van Docker-objecten. Caddy, databases en apps hebben `restart: unless-stopped`.

### 3.3 Desired state vs. actual state

- **Desired** = SQLite (wat de gebruiker heeft ingesteld + welke deployment "live" moet zijn).
- **Actual** = Docker, uitgelezen via labels. Elk object dat Dockyard maakt draagt minimaal:

```
dev.dockyard.managed=true
dev.dockyard.server=<server_id>
dev.dockyard.project=<project_id>
dev.dockyard.service=<service_id>
dev.dockyard.kind=app|database|redis|bucket|system
dev.dockyard.deployment=<deployment_id>   (apps)
```

- De **reconciler** (elke 30 s + bij opstarten + na elke job) vergelijkt beide en:
  - start containers die horen te draaien maar niet draaien (na crash-loops: backoff + event "needs attention"),
  - ruimt verweesde containers van mislukte/afgebroken deploys op,
  - pusht de volledige Caddy-config opnieuw als de hash afwijkt,
  - markeert deployments die tijdens een herstart in `building/starting` bleven hangen als `failed (interrupted)`.

Regel: **de reconciler doet nooit iets destructiefs met volumes.** Volumes worden alleen verwijderd bij expliciet "delete service" met bevestiging.

### 3.4 Namen en netwerken

| Object | Naam |
|---|---|
| Edge-netwerk | `dy-edge` (Caddy + alles met een publiek domein) |
| Projectnetwerk | `dy-p-<project_short>` (isolatie tussen projecten) |
| App-container | `dy-<service_slug>-<deploy_short>` met netwerk-alias `<service_slug>` binnen het projectnetwerk |
| Database-container | `dy-<service_slug>` (vaste naam, geen blue/green) |
| Volume | `dy-<service_short>-data` |
| Image | `dockyard/<service_short>:<deploy_short>` |
| Systeemcontainers | `dy-proxy`, `dy-buildkit`, `dy-garage` |

Interne hostnamen voor gebruikers zijn **leesbaar**: een app in project "Client X" bereikt de database op `db:5432` (de service-slug), precies zoals het design toont (`postgres · internal:5432`).

### 3.5 Proxy: Caddy via admin-API

- Dockyard rendert de **volledige** Caddy JSON-config uit SQLite en doet `POST /load` op de admin-socket (`/var/lib/dockyard/caddy/admin.sock`, gemount in de container). Caddy valideert atomisch; bij een ongeldige config blijft de oude actief en logt Dockyard een event.
- Geen Caddyfile, geen partial PATCHes: altijd volledige config → idempotent en makkelijk te debuggen (`dockyard proxy dump`).
- Certificaten in volume `dy-proxy-data`; ACME HTTP-01 standaard, e-mail uit de instance-settings.
- Routes per domein:
  - **app**: `reverse_proxy <container-naam>:<poort>` op `dy-edge`, met `health_uri` optioneel en `lb_try_duration 5s` zodat een switch geen 502 geeft.
  - **static site**: `file_server` met `root /srv/sites/<site_id>/current`, optioneel SPA-fallback (`try_files {path} /index.html`), cache-headers voor gehashte assets, `encode zstd gzip`.
  - **redirects**: `www` ↔ apex, http → https (Caddy default).
- Access logs: JSON naar `/var/lib/dockyard/caddy/access.log` met rotatie (roll_size 50 MB, roll_keep 3). Dockyard tailt dit bestand voor analytics (§11.3).
- **Platformdomein**: als geen wildcard-domein is ingesteld, krijgt elke service `<slug>-<rand4>.<ip-met-streepjes>.sslip.io` → direct werkend HTTPS zonder DNS-werk. Met een wildcard-domein (bijv. `*.apps.example.com`) wordt het `<slug>-<rand4>.apps.example.com`.

### 3.6 Zero-downtime app-deploy (blue/green)

```
1. build image dockyard/<svc>:<deploy>
2. start nieuwe container  dy-<svc>-<deploy>  (netwerken: dy-p-<proj> [+ dy-edge])
3. healthcheck: TCP-connect op poort, daarna optioneel HTTP GET healthPath → 2xx/3xx
   timeout instelbaar (default 60 s), interval 2 s
4a. healthy  → Caddy-upstream switchen naar nieuwe container (POST /load)
             → netwerk-alias <svc> verplaatsen naar nieuwe container
             → oude container: SIGTERM, grace 30 s (instelbaar), daarna verwijderen
             → deployment = live, vorige = superseded
4b. unhealthy → nieuwe container stoppen + logs bewaren → deployment = failed
             → oude blijft live, event "deploy failed" + link naar build/run-log
```

Apps zonder publiek domein (workers) volgen hetzelfde pad zonder de Caddy-stap; optie "stop old before start new" voor workers die geen dubbele instanties mogen hebben (bijv. cron-achtige jobs).

Rollback = "redeploy" van een eerdere deployment: het image bestaat nog (we bewaren de laatste N=5 images per service), dus geen nieuwe build nodig.

---

## 4. Tech-stack (met motivatie)

### 4.1 Backend (Go)

| Onderdeel | Keuze | Waarom |
|---|---|---|
| Taal | Go (nieuwste stabiele, ≥ 1.25) | Statische binary, laag geheugengebruik, Docker-ecosysteem is Go |
| HTTP | `net/http` met de method+pattern `ServeMux` | Geen framework nodig |
| API-schema | **Huma v2** (MIT) bovenop `net/http` | Genereert OpenAPI 3.1 uit Go-types → getypeerde TS-client voor de UI. Validatie gratis |
| Database | SQLite via **`modernc.org/sqlite`** (pure Go, geen CGO) | Cross-compile naar arm64 zonder toolchain-gedoe. WAL, `busy_timeout=5000`, `foreign_keys=on`, één writer-connectie + read-pool |
| Queries | **sqlc** (gegenereerde type-safe Go) | Geen ORM, wel compile-time checks |
| Migraties | **goose** met embedded SQL-files | Simpel, forward-only in productie |
| Docker | **`github.com/moby/moby/client`** + `github.com/moby/moby/api` | Sinds Docker v29 is de SDK opgesplitst in deze modules; `github.com/docker/docker` krijgt geen v29-releases |
| Builds | `docker buildx build` als subprocess + **Railpack**-CLI (`railpack prepare`) | Zelfde pad als Railway in productie aanbeveelt; streambare logs; BuildKit-cache |
| Proxy-config | Eigen Go-structs voor Caddy JSON (alleen de gebruikte modules) | Geen afhankelijkheid op Caddy als library |
| Host-metrics | `/proc` en `/sys` direct lezen (kleine eigen package) of `gopsutil/v4` | Weinig code, geen CGO |
| Wachtwoorden | `golang.org/x/crypto/argon2` (argon2id) | Moderne standaard |
| TOTP | `github.com/pquerna/otp` | Bewezen, klein |
| Secrets at rest | AES-256-GCM met master-key in `/etc/dockyard/secret.key` (0600) | Env-vars, DB-wachtwoorden, S3-keys, GitHub App-key versleuteld in SQLite |
| Logging | `log/slog` (JSON) → journald | Standaard, structured |
| CLI | `cobra` of eigen minimal `flag`-subcommands | Klein houden |
| Self-update | Eigen updater: GitHub Releases API + `sha256` + **cosign/sigstore**-verificatie | Zie §13 |
| Release-builds | **GoReleaser** | linux/amd64 + arm64, checksums, SBOM, signatures |

### 4.2 Frontend (control plane)

| Onderdeel | Keuze | Toelichting |
|---|---|---|
| Framework | **Next.js (App Router, nieuwste stabiele) met `output: 'export'`** | Puur statische export, ingebed in de Go-binary. Geen Node.js op de server |
| UI-componenten | **shadcn/ui met Base UI als primitive-laag** | shadcn ondersteunt sinds begin 2026 Base UI naast Radix. Bij `init` Base UI kiezen; componenten in `web/components/ui` zijn eigen code |
| Verbod op Radix | CI-check: `pnpm why '@radix-ui/*'` moet leeg zijn (zie §14.4) | Let op: `cmdk` en `vaul` trekken Radix binnen → **niet gebruiken** |
| Styling | **Tailwind CSS v4** met design-tokens als CSS-variabelen (§9) | Light mode only |
| Charts | **Apache ECharts** via `echarts/core` (tree-shaken: Line, Bar, Grid, Tooltip, DataZoom) + eigen `<EChart>`-wrapper | Geen `echarts-for-react` nodig; eigen wrapper is ~40 regels met `ResizeObserver` |
| Data | **TanStack Query** + **openapi-fetch** met types uit `openapi-typescript` | Types gegenereerd uit de Huma-spec, nooit handmatig |
| Realtime | `EventSource` (SSE) hook die TanStack Query-caches invalideert / logs streamt | |
| Formulieren | `react-hook-form` + `zod` | |
| Command palette (⌘K) | Base UI `Dialog` + Base UI `Autocomplete`/`Combobox` | Vervangt cmdk |
| Toasts | `sonner` (geen Radix-dependency) | |
| Iconen | `lucide-react`, stroke-width 1.8 (matcht design) | |
| Fonts | `geist`-package (Geist + Geist Mono), **self-hosted** | Geen Google Fonts-calls vanaf een self-hosted panel (privacy/offline) |
| Lint/format | Biome | Eén tool, snel |
| Package manager | pnpm | |

**Static export en dynamische routes.** `output: 'export'` ondersteunt geen runtime-routes. Oplossing (bewezen patroon):

1. Dynamische segmenten krijgen precies één prerender: `generateStaticParams() → [{ id: '_' }]`, `dynamicParams = false`.
2. Pagina's zijn client components en lezen het echte ID via een helper `useRouteParam('id')` die `usePathname()` parseert (niet `useParams()`, dat bij hydration `_` teruggeeft).
3. De Go-server heeft een **route-tabel** die `/apps/abc123/logs` serveert met `apps/_/logs/index.html` en de bijbehorende RSC-payload (`…/index.txt`) met de `_`-variant. Daarmee werken zowel harde refreshes als client-side navigatie via `next/link`.
4. Alle overige onbekende paden → `404.html`.

Als dit patroon in de praktijk wringt, is de fallback Vite + TanStack Router met exact dezelfde componenten; de componentcode hoeft daarvoor niet te veranderen. Neem dit besluit uiterlijk aan het eind van milestone M1.

### 4.3 Systeemcomponenten (containers, versies gepind per Dockyard-release)

| Container | Image | Rol |
|---|---|---|
| `dy-proxy` | `caddy:2-alpine` (gepinde digest) | Proxy, TLS, static files |
| `dy-buildkit` | `moby/buildkit` (gepind) | BuildKit-daemon voor Railpack-builds en cache |
| `dy-garage` | `dxflrs/garage` v2.x (gepind) | S3-server, alleen na "Install Garage" |

Gebruikersservices: officiële images `postgres:<major>`, `valkey/valkey:<major>` (Redis-compatible, BSD), later `mysql`, `mariadb`.

---

## 5. Repo-structuur

Eén repo, één Go-module, platte structuur. Alles wat niet publiek API is, zit in `internal/`.

```
dockyard/
├── cmd/
│   └── dockyard/
│       └── main.go              # CLI-entrypoint: serve, init, admin, update, doctor, version
├── internal/
│   ├── api/                     # Huma-handlers per resource, middleware (auth, csrf, ratelimit)
│   ├── auth/                    # users, sessions, argon2id, totp, invites, github oauth, rbac
│   ├── build/                   # railpack, dockerfile, static-site builder, log streaming
│   ├── config/                  # env/flags, paden, defaults
│   ├── crypto/                  # AES-GCM secrets, master key
│   ├── db/
│   │   ├── migrations/          # 0001_init.sql …  (goose, embedded)
│   │   ├── queries/             # *.sql voor sqlc
│   │   └── sqlc/                # gegenereerde code (gecommit)
│   ├── deploy/                  # deploy state machine, blue/green, rollback
│   ├── docker/                  # dunne wrapper rond moby client + labels + naming
│   ├── events/                  # event log + pub/sub voor SSE
│   ├── github/                  # GitHub App manifest-flow, webhooks, clone tokens
│   ├── jobs/                    # SQLite-backed queue, worker pools, retries
│   ├── metrics/                 # host + container collectors, rollups, caddy log tailer
│   ├── proxy/                   # Caddy JSON-model, renderer, admin-socket client
│   ├── reconcile/               # desired vs actual
│   ├── services/                # app, database (postgres, redis), site, bucket (garage)
│   ├── storage/                 # garage admin-API client
│   ├── updater/                 # release check, verificatie, swap, rollback
│   └── server/                  # http.Server wiring, SPA route table, static assets
├── web/                         # Next.js control plane
│   ├── app/                     # routes (zie §10)
│   ├── components/
│   │   ├── ui/                  # shadcn (Base UI) componenten — eigen code
│   │   └── dockyard/            # AppShell, Sidebar, StatusBadge, StatCard, ResourceRow, LogViewer, EChart …
│   ├── lib/                     # api client (gegenereerd), sse, format, route-param helper
│   ├── styles/                  # globals.css met tokens
│   ├── embed.go                 # package web: //go:embed all:out
│   ├── next.config.ts
│   └── package.json
├── scripts/
│   ├── install.sh               # get.dockyard.dev
│   ├── uninstall.sh
│   └── dev-vm.sh                # Multipass/Lima VM voor lokaal testen
├── packaging/
│   └── dockyard.service         # systemd unit
├── docs/                        # Markdown-docs (later een docs-site)
│   ├── architecture.md
│   ├── install.md
│   └── adr/                     # Architecture Decision Records (0001-sqlite.md, …)
├── e2e/                         # Playwright + shell-based install-tests
├── .github/
│   ├── workflows/ci.yml  release.yml  e2e.yml
│   ├── ISSUE_TEMPLATE/  PULL_REQUEST_TEMPLATE.md
├── .goreleaser.yaml
├── Makefile
├── go.mod / go.sum
├── LICENSE                      # Apache-2.0
├── NOTICE
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md                  # responsible disclosure
└── CODE_OF_CONDUCT.md           # Contributor Covenant
```

**Embedding**: `web/embed.go` is een Go-package `web` met `//go:embed all:out` en een `FS()`-functie. `make build` draait eerst `pnpm --dir web build` (export naar `web/out`) en daarna `go build`. In CI is `web/out` een build-artefact; niet committen.

**Makefile-targets**: `make dev`, `make build`, `make test`, `make lint`, `make generate` (sqlc + openapi + TS-types), `make e2e`, `make release-snapshot`.

---

## 6. Installatie en server-lifecycle

### 6.1 Eén commando

```bash
curl -fsSL https://get.dockyard.dev | sudo sh
```

`install.sh` (POSIX sh, idempotent, opnieuw draaien = repareren/upgraden):

1. **Preflight**: root, OS (Ubuntu 22.04/24.04, Debian 12/13; anders waarschuwing + `--force`), arch (amd64/arm64), vrije poorten 80/443/8000, ≥ 10 GB vrije disk, systemd aanwezig.
2. **Docker**: als niet aanwezig → officiële apt-repo van Docker installeren (`docker-ce`, `containerd.io`, `docker-buildx-plugin`). Als wel aanwezig: versie ≥ 27 vereisen.
3. **Docker daemon-tuning** (merge, niet overschrijven, van `/etc/docker/daemon.json`): `log-driver: json-file`, `log-opts: {max-size: 10m, max-file: 3}`, `live-restore: true` (containers blijven draaien tijdens dockerd-restart). Alleen herstarten als er iets wijzigt en er nog geen workloads draaien; anders melden.
4. **Swap**: bij < 2 GB RAM en geen swap → 2 GB swapfile + `vm.swappiness=10`.
5. **Binary**: download `dockyard_<ver>_linux_<arch>.tar.gz` + `checksums.txt` + signature van GitHub Releases. Verifieer sha256; verifieer cosign-signature als `cosign` beschikbaar is (anders checksum + HTTPS, en de binary verifieert zichzelf bij de eerste update). Installeer naar `/usr/local/bin/dockyard`.
6. **Paden**: `/etc/dockyard/` (0700), `/var/lib/dockyard/` (data), `/var/log/dockyard/` niet nodig (journald).
7. **`dockyard init`**: maakt master-key, SQLite + migraties, netwerk `dy-edge`, start `dy-proxy` en `dy-buildkit`, genereert een **eenmalige setup-token**.
8. **systemd**: `packaging/dockyard.service` installeren, `enable --now`.
9. **Firewall**: als `ufw` actief is → 22, 80, 443, 8000 openzetten en melden. Waarschuwing in docs: Docker-gepubliceerde poorten omzeilen ufw; daarom publiceert Dockyard standaard **geen** databasepoorten.
10. **Output**:

```
✓ Dockyard v0.1.0 is running.

  Open:   http://203.0.113.10:8000/setup?token=8f3k-29xq-a71m
  Token is valid for 60 minutes and can be used once.
  Lost it?  sudo dockyard setup-token
```

Flags: `--version v0.1.2`, `--channel stable|beta`, `--no-docker-install`, `--dashboard-domain dockyard.example.com`, `--force`.

### 6.2 systemd-unit

```ini
[Unit]
Description=Dockyard
After=network-online.target docker.service
Requires=docker.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/dockyard serve
Restart=always
RestartSec=2
EnvironmentFile=-/etc/dockyard/dockyard.env
LimitNOFILE=65536
# Hardening (draait als root i.v.m. docker.sock — docker-groep is toch root-equivalent)
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=true
PrivateTmp=true
ReadWritePaths=/var/lib/dockyard /etc/dockyard /usr/local/bin

[Install]
WantedBy=multi-user.target
```

`Type=notify` + `sd_notify(READY=1)` pas na: DB open, migraties ok, Docker bereikbaar, proxy-config gepusht. Watchdog (`WatchdogSec=30s`) met heartbeat vanuit de main loop.

### 6.3 Onboarding (eerste keer)

Setup-wizard in de UI (scherm nog te ontwerpen, §17):

1. **Owner-account**: naam, e-mail, wachtwoord (min. 12 tekens, zxcvbn-score ≥ 3).
2. **Instance**: dashboard-domein (optioneel; anders blijft het op `IP:8000`), ACME-e-mail, wildcard-domein voor apps (optioneel → anders sslip.io).
3. **GitHub** (optioneel, kan later): "Create GitHub App" via manifest-flow.
4. Klaar → Dashboard met lege-staat-CTA's "Create application", "Create database", "Create static site".

### 6.4 CLI

| Commando | Doel |
|---|---|
| `dockyard serve` | Start de server (systemd) |
| `dockyard init` | Eerste setup (idempotent) |
| `dockyard setup-token` | Nieuwe eenmalige token (alleen zolang er geen owner is) |
| `dockyard admin reset-password <email>` | Nood-reset vanaf de server |
| `dockyard admin disable-2fa <email>` | Nood-uitschakelen 2FA |
| `dockyard update [--version]` | Handmatige update |
| `dockyard doctor` | Checkt Docker, poorten, DNS van domeinen, disk, proxy-config, klok (ACME), versies; geeft concrete fixes |
| `dockyard proxy dump` | Huidige Caddy-config tonen |
| `dockyard backup` | Snapshot van control-plane DB (zie §6.5) |
| `dockyard version` | Versie, commit, builddatum |

### 6.5 Control-plane self-backup (wél in MVP)

Klein en essentieel: dagelijks `VACUUM INTO /var/lib/dockyard/backups/dockyard-YYYYMMDD.db`, 7 bewaren, plus vóór elke update en elke migratie. Dit staat los van de (latere) database-backups van gebruikers.

### 6.6 Uninstall

`scripts/uninstall.sh`: stopt service, verwijdert systeemcontainers; **vraagt** of gebruikerscontainers/volumes/data weg moeten (default: nee, alleen platform verwijderen).

---

## 7. Domeinmodel en database

### 7.1 Begrippen

- **Server**: de machine. MVP heeft er precies één (`local`), maar elke resource heeft `server_id`.
- **Project**: groepering (bijv. "Client X"). Heeft een Docker-netwerk.
- **Environment**: binnen een project (`production` standaard; `staging` later). Design toont env-chips per project.
- **Service**: iets dat draait of geserveerd wordt. `kind` ∈ `app`, `static_site`, `postgres`, `redis`, `bucket`.
- **Deployment**: één poging om een app of static site naar een nieuwe versie te brengen.
- **Domain**: hostnaam gekoppeld aan een service.
- **Variable**: env-var op service-niveau, plain of secret, build-time en/of runtime.

### 7.2 Schema (eerste migratie, ingekort)

```sql
-- identiteit & toegang
CREATE TABLE users (
  id TEXT PRIMARY KEY,                 -- ULID
  email TEXT NOT NULL UNIQUE COLLATE NOCASE,
  name TEXT NOT NULL,
  password_hash TEXT,                  -- NULL bij alleen-GitHub-login
  role TEXT NOT NULL CHECK (role IN ('owner','admin','member','viewer')),
  totp_secret_enc BLOB,                -- AES-GCM
  totp_enabled_at INTEGER,
  avatar_url TEXT,
  disabled_at INTEGER,
  created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL
);
CREATE TABLE sessions (
  id_hash BLOB PRIMARY KEY,            -- sha256 van token; token alleen in cookie
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_agent TEXT, ip TEXT, location TEXT,
  mfa_passed INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL, last_seen_at INTEGER NOT NULL, expires_at INTEGER NOT NULL
);
CREATE TABLE recovery_codes (user_id TEXT, code_hash BLOB, used_at INTEGER);
CREATE TABLE invites (id TEXT PRIMARY KEY, email TEXT, role TEXT, token_hash BLOB, invited_by TEXT, expires_at INTEGER, accepted_at INTEGER);
CREATE TABLE password_resets (token_hash BLOB PRIMARY KEY, user_id TEXT, expires_at INTEGER, used_at INTEGER);
CREATE TABLE oauth_identities (provider TEXT, provider_user_id TEXT, user_id TEXT, PRIMARY KEY (provider, provider_user_id));
CREATE TABLE project_members (project_id TEXT, user_id TEXT, role TEXT, PRIMARY KEY (project_id, user_id)); -- v0.2: per-project rechten

-- infrastructuur
CREATE TABLE servers (id TEXT PRIMARY KEY, name TEXT, kind TEXT DEFAULT 'local', public_ipv4 TEXT, public_ipv6 TEXT, created_at INTEGER);
CREATE TABLE projects (id TEXT PRIMARY KEY, server_id TEXT NOT NULL, name TEXT NOT NULL, slug TEXT NOT NULL UNIQUE, description TEXT, created_at INTEGER, updated_at INTEGER);
CREATE TABLE environments (id TEXT PRIMARY KEY, project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE, name TEXT NOT NULL, UNIQUE(project_id, name));

CREATE TABLE services (
  id TEXT PRIMARY KEY,
  environment_id TEXT NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  kind TEXT NOT NULL CHECK (kind IN ('app','static_site','postgres','redis','bucket')),
  name TEXT NOT NULL, slug TEXT NOT NULL,                 -- slug = interne hostnaam
  desired_state TEXT NOT NULL DEFAULT 'running',         -- running | stopped
  config JSON NOT NULL,                                  -- kind-specifiek, gevalideerd in Go
  live_deployment_id TEXT,
  created_at INTEGER, updated_at INTEGER,
  UNIQUE(environment_id, slug)
);

CREATE TABLE deployments (
  id TEXT PRIMARY KEY, service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  status TEXT NOT NULL,      -- queued|waiting|cloning|building|starting|healthchecking|live|failed|cancelled|superseded
  trigger TEXT NOT NULL,     -- push|manual|redeploy|rollback|config
  triggered_by TEXT,         -- user_id of 'github'
  commit_sha TEXT, commit_message TEXT, commit_author TEXT, branch TEXT,
  image_ref TEXT, release_path TEXT,                     -- app resp. static site
  error TEXT, build_log_path TEXT,
  queued_at INTEGER, started_at INTEGER, finished_at INTEGER
);

CREATE TABLE domains (
  id TEXT PRIMARY KEY, service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  hostname TEXT NOT NULL UNIQUE COLLATE NOCASE,
  is_platform INTEGER NOT NULL DEFAULT 0, is_primary INTEGER NOT NULL DEFAULT 0,
  redirect_to TEXT,                                       -- bv. www → apex
  dns_status TEXT, dns_checked_at INTEGER,                -- ok|pending|wrong_target
  cert_status TEXT, cert_expires_at INTEGER
);

CREATE TABLE variables (
  id TEXT PRIMARY KEY, service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
  key TEXT NOT NULL, value_enc BLOB NOT NULL,
  is_secret INTEGER NOT NULL DEFAULT 0,
  at_build INTEGER NOT NULL DEFAULT 1, at_runtime INTEGER NOT NULL DEFAULT 1,
  UNIQUE(service_id, key)
);

CREATE TABLE git_sources (id TEXT PRIMARY KEY, kind TEXT, github_app_id INTEGER, installation_id INTEGER, private_key_enc BLOB, webhook_secret_enc BLOB, client_id TEXT, client_secret_enc BLOB, slug TEXT, created_at INTEGER);

-- werk & observability
CREATE TABLE jobs (id TEXT PRIMARY KEY, kind TEXT, payload JSON, status TEXT, attempts INTEGER DEFAULT 0, run_after INTEGER, locked_until INTEGER, last_error TEXT, created_at INTEGER);
CREATE TABLE events (id TEXT PRIMARY KEY, level TEXT, category TEXT, title TEXT, detail TEXT, project_id TEXT, service_id TEXT, deployment_id TEXT, acknowledged_at INTEGER, created_at INTEGER);
CREATE TABLE audit_log (id TEXT PRIMARY KEY, user_id TEXT, action TEXT, target_kind TEXT, target_id TEXT, meta JSON, ip TEXT, created_at INTEGER);
CREATE TABLE settings (key TEXT PRIMARY KEY, value JSON);

-- metrics (rollups; zie §11)
CREATE TABLE metric_host (ts INTEGER, res INTEGER, cpu REAL, mem_used INTEGER, mem_total INTEGER, disk_used INTEGER, disk_total INTEGER, net_rx INTEGER, net_tx INTEGER, load1 REAL, PRIMARY KEY (res, ts)) WITHOUT ROWID;
CREATE TABLE metric_container (ts INTEGER, res INTEGER, service_id TEXT, cpu REAL, mem INTEGER, mem_limit INTEGER, net_rx INTEGER, net_tx INTEGER, restarts INTEGER, PRIMARY KEY (service_id, res, ts)) WITHOUT ROWID;
CREATE TABLE metric_http (ts INTEGER, res INTEGER, service_id TEXT, requests INTEGER, s2xx INTEGER, s3xx INTEGER, s4xx INTEGER, s5xx INTEGER, bytes INTEGER, p50 REAL, p90 REAL, p95 REAL, p99 REAL, PRIMARY KEY (service_id, res, ts)) WITHOUT ROWID;
```

`config` per kind (Go-structs, JSON in DB):

```go
type AppConfig struct {
    Source      Source        // github{repo, branch, root_dir} | git{url, branch, deploy_key_id} | image{ref, registry_cred_id}
    Builder     string        // "railpack" | "dockerfile"
    Dockerfile  string        // pad, default "Dockerfile"
    BuildCmd    string        // optionele override (railpack)
    StartCmd    string        // optionele override
    Port        int           // default: EXPOSE uit image, anders 3000
    HealthPath  string        // "" = alleen TCP-check
    Replicas    int           // MVP: 1
    CPULimit    float64       // 0 = geen limiet
    MemLimitMB  int           // default 512 (instelbaar), zichtbaar in design ("256 MB limit")
    AutoDeploy  bool
    Volumes     []VolumeMount // persistente mounts
    StopGrace   int           // seconden
}

type StaticSiteConfig struct {
    Source        Source
    NodeVersion   string // default LTS; ook "none" voor pure HTML
    InstallCmd    string // auto: pnpm/npm/yarn/bun o.b.v. lockfile
    BuildCmd      string // auto uit package.json "build"
    OutputDir     string // auto-detect: dist, out, build, public, _site
    SPAFallback   bool
    GitLFS        bool
    AutoDeploy    bool
    BasicAuth     *BasicAuth // tab "Access" (v0.2)
}

type PostgresConfig struct { Version int; Database, User string; PasswordEnc []byte; MemLimitMB int; PublicPort int /* 0 = niet publiek */ }
type RedisConfig    struct { Version int; PasswordEnc []byte; Persistence bool; MaxMemoryMB int }
type BucketConfig   struct { BucketName string; AccessKeyID string; SecretKeyEnc []byte; PublicRead bool }
```

### 7.3 Verwijsbare variabelen ("service links")

Elke service publiceert **afgeleide variabelen**. In een app kan de gebruiker schrijven:

```
DATABASE_URL=${{db.DATABASE_URL}}
REDIS_URL=${{cache.REDIS_URL}}
S3_ENDPOINT=${{uploads.S3_ENDPOINT}}
```

| Kind | Gepubliceerde variabelen |
|---|---|
| postgres | `DATABASE_URL`, `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE` |
| redis | `REDIS_URL`, `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` |
| bucket | `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_PUBLIC_URL` |
| app | `PRIVATE_HOST` (`<slug>:<port>`), `PUBLIC_URL` (primair domein) |

Regels: alleen binnen hetzelfde project resolvebaar; resolutie bij elke deploy; cycli → validatiefout. UI: bij "Create database" vanuit een project biedt Dockyard aan om `DATABASE_URL` meteen aan geselecteerde apps toe te voegen. Bij het aanmaken verbindt Dockyard de app ook met het netwerk van Garage als er een bucket-link is.

**Next.js-specifiek**: variabelen met prefix `NEXT_PUBLIC_` krijgen automatisch `at_build=1`; UI toont een hint dat wijzigen een rebuild vraagt (niet alleen een restart).

---

## 8. Gebruikersbeheer en security

### 8.1 Rollen (MVP: instance-breed)

| Recht | owner | admin | member | viewer |
|---|---|---|---|---|
| Alles lezen (geen secrets) | ✓ | ✓ | ✓ | ✓ |
| Secrets tonen | ✓ | ✓ | ✓ | – |
| Deploy, restart, logs | ✓ | ✓ | ✓ | – |
| Services/projecten aanmaken/verwijderen | ✓ | ✓ | ✓ | – |
| Domeinen, GitHub-koppeling, instance-settings | ✓ | ✓ | – | – |
| Users uitnodigen/rollen wijzigen | ✓ | ✓ (niet owner) | – | – |
| Updates, server-cleanup | ✓ | ✓ | – | – |
| Ownership overdragen | ✓ | – | – | – |

RBAC als één Go-functie `can(user, action, target)`; elke Huma-operation declareert een `action`. `project_members` staat al in het schema voor per-project rechten in v0.2.

### 8.2 Authenticatie

- **Wachtwoord**: argon2id (m=64 MiB, t=3, p=2; parameters in de hash zodat ze later te verhogen zijn).
- **Sessies**: random 32-byte token in cookie `dy_session` (`HttpOnly`, `Secure` zodra HTTPS, `SameSite=Lax`, 30 dagen sliding, 12 uur zonder "remember"). In DB alleen de sha256. Design-scherm "Active sessions" toont device, locatie (GeoIP optioneel/uit; anders alleen IP), laatst actief, "Sign out".
- **2FA**: TOTP (RFC 6238) + 10 recovery codes. Na wachtwoord → tussenstap; sessie heeft `mfa_passed`. Owner kan 2FA verplicht stellen voor iedereen.
- **GitHub-login** ("Continue with GitHub" in design): via de OAuth-credentials van de GitHub App. Alleen zichtbaar als geconfigureerd. Koppelen aan bestaand account via verified e-mail óf expliciet vanuit Settings. Geen automatische account-aanmaak zonder invite.
- **Wachtwoord vergeten**: werkt alleen als SMTP is ingesteld (settings). Zonder SMTP toont het scherm: "Ask an admin, or run `sudo dockyard admin reset-password` on the server".
- **Invites**: link met token (48 uur geldig); e-mail als SMTP beschikbaar, anders kopieerbare link.
- **Brute force**: rate limit per IP + per account (bijv. 10/min, exponentiële backoff), constant-time vergelijkingen, generieke foutmelding ("Incorrect email or password", zoals in design).

### 8.3 Overige security-maatregelen

- **CSRF**: `SameSite=Lax` + `Origin`/`Sec-Fetch-Site`-check op alle muterende requests.
- **Headers**: strikte CSP (`default-src 'self'`; geen externe fonts/CDN's), `X-Frame-Options: DENY`, `Referrer-Policy: same-origin`.
- **Secrets**: AES-256-GCM met master-key; nooit in logs (build-log-scrubber vervangt secret-waarden door `••••`); secrets naar Railpack/BuildKit als build-secrets, niet als build-args.
- **Containers van gebruikers**: nooit `docker.sock` mounten, `no-new-privileges`, standaard geheugenlimiet, geen `--privileged` in de UI.
- **Netwerkisolatie**: projecten zien elkaar niet (eigen netwerk); alleen services met domein zitten op `dy-edge`.
- **Databases**: standaard niet publiek. "Make public" (v0.2) publiceert een poort met waarschuwing + sterk wachtwoord + optioneel IP-allowlist.
- **Webhooks**: HMAC-SHA256 (`X-Hub-Signature-256`) verplicht.
- **Audit log** van alle muterende acties.
- **SECURITY.md** met disclosure-adres en 90-dagenbeleid.

---

## 9. Design-systeem → code

Light mode only. Tokens rechtstreeks overgenomen uit het canvas.

### 9.1 Tokens (`web/styles/globals.css`)

```css
@import "tailwindcss";

@theme {
  /* surfaces */
  --color-canvas: #FAF9F8;        /* app-achtergrond, sidebar */
  --color-surface: #FFFFFF;       /* kaarten, tabellen */
  --color-subtle: #F1F0ED;        /* chips, nav-hover */
  --color-row-hover: #F6F5F2;
  --color-border: #ECEAE6;        /* kaarten, dividers */
  --color-input: #E1DFDA;         /* inputs, search */
  --color-border-strong: #CBD0D8; /* secundaire knoppen */

  /* tekst */
  --color-ink: #0B1F3A;           /* primaire tekst, primaire knop */
  --color-ink-hover: #1A3358;
  --color-ink-2: #3A4A63;         /* nav, chip-tekst */
  --color-muted: #5A6577;         /* labels, meta (≥ 4.5:1 op canvas) */

  /* merk */
  --color-brand: #8EBEFF;         /* banner, accent */
  --color-brand-soft: #E3F0FF;    /* actieve nav, icon tiles */
  --color-brand-mark: #3B93F5;    /* logo */
  --color-link: #0B57C9;

  /* status */
  --color-ok: #1A7F4B;   --color-ok-fg: #176B3D;   --color-ok-bg: #E6F4EC;
  --color-warn: #D98A00; --color-warn-fg: #8A5200; --color-warn-bg: #FBF0DC;
  --color-err: #B3261E;  --color-err-fg: #B3261E;  --color-err-bg: #FBE8E6;
  --color-idle: #6B7686; --color-idle-fg: #4A5361; --color-idle-bg: #ECEEF1;

  /* radius */
  --radius-chip: 6px;
  --radius-control: 10px;   /* knoppen, inputs, nav-items */
  --radius-tile: 12px;      /* icon tiles */
  --radius-card-sm: 16px;
  --radius-card: 20px;

  /* type */
  --font-sans: "Geist", system-ui, sans-serif;
  --font-mono: "Geist Mono", ui-monospace, monospace;
}
```

shadcn-variabelen (`--background`, `--primary`, `--border`, `--ring`, …) mappen op deze tokens, zodat shadcn-componenten automatisch in de Dockyard-stijl renderen.

### 9.2 Regels uit het design (als component-API vastleggen)

| Element | Specificatie |
|---|---|
| Control-hoogte | 44 px (knoppen, inputs, nav-items, search) |
| Primaire knop | bg `ink`, tekst wit, **mono, uppercase, 12 px, letter-spacing .02em**, radius 10 |
| Secundaire knop | transparant, 1.5 px border `border-strong` (licht/middengrijs, nooit donker), zelfde mono-uppercase tekst |
| Labels/eyebrows | mono 11–12 px uppercase, letter-spacing .06em, kleur `muted` |
| Chips | radius 6, bg `subtle`, mono 11 px uppercase |
| Status-badge | dot 6 px + mono-uppercase label, kleuren uit status-tokens (Healthy, Deploying, Failing, Stopped, Live, Building, Superseded) |
| H1 | 44 px / 600 / letter-spacing −0.04em / line-height 1.04 |
| H2 | 20 px / 600 / −0.03em |
| Stat-card | radius 20, padding 20, label mono, waarde 40 px/600 |
| Lijstrij | icon tile 40×40 (radius 12, bg `brand-soft`, icoon `link`), naam (600), meta (13 px muted), rechts commit-sha-chip + bericht + tijd |
| Layout | sidebar 232 px, header 64 px met breadcrumb links en ⌘K-search (280 px) + "Docs" rechts, main padding 32 |
| Banner | bovenaan, bg `brand`, "Open source"-chip, sluitbaar (dismiss opslaan per user) |
| Sidebar-voet | kaart "Server healthy" + versie-chip + "Up to date"/"Update available" |
| Navigatie | Dashboard, Projects, dan **Applications, Databases, Static sites, Object storage** als volwaardige items; extra ruimte; label "Server"; **Analytics**, **Settings** |
| Header | geen "Add service"-knop; geen Production/Staging-switch op de projectpagina |
| Toegankelijkheid | echte `<button>`/`<a>`, `aria-current="page"`, `aria-label` op icon-only knoppen, focus-ring met `brand` |

### 9.3 Componentenbibliotheek (`web/components/dockyard`)

`AppShell`, `Sidebar`, `TopBar`, `Breadcrumbs`, `CommandPalette`, `AnnouncementBanner`, `PageHeader` (eyebrow + H1 + subtitle + action), `StatCard`, `StatGrid`, `ResourceList` + `ResourceRow`, `StatusBadge`, `CommitChip`, `EmptyState`, `Tabs` (detail-tabs), `KeyValueList`, `LogViewer` (virtueel gescrold, filter, pause/resume, wrap, UTC-toggle), `EnvEditor` (reveal secrets, paste .env), `DomainCard` (DNS-instructies + "Check again"), `DangerZone`, `EChart`, `TimeRangeToggle` (1h/24h/7d/30d), `ServerHealthCard`, `CreateServiceDialog`.

---

## 10. Schermen → routes → API

### 10.1 Routekaart (Next.js `app/`)

| Design-scherm | Route | Belangrijkste API-calls |
|---|---|---|
| Sign in | `/login` | `POST /auth/login`, `POST /auth/mfa`, `GET /auth/github/start` |
| Forgot password | `/forgot-password`, `/reset-password` | `POST /auth/password/forgot`, `POST /auth/password/reset` |
| — (nieuw) Setup wizard | `/setup` | `POST /setup` |
| — (nieuw) Accept invite | `/invite/[token]` | `POST /auth/invites/accept` |
| Dashboard | `/` | `GET /overview` (stats, attention, recent, deploys, usage sparklines, backups) |
| Projects | `/projects` | `GET /projects`, `GET /events?limit=…` |
| Project overview | `/projects/[id]` | `GET /projects/{id}`, `GET /projects/{id}/services` |
| Applications | `/apps` | `GET /services?kind=app` |
| — App detail | `/apps/[id]` (+ `/deployments`, `/logs`, `/metrics`, `/variables`, `/domains`, `/settings`) | zie §10.2 |
| Databases | `/databases` | `GET /services?kind=postgres,redis` |
| — Database detail | `/databases/[id]` (+ `/logs`, `/settings`) | connection info, credentials (reveal), restart |
| Static sites | `/sites` | `GET /services?kind=static_site` + 24h-traffic |
| Static site · Overview/Deployments/Logs/Analytics/Environment/Domains/Settings | `/sites/[id]`, `/sites/[id]/deployments`, … | zie §10.2 |
| Object storage | `/storage` | `GET /storage` (Garage-status, buckets), `POST /storage/install` |
| — Bucket detail | `/storage/[id]` | keys, endpoint, public toggle |
| Analytics (server) | `/server` | `GET /server/metrics?range=24h`, `GET /events?category=…`, `GET /server/containers`, `GET /server/cleanup` |
| Settings · Profile & security | `/settings` | profile, password, 2FA, sessions |
| — (nieuw) Settings · Team | `/settings/team` | users, invites, rollen |
| — (nieuw) Settings · Instance | `/settings/instance` | domein, ACME-mail, wildcard, SMTP, updates, GitHub |
| Search (⌘K) | overlay in `AppShell` | `GET /search?q=` (projecten, services, pagina's, acties) |

### 10.2 REST API (`/api/v1`, JSON, cookie-auth)

```
Auth
POST   /auth/login                    {email, password, remember} → 200 | 202 {mfa_required}
POST   /auth/mfa                      {code | recovery_code}
POST   /auth/logout
GET    /auth/me
POST   /auth/password/forgot | /auth/password/reset
GET    /auth/github/start | /auth/github/callback
POST   /auth/invites/accept
POST   /setup                         (alleen met geldige setup-token, alleen zonder owner)

Account
PATCH  /me                            {name, email}
POST   /me/password
POST   /me/2fa/setup → {otpauth_url, qr_svg}   POST /me/2fa/enable {code}   DELETE /me/2fa
GET    /me/sessions                   DELETE /me/sessions/{id}

Team (admin)
GET    /users                         PATCH /users/{id} {role, disabled}   DELETE /users/{id}
GET    /invites                       POST /invites {email, role}           DELETE /invites/{id}

Projects
GET    /projects                      POST /projects                        GET|PATCH|DELETE /projects/{id}

Services (gemeenschappelijk)
GET    /services?kind=&project=
POST   /services                      {kind, project_id, environment, name, config}
GET|PATCH|DELETE /services/{id}
POST   /services/{id}/start | /stop | /restart
GET    /services/{id}/variables       PUT /services/{id}/variables  (bulk, ook "paste .env")
GET    /services/{id}/variables/reveal                                  (audit-logged)
GET    /services/{id}/domains         POST /services/{id}/domains   DELETE /domains/{id}
POST   /domains/{id}/check            → DNS-status + verwachte records

Deployments (app, static_site)
GET    /services/{id}/deployments
POST   /services/{id}/deployments     {ref?}                → nieuwe deploy
POST   /deployments/{id}/cancel | /redeploy | /promote (rollback)
GET    /deployments/{id}/log          (volledig, text/plain)

Logs & metrics
GET    /services/{id}/logs?since=&tail=&stream=runtime|access|build
GET    /services/{id}/metrics?range=1h|24h|7d|30d
GET    /services/{id}/analytics?range=…   (static sites + apps met domein: requests, latency pXX, status, bandwidth)

Server
GET    /overview
GET    /server                        (specs, uptime, versie, update-status)
GET    /server/metrics?range=…
GET    /server/containers
GET    /server/cleanup                → {build_cache, dangling_images, stopped_containers, old_releases}
POST   /server/cleanup                {targets:[…]}
GET    /events?category=all|issues|deploys|system&cursor=
POST   /events/{id}/ack

Storage
GET    /storage                       POST /storage/install {provider:"garage"}
(buckets zijn services met kind=bucket)

GitHub
POST   /github/app/manifest           → redirect-URL naar GitHub
GET    /github/app/callback
GET    /github/repos?q=               GET /github/repos/{owner}/{repo}/branches
POST   /hooks/github                  (publiek, HMAC)

Instance
GET|PATCH /settings/instance
POST   /settings/smtp/test
GET    /updates                       POST /updates/apply

Realtime
GET    /stream?topics=events,deployments:{id},logs:{service_id}   (SSE)

Search
GET    /search?q=
```

Conventies: ULID-ID's, `snake_case` JSON, foutformaat RFC 9457 (`application/problem+json`, Huma default), cursor-paginatie (`?cursor=&limit=`), tijden als Unix-ms in de API en relatief geformatteerd in de UI ("12 min ago").

---

## 11. Kernfunctionaliteit in detail

### 11.1 Applicaties

**Bronnen**
1. **GitHub (GitHub App)**: repo + branch + root directory (monorepo). Auto-deploy op push naar de branch.
2. **Publieke Git-URL**: zonder auth.
3. **Private Git zonder GitHub App**: deploy key (Dockyard genereert ed25519, toont public key).
4. **Docker image**: `ghcr.io/org/app:tag`, optioneel registry-credentials. "Redeploy" pullt opnieuw.

**Build-pipeline** (job `build`, concurrency standaard 1, instelbaar):

```
clone (shallow, depth=1; LFS optioneel)  → /var/lib/dockyard/builds/<deploy>/src
detect builder:
  Dockerfile aanwezig + builder=dockerfile  → docker buildx build -f <Dockerfile>
  anders railpack:
     railpack prepare src --plan-out plan.json --info-out info.json [--env K=V …]
     docker buildx build --builder dy-buildkit \
        --build-arg BUILDKIT_SYNTAX=ghcr.io/railwayapp/railpack-frontend:<gepinde versie> \
        --build-arg cache-key=<service_id> \
        --secret id=K,env=K … \
        -f plan.json -t dockyard/<svc>:<deploy> --load src
stream stdout/stderr → logbestand + SSE (met secret-scrubber)
cleanup src-map (altijd, ook bij falen)
```

- `railpack prepare` exit code 75 → retry; 1 → definitief falen; fouten uitlezen uit `info.json`.
- **Next.js**: Railpack detecteert Next en zet build/start correct. Docs-tip: `output: 'standalone'` in `next.config` voor kleinere images. Dockyard detecteert ontbrekende `standalone` en toont een niet-blokkerende hint.
- Build-timeouts (default 20 min), disk-guard (weiger build bij < 10 % vrije disk, met event + link naar Docker cleanup).
- Image-retentie: laatste 5 per service; oudere images opgeruimd door de `prune`-job.

**Runtime**: zie §3.6. Restart policy `unless-stopped`, geheugenlimiet, logs via json-file met rotatie. Persistente volumes per app optioneel (`/data` enz.).

### 11.2 Static sites

- Bronnen: GitHub, Git-URL.
- Build in een **wegwerp-container** (`node:<lts>-alpine` of op basis van `.nvmrc`/`engines`), met install/build-commando's uit settings of auto-detect (lockfile → pnpm/npm/yarn/bun; `package.json` `build`-script).
- Output-dir auto-detect: `dist`, `out`, `build`, `public`, `_site`. Leeg/ontbrekend → duidelijke fout.
- **Release**: kopieer output naar `/var/lib/dockyard/sites/<site>/releases/<deploy>/` → atomische `rename` van symlink `current` → Caddy serveert direct, geen reload nodig. Rollback = symlink terugzetten (milliseconden).
- Retentie: laatste 5 releases.
- Env vars zijn alleen build-time (design: "Injected at build time").
- Tab **Access** (v0.2): basic auth / wachtwoord op de site via Caddy.
- Optionele `_headers`/`_redirects`-bestanden (Netlify-formaat) → vertaald naar Caddy-routes. Nice-to-have na MVP.

### 11.3 Logs en analytics

| Bron | Opslag | Weergave |
|---|---|---|
| Build-logs | Bestand per deployment (`/var/lib/dockyard/logs/deploy/<id>.log`), 30 dagen | Deploy-detail, live via SSE |
| Runtime-logs (apps, DB's) | Docker json-file (rotatie 3×10 MB) | `ContainerLogs` met `follow` → SSE |
| Access-logs (apps met domein, static sites) | Caddy JSON-log; Dockyard tailt het bestand (inotify + positie-checkpoint) | Logs-tab "Access"; geaggregeerd voor Analytics |

**Analytics-aggregatie**: per minuut per service: requests, 2xx/3xx/4xx/5xx, bytes, latency-percentielen (t-digest of vaste histogram-buckets per minuut). Rollups: 1 min (48 uur bewaren), 5 min (7 dagen), 1 uur (90 dagen). Grafieken uit het design (requests/min, response time Avg/P90/P95/P99, status codes, bandwidth) komen hier direct uit. Geen cookies, geen IP-opslag, geen bezoekersprofielen.

### 11.4 Server-analytics

- **Host-collector** (elke 10 s, opgeslagen per minuut): CPU %, load, geheugen, disk per mountpoint, netwerk rx/tx.
- **Container-collector** (elke 15 s): Docker stats one-shot per container (CPU, mem/limit, net, restarts).
- **Events** (design "Important events"): deploy started/succeeded/failed, container restarted (OOM via `docker events` `oom`/`die`), certificate renewed/failing, Dockyard updated, backup done (v0.2), disk > 85 %, build cache > X GB. Categorieën: deploys, issues, system.
- **Needs attention**: open events met level warning/error die niet ge-ackt zijn (bijv. "worker keeps hitting its memory limit" → knop "Open service").
- **Docker cleanup**: overzicht van build cache, dangling images, gestopte niet-beheerde containers, oude releases → "Review and clean up". Raakt nooit draaiende services of volumes.

### 11.5 Databases

- **Postgres** (major 16/17/18) en **Redis-compatible** (Valkey 8; label "Redis" in UI). Officiële images, eigen volume, gegenereerd wachtwoord (32 tekens), healthcheck (`pg_isready` / `PING`).
- Detail: connection strings (intern), credentials met reveal, logs, restart, versie, geheugenlimiet.
- Geen blue/green: config-wijziging = gecontroleerde restart.
- **Major upgrades**: niet in MVP (docs: dump/restore).
- **Backups**: v0.2 (§15). Design-tekst "with backups you can restore" en de backup-widget op het dashboard: in MVP verbergen of tonen als "Coming soon".

### 11.6 Object storage (Garage)

- Lege staat (design): "No buckets yet" + providerkeuze. **Install Garage** → Dockyard start `dy-garage` (single node, `replication_factor = 1`, metadata en data op `/var/lib/dockyard/garage`), genereert `rpc_secret` en admin-token, wijst layout toe.
- Buckets en access keys via de **Garage admin-API** (v2) vanuit `internal/storage`.
- Intern endpoint `http://dy-garage:3900` (Garage wordt aan projectnetwerken gekoppeld zodra een project een bucket heeft); optioneel publiek via `s3.<instance-domein>` (Caddy-route) voor presigned uploads vanuit de browser.
- Publieke read (statische assets) per bucket via Garage website-mode + Caddy-route.
- Metrics: opslag, objecten, requests (Garage metrics-endpoint).

---

## 12. Belangrijke afwijking van het design: MinIO

Het Object storage-scherm biedt **Garage (Recommended)** en **MinIO (Popular)**. Advies: **MinIO uit het design halen.** De community-editie van MinIO stopte in oktober 2025 met publieke binaries en Docker-images, ging in december 2025 in maintenance mode en de repository is in 2026 gearchiveerd. Een nieuw platform zou daarmee een bevroren, source-only dependency installeren.

Voorstel: alleen Garage in de MVP; de tweede kaart wordt **SeaweedFS** (Apache-2.0, volwassen) als "later" of verdwijnt. RustFS is interessant maar nog te jong voor een default.

---

## 13. Updates en releases

### 13.1 Releaseproces

- Tags `vX.Y.Z` (semver). GoReleaser bouwt `linux/amd64` + `linux/arm64`, maakt `checksums.txt`, SBOM (syft), en ondertekent met **cosign keyless** (GitHub OIDC). Release notes uit conventional commits.
- Kanalen: `stable` (GitHub "latest"), `beta` (pre-releases).
- Elke release pint digests van `caddy`, `buildkit`, `railpack-frontend`, `garage` in een embedded `components.json`.

### 13.2 Self-update flow

```
check (1×/24h + handmatig): GET api.github.com/repos/<org>/dockyard/releases/latest
  → nieuwere versie? event "Update available" + chip in sidebar (design: "Up to date")
apply (handmatig in UI of auto als ingesteld, met onderhoudsvenster):
  1. download tarball + checksums + .sig/.cert
  2. verifieer sha256 én cosign-signature (identiteit = release-workflow van deze repo)
  3. VACUUM INTO pre-update backup
  4. schrijf nieuwe binary naar /usr/local/bin/dockyard.new, fsync, rename (atomisch)
     bewaar vorige als dockyard.prev
  5. systemctl restart dockyard   (workloads + Caddy draaien door)
  6. nieuwe versie: migraties → health → sd_notify READY
  7. updater-watchdog (systemd OnFailure-unit dockyard-rollback.service):
     als READY niet binnen 60 s → dockyard.prev terugzetten + pre-update DB terugzetten + event
  8. systeemcontainers alleen vervangen als hun gepinde digest wijzigt. Caddy: graceful stop
     + start duurt ~1 s; dit staat in de release notes ("proxy restart, ~1 s"). Zonder
     digest-wijziging blijft de proxy ongemoeid.
```

Database-migraties zijn **forward-only en backwards-compatible voor één minor-versie** (expand/contract), zodat een rollback naar `prev` met de nieuwe DB nog werkt; anders wordt de pre-update DB teruggezet.

---

## 14. Ontwikkeling, testen en CI

### 14.1 Lokaal ontwikkelen

- Dockyard heeft echte Linux + Docker nodig. Aanbevolen: `scripts/dev-vm.sh` (Multipass/Lima VM met Ubuntu 24.04) of een goedkope dev-VPS. Docker Desktop/OrbStack werkt voor het meeste, behalve host-paden en `/proc`-metrics.
- `make dev`:
  - Go-server met `DOCKYARD_DEV_UI=http://localhost:3000` → niet-API-requests worden ge-reverse-proxied naar de Next.js dev-server. Eén origin, cookies werken, HMR werkt.
  - `pnpm dev` in `web/`.
- `make generate`: sqlc → Go; Huma → `openapi.json` → `openapi-typescript` → `web/lib/api/schema.d.ts`. CI faalt als gegenereerde bestanden niet up-to-date zijn.
- Seed-commando `dockyard dev seed` vult een DB met de demo-data uit het design (Client X, Webshop, …) voor UI-werk zonder echte containers.

### 14.2 Tests

| Laag | Wat | Tooling |
|---|---|---|
| Unit | config-validatie, variabele-resolutie, Caddy-renderer (golden files), RBAC, rollups, state machine | `go test`, golden files |
| Integratie | echte Docker: netwerk/volume/containers aanmaken, blue/green, reconciler-drift, Caddy laden | `go test -tags=integration` op GitHub Actions (Docker beschikbaar) |
| E2E | `install.sh` op de CI-runner → setup → deploy voorbeeld-Next.js-app + Postgres + static site → HTTP-checks → update naar snapshot-build → workloads bleven up | shell + Playwright |
| UI | componenten en kritieke flows (login + 2FA, create app, env editor) | Playwright |
| Chaos-light | `kill -9 dockyard` tijdens build/deploy → na herstart consistente staat, geen downtime van live service | integratietest |

Voorbeeldrepo's in `e2e/fixtures/` (Next.js app-router met Prisma/Drizzle → Postgres, Vite static site, Dockerfile-app).

### 14.3 CI-workflows

- `ci.yml` (PR): `go vet`, `staticcheck`, `golangci-lint`, unit tests, `pnpm lint/typecheck/build`, generated-files-check, **Radix-check**, licentie-check van dependencies (geen GPL/AGPL in de binary).
- `e2e.yml` (main + nightly): integratie + install-e2e.
- `release.yml` (tags): GoReleaser + cosign + publiceren van `install.sh` naar `get.dockyard.dev` (GitHub Pages of Cloudflare Pages).

### 14.4 Radix-guard

```bash
# scripts/check-no-radix.sh
if pnpm --dir web why '@radix-ui/*' --json | grep -q '"name"'; then
  echo "Radix UI dependency found; use Base UI primitives instead"; exit 1
fi
```

### 14.5 Bijdragen (open source)

- `CONTRIBUTING.md`: dev-setup, commit-conventie (Conventional Commits), "één PR = één onderwerp", ADR voor architectuurwijzigingen.
- Issue-templates: bug (met `dockyard doctor`-output), feature request.
- Labels: `good first issue`, `help wanted`, `area/api`, `area/ui`, `area/proxy`, …
- DCO sign-off (lichter dan een CLA) via GitHub-app of check.

---

## 15. Roadmap en milestones

| Milestone | Inhoud | Klaar als |
|---|---|---|
| **M0 · Fundament** | repo, licentie, CI, Go-skelet, SQLite + migraties, `install.sh`, systemd, setup-token, setup-wizard, login/logout, sessies, AppShell + sidebar + design-tokens, ⌘K-skelet | Verse VPS → één commando → inloggen in een leeg dashboard |
| **M1 · Apps uit image/Dockerfile** | projects, services (app), Docker-laag + labels, Caddy-integratie, domeinen + HTTPS + sslip.io, blue/green, runtime-logs (SSE), env vars (versleuteld), reconciler | `nginx`-image én een Dockerfile-repo live op eigen domein, zero-downtime redeploy |
| **M2 · GitHub + Railpack + static sites** | GitHub App manifest-flow, repo-picker, webhooks/auto-deploy, Railpack-builds, build-logs, static sites met releases/rollback, deployment-history, rollback | Next.js-repo live via push-to-deploy; Vite-site live met rollback |
| **M3 · Databases + storage** | Postgres, Valkey, verwijsbare variabelen, Garage-installatie, buckets/keys | Next.js + Postgres + S3-uploads end-to-end werkend |
| **M4 · Observability + team** | host/container-metrics + rollups, events + needs attention, Analytics-pagina (ECharts), static-site analytics uit Caddy-logs, Docker cleanup, team/invites/rollen, 2FA, sessions-overzicht, audit log, search | Alle 19 design-schermen functioneel |
| **M5 · Release v0.1.0** | self-update + rollback, `doctor`, docs (install, eerste app, Next.js-guide, troubleshooting), e2e groen, security-review | Publieke release + announcement |
| **v0.2** | database-backups (pg_dump → S3/Garage, schema + restore), notificaties, static-site Access (basic auth), web-terminal, preview deploys, per-project rechten, API-tokens | |
| **v0.3** | Multi-server via `dockyard agent` (mTLS-tunnel, geen SSH-shellcommando's), Compose-templates, MySQL/MariaDB | |

**Backups (v0.2, alvast ontworpen)**: job `backup` per database-schema (cron-expressie), `pg_dump -Fc` gestreamd naar S3-compatible target (Garage lokaal of extern), retentie (dagelijks 7, wekelijks 4), restore naar dezelfde of nieuwe database, verificatie door periodiek een restore naar een tijdelijke container. Design-widget "Backups · Scheduled to S3 · Next run 03:00 · Restore" past hier precies op.

---

## 16. Licentie

**Apache License 2.0.**

- Simpel en breed geaccepteerd; ook gebruikt door Coolify en Dokploy, dus vertrouwd in deze niche.
- Expliciete patentlicentie (beter dan MIT voor infra-software waar bedrijven op bouwen).
- Laat bedrijven Dockyard vrij gebruiken en aanpassen, wat adoptie helpt.

Afweging: als het risico dat een partij Dockyard als concurrerende gehoste dienst aanbiedt zwaarder weegt dan adoptie, is AGPL-3.0 het alternatief. Voor een project dat vooral adoptie en contributors zoekt is Apache-2.0 de beste fit. Let erop dat ingebedde dependencies licentie-compatibel zijn (Caddy en BuildKit: Apache-2.0; Garage: AGPL-3.0, maar draait als **losse container** en wordt niet meegelinkt, dus geen probleem voor de licentie van Dockyard; vermeld het wel in `NOTICE`/docs).

---

## 17. Design-gaten (nog te ontwerpen vóór/tijdens bouw)

| Ontbreekt | Nodig voor | Prioriteit |
|---|---|---|
| Setup-wizard (owner, domein, GitHub) | M0 | Hoog |
| App detail (Overview/Deployments/Logs/Metrics/Variables/Domains/Settings) | M1 | Hoog — kan het static-site-patroon hergebruiken |
| Create-flows: application (bron kiezen → repo → config), database, static site, bucket | M1–M3 | Hoog |
| Database detail (connection info, credentials, logs, settings) | M3 | Middel |
| Bucket detail (keys, endpoint, public) | M3 | Middel |
| Settings · Team (users, invites, rollen) en Settings · Instance (domein, SMTP, updates, GitHub) | M4 | Hoog (user management is verplicht) |
| 2FA-stap bij login, invite-acceptatie, reset-password-scherm | M4 | Middel |
| Deployment-detail met build-log | M2 | Middel |
| Lege staten voor Dashboard/Projects bij nul services | M0 | Middel |
| Update-flow (modal met release notes, voortgang, resultaat) | M5 | Laag |
| MinIO-kaart vervangen (§12) | M3 | Laag |

---

## 18. Open beslissingen voor Joey

1. **Domein voor installer en docs**: `get.dockyard.dev`/`dockyard.dev` beschikbaar? (bepaalt URL's in `install.sh` en de banner "Learn more").
2. **GitHub-organisatie** voor de repo (bijv. `github.com/dockyard-dev/dockyard`).
3. **Redis vs. Valkey**: voorstel is Valkey-image met UI-label "Redis" en hint "Redis-compatible (Valkey)". Akkoord?
4. **Dashboard-poort**: `8000` (zoals Coolify) of `3000`? Voorstel 8000.
5. **Auto-updates standaard aan of uit**? Voorstel: check aan, toepassen handmatig (met optie "auto-apply patch releases").
6. **Telemetrie**: voorstel geen, of strikt opt-in anonieme installatie-ping. Expliciet in README.
7. **Next.js static-export-patroon** (§4.2) evalueren aan het eind van M1; fallback is Vite + TanStack Router.

---

## 19. Definition of done voor v0.1.0

- [ ] `curl -fsSL https://get.dockyard.dev | sudo sh` op verse Ubuntu 24.04 én Debian 12, amd64 én arm64, < 3 min.
- [ ] Setup-token vereist; zonder token geen owner-claim.
- [ ] Next.js-repo (app router) + Postgres + Garage-bucket live op eigen domein met HTTPS via push-to-deploy.
- [ ] Static site (Vite) live, rollback < 1 s.
- [ ] Redeploy zonder één mislukte request (getest met load tijdens deploy).
- [ ] `kill -9` van Dockyard tijdens build → live services onaangetast, deployment na herstart netjes `failed (interrupted)`.
- [ ] Self-update naar nieuwe versie + automatische rollback bij kapotte release, beide getest.
- [ ] Users: invite, rollen, 2FA, sessies intrekken, wachtwoord-reset via CLI.
- [ ] Idle-footprint gemeten: Dockyard < 60 MB RSS, totaal platform < 150 MB.
- [ ] Geen `@radix-ui/*` in de dependency tree; geen externe requests vanuit de UI.
- [ ] Docs: installatie, eerste app, Next.js-gids, domeinen/DNS, troubleshooting (`dockyard doctor`).
- [ ] LICENSE, NOTICE, CONTRIBUTING, SECURITY, CODE_OF_CONDUCT aanwezig.

---

## Bijlage A — Referenties

- Coolify proxy-documentatie: https://coolify.io/docs/core/networking/proxy/overview
- Dokploy architectuur: https://docs.dokploy.com/docs/core/architecture
- Railpack introductie (opvolger van Nixpacks): https://blog.railway.com/p/introducing-railpack
- Railpack in productie (prepare + BuildKit-frontend): https://railpack.com/platforms/running-railpack-in-production
- Docker SDK v29-modulesplitsing (`moby/moby/client`, `moby/moby/api`): https://gitlab.com/gitlab-org/gitlab-runner/-/merge_requests/6902
- shadcn/ui met Base UI: https://www.pkgpulse.com/guides/shadcn-ui-vs-base-ui-vs-radix-components-2026 · https://github.com/shadcn-ui/ui/discussions/9562
- Status MinIO CE 2026 en alternatieven: https://glukhov.org/data-infrastructure/object-storage/minio-dead/
- Garage: https://garagehq.deuxfleurs.fr

## Bijlage B — Voorbeeld Caddy-config (gerenderd door Dockyard)

```json
{
  "admin": { "listen": "unix//run/caddy/admin.sock" },
  "logging": { "logs": { "access": {
    "writer": { "output": "file", "filename": "/var/log/caddy/access.log", "roll_size_mb": 50, "roll_keep": 3 },
    "encoder": { "format": "json" }, "include": ["http.log.access.dy"] } } },
  "apps": {
    "http": { "servers": { "dy": {
      "listen": [":443", ":80"],
      "logs": { "default_logger_name": "dy" },
      "routes": [
        { "match": [{ "host": ["api.clientx.nl"] }],
          "handle": [{ "handler": "reverse_proxy",
                       "upstreams": [{ "dial": "dy-api-01jc9k:3000" }],
                       "load_balancing": { "try_duration": "5s" } }],
          "terminal": true },
        { "match": [{ "host": ["clientx.nl"] }],
          "handle": [
            { "handler": "vars", "dy_service": "svc_01jc…" },
            { "handler": "encode", "encodings": { "zstd": {}, "gzip": {} } },
            { "handler": "file_server", "root": "/srv/sites/svc_01jc…/current" } ],
          "terminal": true }
      ] } } },
    "tls": { "automation": { "policies": [{ "issuers": [{ "module": "acme", "email": "ops@example.com" }] }] } }
  }
}
```

## Bijlage C — Deployment-statusmachine

```
queued ──► waiting ──► cloning ──► building ──► starting ──► healthchecking ──► live
   │          │           │           │            │              │               │
   └──────────┴───────────┴───────────┴────────────┴──────────────┴──► failed     └──► superseded (bij volgende live)
   └──► cancelled (door gebruiker of nieuwere push met "cancel older builds")
```

`waiting` = admission control wacht op RAM-budget (build resilience). Zie [`docs/project/build-resilience.md`](project/build-resilience.md).

Overgangen gebeuren alleen in `internal/deploy` in een DB-transactie; elke overgang publiceert een SSE-event `deployment.updated` en, bij `live`/`failed`, een `events`-record.