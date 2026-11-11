"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DatabaseConnectionCards } from "@/components/dockyard/database-connection-cards";
import { DatabaseDetailSettings } from "@/components/dockyard/database-detail-settings";
import { DatabaseQueryPanel } from "@/components/dockyard/database-query-panel";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { ServiceConnectedApplicationsPanel } from "@/components/dockyard/service-connected-applications-panel";
import { ServiceLogsPanel } from "@/components/dockyard/service-logs-panel";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Tabs } from "@/components/dockyard/ui/tabs";
import {
  type DatabaseRow,
  listDatabases,
  restartDatabaseService,
} from "@/lib/api";
import {
  isUnassignedProject,
  projectDisplayName,
} from "@/lib/project-optional";
import { databaseDetailBreadcrumbs } from "@/lib/resource-breadcrumbs";
import {
  projectPathLabel,
  resourceStatusDisplay,
} from "@/lib/resource-dashboard";
import {
  persistDetailTab,
  RESOURCE_DETAIL_CARD,
  readInitialDetailTab,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";
import { databaseEngineLabel } from "@/lib/service-display";
import { useRouteParam } from "@/lib/use-route-param";

const DETAIL_TABS = [
  { id: "overview", label: "Overview" },
  { id: "query", label: "Query" },
  { id: "logs", label: "Logs" },
  { id: "settings", label: "Settings" },
] as const;

type TabId = (typeof DETAIL_TABS)[number]["id"];

function engineVersionChip(row: DatabaseRow): string {
  const version = (row.config as { version?: number }).version;
  if (row.kind === "postgres") return `PG ${version ?? 16}`;
  if (row.kind === "redis") return `REDIS ${version ?? 7}`;
  return databaseEngineLabel(row.kind).toUpperCase();
}

function internalEndpoint(row: DatabaseRow): string {
  if (row.kind === "postgres") return `${row.slug}:5432`;
  return `${row.slug}:6379`;
}

function memoryLimitLabel(
  config: Record<string, unknown>,
  kind: string,
): string {
  if (kind === "postgres") {
    const mb = (config as { mem_limit_mb?: number }).mem_limit_mb;
    return mb ? `${mb} MB` : "512 MB default";
  }
  const mb = (config as { max_memory_mb?: number }).max_memory_mb;
  return mb ? `${mb} MB` : "256 MB default";
}

export function DatabaseDetailPage() {
  const slug = useRouteParam("slug");
  const [row, setRow] = useState<DatabaseRow | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<TabId>("overview");
  const [restartBusy, setRestartBusy] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!slug) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const { databases } = await listDatabases();
      const match = databases.find((d) => d.slug === slug) ?? null;
      setRow(match);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load database");
    } finally {
      setLoading(false);
    }
  }, [slug]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    setTab(
      readInitialDetailTab(
        DETAIL_TABS.map((t) => t.id),
        "overview",
      ),
    );
  }, []);

  function selectTab(next: TabId) {
    setTab(next);
    persistDetailTab(next);
  }

  const status = useMemo(
    () =>
      row
        ? resourceStatusDisplay(row.desired_state, "database")
        : { label: "Unknown", variant: "idle" as const },
    [row],
  );

  const title = row?.name ?? "Database";
  const breadcrumbItems = useMemo(
    () =>
      databaseDetailBreadcrumbs(
        row ? { name: row.project_name, slug: row.project_slug } : undefined,
        title,
      ),
    [row, title],
  );

  async function onRestart() {
    if (!row) return;
    setRestartBusy(true);
    setActionError(null);
    try {
      await restartDatabaseService(row.id);
      await refresh();
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Restart failed");
    } finally {
      setRestartBusy(false);
    }
  }

  return (
    <DashboardShell title={title} breadcrumbItems={breadcrumbItems}>
      {loading ? (
        <DockyrdPageLoading />
      ) : error ? (
        <p className="text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : !row ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Database not found.{" "}
          <Link href="/databases" className="underline">
            Back to databases
          </Link>
        </p>
      ) : (
        <div className="space-y-6">
          <header className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div className="flex flex-wrap items-center gap-3">
                <h1 className="text-2xl font-semibold tracking-tight">
                  {row.name}
                </h1>
                <StatusBadge label={status.label} variant={status.variant} />
                <Chip>{engineVersionChip(row)}</Chip>
              </div>
              <p
                className="mt-2 text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {isUnassignedProject(row.project_slug) ? (
                  <span>
                    {projectPathLabel(row.project_name, row.project_slug)}
                  </span>
                ) : (
                  <Link
                    href={`/projects/${row.project_slug}`}
                    className="hover:underline"
                    style={{ color: "var(--link-blue)" }}
                  >
                    {projectPathLabel(row.project_name, row.project_slug)}
                  </Link>
                )}
                <span className="mx-2">·</span>
                <span className="font-mono">{internalEndpoint(row)}</span>
              </p>
            </div>
            <Button
              type="button"
              variant="outline"
              disabled={restartBusy || row.desired_state !== "running"}
              onClick={() => void onRestart()}
            >
              {restartBusy ? "Restarting…" : "Restart"}
            </Button>
          </header>
          {actionError ? (
            <p className="text-sm text-red-600" role="alert">
              {actionError}
            </p>
          ) : null}

          <Tabs
            activeId={tab}
            onChange={(id) => selectTab(id as TabId)}
            tabs={[...DETAIL_TABS]}
          />

          {tab === "overview" ? (
            <div className="space-y-6">
              <section
                className={RESOURCE_DETAIL_CARD}
                style={resourceDetailCardStyle()}
                aria-label="Details"
              >
                <h2 className="text-lg font-semibold tracking-tight">
                  Details
                </h2>
                <dl className="mt-4">
                  <KeyValueRow
                    label="Engine"
                    value={databaseEngineLabel(row.kind)}
                  />
                  <KeyValueRow label="Version" value={engineVersionChip(row)} />
                  <KeyValueRow
                    label="Project"
                    value={
                      isUnassignedProject(row.project_slug) ? (
                        projectDisplayName(row.project_name, row.project_slug)
                      ) : (
                        <Link
                          href={`/projects/${row.project_slug}`}
                          className="hover:underline"
                          style={{ color: "var(--link-blue)" }}
                        >
                          {row.project_name}
                        </Link>
                      )
                    }
                  />
                  <KeyValueRow
                    label="Internal host"
                    value={
                      <span className="font-mono text-sm">
                        {internalEndpoint(row)}
                      </span>
                    }
                  />
                  <KeyValueRow
                    label="Status"
                    value={
                      <StatusBadge
                        label={status.label}
                        variant={status.variant}
                      />
                    }
                  />
                  <KeyValueRow
                    label="Memory limit"
                    value={memoryLimitLabel(row.config, row.kind)}
                  />
                </dl>
              </section>
              <DatabaseConnectionCards serviceId={row.id} kind={row.kind} />
              <ServiceConnectedApplicationsPanel providerServiceId={row.id} />
            </div>
          ) : null}

          {tab === "logs" ? <ServiceLogsPanel serviceId={row.id} /> : null}

          {tab === "query" && row.kind === "postgres" ? (
            <DatabaseQueryPanel serviceId={row.id} />
          ) : null}
          {tab === "query" && row.kind !== "postgres" ? (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              Query studio is available for Postgres databases only.
            </p>
          ) : null}

          {tab === "settings" ? <DatabaseDetailSettings row={row} /> : null}
        </div>
      )}
    </DashboardShell>
  );
}
