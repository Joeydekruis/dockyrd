"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { AddProjectResourceModal } from "@/components/dockyard/add-project-resource-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { ProjectHardwareUsageSection } from "@/components/dockyard/project-hardware-usage-section";
import {
  type ProjectResourceRow,
  ProjectResourceTable,
} from "@/components/dockyard/project-resource-table";
import {
  StaticSiteCreateProgressModal,
  type StaticSiteCreateResult,
} from "@/components/dockyard/static-site-create-progress-modal";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { Tabs } from "@/components/dockyard/ui/tabs";
import type { TimeRange } from "@/components/dockyard/ui/time-range-toggle";
import { formatBytes } from "@/lib/analytics-stats";
import {
  type Deployment,
  getServerContainerMetrics,
  getServerInfo,
  listDeployments,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
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
  const [addOpen, setAddOpen] = useState(false);
  const [progressOpen, setProgressOpen] = useState(false);
  const [progressContext, setProgressContext] =
    useState<StaticSiteCreateResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!slug) {
      setLoading(false);
      return;
    }
    setLoading(true);
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
  }, [slug]);

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
      })),
    [services, deployments, metricsByService],
  );

  const inProduction = services.filter(
    (s) => s.desired_state === "running",
  ).length;

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

      {loading ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
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
            action={
              <Button
                type="button"
                iconLeft={Plus}
                onClick={() => setAddOpen(true)}
              >
                Add resource
              </Button>
            }
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
                Name, slug, and environment settings for this project will live
                here.
              </p>
            </section>
          ) : (
            <>
              <section className="mt-8" aria-label="Resources">
                <div className="mb-4">
                  <h2 className="text-lg font-semibold">Resources</h2>
                  <p
                    className="mt-1 text-sm"
                    style={{ color: "var(--muted-text)" }}
                  >
                    Everything that runs in this project.
                  </p>
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
                    <Button
                      type="button"
                      className="mt-6"
                      iconLeft={Plus}
                      onClick={() => setAddOpen(true)}
                    >
                      Add resource
                    </Button>
                  </div>
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
        onStaticSiteCreated={(result) => {
          setProgressContext(result);
          setProgressOpen(true);
        }}
      />

      <StaticSiteCreateProgressModal
        open={progressOpen}
        context={progressContext}
        onClose={() => setProgressOpen(false)}
        onEditSettings={() => setProgressOpen(false)}
        onFinished={() => void refresh()}
      />
    </DashboardShell>
  );
}
