"use client";

import { Globe } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { AddCard } from "@/components/dockyard/ui/add-card";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { StatCard } from "@/components/dockyard/ui/stat-card";
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
import {
  filterByQuery,
  serviceStatusDisplay,
} from "@/lib/service-display";

type StaticRow = {
  project: Project;
  service: Service;
  deployments: Deployment[];
};

function staticSiteMeta(service: Service, project: Project): string {
  const cfg = service.config as {
    source?: { repo?: string; branch?: string };
  };
  const repo = cfg.source?.repo;
  const branch = cfg.source?.branch ?? "main";
  if (repo) return `${project.name} · ${repo} · ${branch}`;
  return project.name;
}

export function StaticSitesPage() {
  const [rows, setRows] = useState<StaticRow[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [repo, setRepo] = useState("");
  const [repos, setRepos] = useState<{ full_name: string }[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

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

  const countLabel =
    filtered.length === rows.length
      ? `${rows.length} total`
      : `${filtered.length} of ${rows.length}`;

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
      setShowCreate(false);
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

  const createForm = (
    <form
      onSubmit={onCreate}
      className="grid gap-3 rounded-[20px] border p-4 md:grid-cols-2"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <SelectField
        label="Project"
        placeholder="Select project"
        value={projectId}
        onChange={setProjectId}
        options={projects.map((p) => ({ value: p.id, label: p.name }))}
      />
      <label className="flex flex-col gap-1 text-sm">
        <span
          className="font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Site name
        </span>
        <input
          className="field-input"
          value={name}
          onChange={(ev) => setName(ev.target.value)}
        />
      </label>
      <label className="flex flex-col gap-1 text-sm md:col-span-2">
        <span
          className="font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          GitHub repository
        </span>
        <input
          list="github-repos"
          className="field-input"
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
      <div className="flex flex-wrap gap-2 md:col-span-2">
        <Button type="submit">Create static site</Button>
        <Button
          type="button"
          variant="outline"
          onClick={() => setShowCreate(false)}
        >
          Cancel
        </Button>
      </div>
    </form>
  );

  return (
    <DashboardShell title="Static sites">
      <PageHeader
        title="Static sites"
        subtitle="Git-backed static exports with zero-downtime releases."
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section aria-label="Summary" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total" value={loading ? "—" : rows.length} />
        <StatCard label="Live" value={loading ? "—" : live} />
        <StatCard label="Projects" value={loading ? "—" : projectCount} />
        <StatCard
          label="Auto deploy"
          value={loading ? "—" : rows.length}
          hint="On push to main"
        />
      </section>

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : rows.length === 0 ? (
        <div className="mt-10 space-y-4">
          {!showCreate ? (
            <>
              <EmptyState
                title="No static sites yet"
                description="Connect a GitHub repo and publish build output to the edge proxy."
              />
              <AddCard
                title="Create static site"
                description="Pick a project and repository."
                onClick={() => setShowCreate(true)}
              />
            </>
          ) : (
            createForm
          )}
        </div>
      ) : (
        <ResourceListSection
          title="All static sites"
          countLabel={countLabel}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            showCreate ? (
              createForm
            ) : (
              <AddCard
                title="Create static site"
                description="Add another site from GitHub."
                onClick={() => setShowCreate(true)}
              />
            )
          }
        >
          {filtered.length === 0 ? (
            <p className="py-6 text-center text-sm" style={{ color: "var(--muted-text)" }}>
              No static sites match your filter.
            </p>
          ) : (
            filtered.map(({ project, service, deployments }) => {
              const status = serviceStatusDisplay(service.desired_state);
              const latest = deployments[0];
              return (
                <ResourceRow
                  key={service.id}
                  name={service.name}
                  href={`/projects/${project.slug}`}
                  meta={staticSiteMeta(service, project)}
                  subMeta={
                    latest
                      ? `Latest deploy · ${latest.status}${latest.release_path ? ` · ${latest.release_path}` : ""}`
                      : "No deployments yet"
                  }
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Globe}
                  trailing={
                    <div className="flex flex-wrap gap-2">
                      <button
                        type="button"
                        onClick={() => void onDeploy(service.id)}
                        className="rounded-lg border px-3 py-1.5 font-mono text-[11px] uppercase tracking-wide"
                        style={{ borderColor: "var(--border)" }}
                      >
                        Deploy
                      </button>
                      {latest?.status === "success" && latest.release_path ? (
                        <button
                          type="button"
                          className="rounded-lg border px-3 py-1.5 font-mono text-[11px] uppercase tracking-wide"
                          style={{ borderColor: "var(--border)" }}
                          onClick={() =>
                            void promoteDeployment(latest.id).then(() =>
                              refresh(),
                            )
                          }
                        >
                          Promote
                        </button>
                      ) : null}
                    </div>
                  }
                />
              );
            })
          )}
        </ResourceListSection>
      )}
    </DashboardShell>
  );
}
