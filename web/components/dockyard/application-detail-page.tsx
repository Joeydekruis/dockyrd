"use client";

import { Monitor } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { ApplicationConnectionsFlow } from "@/components/dockyard/application-connections-flow";
import { ApplicationEnvironmentTab } from "@/components/dockyard/application-environment-tab";
import { ApplicationProcessesTab } from "@/components/dockyard/application-processes-tab";
import { ApplicationSettingsTab } from "@/components/dockyard/application-settings-tab";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DeploymentResourceStrip } from "@/components/dockyard/deployment-resource-strip";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { PendingEnvRedeployBanner } from "@/components/dockyard/pending-env-redeploy-banner";
import { ServiceDomainsTab } from "@/components/dockyard/service-domains-tab";
import { ServicePreviewsTab } from "@/components/dockyard/service-previews-tab";
import { ServiceSecurityTab } from "@/components/dockyard/service-security-tab";
import {
  ServiceAccessLogsPanel,
  ServiceTrafficAnalyticsPanel,
} from "@/components/dockyard/service-traffic-panel";
import { ApplicationDeploymentTable } from "@/components/dockyard/ui/application-deployment-table";
import { Button } from "@/components/dockyard/ui/button";
import { ConfirmDialog } from "@/components/dockyard/ui/confirm-dialog";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import {
  SettingsBranchPill,
  SettingsCopyPill,
} from "@/components/dockyard/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Tabs } from "@/components/dockyard/ui/tabs";
import {
  createDeployment,
  type Deployment,
  listDeployments,
  listProjects,
  listServiceDomains,
  listServiceLinks,
  listServices,
  listServiceVariables,
  type Project,
  putServiceVariables,
  type Service,
  type ServiceDomain,
  type ServiceLink,
  type ServiceVariable,
} from "@/lib/api";
import { hasDeploymentResourceMetrics } from "@/lib/build-resilience";
import { hasPendingEnvRedeploy } from "@/lib/env-var-ref";
import { isUnassignedProject } from "@/lib/project-optional";
import { applicationDetailBreadcrumbs } from "@/lib/resource-breadcrumbs";
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
  deploymentCommitShaShort,
  deploymentIsInProgress,
  deploymentSourceTitle,
  primaryDomainHost,
  staticSiteDeploymentStatusDisplay,
  visitSiteHref,
} from "@/lib/static-site-display";
import { useDeploymentPoll } from "@/lib/use-deployment-poll";
import { useRouteParam } from "@/lib/use-route-param";

const DETAIL_TABS = [
  { id: "overview", label: "Overview" },
  { id: "deployments", label: "Deployments" },
  { id: "processes", label: "Processes" },
  { id: "environment", label: "Environment" },
  { id: "domains", label: "Domains" },
  { id: "logs", label: "Logs" },
  { id: "analytics", label: "Analytics" },
  { id: "previews", label: "Previews" },
  { id: "security", label: "Security" },
  { id: "settings", label: "Settings" },
] as const;

type TabId = (typeof DETAIL_TABS)[number]["id"];

type AppRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
  domains: ServiceDomain[];
  variables: ServiceVariable[];
  links: ServiceLink[];
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

function appPort(config: Record<string, unknown>): number {
  const port = (config as { port?: number }).port;
  return port && port > 0 ? port : 80;
}

export function ApplicationDetailPage() {
  const router = useRouter();
  const slug = useRouteParam("slug");
  const [row, setRow] = useState<AppRow | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<TabId>("overview");
  const [revealedSecrets, setRevealedSecrets] = useState(false);
  const [deployBusy, setDeployBusy] = useState(false);
  const [deployConfirmOpen, setDeployConfirmOpen] = useState(false);
  const [bannerDismissed, setBannerDismissed] = useState(false);

  const refresh = useCallback(
    async (opts?: { silent?: boolean }) => {
      if (!slug) {
        setLoading(false);
        return;
      }
      if (!opts?.silent) setLoading(true);
      try {
        const { projects } = await listProjects();
        for (const project of projects) {
          const { services } = await listServices(project.id, "app");
          const service = services.find((s) => s.slug === slug);
          if (!service) continue;
          const [{ deployments }, { domains }, { variables }, linksRes] =
            await Promise.all([
              listDeployments(service.id),
              listServiceDomains(service.id),
              listServiceVariables(service.id, revealedSecrets),
              listServiceLinks(service.id),
            ]);
          setRow({
            project,
            service,
            deployments,
            domains,
            variables,
            links: linksRes.outgoing,
          });
          setError(null);
          if (!opts?.silent) setLoading(false);
          return;
        }
        setRow(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load application");
      } finally {
        if (!opts?.silent) setLoading(false);
      }
    },
    [slug, revealedSecrets],
  );

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const t = params.get("tab") as TabId | null;
    if (t && DETAIL_TABS.some((x) => x.id === t)) {
      setTab(t);
    } else {
      setTab(
        readInitialDetailTab(
          DETAIL_TABS.map((x) => x.id),
          "overview",
        ),
      );
    }
  }, []);

  function selectTab(next: TabId) {
    setTab(next);
    persistDetailTab(next);
    const url = new URL(window.location.href);
    url.searchParams.set("tab", next);
    window.history.replaceState(
      null,
      "",
      `${url.pathname}?${url.searchParams.toString()}`,
    );
  }

  const linkedSlugs = useMemo(
    () => new Set(row?.links.map((l) => l.peer.slug) ?? []),
    [row?.links],
  );

  const appStatus = useMemo(
    () =>
      row
        ? resourceStatusDisplay(row.service.desired_state, "application")
        : { label: "Unknown", variant: "idle" as const },
    [row],
  );

  const primaryHost = row ? primaryDomainHost(row.domains) : undefined;
  const visitHref = visitSiteHref(primaryHost);
  const latest = row?.deployments[0];

  const deploymentsPollActive = useMemo(() => {
    if (!row) return false;
    const inProgress = row.deployments.some((d) =>
      deploymentIsInProgress(d.status),
    );
    if (!inProgress) return false;
    return tab === "deployments" || tab === "overview";
  }, [row, tab]);

  const pollDeployments = useCallback(() => {
    void refresh({ silent: true });
  }, [refresh]);

  useDeploymentPoll(deploymentsPollActive, pollDeployments);
  const port = row ? appPort(row.service.config) : 80;
  const source = row?.service.config.source as
    | { repo?: string; branch?: string }
    | undefined;

  const showOverviewBanner = useMemo(() => {
    if (!row || bannerDismissed) return false;
    const live = row.deployments[0];
    const liveAt =
      live?.finished_at ?? live?.started_at ?? live?.queued_at ?? 0;
    return hasPendingEnvRedeploy({
      variables: row.variables,
      linkedSlugs,
      serviceUpdatedAt: row.service.updated_at,
      liveDeploymentAt: liveAt,
    });
  }, [bannerDismissed, linkedSlugs, row]);

  function requestDeploy() {
    setDeployConfirmOpen(true);
  }

  async function onDeployNow() {
    if (!row) return;
    setDeployBusy(true);
    setError(null);
    try {
      const dep = await createDeployment(row.service.id);
      router.push(`/applications/${row.service.slug}/deployments/${dep.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Deploy failed");
    } finally {
      setDeployBusy(false);
    }
  }

  function confirmDeploy() {
    setDeployConfirmOpen(false);
    void onDeployNow();
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

  const title = row?.service.name ?? "Application";
  const breadcrumbItems = useMemo(
    () =>
      applicationDetailBreadcrumbs(
        row?.project
          ? { name: row.project.name, slug: row.project.slug }
          : undefined,
        title,
      ),
    [row?.project, title],
  );

  return (
    <DashboardShell title={title} breadcrumbItems={breadcrumbItems}>
      {loading ? (
        <DockyrdPageLoading />
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
          {showOverviewBanner ? (
            <PendingEnvRedeployBanner
              onDeploy={requestDeploy}
              deployBusy={deployBusy}
              onDismiss={() => setBannerDismissed(true)}
            />
          ) : null}

          <header className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex flex-wrap items-start gap-4">
              <span
                className="flex size-14 shrink-0 items-center justify-center rounded-2xl"
                style={{ background: "#E3F0FF", color: "#0B57C9" }}
                aria-hidden
              >
                <Monitor className="size-7" strokeWidth={1.8} />
              </span>
              <div>
                <div className="flex flex-wrap items-center gap-3">
                  <h1 className="text-3xl font-semibold tracking-tight">
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
                  {isUnassignedProject(row.project.slug) ? (
                    <span>
                      {projectPathLabel(row.project.name, row.project.slug)}
                    </span>
                  ) : (
                    <Link
                      href={`/projects/${row.project.slug}`}
                      className="hover:underline"
                      style={{ color: "var(--link-blue)" }}
                    >
                      {projectPathLabel(row.project.name, row.project.slug)}
                    </Link>
                  )}
                </p>
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              <Button
                type="button"
                variant="outline"
                disabled={deployBusy}
                onClick={requestDeploy}
              >
                {deployBusy ? "Deploying…" : "Deploy now"}
              </Button>
              {visitHref ? (
                <Button
                  href={visitHref}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Visit site
                </Button>
              ) : null}
            </div>
          </header>

          <Tabs
            activeId={tab}
            onChange={(id) => selectTab(id as TabId)}
            tabs={[...DETAIL_TABS]}
          />

          {tab === "overview" ? (
            <div className="space-y-6">
              {latest ? (
                <section
                  className={RESOURCE_DETAIL_CARD}
                  style={resourceDetailCardStyle()}
                  aria-label="Current deployment"
                >
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div>
                      <h2 className="text-lg font-semibold tracking-tight">
                        Current deployment
                      </h2>
                      <p
                        className="mt-1 text-sm"
                        style={{ color: "var(--muted-text)" }}
                      >
                        {deploymentSourceTitle(latest)}
                      </p>
                    </div>
                    <StatusBadge
                      label={
                        staticSiteDeploymentStatusDisplay(latest.status).label
                      }
                      variant={
                        staticSiteDeploymentStatusDisplay(latest.status).variant
                      }
                    />
                  </div>
                  <div className="mt-4 flex flex-wrap items-center gap-2 font-mono text-[12px]">
                    <SettingsBranchPill
                      branch={latest.branch?.trim() || "main"}
                    />
                    <SettingsCopyPill
                      value={deploymentCommitShaShort(latest)}
                    />
                    <span style={{ color: "var(--muted-text)" }}>
                      {formatTimeAgo(
                        latest.finished_at ??
                          latest.started_at ??
                          latest.queued_at,
                      ) ?? "—"}
                    </span>
                  </div>
                  {hasDeploymentResourceMetrics(latest) ? (
                    <DeploymentResourceStrip
                      deployment={latest}
                      className="mt-4"
                    />
                  ) : null}
                  <Button
                    type="button"
                    variant="outline"
                    className="mt-4"
                    href={`/applications/${row.service.slug}/deployments/${latest.id}`}
                  >
                    View deployment
                  </Button>
                </section>
              ) : null}

              <ApplicationConnectionsFlow
                service={row.service}
                project={row.project}
                domains={row.domains}
                port={port}
                onChanged={() => void refresh()}
                onGoDomains={() => selectTab("domains")}
                onGoLogs={() => selectTab("logs")}
              />

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
                    onClick={() => selectTab("deployments")}
                  >
                    View all
                  </Button>
                </div>
                <div className="mt-2">
                  <ApplicationDeploymentTable
                    deployments={row.deployments.slice(0, 5)}
                    appSlug={row.service.slug}
                    sourceRepo={source?.repo}
                    defaultBranch={source?.branch ?? "main"}
                    compact
                  />
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

              <section
                className={RESOURCE_DETAIL_CARD}
                style={resourceDetailCardStyle()}
                aria-label="Processes teaser"
              >
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <h2 className="text-lg font-semibold tracking-tight">
                      Processes
                    </h2>
                    <p
                      className="mt-1 text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      Primary web process on port {port}.
                    </p>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => selectTab("processes")}
                  >
                    Open processes
                  </Button>
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
                  variant="outline"
                  disabled={deployBusy}
                  onClick={requestDeploy}
                >
                  {deployBusy ? "Deploying…" : "Deploy now"}
                </Button>
              </div>
              <div className="mt-2">
                <ApplicationDeploymentTable
                  deployments={row.deployments}
                  appSlug={row.service.slug}
                  sourceRepo={source?.repo}
                  defaultBranch={source?.branch ?? "main"}
                />
                {row.deployments.length === 0 ? (
                  <div className="py-8 text-center">
                    <EmptyState
                      title="No deployments"
                      description="Trigger a deploy to roll out a new version."
                    />
                    <Button
                      type="button"
                      variant="outline"
                      className="mt-4"
                      disabled={deployBusy}
                      onClick={requestDeploy}
                    >
                      {deployBusy ? "Deploying…" : "Deploy now"}
                    </Button>
                  </div>
                ) : null}
              </div>
            </section>
          ) : null}

          {tab === "processes" ? (
            <ApplicationProcessesTab
              service={row.service}
              port={port}
              onViewLogs={() => selectTab("logs")}
            />
          ) : null}

          {tab === "environment" ? (
            <ApplicationEnvironmentTab
              service={row.service}
              variables={row.variables}
              linkedSlugs={linkedSlugs}
              liveDeploymentAt={
                row.deployments[0]?.finished_at ??
                row.deployments[0]?.started_at ??
                row.deployments[0]?.queued_at
              }
              revealedSecrets={revealedSecrets}
              onToggleRevealSecrets={() => setRevealedSecrets((v) => !v)}
              onAddVariables={onAddVariables}
              onRefresh={refresh}
              onDeploy={requestDeploy}
              deployBusy={deployBusy}
              bannerDismissed={bannerDismissed}
              onDismissBanner={() => setBannerDismissed(true)}
            />
          ) : null}

          {tab === "domains" ? (
            <ServiceDomainsTab
              serviceId={row.service.id}
              domains={row.domains}
              onRefresh={refresh}
              onError={setError}
              layout="app"
              cardClassName={RESOURCE_DETAIL_CARD}
              cardStyle={resourceDetailCardStyle()}
            />
          ) : null}

          {tab === "logs" ? (
            <ServiceAccessLogsPanel serviceId={row.service.id} />
          ) : null}

          {tab === "analytics" ? (
            <ServiceTrafficAnalyticsPanel serviceId={row.service.id} />
          ) : null}

          {tab === "previews" ? (
            <ServicePreviewsTab serviceId={row.service.id} onError={setError} />
          ) : null}

          {tab === "security" ? (
            <ServiceSecurityTab serviceId={row.service.id} kind="app" />
          ) : null}

          {tab === "settings" ? (
            <ApplicationSettingsTab
              service={row.service}
              project={row.project}
              onUpdated={refresh}
              onError={setError}
            />
          ) : null}

          <ConfirmDialog
            open={deployConfirmOpen}
            title="Deploy now?"
            description="This starts a new deployment and may replace the currently live version."
            confirmLabel="Deploy"
            onConfirm={confirmDeploy}
            onCancel={() => setDeployConfirmOpen(false)}
          />
        </div>
      )}
    </DashboardShell>
  );
}
