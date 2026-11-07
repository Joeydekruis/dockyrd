# Next.js on dockyrd

dockyrd deploys Node apps with **Railpack** (Git push) or a **Dockerfile**. Next.js App Router projects work well with Railpack defaults.

## Recommended stack

- Managed **Postgres** in dockyrd (see [first app](./first-app.md))
- Environment variables for `DATABASE_URL` at **runtime**
- GitHub App connected under **Instance settings**

## Push-to-deploy

1. Connect the GitHub App and grant access to your repository.
2. **Create application** → GitHub → select the repo and branch.
3. Set the public hostname under **Domains**.
4. Push to the tracked branch; watch build logs on the deployment.

Railpack detects Next.js and runs `next build` with the appropriate start command for production.

## Dockerfile alternative

If you need full control, add a Dockerfile to the repo root:

```dockerfile
FROM node:22-bookworm-slim AS build
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile
COPY . .
RUN pnpm build

FROM node:22-bookworm-slim
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

Enable **standalone** output in `next.config.ts` when using this pattern:

```ts
const nextConfig = {
  output: "standalone",
};
export default nextConfig;
```

## Static export (not for SSR apps)

The **Dockyrd dashboard** itself uses `output: 'export'` because it is a static admin UI embedded in the Go binary. User-facing Next.js apps with SSR, server actions, or dynamic routes should **not** use static export — deploy them as Node containers instead.

## Variables

| Variable | Typical source |
|----------|----------------|
| `DATABASE_URL` | Postgres service connection string |
| `NEXTAUTH_SECRET` / session secrets | dockyrd secret variable |
| Public URL | Derived from assigned domain |

Use reference variables (`${{postgres-slug.DATABASE_URL}}`) between services in the same project.

## Troubleshooting

Build failures often come from missing env at build time — mark required vars as available during build in the service settings. See [troubleshooting](./troubleshooting.md).
