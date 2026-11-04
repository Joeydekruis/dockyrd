"use client";

import { ExternalLink } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AddEnvVariableModal } from "@/components/dockyard/add-env-variable-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { CommitChip } from "@/components/dockyard/ui/commit-chip";
import { DnsRecordBlock } from "@/components/dockyard/ui/dns-record-block";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { EnvVarRow } from "@/components/dockyard/ui/env-var-row";
import { Field } from "@/components/dockyard/ui/field";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Switch } from "@/components/dockyard/ui/switch";
import { Tabs } from "@/components/dockyard/ui/tabs";
import {
  createDeployment,
  createServiceDomain,
  type Deployment,
  listDeployments,
  listProjects,
  listServiceDomains,
  listServices,
  listServiceVariables,
  type Project,
  putServiceVariables,
  type Service,
  type ServiceDomain,
  type ServiceVariable,
} from "@/lib/api";
import { formatTimeAgo, resourceStatusDisplay } from "@/lib/resource-dashboard";
import {
  deploymentDurationSec,
  deploymentStatusDisplay,
  domainStatusDisplay,
  formatDeploymentWhen,
  formatDurationSec,
  parseStaticSiteConfig,
  primaryDomainHost,
  splitBuildCmd,
  visitSiteHref,
} from "@/lib/static-site-display";
import { useRouteParam } from "@/lib/use-route-param";

const DETAIL_TABS = [
  { id: "overview", label: "Overview" },
  { id: "deployments", label: "Deployments" },
  { id: "logs", label: "Logs" },
  { id: "analytics", label: "Analytics" },
  { id: "environment", label: "Environment" },
  { id: "domains", label: "Domains" },
  { id: "settings", label: "Settings" },
] as const;

type TabId = (typeof DETAIL_TABS)[number]["id"];

type SiteRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
  domains: ServiceDomain[];
  variables: ServiceVariable[];
};

function cardClassName() {
  return "rounded-[20px] border px-6 py-5";
}

function DeploymentListItem({
  deployment,
  siteSlug,
}: {
  deployment: Deployment;
  siteSlug: string;
}) {
  const st = deploymentStatusDisplay(deployment.status);
  const dur = deploymentDurationSec(deployment);
  const when = formatDeploymentWhen(
    deployment.finished_at ?? deployment.started_at ?? deployment.queued_at,
  );
  const meta = `${deployment.trigger}${dur !== undefined ? ` · ${formatDurationSec(dur)}` : ""}`;

  return (
    <Link
      href={`/static-sites/${siteSlug}/deployments/${deployment.id}`}
      className="flex flex-wrap items-center gap-3 border-t py-4 first:border-t-0 hover:bg-[var(--hover-row)]"
      style={{ borderColor: "var(--border)" }}
    >
      <StatusBadge label={st.label} variant={st.variant} />
      <CommitChip
        sha={deployment.commit_sha ?? "pending"}
        message={deployment.branch ?? deployment.trigger}
      />
      <span
        className="min-w-0 flex-1 truncate text-[13px]"
        style={{ color: "var(--muted-text)" }}
      >
        {meta}
      </span>
      <span
        className="shrink-0 font-mono text-[12px] tabular-nums"
        style={{ color: "var(--muted-text)" }}
      >
        {when}
      </span>
    </Link>
  );
}

export function StaticSiteDetailPage() {
  const slug = useRouteParam("slug");
  const [row, setRow] = useState<SiteRow | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<TabId>("overview");
  const [revealedSecrets, setRevealedSecrets] = useState(false);
  const [addVarOpen, setAddVarOpen] = useState(false);
  const [newDomain, setNewDomain] = useState("");
  const [domainBusy, setDomainBusy] = useState(false);
  const [deployBusy, setDeployBusy] = useState(false);

  const refresh = useCallback(async () => {
    if (!slug) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const { projects } = await listProjects();
      for (const project of projects) {
        const { services } = await listServices(project.id, "static_site");
        const service = services.find((s) => s.slug === slug);
        if (!service) continue;
        const [{ deployments }, { domains }, { variables }] = await Promise.all(
          [
            listDeployments(service.id),
            listServiceDomains(service.id),
            listServiceVariables(service.id, revealedSecrets),
          ],
        );
        setRow({ project, service, deployments, domains, variables });
        setError(null);
        setLoading(false);
        return;
      }
      setRow(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load static site");
    } finally {
      setLoading(false);
    }
  }, [slug, revealedSecrets]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("tab") as TabId | null;
    if (t && DETAIL_TABS.some((x) => x.id === t)) {
      setTab(t);
    }
  }, []);

  function selectTab(next: TabId) {
    setTab(next);
    const url = new URL(window.location.href);
    url.searchParams.set("tab", next);
    window.history.replaceState(
      null,
      "",
      `${url.pathname}?${url.searchParams.toString()}`,
    );
  }

  const cfg = useMemo(
    () => (row ? parseStaticSiteConfig(row.service) : null),
    [row],
  );
  const siteStatus = useMemo(
    () =>
      row
        ? resourceStatusDisplay(row.service.desired_state, "static")
        : { label: "Unknown", variant: "idle" as const },
    [row],
  );
  const primaryHost = row ? primaryDomainHost(row.domains) : undefined;
  const visitHref = visitSiteHref(primaryHost);
  const latest = row?.deployments[0];
  const buildParts = cfg ? splitBuildCmd(cfg.buildCmd ?? "") : null;

  async function onDeployNow() {
    if (!row) return;
    setDeployBusy(true);
    setError(null);
    try {
      await createDeployment(row.service.id);
      await refresh();
      selectTab("deployments");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Deploy failed");
    } finally {
      setDeployBusy(false);
    }
  }

  async function onAddVariables(
    entries: {
      key: string;
      value: string;
      is_secret: boolean;
      at_build: boolean;
      at_runtime: boolean;
    }[],
  ) {
    if (!row) return;
    const merged = [...row.variables];
    for (const e of entries) {
      if (!merged.some((v) => v.key === e.key)) {
        merged.push({
          key: e.key,
          value: e.value,
          is_secret: e.is_secret,
          at_build: e.at_build,
          at_runtime: e.at_runtime,
        });
      }
    }
    await putServiceVariables(
      row.service.id,
      merged.map((v) => ({
        key: v.key,
        value: v.value ?? "",
        is_secret: v.is_secret,
        at_build: v.at_build,
        at_runtime: v.at_runtime,
      })),
    );
    await refresh();
  }

  async function onAddDomain(e: React.FormEvent) {
    e.preventDefault();
    if (!row) return;
    const host = newDomain.trim();
    if (!host) return;
    setDomainBusy(true);
    setError(null);
    try {
      await createServiceDomain(row.service.id, {
        hostname: host,
        primary: true,
      });
      setNewDomain("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not add domain");
    } finally {
      setDomainBusy(false);
    }
  }

  async function ensurePlatformDomain() {
    if (!row) return;
    setDomainBusy(true);
    try {
      await createServiceDomain(row.service.id, {
        platform: true,
        primary: !primaryHost,
      });
      await refresh();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Could not add platform domain",
      );
    } finally {
      setDomainBusy(false);
    }
  }

  const title = row?.service.name ?? "Static site";

  return (
    <DashboardShell title={title} breadcrumb="Static sites">
      {loading ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : error ? (
        <p className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : !row ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Static site not found.
        </p>
      ) : (
        <div className="space-y-6">
          <header className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div className="flex flex-wrap items-center gap-3">
                <h1 className="text-2xl font-semibold tracking-tight">
                  {row.service.name}
                </h1>
                <StatusBadge
                  label={siteStatus.label}
                  variant={siteStatus.variant}
                />
              </div>
              <p
                className="mt-2 flex flex-wrap items-center gap-2 text-[13px] font-mono uppercase tracking-wide"
                style={{ color: "var(--muted-text)" }}
              >
                <span>{cfg?.source?.repo ?? "—"}</span>
                <span>·</span>
                <span>{cfg?.source?.branch ?? "main"}</span>
                {primaryHost ? (
                  <>
                    <span>·</span>
                    <a
                      href={visitHref}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 normal-case hover:underline"
                      style={{ color: "var(--link-blue)" }}
                    >
                      {primaryHost}
                      <ExternalLink className="size-3.5" strokeWidth={1.8} />
                    </a>
                  </>
                ) : null}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              {visitHref ? (
                <Button href={visitHref} variant="outline" target="_blank">
                  Visit site
                </Button>
              ) : null}
              <Button
                type="button"
                disabled={deployBusy}
                onClick={() => void onDeployNow()}
              >
                Deploy now
              </Button>
            </div>
          </header>

          <Tabs
            activeId={tab}
            onChange={(id) => selectTab(id as TabId)}
            tabs={[...DETAIL_TABS]}
          />

          {tab === "overview" ? (
            <div className="grid gap-6 lg:grid-cols-2">
              <section
                className={cardClassName()}
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
                aria-label="Details"
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Details
                </h2>
                <dl className="mt-4">
                  <KeyValueRow
                    label="Domain"
                    value={
                      primaryHost ? (
                        <a
                          href={visitHref}
                          className="hover:underline"
                          style={{ color: "var(--link-blue)" }}
                        >
                          {primaryHost}
                        </a>
                      ) : (
                        "—"
                      )
                    }
                  />
                  <KeyValueRow
                    label="Status"
                    value={
                      <StatusBadge
                        label={siteStatus.label}
                        variant={siteStatus.variant}
                      />
                    }
                  />
                  <KeyValueRow
                    label="Source"
                    value={`${cfg?.source?.repo ?? "—"} · ${(cfg?.source?.branch ?? "main").toUpperCase()}`}
                  />
                  <KeyValueRow
                    label="Last deploy"
                    value={
                      latest?.commit_sha
                        ? `${latest.commit_sha.slice(0, 7)} · ${formatTimeAgo(latest.finished_at ?? latest.started_at) ?? "—"}`
                        : "—"
                    }
                  />
                </dl>
                {visitHref ? (
                  <Button
                    href={visitHref}
                    variant="outline"
                    className="mt-4"
                    target="_blank"
                  >
                    Visit site
                  </Button>
                ) : null}
              </section>
              <section
                className={cardClassName()}
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
                aria-label="Recent deployments"
              >
                <div className="flex items-center justify-between gap-3">
                  <h2 className="text-lg font-semibold tracking-tight">
                    Recent deployments
                  </h2>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={deployBusy}
                    onClick={() => void onDeployNow()}
                  >
                    Deploy now
                  </Button>
                </div>
                <div className="mt-2">
                  {row.deployments.slice(0, 5).map((d) => (
                    <DeploymentListItem
                      key={d.id}
                      deployment={d}
                      siteSlug={row.service.slug}
                    />
                  ))}
                  {row.deployments.length === 0 ? (
                    <p
                      className="py-6 text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      No deployments yet.
                    </p>
                  ) : null}
                </div>
              </section>
            </div>
          ) : null}

          {tab === "deployments" ? (
            <section
              className={cardClassName()}
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
            >
              <div className="flex flex-wrap items-center justify-between gap-3">
                <h2 className="text-lg font-semibold tracking-tight">
                  Deployments
                </h2>
                <Button
                  type="button"
                  disabled={deployBusy}
                  onClick={() => void onDeployNow()}
                >
                  Deploy now
                </Button>
              </div>
              <div className="mt-2">
                {row.deployments.map((d) => (
                  <DeploymentListItem
                    key={d.id}
                    deployment={d}
                    siteSlug={row.service.slug}
                  />
                ))}
                {row.deployments.length === 0 ? (
                  <div className="py-8 text-center">
                    <EmptyState
                      title="No deployments"
                      description="Trigger a deploy to build and publish this site."
                    />
                    <Button
                      type="button"
                      className="mt-4"
                      disabled={deployBusy}
                      onClick={() => void onDeployNow()}
                    >
                      Deploy now
                    </Button>
                  </div>
                ) : null}
              </div>
            </section>
          ) : null}

          {tab === "logs" || tab === "analytics" ? (
            <EmptyState
              title={tab === "logs" ? "Access logs" : "Analytics"}
              description={
                tab === "logs"
                  ? "HTTP access logs for static sites will appear here once Caddy log streaming is wired to the panel."
                  : "Per-site traffic charts will appear here in a future release."
              }
            />
          ) : null}

          {tab === "environment" ? (
            <section
              className={cardClassName()}
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h2 className="text-lg font-semibold tracking-tight">
                    Environment variables
                  </h2>
                  <p
                    className="mt-1 text-sm"
                    style={{ color: "var(--muted-text)" }}
                  >
                    Injected at build time.
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setRevealedSecrets((v) => !v);
                    }}
                  >
                    {revealedSecrets ? "Hide secrets" : "Reveal secrets"}
                  </Button>
                  <Button type="button" onClick={() => setAddVarOpen(true)}>
                    Add variable
                  </Button>
                </div>
              </div>
              <div className="mt-4">
                {row.variables.map((v) => (
                  <EnvVarRow
                    key={v.key}
                    name={v.key}
                    value={v.value ?? ""}
                    secret={v.is_secret}
                  />
                ))}
                {row.variables.length === 0 ? (
                  <p
                    className="py-4 text-sm"
                    style={{ color: "var(--muted-text)" }}
                  >
                    No variables yet.
                  </p>
                ) : null}
              </div>
            </section>
          ) : null}

          {tab === "domains" ? (
            <div className="space-y-6">
              <section
                className={cardClassName()}
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Platform domain
                </h2>
                {row.domains
                  .filter((d) => d.is_platform)
                  .map((d) => {
                    const st = domainStatusDisplay(d);
                    return (
                      <div
                        key={d.id}
                        className="mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-4"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <span className="font-mono text-sm">{d.hostname}</span>
                        <StatusBadge label={st.label} variant={st.variant} />
                      </div>
                    );
                  })}
                {row.domains.every((d) => !d.is_platform) ? (
                  <div className="mt-4">
                    <p
                      className="text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      No platform domain on file — HTTPS routing may not work
                      until one is added.
                    </p>
                    <Button
                      type="button"
                      variant="outline"
                      className="mt-3"
                      disabled={domainBusy}
                      onClick={() => void ensurePlatformDomain()}
                    >
                      Add platform domain
                    </Button>
                  </div>
                ) : null}
              </section>
              <section
                className={cardClassName()}
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <h2 className="text-lg font-semibold tracking-tight">
                    Custom domains
                  </h2>
                </div>
                {row.domains
                  .filter((d) => !d.is_platform)
                  .map((d) => {
                    const st = domainStatusDisplay(d);
                    return (
                      <div
                        key={d.id}
                        className="mt-4 border-t pt-4"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <div className="flex flex-wrap items-center gap-3">
                          <span className="font-mono text-sm">
                            {d.hostname}
                          </span>
                          {d.is_primary ? (
                            <span
                              className="font-mono text-[11px] uppercase tracking-widest"
                              style={{ color: "var(--muted-text)" }}
                            >
                              Primary
                            </span>
                          ) : null}
                          <StatusBadge label={st.label} variant={st.variant} />
                          {d.cert_status === "active" ? (
                            <span
                              className="text-[13px]"
                              style={{ color: "var(--muted-text)" }}
                            >
                              HTTPS · Let&apos;s Encrypt
                            </span>
                          ) : null}
                        </div>
                        {st.label === "Pending DNS" ? (
                          <div className="mt-3 space-y-2">
                            <DnsRecordBlock
                              type="A"
                              name="@"
                              value="Your server public IPv4"
                            />
                            <p
                              className="text-[13px]"
                              style={{ color: "var(--muted-text)" }}
                            >
                              Point DNS at this server, then reload this page.
                              Certificates issue automatically once HTTP-01 can
                              reach Caddy on ports 80 and 443.
                            </p>
                          </div>
                        ) : null}
                      </div>
                    );
                  })}
                <form
                  className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-end"
                  onSubmit={(e) => void onAddDomain(e)}
                >
                  <Field
                    label="Custom domain"
                    value={newDomain}
                    onChange={(e) => setNewDomain(e.target.value)}
                    placeholder="clientx.nl"
                    className="flex-1"
                  />
                  <Button type="submit" disabled={domainBusy}>
                    Add domain
                  </Button>
                </form>
              </section>
            </div>
          ) : null}

          {tab === "settings" ? (
            <div className="space-y-6">
              {[
                {
                  title: "General",
                  fields: [{ label: "Name", value: row.service.name }],
                },
                {
                  title: "Source",
                  fields: [
                    { label: "Repository", value: cfg?.source?.repo ?? "—" },
                    { label: "Branch", value: cfg?.source?.branch ?? "main" },
                  ],
                },
                {
                  title: "Build settings",
                  fields: [
                    {
                      label: "Install command",
                      value: buildParts?.install ?? "—",
                    },
                    { label: "Build command", value: buildParts?.build ?? "—" },
                    {
                      label: "Output directory",
                      value: cfg?.outputDir ?? "dist",
                    },
                  ],
                },
              ].map((block) => (
                <section
                  key={block.title}
                  className={cardClassName()}
                  style={{
                    borderColor: "var(--border)",
                    background: "var(--surface)",
                  }}
                >
                  <h2 className="text-lg font-semibold tracking-tight">
                    {block.title}
                  </h2>
                  <div className="mt-4 space-y-4">
                    {block.fields.map((f) => (
                      <Field
                        key={f.label}
                        label={f.label}
                        value={f.value}
                        readOnly
                      />
                    ))}
                  </div>
                </section>
              ))}
              <section
                className={cardClassName()}
                style={{
                  borderColor: "var(--border)",
                  background: "var(--surface)",
                }}
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Deploys
                </h2>
                <div className="mt-4 space-y-4">
                  <Switch
                    label="Auto-deploy"
                    checked={cfg?.autoDeploy ?? false}
                    onChange={() => {}}
                    disabled
                  />
                  <p
                    className="text-[13px]"
                    style={{ color: "var(--muted-text)" }}
                  >
                    Deploy on every push to {cfg?.source?.branch ?? "main"}.
                  </p>
                  <Switch
                    label="Single-page app fallback"
                    checked={cfg?.spaFallback ?? false}
                    onChange={() => {}}
                    disabled
                  />
                </div>
              </section>
            </div>
          ) : null}

          <AddEnvVariableModal
            open={addVarOpen}
            existingKeys={row.variables.map((v) => v.key)}
            onClose={() => setAddVarOpen(false)}
            onAdd={onAddVariables}
          />
        </div>
      )}
    </DashboardShell>
  );
}
