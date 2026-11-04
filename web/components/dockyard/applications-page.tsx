"use client";

import { Box } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { EmptyState } from "@/components/dockyard/ui/empty-state";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { type ApplicationRow, listApplications } from "@/lib/api";
import {
  filterByQuery,
  serviceStatusDisplay,
} from "@/lib/service-display";

export function ApplicationsPage() {
  const [rows, setRows] = useState<ApplicationRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { applications } = await listApplications();
      setRows(applications);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load applications");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const filtered = useMemo(
    () =>
      filterByQuery(rows, query, (app) => [
        app.name,
        app.slug,
        app.project_name,
        app.project_slug,
      ]),
    [rows, query],
  );

  const running = rows.filter((a) => a.desired_state === "running").length;
  const stopped = rows.filter((a) => a.desired_state === "stopped").length;
  const projectCount = new Set(rows.map((a) => a.project_id)).size;

  const countLabel =
    filtered.length === rows.length
      ? `${rows.length} total`
      : `${filtered.length} of ${rows.length}`;

  return (
    <DashboardShell title="Applications">
      <PageHeader
        title="Applications"
        subtitle="All container apps across your projects, with deploy state at a glance."
      />

      {error ? (
        <p className="mb-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}

      <section aria-label="Summary" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard label="Total" value={loading ? "—" : rows.length} />
        <StatCard
          label="Running"
          value={loading ? "—" : running}
          hint="Desired state running"
        />
        <StatCard label="Stopped" value={loading ? "—" : stopped} />
        <StatCard
          label="Projects"
          value={loading ? "—" : projectCount}
          hint="With at least one app"
        />
      </section>

      {loading ? (
        <p className="mt-10 text-sm" style={{ color: "var(--muted-text)" }}>
          Loading…
        </p>
      ) : rows.length === 0 ? (
        <div className="mt-10">
          <EmptyState
            title="No applications yet"
            description="Create an app service inside a project to see it listed here."
            actionLabel="Open projects"
            actionHref="/projects"
          />
        </div>
      ) : (
        <ResourceListSection
          title="All applications"
          countLabel={countLabel}
          filterValue={query}
          onFilterChange={setQuery}
        >
          {filtered.length === 0 ? (
            <p className="py-6 text-center text-sm" style={{ color: "var(--muted-text)" }}>
              No applications match your filter.
            </p>
          ) : (
            filtered.map((app) => {
              const status = serviceStatusDisplay(app.desired_state);
              return (
                <ResourceRow
                  key={app.id}
                  name={app.name}
                  href={`/projects/${app.project_slug}`}
                  meta={`${app.project_name} · ${app.slug}`}
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Box}
                />
              );
            })
          )}
        </ResourceListSection>
      )}
    </DashboardShell>
  );
}
