"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { ServiceConnectionPanel } from "@/components/dockyard/service-connection-panel";
import { ServiceLogsPanel } from "@/components/dockyard/service-logs-panel";
import { Chip } from "@/components/dockyard/ui/chip";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";
import { SettingsDetailBlock } from "@/components/dockyard/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import { Tabs } from "@/components/dockyard/ui/tabs";
import { type DatabaseRow, listDatabases } from "@/lib/api";
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

  return (
    <DashboardShell title={title} breadcrumb="Databases">
      {loading ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
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
                <Link
                  href={`/projects/${row.project_slug}`}
                  className="hover:underline"
                  style={{ color: "var(--link-blue)" }}
                >
                  {projectPathLabel(row.project_name)}
                </Link>
                <span className="mx-2">·</span>
                <span className="font-mono">{internalEndpoint(row)}</span>
              </p>
            </div>
          </header>

          <Tabs
            activeId={tab}
            onChange={(id) => selectTab(id as TabId)}
            tabs={[...DETAIL_TABS]}
          />

          {tab === "overview" ? (
            <div className="grid gap-6 lg:grid-cols-2">
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
                      <Link
                        href={`/projects/${row.project_slug}`}
                        className="hover:underline"
                        style={{ color: "var(--link-blue)" }}
                      >
                        {row.project_name}
                      </Link>
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
              <ServiceConnectionPanel serviceId={row.id} />
            </div>
          ) : null}

          {tab === "logs" ? <ServiceLogsPanel serviceId={row.id} /> : null}

          {tab === "settings" ? (
            <section
              className={RESOURCE_DETAIL_CARD}
              style={resourceDetailCardStyle()}
              aria-label="Settings"
            >
              <h2 className="text-lg font-semibold tracking-tight">Settings</h2>
              <SettingsDetailBlock
                title="Engine"
                description="Major version upgrades require dump and restore — not available in the panel yet."
                rows={[
                  {
                    label: "Type",
                    value: databaseEngineLabel(row.kind),
                  },
                  {
                    label: "Version",
                    value: engineVersionChip(row),
                  },
                  {
                    label: "Memory",
                    value: memoryLimitLabel(row.config, row.kind),
                  },
                ]}
                actionDisabled
              />
              <SettingsDetailBlock
                title="Backups"
                description="Scheduled backups and restore will ship in a later release."
                rows={[
                  {
                    label: "Schedule",
                    value: "Not configured",
                  },
                ]}
                actionDisabled
              />
            </section>
          ) : null}
        </div>
      )}
    </DashboardShell>
  );
}
