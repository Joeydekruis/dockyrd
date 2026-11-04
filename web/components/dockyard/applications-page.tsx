"use client";

import { Box, Plus } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { CreateApplicationModal } from "@/components/dockyard/create-application-modal";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import {
  type ResourceProvisionContext,
  ResourceProvisioningModal,
} from "@/components/dockyard/resource-provisioning-modal";
import { Button } from "@/components/dockyard/ui/button";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { ResourceListSection } from "@/components/dockyard/ui/resource-list-section";
import { ResourceRow } from "@/components/dockyard/ui/resource-row";
import { StatCard } from "@/components/dockyard/ui/stat-card";
import { type ApplicationRow, listApplications } from "@/lib/api";
import {
  formatResourceCount,
  projectPathLabel,
  resourceStatusDisplay,
  summarizeApplicationStats,
} from "@/lib/resource-dashboard";
import { filterByQuery } from "@/lib/service-display";
import { serviceDetailHref } from "@/lib/service-routes";

export function ApplicationsPage() {
  const [rows, setRows] = useState<ApplicationRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [provisionOpen, setProvisionOpen] = useState(false);
  const [provisionContext, setProvisionContext] =
    useState<ResourceProvisionContext | null>(null);

  const refresh = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setLoading(true);
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

  const onProvisionFinished = useCallback(() => {
    void refresh({ silent: true });
  }, [refresh]);

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

  const stats = summarizeApplicationStats(rows);
  const dash = loading && rows.length === 0 ? "—" : undefined;
  const showInitialLoading = loading && rows.length === 0;

  return (
    <DashboardShell title="Applications" breadcrumb="Projects">
      <PageHeader
        eyebrow="All projects"
        title="Applications"
        subtitle="Apps deployed from GitHub, a Dockerfile or an image."
        action={
          <Button
            type="button"
            iconLeft={Plus}
            onClick={() => setCreateOpen(true)}
          >
            Create application
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
          value={dash ?? stats.total}
          hint={
            showInitialLoading
              ? undefined
              : `across ${stats.projectCount} projects`
          }
        />
        <StatCard
          label="Healthy"
          value={dash ?? stats.healthy}
          hint="running normally"
        />
        <StatCard
          label="In progress"
          value={dash ?? stats.inProgress}
          hint="deploying now"
        />
        <StatCard
          label="Needs attention"
          value={dash ?? stats.needsAttention}
          hint={showInitialLoading ? undefined : stats.attentionHint}
        />
      </section>

      {showInitialLoading ? (
        <DockyrdPageLoading className="mt-2" />
      ) : rows.length === 0 ? (
        <section
          className="mt-10 rounded-[20px] border px-8 py-12 text-center"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <h2 className="text-xl font-semibold tracking-tight">
            No applications yet
          </h2>
          <p
            className="mx-auto mt-2 max-w-md text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            Create an app inside a project to see it listed here.
          </p>
          <Button
            type="button"
            className="mt-6"
            onClick={() => setCreateOpen(true)}
          >
            Create application
          </Button>
        </section>
      ) : (
        <ResourceListSection
          title="All applications"
          countLabel={formatResourceCount(rows.length, "application")}
          filterValue={query}
          onFilterChange={setQuery}
        >
          {filtered.length === 0 ? (
            <p
              className="py-6 text-center text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              No applications match your filter.
            </p>
          ) : (
            filtered.map((app) => {
              const status = resourceStatusDisplay(
                app.desired_state,
                "application",
              );
              return (
                <ResourceRow
                  key={app.id}
                  name={app.name}
                  href={serviceDetailHref("app", app.slug)}
                  meta={projectPathLabel(app.project_name)}
                  internalOnly
                  statusLabel={status.label}
                  statusVariant={status.variant}
                  icon={Box}
                />
              );
            })
          )}
        </ResourceListSection>
      )}

      <CreateApplicationModal
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
        onFinished={onProvisionFinished}
      />
    </DashboardShell>
  );
}
