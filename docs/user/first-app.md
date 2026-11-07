# Your first app

This guide assumes dockyrd is installed and you completed setup.

## 1. Create a project

In the dashboard, open **Projects → Create project**. Projects group services and networks.

## 2. Add an application

**Create application** and choose a source:

- **GitHub** — connect the GitHub App, pick a repo, push to deploy (Railpack or Dockerfile).
- **Docker image** — deploy a public image (M1 path).
- **Static site** — build output from Git or upload (M2 path).

Set a subdomain or custom domain under **Domains**. dockyrd configures HTTPS via Caddy.

## 3. Add Postgres (optional)

**Databases → Create** provisions Postgres in Docker on the project network. Copy connection variables into your app under **Variables**, or use reference syntax `${{service-slug.VAR}}` between services.

## 4. Object storage (optional)

Install Garage from **Object storage**, create a bucket, and attach keys to apps that need S3-compatible uploads.

## 5. Deploy and verify

Trigger a deploy from the app detail or via `git push`. Watch build logs and the **Deployments** tab. Live traffic switches with zero downtime for container apps.

## Next.js

Use App Router with environment variables pointing at the managed Postgres URL. Ensure `output: 'standalone'` or Railpack defaults for Node apps — see briefing §4 for the recommended static-export pattern for the **dockyrd panel itself** (not necessarily your apps).

For DNS and TLS tips, see [troubleshooting](./troubleshooting.md).
