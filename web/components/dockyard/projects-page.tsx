"use client";

import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateProjectModal } from "@/components/dockyard/create-project-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { Button } from "@/components/dockyard/ui/button";
import { ProjectCard } from "@/components/dockyard/ui/project-card";
import { listProjects, listServices, type Project } from "@/lib/api";
import {
  buildProjectOverview,
  type ProjectOverview,
} from "@/lib/projects-overview";
import { formatResourceCount } from "@/lib/resource-dashboard";

export function ProjectsPage() {
  const router = useRouter();
  const [overviews, setOverviews] = useState<ProjectOverview[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [createOpen, setCreateOpen] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { projects } = await listProjects();
      const built = await Promise.all(
        projects.map(async (project) => {
          const { services } = await listServices(project.id);
          return buildProjectOverview(project, services);
        }),
      );
      built.sort((a, b) => b.project.updated_at - a.project.updated_at);
      setOverviews(built);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load projects");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const totals = useMemo(() => {
    const projectCount = overviews.length;
    const serviceCount = overviews.reduce((n, o) => n + o.serviceCount, 0);
    return { projectCount, serviceCount };
  }, [overviews]);

  function openCreate() {
    setCreateOpen(true);
  }

  function onProjectCreated(project: Project) {
    void refresh();
    router.push(`/projects/${project.slug}`);
  }

  return (
    <DashboardShell title="Projects">
      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section aria-label="All projects" className="space-y-6">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <h2 className="text-[36px] font-semibold leading-[1.04] tracking-[-0.04em]">
            Your projects
          </h2>
          <div className="flex flex-wrap items-center gap-4">
            {!loading ? (
              <p
                className="font-mono text-[11px] uppercase tracking-widest"
                style={{ color: "var(--muted-text)" }}
              >
                {formatResourceCount(totals.projectCount, "project")}
                {" · "}
                {formatResourceCount(totals.serviceCount, "service")}
              </p>
            ) : null}
            <Button type="button" iconLeft={Plus} onClick={openCreate}>
              Create project
            </Button>
          </div>
        </div>

        {loading ? (
          <DockyrdPageLoading />
        ) : overviews.length === 0 ? (
          <section
            className="rounded-[20px] border px-8 py-12 text-center"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
            }}
          >
            <h3 className="text-xl font-semibold tracking-tight">
              No projects yet
            </h3>
            <p
              className="mx-auto mt-2 max-w-md text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              Create a project to group apps, databases, static sites, and
              buckets on your server.
            </p>
            <Button type="button" className="mt-6" onClick={openCreate}>
              Create project
            </Button>
          </section>
        ) : (
          <ul className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {overviews.map((row) => (
              <li key={row.project.id}>
                <ProjectCard
                  name={row.project.name}
                  href={`/projects/${row.project.slug}`}
                  statusLabel={row.status.label}
                  statusVariant={row.status.variant}
                  environments={row.environments}
                  resources={row.resources}
                  extraResourceCount={row.extraResourceCount}
                  activityLine={row.activityLine}
                />
              </li>
            ))}
          </ul>
        )}
      </section>

      <CreateProjectModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={onProjectCreated}
      />
    </DashboardShell>
  );
}
