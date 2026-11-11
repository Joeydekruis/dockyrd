"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { CreateProjectModal } from "@/components/dockyard/create-project-modal";
import { DashboardActivityPanel } from "@/components/dockyard/dashboard-activity-panel";
import { DashboardMonitoringSection } from "@/components/dockyard/dashboard-monitoring-section";
import { DashboardProjectTiles } from "@/components/dockyard/dashboard-project-tiles";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { ProjectsWelcomeHero } from "@/components/dockyard/projects-welcome-hero";
import {
  getServerInfo,
  listProjects,
  listServices,
  type Project,
} from "@/lib/api";
import {
  buildProjectOverview,
  type ProjectOverview,
} from "@/lib/projects-overview";

export function HomeDashboard() {
  const router = useRouter();
  const [overviews, setOverviews] = useState<ProjectOverview[]>([]);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);
  const [serverEyebrow, setServerEyebrow] = useState<string | undefined>();

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [{ projects }, server] = await Promise.all([
        listProjects(),
        getServerInfo().catch(() => null),
      ]);
      const built = await Promise.all(
        projects.map(async (project) => {
          const { services } = await listServices(project.id);
          return buildProjectOverview(project, services);
        }),
      );
      built.sort((a, b) => b.project.updated_at - a.project.updated_at);
      setOverviews(built);
      if (server?.hostname?.trim()) {
        setServerEyebrow(`${server.hostname.trim()} · Connected`);
      }
    } catch {
      setOverviews([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  function onProjectCreated(project: Project) {
    void refresh();
    router.push(`/projects/${project.slug}`);
  }

  return (
    <DashboardShell title="Dashboard" breadcrumb="Home">
      <ProjectsWelcomeHero
        eyebrow={serverEyebrow}
        onCreateProject={() => setCreateOpen(true)}
      />

      <section aria-labelledby="dashboard-projects-heading" className="mb-8">
        <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
          <h2
            id="dashboard-projects-heading"
            className="font-mono text-xs font-medium uppercase tracking-widest"
            style={{ color: "var(--muted-text)" }}
          >
            Projects
          </h2>
          <Link
            href="/projects"
            className="text-sm font-medium hover:underline"
            style={{ color: "var(--link-blue)" }}
          >
            All projects →
          </Link>
        </div>
        {loading ? (
          <DockyrdPageLoading />
        ) : (
          <DashboardProjectTiles projects={overviews} />
        )}
      </section>

      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <h2
          className="font-mono text-xs font-medium uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Monitoring
        </h2>
        <Link
          href="/analytics"
          className="text-sm font-medium hover:underline"
          style={{ color: "var(--link-blue)" }}
        >
          Open server analytics →
        </Link>
      </div>

      <div className="flex flex-col gap-6 xl:flex-row xl:items-start">
        <div className="min-w-0 flex-1">
          <DashboardMonitoringSection />
        </div>
        <DashboardActivityPanel className="w-full xl:max-w-[min(100%,480px)]" />
      </div>

      <CreateProjectModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={onProjectCreated}
      />
    </DashboardShell>
  );
}
