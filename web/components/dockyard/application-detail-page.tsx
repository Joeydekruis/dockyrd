"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AddEnvVariableModal } from "@/components/dockyard/add-env-variable-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { ServiceLogsPanel } from "@/components/dockyard/service-logs-panel";
import { Button } from "@/components/dockyard/ui/button";
import { CommitChip } from "@/components/dockyard/ui/commit-chip";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { EnvVarRow } from "@/components/dockyard/ui/env-var-row";
import { Field } from "@/components/dockyard/ui/field";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
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
import {
  formatTimeAgo,
  projectPathLabel,
  resourceStatusDisplay,
} from "@/lib/resource-dashboard";
import {
  persistDetailTab,
  RESOURCE_DETAIL_CARD,
  readInitialDetailTab,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";
import {
  deploymentDurationSec,
  deploymentStatusDisplay,
  domainStatusDisplay,
  formatDeploymentWhen,
  formatDurationSec,
} from "@/lib/static-site-display";
import { useRouteParam } from "@/lib/use-route-param";

const DETAIL_TABS = [
  { id: "overview", label: "Overview" },
  { id: "deployments", label: "Deployments" },
  { id: "logs", label: "Logs" },
  { id: "environment", label: "Environment" },
  { id: "domains", label: "Domains" },
  { id: "settings", label: "Settings" },
] as const;

type TabId = (typeof DETAIL_TABS)[number]["id"];

type AppRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
  domains: ServiceDomain[];
  variables: ServiceVariable[];
};

function appSourceSummary(config: Record<string, unknown>): string {
  const source = config.source as
    | { type?: string; ref?: string; repo?: string; branch?: string }
    | undefined;
  if (source?.repo) {
    const branch = source.branch ?? "main";
    return `${source.repo} · ${branch}`;
  }
  if (source?.ref) return source.ref;
  if (source?.type) return source.type;
  return "—";
}

function appPort(config: Record<string, unknown>): string {
  const port = (config as { port?: number }).port;
  return port ? String(port) : "—";
}

function DeploymentRow({ deployment }: { deployment: Deployment }) {
  const st = deploymentStatusDisplay(deployment.status);
  const dur = deploymentDurationSec(deployment);
  const when = formatDeploymentWhen(
    deployment.finished_at ?? deployment.started_at ?? deployment.queued_at,
  );
  const meta = `${deployment.trigger}${dur !== undefined ? ` · ${formatDurationSec(dur)}` : ""}`;

  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-4 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <StatusBadge label={st.label} variant={st.variant} />
      <CommitChip
        sha={deployment.commit_sha ?? deployment.image_ref ?? "—"}
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
    </div>
  );
}

export function ApplicationDetailPage() {
  const slug = useRouteParam("slug");
  const [row, setRow] = useState<AppRow | null>(null);
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
        const { services } = await listServices(project.id, "app");
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
      setError(e instanceof Error ? e.message : "Failed to load application");
    } finally {
      setLoading(false);
    }
  }, [slug, revealedSecrets]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    setTab(
      readInitialDetailTab(
        DETAIL_TABS.map((t) => t.id),
        "overview",
      ),
    );
  }, []);

  function selectTab(next: TabId) {
    setTab(next);
    persistDetailTab(next);
  }

  const appStatus = useMemo(
    () =>
      row
        ? resourceStatusDisplay(row.service.desired_state, "application")
        : { label: "Unknown", variant: "idle" as const },
    [row],
  );

  const primaryHost = row?.domains.find((d) => d.is_primary)?.hostname;
  const latest = row?.deployments[0];

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

  const title = row?.service.name ?? "Application";

  return (
    <DashboardShell title={title} breadcrumb="Applications">
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
          Application not found.{" "}
          <Link href="/applications" className="underline">
            Back to applications
          </Link>
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
                  label={appStatus.label}
                  variant={appStatus.variant}
                />
              </div>
              <p
                className="mt-2 flex flex-wrap items-center gap-2 text-[13px] font-mono uppercase tracking-wide"
                style={{ color: "var(--muted-text)" }}
              >
                <span>{appSourceSummary(row.service.config)}</span>
                {primaryHost ? (
                  <>
                    <span>·</span>
                    <span className="normal-case">{primaryHost}</span>
                  </>
                ) : null}
              </p>
              <p
                className="mt-1 text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                <Link
                  href={`/projects/${row.project.slug}`}
                  className="hover:underline"
                  style={{ color: "var(--link-blue)" }}
                >
                  {projectPathLabel(row.project.name)}
                </Link>
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
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
                className={RESOURCE_DETAIL_CARD}
                style={resourceDetailCardStyle()}
                aria-label="Details"
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Details
                </h2>
                <dl className="mt-4">
                  <KeyValueRow
                    label="Source"
                    value={appSourceSummary(row.service.config)}
                  />
                  <KeyValueRow
                    label="Port"
                    value={appPort(row.service.config)}
                  />
                  <KeyValueRow
                    label="Project"
                    value={
                      <Link
                        href={`/projects/${row.project.slug}`}
                        className="hover:underline"
                        style={{ color: "var(--link-blue)" }}
                      >
                        {row.project.name}
                      </Link>
                    }
                  />
                  <KeyValueRow
                    label="Status"
                    value={
                      <StatusBadge
                        label={appStatus.label}
                        variant={appStatus.variant}
                      />
                    }
                  />
                  <KeyValueRow
                    label="Last deploy"
                    value={
                      latest?.commit_sha || latest?.image_ref
                        ? `${(latest.commit_sha ?? latest.image_ref ?? "").slice(0, 7)} · ${formatTimeAgo(latest.finished_at ?? latest.started_at) ?? "—"}`
                        : "—"
                    }
                  />
                </dl>
              </section>
              <section
                className={RESOURCE_DETAIL_CARD}
                style={resourceDetailCardStyle()}
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
                    <DeploymentRow key={d.id} deployment={d} />
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
              className={RESOURCE_DETAIL_CARD}
              style={resourceDetailCardStyle()}
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
                  <DeploymentRow key={d.id} deployment={d} />
                ))}
                {row.deployments.length === 0 ? (
                  <div className="py-8 text-center">
                    <EmptyState
                      title="No deployments"
                      description="Trigger a deploy to roll out a new version."
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

          {tab === "logs" ? (
            <ServiceLogsPanel serviceId={row.service.id} />
          ) : null}

          {tab === "environment" ? (
            <section
              className={RESOURCE_DETAIL_CARD}
              style={resourceDetailCardStyle()}
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
                    Injected at runtime in the container.
                  </p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setRevealedSecrets((v) => !v)}
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
                className={RESOURCE_DETAIL_CARD}
                style={resourceDetailCardStyle()}
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Domains
                </h2>
                {row.domains.length === 0 ? (
                  <p
                    className="mt-4 text-sm"
                    style={{ color: "var(--muted-text)" }}
                  >
                    No domains yet. Add a hostname to expose this app through
                    Caddy.
                  </p>
                ) : (
                  row.domains.map((d) => {
                    const st = domainStatusDisplay(d);
                    return (
                      <div
                        key={d.id}
                        className="mt-4 border-t pt-4 first:mt-0 first:border-t-0 first:pt-0"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <div className="flex flex-wrap items-center gap-2">
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
                        </div>
                      </div>
                    );
                  })
                )}
                <form
                  className="mt-6 flex flex-col gap-3 sm:flex-row sm:items-end"
                  onSubmit={(e) => void onAddDomain(e)}
                >
                  <Field
                    label="Custom domain"
                    value={newDomain}
                    onChange={(e) => setNewDomain(e.target.value)}
                    placeholder="app.example.com"
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
            <section
              className={RESOURCE_DETAIL_CARD}
              style={resourceDetailCardStyle()}
            >
              <h2 className="text-lg font-semibold tracking-tight">Settings</h2>
              <p
                className="mt-2 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                Image, GitHub, and health-check settings for container apps will
                expand here. Configure deploy source from the{" "}
                <Link
                  href={`/projects/${row.project.slug}`}
                  className="underline"
                  style={{ color: "var(--link-blue)" }}
                >
                  project page
                </Link>{" "}
                for now.
              </p>
              <dl className="mt-4">
                <KeyValueRow label="Slug" value={row.service.slug} />
                <KeyValueRow
                  label="Desired state"
                  value={row.service.desired_state}
                />
              </dl>
            </section>
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
