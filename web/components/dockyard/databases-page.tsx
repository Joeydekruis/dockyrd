"use client";

import { Database } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { SelectField } from "@/components/dockyard/ui/select-field";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { AddCard } from "@/components/dockyard/ui/add-card";
import {
  createService,
  type DatabaseRow,
  getServiceConnection,
  listDatabases,
  listProjects,
  type Project,
} from "@/lib/api";
import {
  databaseEngineLabel,
  filterByQuery,
  serviceStatusDisplay,
} from "@/lib/service-display";

export function DatabasesPage() {
  const [rows, setRows] = useState<DatabaseRow[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [name, setName] = useState("");
  const [kind, setKind] = useState<"postgres" | "redis">("postgres");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [connection, setConnection] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [showCreate, setShowCreate] = useState(false);

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

  const filtered = useMemo(
    () =>
      filterByQuery(rows, query, (row) => [
        row.name,
        row.slug,
        row.project_name,
        row.kind,
      ]),
    [rows, query],
  );

  const postgresCount = rows.filter((r) => r.kind === "postgres").length;
  const redisCount = rows.filter((r) => r.kind === "redis").length;
  const running = rows.filter((r) => r.desired_state === "running").length;

  const countLabel =
    filtered.length === rows.length
      ? `${rows.length} total`
      : `${filtered.length} of ${rows.length}`;

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
      setShowCreate(false);
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
      <PageHeader
        title="Databases"
        subtitle="Managed Postgres and Redis instances with connection strings for your apps."
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section aria-label="Summary" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total" value={loading ? "—" : rows.length} />
        <StatCard label="Postgres" value={loading ? "—" : postgresCount} />
        <StatCard label="Redis" value={loading ? "—" : redisCount} />
        <StatCard
          label="Running"
          value={loading ? "—" : running}
          hint="Containers up"
        />
      </section>

      {connection ? (
        <pre className="mt-6 overflow-x-auto rounded-[20px] border p-4 font-mono text-xs">
          {connection}
        </pre>
      ) : null}

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : rows.length === 0 ? (
        <div className="mt-10 space-y-4">
          {!showCreate ? (
            <>
              <EmptyState
                title="No databases yet"
                description="Provision Postgres or Redis for a project. Connection strings stay encrypted until you reveal them."
              />
              <AddCard
                title="Create database"
                description="Pick a project, engine, and name."
                onClick={() => setShowCreate(true)}
              />
            </>
          ) : (
            <form
              onSubmit={onCreate}
              className="grid gap-3 rounded-[20px] border p-4 md:grid-cols-2"
              style={{
                borderColor: "var(--border)",
                background: "var(--surface)",
              }}
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
              <label className="flex flex-col gap-1 text-sm md:col-span-2">
                <span
                  className="font-mono text-[11px] uppercase tracking-widest"
                  style={{ color: "var(--muted-text)" }}
                >
                  Name
                </span>
                <input
                  className="field-input"
                  value={name}
                  onChange={(ev) => setName(ev.target.value)}
                  placeholder="Primary database"
                />
              </label>
              <div className="flex flex-wrap gap-2 md:col-span-2">
                <Button type="submit">Create database</Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowCreate(false)}
                >
                  Cancel
                </Button>
              </div>
            </form>
          )}
        </div>
      ) : (
        <ResourceListSection
          title="All databases"
          countLabel={countLabel}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            showCreate ? (
              <form
                onSubmit={onCreate}
                className="grid gap-3 rounded-[20px] border p-4 md:grid-cols-2"
                style={{ borderColor: "var(--border)", background: "var(--surface)" }}
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
                <label className="flex flex-col gap-1 text-sm md:col-span-2">
                  <span
                    className="font-mono text-[11px] uppercase tracking-widest"
                    style={{ color: "var(--muted-text)" }}
                  >
                    Name
                  </span>
                  <input
                    className="field-input"
                    value={name}
                    onChange={(ev) => setName(ev.target.value)}
                    placeholder="Primary database"
                  />
                </label>
                <div className="flex flex-wrap gap-2 md:col-span-2">
                  <Button type="submit">Create database</Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setShowCreate(false)}
                  >
                    Cancel
                  </Button>
                </div>
              </form>
            ) : (
              <AddCard
                title="Create database"
                description="Add Postgres or Redis to a project."
                onClick={() => setShowCreate(true)}
              />
            )
          }
        >
          {filtered.length === 0 ? (
            <p className="py-6 text-center text-sm" style={{ color: "var(--muted-text)" }}>
              No databases match your filter.
            </p>
          ) : (
            filtered.map((row) => {
              const status = serviceStatusDisplay(row.desired_state);
              const version =
                row.kind === "postgres"
                  ? String((row.config as { version?: number }).version ?? 17)
                  : String((row.config as { version?: number }).version ?? 8);
              return (
                <ResourceRow
                  key={row.id}
                  name={row.name}
                  href={`/projects/${row.project_slug}`}
                  meta={`${row.project_name} · ${row.slug}`}
                  subMeta={`Backups · managed on this server`}
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Database}
                  trailing={
                    <div className="flex flex-wrap items-center gap-2">
                      <Chip>{databaseEngineLabel(row.kind)} {version}</Chip>
                      <button
                        type="button"
                        className="rounded-lg border px-3 py-1.5 font-mono text-[11px] uppercase tracking-wide"
                        style={{ borderColor: "var(--border)" }}
                        onClick={() => void onReveal(row.id)}
                      >
                        Connection
                      </button>
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
