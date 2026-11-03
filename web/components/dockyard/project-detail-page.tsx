"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  createDeployment,
  createService,
  type Deployment,
  listDeployments,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
import { useRouteParam } from "@/lib/use-route-param";

export function ProjectDetailPage() {
  const slug = useRouteParam("slug");
  const [project, setProject] = useState<Project | null>(null);
  const [services, setServices] = useState<Service[]>([]);
  const [deployments, setDeployments] = useState<Record<string, Deployment[]>>(
    {},
  );
  const [appName, setAppName] = useState("");
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

  const title = useMemo(() => project?.name ?? "Project", [project]);

  async function onAddApp(e: React.FormEvent) {
    e.preventDefault();
    if (!project) return;
    const name = appName.trim();
    if (!name) return;
    try {
      await createService({
        project_id: project.id,
        kind: "app",
        name,
        config: {
          source: { type: "image", ref: "nginx:alpine" },
          port: 80,
        },
      });
      setAppName("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    }
  }

  async function onDeploy(serviceId: string) {
    try {
      await createDeployment(serviceId);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Deploy failed");
    }
  }

  if (!slug) {
    return (
      <DashboardShell title="Project">
        <p className="text-sm text-[var(--muted-foreground)]">
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
    <DashboardShell title={title}>
      <section className="space-y-6" aria-label="Project services">
        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : !project ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            Project not found.
          </p>
        ) : (
          <>
            <form
              onSubmit={onAddApp}
              className="flex flex-col gap-3 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 sm:flex-row sm:items-end"
            >
              <label className="flex flex-1 flex-col gap-1 text-sm">
                <span className="text-[var(--muted-foreground)]">
                  New application
                </span>
                <input
                  className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
                  value={appName}
                  onChange={(ev) => setAppName(ev.target.value)}
                  placeholder="API service"
                />
              </label>
              <button
                type="submit"
                className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
              >
                Add app
              </button>
            </form>
            <ul className="space-y-4">
              {services.map((s) => {
                const deps = deployments[s.id] ?? [];
                const latest = deps[0];
                return (
                  <li
                    key={s.id}
                    className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4"
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div>
                        <h2 className="font-medium">{s.name}</h2>
                        <p className="text-sm text-[var(--muted-foreground)]">
                          {s.kind} · {s.slug}
                        </p>
                      </div>
                      <button
                        type="button"
                        onClick={() => void onDeploy(s.id)}
                        className="rounded-lg border border-[var(--border)] px-3 py-1.5 text-sm"
                      >
                        Deploy
                      </button>
                    </div>
                    {latest ? (
                      <p className="mt-2 font-mono text-xs text-[var(--muted-foreground)]">
                        Latest: {latest.status}
                        {latest.commit_sha
                          ? ` · ${latest.commit_sha.slice(0, 7)}`
                          : ""}
                        {latest.build_log_path ? (
                          <>
                            {" · "}
                            <a
                              className="underline"
                              href={`/api/v1/deployments/${latest.id}/log`}
                              target="_blank"
                              rel="noreferrer"
                            >
                              build log
                            </a>
                          </>
                        ) : null}
                      </p>
                    ) : (
                      <p className="mt-2 text-xs text-[var(--muted-foreground)]">
                        No deployments yet.
                      </p>
                    )}
                  </li>
                );
              })}
            </ul>
            {services.length === 0 ? (
              <p className="text-sm text-[var(--muted-foreground)]">
                No services in this project yet.
              </p>
            ) : null}
          </>
        )}
      </section>
    </DashboardShell>
  );
}
