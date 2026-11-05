"use client";

import { Layers, Plus } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AddProjectResourceModal } from "@/components/dockyard/add-project-resource-modal";
import { CreateApplicationModal } from "@/components/dockyard/create-application-modal";
import { CreateBucketModal } from "@/components/dockyard/create-bucket-modal";
import { CreateComposeModal } from "@/components/dockyard/create-compose-modal";
import { CreateDatabaseModal } from "@/components/dockyard/create-database-modal";
import { CreateStaticSiteWizard } from "@/components/dockyard/create-static-site-wizard";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { ProjectHardwareUsageSection } from "@/components/dockyard/project-hardware-usage-section";
import { ProjectLinksMap } from "@/components/dockyard/project-links-map";
import { ProjectLinksSettings } from "@/components/dockyard/project-links-settings";
import {
  type ProjectResourceRow,
  ProjectResourceTable,
} from "@/components/dockyard/project-resource-table";
import {
  type ResourceProvisionContext,
  ResourceProvisioningModal,
} from "@/components/dockyard/resource-provisioning-modal";
import {
  StaticSiteCreateProgressModal,
  type StaticSiteCreateResult,
} from "@/components/dockyard/static-site-create-progress-modal";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import {
  SplitButton,
  type SplitButtonMenuItem,
} from "@/components/dockyard/ui/split-button";
import { Tabs } from "@/components/dockyard/ui/tabs";
import type { TimeRange } from "@/components/dockyard/ui/time-range-toggle";
import { formatBytes } from "@/lib/analytics-stats";
import {
  type Deployment,
  getMe,
  getServerContainerMetrics,
  getServerInfo,
  listDeployments,
  listProjectLinks,
  listProjects,
  listServices,
  type Project,
  type ProjectLinkEdge,
  type Service,
} from "@/lib/api";
import {
  PROJECT_RESOURCE_KINDS,
  type ProjectResourceKind,
} from "@/lib/project-resource-catalog";
import { useRouteParam } from "@/lib/use-route-param";

const PROJECT_TABS = [
  { id: "resources", label: "Resources" },
  { id: "settings", label: "Project settings" },
];

export function ProjectDetailPage() {
  const slug = useRouteParam("slug");
  const [project, setProject] = useState<Project | null>(null);
  const [services, setServices] = useState<Service[]>([]);
  const [deployments, setDeployments] = useState<Record<string, Deployment[]>>(
    {},
  );
  const [metricsByService, setMetricsByService] = useState<
    Record<string, { cpu: number; mem: number }>
  >({});
  const [serverHostname, setServerHostname] = useState<string | undefined>();
  const [hardwareRange, setHardwareRange] = useState<TimeRange>("24h");
  const [tab, setTab] = useState("resources");
  const [resourceView, setResourceView] = useState<"list" | "map">("list");
  const [projectLinks, setProjectLinks] = useState<ProjectLinkEdge[]>([]);
  const [linkCounts, setLinkCounts] = useState<Record<string, number>>({});
  const [addOpen, setAddOpen] = useState(false);
  const [createAppOpen, setCreateAppOpen] = useState(false);
  const [createComposeOpen, setCreateComposeOpen] = useState(false);
  const [canCreateCompose, setCanCreateCompose] = useState(false);
  const [createDbOpen, setCreateDbOpen] = useState(false);
  const [createBucketOpen, setCreateBucketOpen] = useState(false);
  const [staticWizardOpen, setStaticWizardOpen] = useState(false);
  const [provisionOpen, setProvisionOpen] = useState(false);
  const [provisionContext, setProvisionContext] =
    useState<ResourceProvisionContext | null>(null);
  const [staticProgressOpen, setStaticProgressOpen] = useState(false);
  const [staticProgressContext, setStaticProgressContext] =
    useState<StaticSiteCreateResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const openCreateKind = useCallback((kind: ProjectResourceKind) => {
    switch (kind) {
      case "app":
        setCreateAppOpen(true);
        break;
      case "postgres":
        setCreateDbOpen(true);
        break;
      case "bucket":
        setCreateBucketOpen(true);
        break;
      case "static_site":
        setStaticWizardOpen(true);
        break;
    }
  }, []);

  useEffect(() => {
    void getMe()
      .then((u) =>
        setCanCreateCompose(u.role === "owner" || u.role === "admin"),
      )
      .catch(() => setCanCreateCompose(false));
  }, []);

  const createMenuItems: SplitButtonMenuItem[] = useMemo(() => {
    const base: SplitButtonMenuItem[] = PROJECT_RESOURCE_KINDS.map((item) => ({
      id: item.id,
      label: item.label,
      icon: item.icon,
      onSelect: () => openCreateKind(item.id),
    }));
    if (canCreateCompose) {
      base.push({
        id: "compose-stack",
        label: "Compose stack",
        icon: Layers,
        onSelect: () => setCreateComposeOpen(true),
      });
    }
    return base;
  }, [openCreateKind, canCreateCompose]);

  const projectResourceActions = (
    <div className="flex flex-wrap items-center justify-end gap-2">
      <Button type="button" variant="outline" onClick={() => setAddOpen(true)}>
        Add existing resource
      </Button>
      <SplitButton
        variant="primary"
        iconLeft={Plus}
        menuLabel="Choose resource type"
        menuItems={createMenuItems}
        primaryOpensMenu
      >
        Create resource
      </SplitButton>
    </div>
  );

  const refresh = useCallback(
    async (opts?: { silent?: boolean }) => {
      if (!slug) {
        setLoading(false);
        return;
      }
      if (!opts?.silent) setLoading(true);
      try {
        const { projects } = await listProjects();
        const p = projects.find((x) => x.slug === slug) ?? null;
        setProject(p);
        if (!p) {
          setServices([]);
          setError(null);
          return;
        }
        const { services: svc } = await listServices(p.id);
        setServices(svc);
        try {
          const linkRes = await listProjectLinks(p.id);
          setProjectLinks(linkRes.links);
          setLinkCounts(linkRes.outgoing_counts);
        } catch {
          setProjectLinks([]);
          setLinkCounts({});
        }
        const depMap: Record<string, Deployment[]> = {};
        await Promise.all(
          svc.map(async (s) => {
            const { deployments: d } = await listDeployments(s.id);
            depMap[s.id] = d.slice(0, 5);
          }),
        );
        setDeployments(depMap);
        try {
          const [{ metrics }, info] = await Promise.all([
            getServerContainerMetrics("24h"),
            getServerInfo(),
          ]);
          setServerHostname(info.hostname);
          const byId: Record<string, { cpu: number; mem: number }> = {};
          for (const m of metrics) {
            byId[m.service_id] = {
              cpu: m.cpu_percent,
              mem: m.memory_bytes,
            };
          }
          setMetricsByService(byId);
        } catch {
          setMetricsByService({});
        }
        setError(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load project");
      } finally {
        setLoading(false);
      }
    },
    [slug],
  );

  const onProvisionFinished = useCallback(() => {
    void refresh({ silent: true });
  }, [refresh]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const resourceRows: ProjectResourceRow[] = useMemo(
    () =>
      services.map((service) => ({
        service,
        deployment: deployments[service.id]?.[0],
        cpuPercent: metricsByService[service.id]?.cpu,
        memoryBytes: metricsByService[service.id]?.mem,
        outgoingLinkCount: linkCounts[service.id],
      })),
    [services, deployments, metricsByService, linkCounts],
  );

  const inProduction = services.filter(
    (s) => s.desired_state === "running",
  ).length;
  const showInitialLoading = loading && !project;

  const hardwareTotals = useMemo(() => {
    let cpu = 0;
    let mem = 0;
    for (const s of services) {
      const m = metricsByService[s.id];
      if (!m) continue;
      cpu += m.cpu;
      mem += m.mem;
    }
    return {
      cpu: cpu > 0 ? `${cpu.toFixed(1)}%` : "—",
      memory: mem > 0 ? formatBytes(mem) : "—",
    };
  }, [services, metricsByService]);

  if (!slug) {
    return (
      <DashboardShell title="Project">
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Open a project from{" "}
          <Link href="/projects" className="underline">
            Projects
          </Link>
          .
        </p>
      </DashboardShell>
    );
  }

  return (
    <DashboardShell title={project?.name ?? "Project"} breadcrumb="Projects">
      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      {showInitialLoading ? (
        <DockyrdPageLoading />
      ) : !project ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Project not found.
        </p>
      ) : (
        <>
          <PageHeader
            eyebrow="Project"
            title={project.name}
            subtitle={`${inProduction} resource${inProduction === 1 ? "" : "s"} in production`}
            action={projectResourceActions}
          />

          <Tabs tabs={PROJECT_TABS} activeId={tab} onChange={setTab} />

          {tab === "settings" ? (
            <section
              className="mt-8 rounded-[20px] border px-6 py-8"
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
            >
              <h2 className="text-lg font-semibold">Project settings</h2>
              <p
                className="mt-2 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                Slug: <span className="font-mono">{project.slug}</span>
              </p>
              <ProjectLinksSettings
                project={project}
                onUpdated={() => void refresh()}
              />
            </section>
          ) : (
            <>
              <section className="mt-8" aria-label="Resources">
                <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
                  <div>
                    <h2 className="text-lg font-semibold">Resources</h2>
                    <p
                      className="mt-1 text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      Everything that runs in this project.
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      type="button"
                      variant={resourceView === "list" ? "primary" : "outline"}
                      onClick={() => setResourceView("list")}
                    >
                      List
                    </Button>
                    <Button
                      type="button"
                      variant={resourceView === "map" ? "primary" : "outline"}
                      onClick={() => setResourceView("map")}
                    >
                      Map
                    </Button>
                  </div>
                </div>
                {resourceRows.length === 0 ? (
                  <div
                    className="rounded-[20px] border px-6 py-12 text-center"
                    style={{
                      borderColor: "var(--border)",
                      background: "var(--surface)",
                    }}
                  >
                    <p className="font-medium">No resources yet</p>
                    <p
                      className="mt-2 text-sm"
                      style={{ color: "var(--muted-text)" }}
                    >
                      Add an app, database, static site, or bucket to this
                      project.
                    </p>
                    <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
                      {projectResourceActions}
                    </div>
                  </div>
                ) : resourceView === "map" ? (
                  <ProjectLinksMap links={projectLinks} />
                ) : (
                  <ProjectResourceTable rows={resourceRows} />
                )}
              </section>

              <ProjectHardwareUsageSection
                projectName={project.name}
                serverHostname={serverHostname}
                range={hardwareRange}
                onRangeChange={setHardwareRange}
                cpuSummary={hardwareTotals.cpu}
                memorySummary={hardwareTotals.memory}
              />
            </>
          )}
        </>
      )}

      <AddProjectResourceModal
        open={addOpen}
        project={project}
        onClose={() => setAddOpen(false)}
        onChanged={() => void refresh()}
      />

      <CreateApplicationModal
        open={createAppOpen}
        defaultProjectId={project?.id}
        onClose={() => setCreateAppOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <CreateComposeModal
        open={createComposeOpen}
        defaultProjectId={project?.id}
        onClose={() => setCreateComposeOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <CreateDatabaseModal
        open={createDbOpen}
        defaultProjectId={project?.id}
        onClose={() => setCreateDbOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <CreateBucketModal
        open={createBucketOpen}
        defaultProjectId={project?.id}
        onClose={() => setCreateBucketOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <CreateStaticSiteWizard
        open={staticWizardOpen}
        defaultProjectId={project?.id}
        onClose={() => setStaticWizardOpen(false)}
        onCreated={(result) => {
          setStaticWizardOpen(false);
          setStaticProgressContext(result);
          setStaticProgressOpen(true);
        }}
      />

      <ResourceProvisioningModal
        open={provisionOpen}
        context={provisionContext}
        onClose={() => setProvisionOpen(false)}
        onFinished={onProvisionFinished}
      />

      <StaticSiteCreateProgressModal
        open={staticProgressOpen}
        context={staticProgressContext}
        onClose={() => setStaticProgressOpen(false)}
        onEditSettings={() => setStaticProgressOpen(false)}
        onFinished={onProvisionFinished}
      />
    </DashboardShell>
  );
}
