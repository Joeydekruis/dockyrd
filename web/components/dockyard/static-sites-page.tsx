"use client";

import { Globe, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateStaticSiteWizard } from "@/components/dockyard/create-static-site-wizard";
import {
  StaticSiteCreateProgressModal,
  type StaticSiteCreateResult,
} from "@/components/dockyard/static-site-create-progress-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListAddLink } from "@/components/dockyard/ui/resource-list-add-link";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import {
  type Deployment,
  listDeployments,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
import {
  formatResourceCount,
  formatTimeAgo,
  projectPathLabel,
  resourceStatusDisplay,
} from "@/lib/resource-dashboard";
import { filterByQuery } from "@/lib/service-display";

type StaticRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
};

function staticSitePublicHost(service: Service): string | undefined {
  const cfg = service.config as {
    publicHost?: string;
    domain?: string;
    hostname?: string;
  };
  return cfg.publicHost ?? cfg.domain ?? cfg.hostname;
}

function deploymentSummary(latest?: Deployment) {
  if (!latest?.commit_sha) {
    return { sha: undefined, message: undefined, timeAgo: undefined };
  }
  return {
    sha: latest.commit_sha,
    message: latest.branch ? `on ${latest.branch}` : latest.trigger,
    timeAgo: formatTimeAgo(latest.finished_at ?? latest.started_at),
  };
}

export function StaticSitesPage() {
  const [rows, setRows] = useState<StaticRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [progressContext, setProgressContext] =
    useState<StaticSiteCreateResult | null>(null);
  const [progressOpen, setProgressOpen] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { projects: ps } = await listProjects();
      const collected: StaticRow[] = [];
      for (const p of ps) {
        const { services } = await listServices(p.id, "static_site");
        for (const service of services) {
          const { deployments } = await listDeployments(service.id);
          collected.push({ project: p, service, deployments });
        }
      }
      setRows(collected);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load static sites");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const filtered = useMemo(
    () =>
      filterByQuery(rows, query, ({ project, service }) => [
        service.name,
        project.name,
        project.slug,
        String(
          (service.config as { source?: { repo?: string } }).source?.repo ?? "",
        ),
      ]),
    [rows, query],
  );

  const live = rows.filter(
    ({ service }) => service.desired_state === "running",
  ).length;
  const projectCount = new Set(rows.map((r) => r.project.id)).size;
  const dash = loading ? "—" : undefined;

  return (
    <DashboardShell title="Static sites" breadcrumb="Projects">
      <PageHeader
        eyebrow="All projects"
        title="Static sites"
        subtitle="Built from a Git repo and served from your own server."
        action={
          <Button type="button" iconLeft={Plus} onClick={() => setCreateOpen(true)}>
            Create static site
          </Button>
        }
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section
        aria-label="Summary"
        className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4"
      >
        <StatCard
          label="Total"
          value={dash ?? rows.length}
          hint={loading ? undefined : `across ${projectCount} projects`}
        />
        <StatCard
          label="Live"
          value={dash ?? live}
          hint="all serving traffic"
        />
        <StatCard label="Requests" value={dash ?? "—"} hint="last 24 hours" />
        <StatCard label="Bandwidth" value={dash ?? "—"} hint="last 24 hours" />
      </section>

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : rows.length === 0 ? (
        <section
          className="mt-10 rounded-[20px] border px-8 py-12 text-center"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <h2 className="text-xl font-semibold tracking-tight">
            No static sites yet
          </h2>
          <p
            className="mx-auto mt-2 max-w-md text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            Connect a GitHub repo and publish build output to the edge proxy.
          </p>
          <Button
            type="button"
            className="mt-6"
            iconLeft={Plus}
            onClick={() => setCreateOpen(true)}
          >
            Create static site
          </Button>
        </section>
      ) : (
        <ResourceListSection
          title="All static sites"
          countLabel={formatResourceCount(rows.length, "static site")}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            <ResourceListAddLink
              label="Create static site"
              onClick={() => setCreateOpen(true)}
            />
          }
        >
          {filtered.length === 0 ? (
            <p
              className="py-6 text-center text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              No static sites match your filter.
            </p>
          ) : (
            filtered.map(({ project, service, deployments }) => {
              const status = resourceStatusDisplay(
                service.desired_state,
                "static",
              );
              const latest = deployments[0];
              const deploy = deploymentSummary(latest);
              const publicHost = staticSitePublicHost(service);
              return (
                <ResourceRow
                  key={service.id}
                  name={service.name}
                  href={`/projects/${project.slug}`}
                  meta={projectPathLabel(project.name)}
                  url={publicHost}
                  internalOnly={!publicHost}
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Globe}
                  sha={deploy.sha}
                  commitMessage={deploy.message}
                  timeAgo={deploy.timeAgo}
                />
              );
            })
          )}
        </ResourceListSection>
      )}

      <CreateStaticSiteWizard
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(result) => {
          setProgressContext(result);
          setProgressOpen(true);
        }}
      />

      <StaticSiteCreateProgressModal
        open={progressOpen}
        context={progressContext}
        onClose={() => setProgressOpen(false)}
        onEditSettings={() => {
          setProgressOpen(false);
          setCreateOpen(true);
        }}
        onFinished={() => void refresh()}
      />
    </DashboardShell>
  );
}
