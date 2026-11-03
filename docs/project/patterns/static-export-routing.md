# Static export routing

Next.js uses `output: 'export'`. Dynamic segments prerender a single placeholder id `_`; the Go server maps real URLs to those files and RSC payloads.

**Contract:** briefing §4.2.

## Web

- `generateStaticParams() → [{ id: '_' }]`, `dynamicParams = false`
- Read real ids with **`useRouteParam('id')`** from pathname — not `useParams()` at hydration

## Go (`internal/server`)

- Route table: `/apps/<id>/logs` → `apps/_/logs/index.html` (+ `index.txt` payload)
- Unknown paths → `404.html`

## Fallback

If the pattern fails by end of M1, Vite + TanStack Router with same components — requires ADR.
