# Dockyrd: Terminal-briefing (web-terminal)

> Status: korte bouwbriefing · Datum: 7 november 2026 · Eigenaar: Joey de Kruis
> Bron design: Claude Design-canvas, artboard `Terminal.dc.html`
> Let op: in `DOCKYARD_BRIEFING` §1.2 en §15 staat de web-terminal in **v0.2** ("WebSocket + security review"). Deze briefing beschrijft de smalle versie die we eerder willen doen. Pas die twee plekken aan als dit doorgaat.

---

## 1. Doel

Vanuit het dashboard een shell openen om dingen te testen en te debuggen, zonder SSH. Veilig genoeg om aan teamleden te geven en zonder dat één gekaapte sessie meteen root op de server betekent.

## 2. Drie soorten sessies

| Sessie | Wat het is | Standaard | Wie |
|---|---|---|---|
| **Test shell** | Wegwerpcontainer op de server, in het netwerk van een gekozen project. Gereedschap: `curl`, `dig`, `ping`, `psql`, `valkey-cli`. Geen host-mounts, geen Docker-socket. Wordt verwijderd bij sluiten van de tab. | Aan | admin+ |
| **Container-shell** | `docker exec` in een app- of databasecontainer. Draait als de container-user, niet als root. Databases krijgen snelknoppen "Open psql" / "Open valkey-cli". | Aan | admin+ |
| **Server-shell (host)** | Root-shell op de host. Hier draait ook de `dockyrd` CLI. | **Uit** (opt-in) | alleen owner |

Waarom deze indeling: het panel draait al als root met de Docker-socket, dus een host-shell voegt geen nieuwe rechten toe, maar maakt misbruik wel direct. De test shell dekt het "dingen testen"-gebruik zonder dat risico. Voor de host blijft SSH een volwaardig alternatief.

## 3. Veiligheid

- **Toegang:** owner en admin. Member en viewer krijgen geen terminal (een shell toont ook environment-variabelen).
- **Host-shell:** instance-instelling, standaard uit. Zet de owner hem aan, dan eerst opnieuw bevestigen met 2FA, daarna elke sessie met een waarschuwingsstrook ("Root on the host…"). Naam van de instelling is een voorstel: `terminal_allow_host`.
- **WebSocket:** kortlevend, eenmalig ticket plus **Origin-check**. Er is nu nog geen CSRF/Origin-check op cookie-requests (zie `FEATURES_V01_BRIEFING` §9, punt 2). Dat moet eerst, anders is dit een risico voor cross-site WebSocket hijacking.
- **Audit:** per sessie wie, welk doel (test/container/host), start en stop. Geen toetsaanslagen opslaan (wachtwoorden).
- **Limieten:** idle-timeout (voorstel 15 min), maximale sessieduur, maximum aantal gelijktijdige sessies per gebruiker.
- **Test shell:** `cap_drop: ALL`, `no-new-privileges`, pids-limiet, CPU/geheugen-limiet, geen host-poorten. Zelfde hardening als app-containers (`SECURITY_BRIEFING` §4.8).
- **Containers zijn al gehard** (`cap_drop: ALL`, pids-limiet 512), dus een shell erin kan weinig.

## 4. UI (zoals in het design)

- Termius-achtig: tabbalk met sessies, plus-knop voor een nieuwe sessie, kruisje om te sluiten, statusbalk onderaan.
- Plus-menu in drie groepen: **Test** (Test shell, label "No host access"), **Containers**, **Advanced** (Server-shell).
- Tabs: Server krijgt badge `ROOT` (oranje), Test shell badge `SANDBOX` (groen).
- Waarschuwingsstrook boven de terminal bij Server-shell (oranje) en Test shell (groen, legt uit dat de container verdwijnt).
- Settings-knop (tandwiel) opent een popup met **4 thema's**: Dockyrd dark (standaard), Dockyrd light, Hacker green (dark), Normal light. Het thema geldt voor alle sessies.
- Geen shell in de image (distroless)? Duidelijke melding in plaats van een foutcode.

## 5. Niet in v1

- Bestanden uploaden/downloaden, port-forwarding, gedeelde sessies, sessie-opnames.
- Vrije commando's via de API (geen "run dit commando"-endpoint).
- Multi-server (komt met `dockyard agent`, v0.3).

## 6. Open punten

1. **Design:** de host-shell staat nu altijd aan en is de standaardtab. Met opt-in moet er een staat zijn voor "uitgeschakeld" (uitleg + link naar instelling), en de standaardtab wordt de Test shell of een container.
2. **Test-shell image:** welk image en welke tools? Voorstel: klein Alpine-image met de genoemde tools. Niet besloten.
3. **Thema opslaan:** voorstel per gebruiker in SQLite. Niet besloten.
4. **Waar draait de test shell?** Op het projectnetwerk (`links_mode=strict`): alleen resources die gelinkt zijn aan dat project zijn bereikbaar. Controleren of dat past bij `RESOURCE_LINKING_BRIEFING`.
5. **Roadmap:** web-terminal van v0.2 naar v0.1.x, of toch na de security-review laten?
