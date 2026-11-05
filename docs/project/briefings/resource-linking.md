# Dockyrd: briefing "Resource links" (verbonden services)

> Status: bouwbriefing, Feature-track · Datum: 5 november 2026 · Eigenaar: Joey de Kruis
> Aanleiding: Sevalla's "Internal connections" / "Connected services" vonden we sterk.
> Bronnen: Sevalla-docs (zie bijlage A), codebase `dockyrd` (branch `development`), design-canvas "Dockyrd Dashboard", `DOCKYARD_BRIEFING.md` §3.4, §7.3, §8.3.

---

## 0. Samenvatting

1. **Wat Sevalla doet:** een *internal connection* is een expliciete, door de gebruiker gelegde koppeling tussen twee resources (app ↔ database, app ↔ app) in dezelfde locatie. Eén klik, en (optioneel) worden host/poort/credentials als env-vars in de app gezet. Het overzicht toont dit als "Connected services"-paneel naast het proces.
2. **Veiligheid of gemak?** Beide, maar de UI draait om **gemak**: de hele flow is "kies service → vink env-vars aan → deploy". De **veiligheid** zit eronder: private hostnames zijn alleen bereikbaar voor gekoppelde services, verkeer blijft op het interne netwerk, en database-toegang van buiten staat standaard uit (met IP-allowlist). Sevalla noemt het zelf "sneller en veiliger" en de simpelere route. Dat het een *afdwingend* netwerkbeleid is, staat niet expliciet in de docs; het volgt uit "alleen internal connections kunnen deze hostnames bereiken".
3. **Wat wij al hebben:** `${{slug.VAR}}`-referenties (`internal/refs`), een per-project Docker-netwerk (`dy-p-<slug>`), `ServiceConnectionPanel` voor databases. **Wat ontbreekt:** een expliciet *link*-begrip. Nu kan elke service in een project elke andere bereiken, en elke variabele in dat project mag elke andere service-secret refereren.
4. **Voorstel:** `service_links` als eersteklas object, met drie lagen: **(a) toegang** (per-link Docker-netwerk = echte afscherming), **(b) gemak** (links schrijven *referenties* in plaats van gekopieerde waarden, dus nooit stale), **(c) zichtbaarheid** (Connections-paneel, "Connected applications" bij databases, later een project-kaart).
5. **Waar we beter kunnen dan Sevalla:** variabelen als referenties (geen stale kopie na wachtwoordwijziging), opruimen van variabelen bij ontkoppelen, backfill uit bestaande referenties, en afdwingen dat een referentie alleen naar een gekoppelde service kan wijzen.

---

## 1. Onderzoek: hoe Sevalla het doet

### 1.1 Concept

| Aspect | Sevalla |
|---|---|
| Object | *Internal connection*: één per paar services. Bidirectioneel; opnieuw aanmaken in een van beide richtingen geeft 409 |
| Tussen wat | Applicatie ↔ applicatie, applicatie ↔ database. Static sites zijn geen geldig target (API: `target_type` = `app` of `database`) |
| Scope | Zelfde *location*/cluster. Voor andere locaties: app klonen naar de locatie van de database |
| Rechten | Je hebt update-rechten nodig op **beide** services |
| Aanmaken | Vanuit de app (tab Networking → Connected services → "Add internal connection") **of** vanuit de database (Networking → Connected applications). Ook via API (`POST /applications/{id}/internal-connections {target_id, target_type}`) |
| Voorwaarde | Services hoeven nog niet gedeployed te zijn; verbinding geldt voor alle processen (incl. cron/jobs) |
| Verwijderen | Via ⋯ → Delete of prullenbak bij "Connected applications". Env-vars blijven staan en moeten handmatig weg |

### 1.2 Netwerk en veiligheid

- Elk web-process en elke worker krijgt een **interne hostname** (`<app>-<process>.<app>.svc.cluster.local`). "Only internal connections can reach these hostnames."
- Verkeer "stays inside the data center and doesn't pass through Cloudflare" (zie ook het Networking-blok in de screenshot: Cloudflare, DDoS, CDN → `8080` → web process).
- **Private ports**: extra poorten (bv. metrics) alleen intern bereikbaar voor gekoppelde services.
- **Databases:** externe toegang staat **standaard uit** ("Public networking" Enable/Disable); met IP-allowlist/denylist (403 voor geblokkeerde clients). Internal en external delen dezelfde user/wachtwoord, verschillen in hostname en poort.
- Sevalla zelf: internal is "sneller en veiliger", en de aanbevolen route tussen services op het platform.

### 1.3 Gemak (waar de UI om draait)

Het dialoog "Add internal connection" heeft:
1. **Service** kiezen (alleen: zelfde locatie, niet gesuspend, nog niet gekoppeld).
2. Toggle **"Add environment variables to the application"** (niet automatisch aan).
3. Sleutels **hernoembaar** (bv. naar wat je framework verwacht) en per variabele **runtime / build / beide**.
4. "Add internal connection", daarna **de app deployen** zodat de variabelen binnenkomen.

Variabelen per soort: Postgres/MySQL/MariaDB → `DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD, DB_URL`; Redis/Valkey → `REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_URL`; app → `APP_HOST`.

Zwakke plekken (kansen voor ons):
- De variabelen zijn **kopieën**: na een wachtwoordwijziging moet je ze zelf bijwerken (docs zeggen dit letterlijk).
- Ontkoppelen laat de variabelen **dangling** achter.
- Variabelen refereren kan wel (`${KEY}`, max 10 niveaus), maar niet naar een andere service.

### 1.4 UI-patroon (uit de screenshot)

Overview is een **kaart met drie kolommen** op een stippelraster: *Networking* (Cloudflare: DDoS/CDN/Edge caching) → label `8080` op een oranje streeplijn → **Web process**-kaart (status Running, start command, resources, instances, domein; acties View logs / Web terminal / Settings; "Create new process") → grijze streeplijn → **Connected services**-kaart (lijst: icoon + naam per gekoppelde service, onderaan oranje "Add internal connection"). Rechtsonder de regio. Kernidee: **de koppeling is zichtbaar als lijn**, de lijst is de bediening.

---

## 2. Analyse: waar staan wij nu (codebase)

| Onderdeel | Stand | Bron |
|---|---|---|
| Referenties `${{slug.VAR}}` | Werkt, deploy-tijd resolutie, cycli geweigerd, **hele project** is catalogus | `internal/refs`, `internal/deploy/catalog.go`, `runtime.go` (`runtimeEnv`, `buildArgs`) |
| Netwerk per project | `dy-p-<slug>`; elke app/db zit op het projectnetwerk met alias = slug | `internal/docker/run.go`, `database.go`, `names.go` |
| Garage | `dy-garage` wordt aan het **hele** projectnetwerk gehangen zodra het project een bucket heeft | `internal/docker/garage_network.go`, `reconcile/buckets.go` |
| UI database | `ServiceConnectionPanel` (waarden, reveal, copy) met tekst "same project network" | `web/components/dockyard/service-connection-panel.tsx` |
| UI app | Tabs Overview/Deployments/Logs/Environment/Domains/Settings; geen Connections/Networking | `application-detail-page.tsx` |
| Env-editor | Toevoegen/paste `.env`; geen referentie-picker | `add-env-variable-modal.tsx` |
| Design-canvas | Heeft Application Overview/Processes/Environment, **geen** connections-scherm | canvas "Dockyrd Dashboard" |

### 2.1 Gaten die dit oplevert (belangrijk)

1. **Geen expliciete toegang.** Elke service in een project bereikt elke andere. Een gecompromitteerde app kan de database van een klantsite in hetzelfde project benaderen, ook als ze niets met elkaar te maken hebben.
2. **Referenties lekken secrets binnen een project.** Elke variabele mag `${{db.PGPASSWORD}}` van een willekeurige service gebruiken, ook zonder dat die app die database hoort te gebruiken.
3. **Mogelijk afwijkend van de briefing (nakijken):** `RunApp` koppelt **elke** app aan `dy-edge`, ook zonder domein. Briefing §3.4/§8.3 zegt "alleen services met domein zitten op dy-edge". Als `dy-edge` gedeeld is, kunnen apps uit verschillende projecten elkaar via containernaam bereiken. Dit is een voorwaarde voor betekenisvolle link-isolatie (zie Fase 0).
4. Garage-koppeling is project-breed in plaats van per bucket-gebruiker.

---

## 3. Ontwerpbeslissingen

### 3.1 Wat is een link?

Een link is een **toestemming + gemak + zichtbaarheid** in één object:

- **Toestemming:** alleen gekoppelde services kunnen elkaar bereiken (netwerk) en elkaars gepubliceerde variabelen refereren.
- **Gemak:** bij aanmaken kunnen we referenties-variabelen voor je klaarzetten.
- **Zichtbaarheid:** één plek die laat zien wat waaraan hangt.

### 3.2 Beslissingen (met advies)

| # | Vraag | Advies | Reden |
|---|---|---|---|
| D1 | Gericht of bidirectioneel? | **Gericht** `from` (consument) → `to` (aanbieder), uniek per paar | Voor databases/buckets is de richting duidelijk. App ↔ app: twee links als beide kanten nodig zijn. Variabelen gaan naar de consument |
| D2 | Welke combinaties? | app→postgres, app→redis, app→bucket, app→app. **Niet** static_site (geen runtime; build-time only) | Gelijk aan Sevalla's beperking |
| D3 | Scope | Alleen dezelfde **project + environment** (briefing: `environment_id`) | Past bij het projectmodel; geen cross-project links in MVP |
| D4 | Afdwingen op netwerk? | **Ja, per-link Docker-netwerk** (zie 4.2), met projectnetwerk-vlak verdwijnend | Zonder dit is een link alleen cosmetisch. Dockyrd's belofte is "reliable & safe" |
| D5 | Variabelen als kopie of referentie? | **Referentie** (`DATABASE_URL=${{db.DATABASE_URL}}`) | Nooit stale, secret blijft op één plek, past bij bestaande `refs`-engine |
| D6 | Referentie zonder link? | **Weigeren** bij opslaan én deploy ("`db` is niet gekoppeld aan deze app. Koppel eerst.") met one-click "Koppel en gebruik" | Sluit gat 2 |
| D7 | Ontkoppelen | Dialoog toont variabelen die naar de service wijzen; checkbox "Verwijder deze variabelen" (default aan) | Sevalla laat ze dangling |
| D8 | Rechten | `service.link` = member+ op het project (viewer: alleen lezen). Audit-log + event bij create/delete | Sevalla eist update-recht op beide kanten; wij hebben instance-brede rollen |
| D9 | Bestaande installaties | **Migratie met backfill**: scan variabelen op `${{slug.VAR}}` → maak links aan. Services zonder referenties houden tijdelijk toegang ("legacy open") tot de gebruiker "Strict links" aanzet | Geen breuk bij update |

### 3.3 Strict vs. open

Projectinstelling `links_mode`: `open` (huidig gedrag, projectnetwerk blijft) of `strict` (alleen links). Nieuwe projecten starten `strict`; bestaande projecten `open` met een banner "Beperk toegang tot gekoppelde services" na backfill. Zo is de migratie veilig en blijft de UI eerlijk over wat wordt afgedwongen (badge "Enforced" / "Not enforced").

---

## 4. Technisch ontwerp

### 4.1 Datamodel (goose-migratie `00005_service_links.sql`)

```sql
CREATE TABLE service_links (
  id            TEXT PRIMARY KEY,                       -- ULID
  from_service_id TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,  -- consument (app)
  to_service_id   TEXT NOT NULL REFERENCES services(id) ON DELETE CASCADE,  -- aanbieder
  created_by    TEXT REFERENCES users(id) ON DELETE SET NULL,
  created_at    INTEGER NOT NULL,
  UNIQUE (from_service_id, to_service_id),
  CHECK (from_service_id <> to_service_id)
);
CREATE INDEX idx_links_to ON service_links(to_service_id);

ALTER TABLE variables ADD COLUMN link_id TEXT REFERENCES service_links(id) ON DELETE SET NULL;
ALTER TABLE projects ADD COLUMN links_mode TEXT NOT NULL DEFAULT 'strict' CHECK (links_mode IN ('open','strict'));
```

- Validatie in Go (niet in SQL): beide services in dezelfde environment, kinds volgens D2, geen static_site.
- `variables.link_id` markeert door een link aangemaakte variabelen (UI-chip "via link", cleanup bij ontkoppelen).
- sqlc-queries: `ListLinksForService` (beide richtingen), `CreateLink`, `DeleteLink`, `ListVariablesByLink`, `ListLinkedProviders(consumer)`.

### 4.2 Netwerk-afdwinging

Per link één bridge-netwerk `dy-l-<linkshort>` met precies twee leden:

```
app container   ── dy-l-<link1> ── db container      (alias: db → slug van de aanbieder)
app container   ── dy-l-<link2> ── redis container
app (met domein)── dy-edge ── dy-proxy (Caddy)
```

- Aanbieder-alias in het link-netwerk blijft de **leesbare slug** (`db:5432`), zoals het design nu al toont.
- Blue/green (briefing §3.6): de **nieuwe** container wordt vóór de healthcheck aan dezelfde link-netwerken gehangen; oude container verlaat ze bij het afbouwen. Alias-verplaatsing zoals nu.
- Aanbieders (db/redis) worden aan elk link-netwerk gehangen via `NetworkConnect` (database.go heeft dat patroon al voor het projectnetwerk).
- **Garage:** `dy-garage` wordt per bucket-link aan het link-netwerk gehangen (vervangt `SyncBucketNetworks` op projectniveau).
- **Reconciler:** link-netwerken zijn "desired state" uit `service_links`; reconciler maakt ontbrekende netwerken aan, koppelt containers opnieuw na restart, en ruimt verweesde `dy-l-*` netwerken op (label `dev.dockyard.link=<id>`).
- **Adresruimte (risico):** Docker heeft een beperkt aantal standaard subnetten. Installer moet `default-address-pools` in `daemon.json` meezetten (bijv. `10.200.0.0/16` in /27-blokken: ruim honderden netwerken). De installer merget `daemon.json` al; dit hoort daar bij. Documenteer de limiet in `dockyard doctor`.
- Volgorde bij `strict`: projectnetwerk `dy-p-<slug>` verdwijnt voor nieuwe deploys; in `open` blijft het bestaan naast de link-netwerken.

### 4.3 Variabelen en referenties

- `projectCatalog` wordt `consumerCatalog(consumerID)`: bevat **alleen** gepubliceerde variabelen van services waarmee een link bestaat. Een referentie naar een niet-gekoppelde slug geeft dezelfde fout als een onbekende slug, plus de hint "niet gekoppeld".
- `Resolve` zelf blijft ongewijzigd (alleen de catalogus wordt gefilterd). Cyclus-detectie blijft.
- Validatie bij `PUT /services/{id}/variables`: elke `${{slug.VAR}}` moet naar een gekoppelde service wijzen en een bestaande publicatie-variabele.
- **Standaard variabelen per soort** (bij aanmaken link, aanpasbaar):

| Aanbieder | Voorgestelde variabelen (sleutel = standaard, hernoembaar) |
|---|---|
| postgres | `DATABASE_URL=${{<slug>.DATABASE_URL}}` (aan); `PGHOST/PGPORT/PGUSER/PGPASSWORD/PGDATABASE` (uit) |
| redis | `REDIS_URL=${{<slug>.REDIS_URL}}` (aan); `REDIS_HOST/PORT/PASSWORD` (uit) |
| bucket | `S3_ENDPOINT, S3_REGION, S3_BUCKET, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY` (aan); `S3_PUBLIC_URL` alleen bij public read |
| app | `PRIVATE_HOST=${{<slug>.PRIVATE_HOST}}`, hernoemd naar `<SLUG>_URL` voorstel |

- Sleutelconflict met een bestaande variabele: dialoog toont "bestaat al" en laat hernoemen of overslaan (nooit stil overschrijven).
- `is_secret` automatisch aan voor `*URL` met credentials en `*PASSWORD`/`*SECRET*`. `at_build/at_runtime` per regel; default runtime aan, build aan voor wie dat wil (Sevalla: beide).
- Na aanmaken: banner "Variabelen gewijzigd. Herdeploy om ze toe te passen" met knop **Redeploy** (bestaande `trigger=config`).

### 4.4 API (`/api/v1`, Huma)

```
GET    /services/{id}/links              → { outgoing:[Link], incoming:[Link] }   (Link bevat peer {id,name,kind,slug,status}, variables_count)
GET    /services/{id}/linkable           → kandidaten in dezelfde environment (niet gekoppeld, geldig per D2) + voorgestelde variabelen per kandidaat
POST   /services/{id}/links              {to_service_id, variables:[{key, ref, at_build, at_runtime}]}  → 201 {id} | 409 bestaat al | 422 ongeldig paar
DELETE /links/{id}?remove_variables=true
PATCH  /projects/{id}                    {links_mode}
POST   /projects/{id}/links/backfill     (admin, idempotent; ook automatisch in migratie)
```

- Actions voor RBAC: `service.link.read` (alle rollen), `service.link.write` (member+).
- Vanuit de aanbieder-kant: `POST /services/{db}/links` met `{from_service_id}` als spiegel, zodat het dialoog op database-pagina's dezelfde backend gebruikt.
- SSE: topic `services:{id}` krijgt `link.created`/`link.removed`. Events-tabel: category `system`. Audit: `link.create`, `link.delete`.
- Foutformaat RFC 9457; 409 bij dubbel paar.

### 4.5 Beveiligingspunten

- Geen secret-waarden in link-responses; alleen referentie-strings en maskering (hergebruik reveal-flow en audit).
- Link-aanmaak controleert **server-side** project/environment, kind-combinatie, RBAC, en maakt in één DB-transactie `service_links` + `variables`; daarna job `link.apply` voor Docker-netwerken (idempotent, retry).
- Verwijderen: eerst netwerkverbinding laten vallen (container verliest toegang), dan rijen. Volumes en data worden nooit geraakt.
- `strict` + link verwijderen terwijl app draait: verbinding valt direct weg; de UI waarschuwt "App verliest verbinding met `db` totdat je opnieuw koppelt".

---

## 5. UI-ontwerp (nieuwe schermen en aanpassingen)

Principes uit ons design-contract blijven: Base UI, tokens, control-hoogte 44, primaire/secundaire knoppen in mono-caps, geen "Add service" in de header, geen Production/Staging-switch.

### 5.1 Application · Overview: kaart "Connections"

Naast Details/Recent deployments een nieuwe kaart (zelfde `RESOURCE_DETAIL_CARD`):

```
CONNECTIONS                                        [ + ADD CONNECTION ]
┌────────────────────────────────────────────────────────────────────┐
│ [▣] db            postgres · db:5432        ● Healthy      ⋯      │
│ [▣] cache         redis · cache:6379        ● Healthy      ⋯      │
│ [▣] uploads       bucket · garage           ● Healthy      ⋯      │
└────────────────────────────────────────────────────────────────────┘
 Alleen gekoppelde services zijn bereikbaar voor deze app.   [Enforced]
```

- Rij = `ResourceRow` (icon tile 40×40 `brand-soft`, naam 600, meta 13 px `muted`, `StatusBadge`, ⋯ `PopoverMenu`: *Open service*, *View variables* (springt naar Environment met filter), *Remove connection*).
- Lege staat (`EmptyState`): "Deze app kan nog geen andere resources bereiken. Koppel een database, cache of bucket."
- Rij toont chip `N variables` die naar Environment filtert.
- Badge **Enforced / Not enforced** (`links_mode`), met tooltip en link naar de projectinstelling.

### 5.2 Dialoog "Add connection"

`WizardModalShell`/`Modal`, één stap:

1. **Service** (`SearchSelect`): lijst met icon tile + naam + `kind · interne host`. Uitgesloten: zichzelf, al gekoppelde, static sites. Is de lijst leeg: "Maak eerst een database" met knop naar de create-flow.
2. **Environment variables** (`Switch` "Add environment variables to the application", **standaard aan**). Daaronder een tabel per voorgestelde variabele: *Key* (editbaar), *Value* (read-only referentie-chip, `${{db.DATABASE_URL}}`), `Runtime` / `Build` (checkboxen), regel-toggle om hem over te slaan.
3. Voetregel: "Herdeploy na het opslaan zodat de app de variabelen krijgt." Knoppen: ANNULEREN (secundair) / ADD CONNECTION (primair, mono-caps).

Succes: snackbar "Verbonden met db" + banner op de pagina "Redeploy om variabelen toe te passen" [REDEPLOY].

### 5.3 Database / Redis / Bucket detail: "Connected applications"

Gespiegeld paneel boven of onder het bestaande `ServiceConnectionPanel`:

- Lijst van apps die dit resource gebruiken (rij + ⋯ *Remove*), knop "+ Connect application" (zelfde dialoog, vanuit de andere kant).
- `ServiceConnectionPanel`-tekst aanpassen: "Alleen bereikbaar voor gekoppelde applicaties." in plaats van "same project network".
- Bij verwijderen (ConfirmDialog): "Applicatie X verliest toegang tot deze database." Checkbox "Verwijder N variabelen in X die naar deze database verwijzen" (default aan, lijst met keys uitklapbaar).

### 5.4 Environment-tab

- Variabelen met een referentie krijgen een chip `→ db · DATABASE_URL` (link-icoon) en een gemaskeerde preview; niet-opgeloste of niet-gekoppelde referentie krijgt een waarschuwingschip `Niet gekoppeld` met actie **Koppel en gebruik**.
- Variabelen aangemaakt door een link: chip `via link: db`.
- Referentie-picker: typen van `${{` opent een Base UI `Autocomplete` met alleen gekoppelde services en hun gepubliceerde sleutels. Geen cmdk/Radix.

### 5.5 Create-flows

- **Create database/Redis/bucket** binnen een project: optionele stap "Koppel aan applicaties" (checkboxlijst + de standaardvariabelen). Dit staat al in de briefing §7.3 ("bij Create database biedt Dockyard aan…"); controleer of `create-database-modal.tsx` dit al doet (niet gevonden bij een snelle grep).
- **Create application**: optionele stap "Koppel aan bestaande resources".

### 5.6 Project-pagina: zichtbaarheid (Fase 3)

- Resourcelijst: kolom/chip `2 links`.
- Toggle **List | Map**. Map = deterministische SVG in drie kolommen zoals Sevalla's overzicht: **Internet/Domeinen** → **Applicaties** → **Data** (db, redis, storage); doorgetrokken lijn = publiek verkeer met poortlabel (`:3000`), streeplijn = interne link, klik op een lijn opent "Remove connection". Geen graph-library (lean); eigen component met vaste layout (< 200 regels). Stippenraster als achtergrond, tokens uit het design.

### 5.7 ⌘K en toegankelijkheid

- Commandopalette: acties "Connect <app> to <db>…" en "Show connections of <service>".
- Echte `<button>`/`<a>`, `aria-label` op icon-only, de map heeft een lijstalternatief (zelfde data) voor toetsenbord/screenreader. Statuskleuren altijd met label.

### 5.8 Design-canvas: te tekenen schermen

| Scherm | Prioriteit |
|---|---|
| Application · Overview met Connections-kaart (gevuld + leeg) | Hoog |
| Dialoog Add connection (stap 1 + variabelen) | Hoog |
| Database detail met Connected applications + verwijder-dialoog | Hoog |
| Environment met referentie-chips en "Niet gekoppeld" | Middel |
| Project · Map-weergave | Laag (Fase 3) |

---

## 6. Fasering en acceptatiecriteria

**Fase 0 · Voorwaarden (Direct-track, klein)**
- Controleer en herstel: alleen apps **met domein** op `dy-edge` (`RunApp`). Test: twee apps in twee projecten kunnen elkaar niet via naam bereiken.
- Installer: `default-address-pools` in `daemon.json`; `dockyard doctor` meldt resterende netwerkruimte.

**Fase 1 · Links + gemak (Feature-track)**
- Migratie `service_links`, `variables.link_id`, `projects.links_mode` (default `open` voor bestaand).
- API (4.4), `consumerCatalog`, referentievalidatie (D6), backfill (D9), audit/events.
- UI 5.1–5.5 (zonder netwerkafdwinging; badge "Not enforced").
- Klaar als: DB koppelen aan app geeft `DATABASE_URL` als referentie, redeploy werkt, ontkoppelen ruimt variabelen op, referentie naar niet-gekoppelde service wordt geweigerd. Tests: `refs`/catalogus-filter, API 409/422/RBAC, backfill-idempotentie, UI-flow (Playwright).

**Fase 2 · Afdwingen**
- Per-link netwerken, blue/green-compatibel, Garage per link, reconciler (4.2), `links_mode = strict`.
- Klaar als: in `strict` kan app A de database van app B (zonder link) niet bereiken (integratietest met echte Docker), een redeploy onder load heeft geen mislukte requests, `kill -9 dockyard` laat links intact, reconciler herstelt een handmatig verwijderd link-netwerk binnen 60 s.

**Fase 3 · Zichtbaarheid**
- Project-Map (5.6), ⌘K-acties, link-teller in lijsten.

**Buiten scope (nu):** cross-project links, bidirectionele auto-link, private ports/TCP-proxy per link, links voor static sites, IP-allowlist voor externe databasetoegang (hoort bij het bestaande "Make public"-plan v0.2).

---

## 7. Risico's

| Risico | Mitigatie |
|---|---|
| Docker-netwerk-adresruimte raakt op | Address pools in installer, `doctor`-check, hergebruik van link-netwerken per paar (niet per variabele) |
| Breuk bij update (apps verliezen verbinding) | `open` als default voor bestaande projecten, backfill uit referenties, banner voor opt-in strict |
| Blue/green raakt link-alias | Nieuwe container eerst op link-netwerken vóór healthcheck; integratietest met load |
| Gebruiker verwacht dat links ook bestaande hardgecodeerde credentials afschermen | UI-tekst: afschermen betreft netwerkroute; credentials blijven nodig |
| Complexiteit UI | Eerst lijst + dialoog (Fase 1), kaart pas in Fase 3 |

## 8. Open beslissingen voor Joey

1. **Strict als default voor nieuwe projecten?** Voorstel: ja.
2. **Gericht of bidirectioneel** (D1)? Voorstel: gericht.
3. **Link-aanmaak ook voor `member`?** Voorstel: ja (viewer niet).
4. **Wil je dat ik de design-canvas aanvul** met de schermen uit 5.8 (Overview-kaart, dialoog, Connected applications)? Ik heb dat nog niet gedaan.
5. **Bevestig** dat `RunApp` alle apps op `dy-edge` zet in de huidige release (Fase 0). Ik las dit in de code, maar draaide het niet.

---

## Bijlage A: bronnen

- Sevalla, Networking (internal connections, private hostnames): https://docs.sevalla.com/applications/networking
- Sevalla, Database networking: https://docs.sevalla.com/databases/networking.md
- Sevalla, Connect to a database (internal vs external): https://docs.sevalla.com/databases/connect
- Sevalla, Environment variables (connectievariabelen, `${KEY}`-referenties): https://docs.sevalla.com/applications/environment-variables.md
- Sevalla API, Create internal connection: https://api-docs.sevalla.com/v3/applications/create-internal-connection
- Intern: `DOCKYARD_BRIEFING.md` §3.4, §3.6, §7.3, §8.3; code `internal/refs`, `internal/deploy/catalog.go`, `internal/docker/run.go`, `internal/docker/database.go`, `web/components/dockyard/service-connection-panel.tsx`, `application-detail-page.tsx`.
