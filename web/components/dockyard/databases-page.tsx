"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  createService,
  type DatabaseRow,
  getServiceConnection,
  listDatabases,
  listProjects,
  type Project,
} from "@/lib/api";

export function DatabasesPage() {
  const [rows, setRows] = useState<DatabaseRow[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [kind, setKind] = useState<"postgres" | "redis">("postgres");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [connection, setConnection] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [{ databases }, { projects: ps }] = await Promise.all([
        listDatabases(),
        listProjects(),
      ]);
      setRows(databases);
      setProjects(ps);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load databases");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  async function onCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!projectId || !name.trim()) return;
    try {
      await createService({
        project_id: projectId,
        kind,
        name: name.trim(),
        config: kind === "postgres" ? { version: 17 } : { version: 8 },
      });
      setName("");
      await refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Create failed");
    }
  }

  async function onReveal(id: string) {
    try {
      const { variables } = await getServiceConnection(id, true);
      setConnection(variables.DATABASE_URL ?? variables.REDIS_URL ?? JSON.stringify(variables));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reveal failed");
    }
  }

  return (
    <DashboardShell title="Databases">
      <section className="space-y-6">
        <form
          onSubmit={onCreate}
          className="flex flex-col gap-3 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 lg:flex-row lg:items-end"
        >
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Project</span>
            <select
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={projectId}
              onChange={(ev) => setProjectId(ev.target.value)}
            >
              <option value="">Select…</option>
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </label>
          <label className="flex flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Engine</span>
            <select
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={kind}
              onChange={(ev) => setKind(ev.target.value as "postgres" | "redis")}
            >
              <option value="postgres">Postgres 17</option>
              <option value="redis">Redis (Valkey 8)</option>
            </select>
          </label>
          <label className="flex flex-1 flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Name</span>
            <input
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={name}
              onChange={(ev) => setName(ev.target.value)}
              placeholder="Primary database"
            />
          </label>
          <button
            type="submit"
            className="rounded-lg bg-[var(--primary)] px-4 py-2 text-sm font-medium text-[var(--primary-foreground)]"
          >
            Create database
          </button>
        </form>

        {error ? (
          <p className="text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        {connection ? (
          <pre className="overflow-x-auto rounded-lg border border-[var(--border)] bg-[var(--surface)] p-3 text-xs">
            {connection}
          </pre>
        ) : null}

        {loading ? (
          <p className="text-sm text-[var(--muted-foreground)]">Loading…</p>
        ) : rows.length === 0 ? (
          <p className="text-sm text-[var(--muted-foreground)]">No managed databases yet.</p>
        ) : (
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {rows.map((row) => (
              <li
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm"
              >
                <div>
                  <p className="font-medium">{row.name}</p>
                  <p className="text-[var(--muted-foreground)]">
                    {row.project_name} · {row.kind} · {row.slug}
                  </p>
                </div>
                <button
                  type="button"
                  className="rounded-lg border border-[var(--border)] px-3 py-1.5 text-xs"
                  onClick={() => void onReveal(row.id)}
                >
                  Reveal connection
                </button>
              </li>
            ))}
          </ul>
        )}
      </section>
    </DashboardShell>
  );
}
