"use client";

import { useCallback, useEffect, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { SelectField } from "@/components/dockyard/ui/select-field";
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
  const [selected, setSelected] = useState<DatabaseRow | null>(null);

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
      setConnection(
        variables.DATABASE_URL ??
          variables.REDIS_URL ??
          JSON.stringify(variables),
      );
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
          <SelectField
            label="Project"
            placeholder="Select…"
            value={projectId}
            onChange={setProjectId}
            options={projects.map((p) => ({ value: p.id, label: p.name }))}
          />
          <SelectField
            label="Engine"
            value={kind}
            onChange={(v) => setKind(v as "postgres" | "redis")}
            options={[
              { value: "postgres", label: "Postgres 17" },
              { value: "redis", label: "Redis (Valkey 8)" },
            ]}
          />
          <label className="flex flex-1 flex-col gap-1 text-sm">
            <span className="text-[var(--muted-foreground)]">Name</span>
            <input
              className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2"
              value={name}
              onChange={(ev) => setName(ev.target.value)}
              placeholder="Primary database"
            />
          </label>
          <Button type="submit">Create database</Button>
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
          <p className="text-sm text-[var(--muted-foreground)]">
            No managed databases yet.
          </p>
        ) : (
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {rows.map((row) => (
              <li
                key={row.id}
                className="flex flex-wrap items-center justify-between gap-2 px-4 py-3 text-sm"
              >
                <button
                  type="button"
                  className="min-w-0 flex-1 text-left"
                  onClick={() => {
                    setSelected(row);
                    setConnection(null);
                  }}
                >
                  <p className="font-medium">{row.name}</p>
                  <p className="text-[var(--muted-foreground)]">
                    {row.project_name} · {row.kind} · {row.slug}
                  </p>
                </button>
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
        {selected ? (
          <section className="mt-4 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 text-sm">
            <h3 className="mb-2 font-medium">Service detail</h3>
            <dl className="grid gap-2 sm:grid-cols-2">
              <div>
                <dt className="text-[var(--muted-foreground)]">Slug</dt>
                <dd className="font-mono">{selected.slug}</dd>
              </div>
              <div>
                <dt className="text-[var(--muted-foreground)]">Engine</dt>
                <dd>{selected.kind}</dd>
              </div>
              <div>
                <dt className="text-[var(--muted-foreground)]">Project</dt>
                <dd>{selected.project_name}</dd>
              </div>
              <div>
                <dt className="text-[var(--muted-foreground)]">State</dt>
                <dd>{selected.desired_state ?? "—"}</dd>
              </div>
            </dl>
            <p className="mt-3 text-[var(--muted-foreground)]">
              Reference in app env:{" "}
              <code className="rounded bg-[var(--muted)]/30 px-1">{`\${{${selected.slug}.DATABASE_URL}}`}</code>{" "}
              or{" "}
              <code className="rounded bg-[var(--muted)]/30 px-1">{`\${{${selected.slug}.REDIS_URL}}`}</code>
            </p>
          </section>
        ) : null}
      </section>
    </DashboardShell>
  );
}
