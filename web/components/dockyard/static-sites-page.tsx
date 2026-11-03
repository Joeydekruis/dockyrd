"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  createDeployment,
  createService,
  type Deployment,
  listDeployments,
  listGitHubRepos,
  listProjects,
  listServices,
  type Project,
  promoteDeployment,
  type Service,
} from "@/lib/api";

type StaticRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
};

export function StaticSitesPage() {
  const [rows, setRows] = useState<StaticRow[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [repo, setRepo] = useState("");
  const [repos, setRepos] = useState<{ full_name: string }[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { projects: ps } = await listProjects();
      setProjects(ps);
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

  useEffect(() => {
    void (async () => {
      try {
        const { repos: r } = await listGitHubRepos();
        setRepos(r);
      } catch {
        setRepos([]);
      }
    })();
  }, []);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!projectId || !name.trim() || !repo.trim()) return;
    try {
      await createService({
        project_id: projectId,
        kind: "static_site",
        name: name.trim(),
        config: {
          source: { type: "github", repo: repo.trim(), branch: "main" },
          autoDeploy: true,
          buildCmd: "npm ci && npm run build",
          outputDir: "dist",
        },
      });
      setName("");
      setRepo("");
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

  async function onPromote(deploymentId: string) {
    try {
      await promoteDeployment(deploymentId);
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Promote failed");
    }
  }

  return (
    <DashboardShell title="Static sites">
      <section className="space-y-6" aria-label="Static sites">
        <form
          onSubmit={onCreate}
          className="grid gap-3 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 md:grid-cols-2"
        >
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Project</span>
            <select
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={projectId}
              onChange={(ev) => setProjectId(ev.target.value)}
            >
              <option value="">Select project</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Site name</span>
            <input
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={name}
              onChange={(ev) => setName(ev.target.value)}
            />
          </label>
          <label className="flex flex-col gap-1 text-sm md:col-span-2">
            <span className="text-[var(--muted-foreground)]">
              GitHub repository
            </span>
            <input
              list="github-repos"
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={repo}
              onChange={(ev) => setRepo(ev.target.value)}
              placeholder="owner/repo"
            />
            <datalist id="github-repos">
              {repos.map((r) => (
                <option key={r.full_name} value={r.full_name} />
              ))}
            </datalist>
          </label>
          <div className="md:col-span-2">
            <button
              type="submit"
              className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
            >
              Create static site
            </button>
          </div>
        </form>

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}

        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : rows.length === 0 ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            No static sites yet.
          </p>
        ) : (
          <ul className="space-y-4">
            {rows.map(({ project, service, deployments }) => (
              <li
                key={service.id}
                className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4"
              >
                <h2 className="font-medium">
                  {service.name}{" "}
                  <span className="text-sm font-normal text-[var(--muted-foreground)]">
                    · {project.name}
                  </span>
                </h2>
                <div className="mt-2 flex flex-wrap gap-2">
                  <button
                    type="button"
                    onClick={() => void onDeploy(service.id)}
                    className="rounded-lg border border-[var(--border)] px-3 py-1.5 text-sm"
                  >
                    Deploy
                  </button>
                </div>
                <ul className="mt-3 space-y-1 font-mono text-xs text-[var(--muted-foreground)]">
                  {deployments.slice(0, 8).map((d) => (
                    <li
                      key={d.id}
                      className="flex flex-wrap items-center gap-2"
                    >
                      <span>
                        {d.status}
                        {d.release_path ? ` · ${d.release_path}` : ""}
                      </span>
                      {d.status === "success" && d.release_path ? (
                        <button
                          type="button"
                          className="rounded border border-[var(--border)] px-2 py-0.5 text-[10px] uppercase"
                          onClick={() => void onPromote(d.id)}
                        >
                          Promote
                        </button>
                      ) : null}
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        )}
      </section>
    </DashboardShell>
  );
}
