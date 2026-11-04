"use client";

import { Database, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateDatabaseModal } from "@/components/dockyard/create-database-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  type ResourceProvisionContext,
  ResourceProvisioningModal,
} from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListAddLink } from "@/components/dockyard/ui/resource-list-add-link";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { type DatabaseRow, listDatabases } from "@/lib/api";
import {
  formatResourceCount,
  projectPathLabel,
  resourceStatusDisplay,
} from "@/lib/resource-dashboard";
import { databaseEngineLabel, filterByQuery } from "@/lib/service-display";

function databaseConnectionMeta(kind: string): string {
  if (kind === "postgres") return "postgres · internal:5432";
  if (kind === "redis") return "redis · internal:6379";
  return kind;
}

function engineVersionChip(row: {
  kind: string;
  config: Record<string, unknown>;
}): string {
  const version = (row.config as { version?: number }).version;
  if (row.kind === "postgres") {
    return `PG ${version ?? 16}`;
  }
  if (row.kind === "redis") {
    return `REDIS ${version ?? 7}`;
  }
  return databaseEngineLabel(row.kind).toUpperCase();
}

export function DatabasesPage() {
  const [rows, setRows] = useState<DatabaseRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [provisionOpen, setProvisionOpen] = useState(false);
  const [provisionContext, setProvisionContext] =
    useState<ResourceProvisionContext | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { databases } = await listDatabases();
      setRows(databases);
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
  const projectCount = new Set(rows.map((r) => r.project_id)).size;
  const dash = loading ? "—" : undefined;

  return (
    <DashboardShell title="Databases" breadcrumb="Projects">
      <PageHeader
        eyebrow="All projects"
        title="Databases"
        subtitle="Postgres and Redis, with backups you can restore."
        action={
          <Button
            type="button"
            iconLeft={Plus}
            onClick={() => setCreateOpen(true)}
          >
            Create database
          </Button>
        }
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section
        aria-label="Summary"
        className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4"
      >
        <StatCard
          label="Total"
          value={dash ?? rows.length}
          hint={loading ? undefined : `across ${projectCount} projects`}
        />
        <StatCard
          label="Postgres"
          value={dash ?? postgresCount}
          hint="managed on this server"
        />
        <StatCard label="Redis" value={dash ?? redisCount} hint="in memory" />
        <StatCard label="Backups" value={dash ?? "—"} hint="last run —" />
      </section>

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : rows.length === 0 ? (
        <section
          className="mt-10 rounded-[20px] border px-8 py-12 text-center"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <h2 className="text-xl font-semibold tracking-tight">
            No databases yet
          </h2>
          <p
            className="mx-auto mt-2 max-w-md text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            Provision Postgres or Redis for a project. Connection strings stay
            encrypted until you reveal them from the project.
          </p>
          <Button
            type="button"
            className="mt-6"
            iconLeft={Plus}
            onClick={() => setCreateOpen(true)}
          >
            Create database
          </Button>
        </section>
      ) : (
        <ResourceListSection
          title="All databases"
          countLabel={formatResourceCount(rows.length, "database")}
          filterValue={query}
          onFilterChange={setQuery}
          footer={
            <ResourceListAddLink
              label="Create database"
              onClick={() => setCreateOpen(true)}
            />
          }
        >
          {filtered.length === 0 ? (
            <p
              className="py-6 text-center text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              No databases match your filter.
            </p>
          ) : (
            filtered.map((row) => {
              const status = resourceStatusDisplay(
                row.desired_state,
                "database",
              );
              return (
                <ResourceRow
                  key={row.id}
                  name={row.name}
                  href={`/projects/${row.project_slug}`}
                  meta={projectPathLabel(row.project_name)}
                  subMeta={databaseConnectionMeta(row.kind)}
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Database}
                  trailing={
                    <div
                      className="flex flex-col items-end gap-1 text-[13px]"
                      style={{ color: "var(--muted-text)" }}
                    >
                      <Chip>{engineVersionChip(row)}</Chip>
                      <span className="hidden lg:inline">daily backups</span>
                      <span className="hidden lg:inline">Backup —</span>
                    </div>
                  }
                />
              );
            })
          )}
        </ResourceListSection>
      )}

      <CreateDatabaseModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(ctx) => {
          setProvisionContext(ctx);
          setProvisionOpen(true);
        }}
      />

      <ResourceProvisioningModal
        open={provisionOpen}
        context={provisionContext}
        onClose={() => setProvisionOpen(false)}
        onFinished={() => void refresh()}
      />
    </DashboardShell>
  );
}
