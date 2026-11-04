"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { CommitChip } from "@/components/dockyard/ui/commit-chip";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  type Deployment,
  fetchDeploymentLog,
  listDeployments,
  listProjects,
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
import {
  deploymentDurationSec,
  deploymentStatusDisplay,
  formatDeploymentWhen,
  formatDurationSec,
} from "@/lib/static-site-display";
import { useRouteParam } from "@/lib/use-route-param";

type Row = { project: Project; service: Service; deployment: Deployment };

export function StaticSiteDeploymentDetailPage() {
  const siteSlug = useRouteParam("slug");
  const deploymentId = useRouteParam("deploymentId");
  const [row, setRow] = useState<Row | null>(null);
  const [log, setLog] = useState("");
  const [loadingLog, setLoadingLog] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!siteSlug || !deploymentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const { projects } = await listProjects();
      for (const project of projects) {
        const { services } = await listServices(project.id, "static_site");
        const service = services.find((s) => s.slug === siteSlug);
        if (!service) continue;
        const { deployments } = await listDeployments(service.id);
        const deployment = deployments.find((d) => d.id === deploymentId);
        if (!deployment) continue;
        setRow({ project, service, deployment });
        setError(null);
        setLoading(false);
        return;
      }
      setRow(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load deployment");
    } finally {
      setLoading(false);
    }
  }, [siteSlug, deploymentId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    if (!deploymentId) return;
    let cancelled = false;
    setLoadingLog(true);
    void fetchDeploymentLog(deploymentId).then((text) => {
      if (cancelled) return;
      setLog(text || "No build log captured for this deployment.");
      setLoadingLog(false);
    });
    return () => {
      cancelled = true;
    };
  }, [deploymentId]);

  const status = useMemo(
    () => (row ? deploymentStatusDisplay(row.deployment.status) : null),
    [row],
  );
  const duration = row ? deploymentDurationSec(row.deployment) : undefined;
  const sha = row?.deployment.commit_sha?.slice(0, 7) ?? "—";

  return (
    <DashboardShell title={sha} breadcrumb="Deployment">
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
          Deployment not found.
        </p>
      ) : (
        <div className="space-y-6">
          <p className="text-sm">
            <Link
              href={`/static-sites/${row.service.slug}?tab=deployments`}
              className="hover:underline"
              style={{ color: "var(--link-blue)" }}
            >
              ← {row.service.name}
            </Link>
          </p>
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div>
              <div className="flex flex-wrap items-center gap-3">
                {status ? (
                  <StatusBadge label={status.label} variant={status.variant} />
                ) : null}
                <CommitChip
                  sha={row.deployment.commit_sha ?? ""}
                  message={row.deployment.branch ?? row.deployment.trigger}
                />
              </div>
              <p
                className="mt-2 text-sm"
                style={{ color: "var(--muted-text)" }}
              >
                {formatDeploymentWhen(
                  row.deployment.finished_at ?? row.deployment.started_at,
                )}
                {duration !== undefined
                  ? ` · ${formatDurationSec(duration)}`
                  : null}
              </p>
            </div>
          </div>
          {row.deployment.error ? (
            <div
              className="rounded-xl border px-4 py-3 text-sm text-red-700"
              style={{
                borderColor: "var(--border)",
                background: "var(--hover-row)",
              }}
              role="alert"
            >
              {row.deployment.error}
            </div>
          ) : null}
          <section
            aria-label="Build log"
            className="overflow-hidden rounded-[20px] border"
            style={{
              borderColor: "var(--border)",
              background: "var(--surface)",
            }}
          >
            <div
              className="flex items-center justify-between border-b px-6 py-4"
              style={{ borderColor: "var(--border)" }}
            >
              <h2 className="text-lg font-semibold tracking-tight">
                Build log
              </h2>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setLoadingLog(true);
                  void fetchDeploymentLog(row.deployment.id).then((text) => {
                    setLog(
                      text || "No build log captured for this deployment.",
                    );
                    setLoadingLog(false);
                  });
                }}
              >
                Refresh
              </Button>
            </div>
            <pre
              className="max-h-[min(60vh,520px)] overflow-auto p-6 font-mono text-[12px] leading-relaxed whitespace-pre-wrap"
              style={{ color: "var(--ink-navy)" }}
            >
              {loadingLog ? "Loading log…" : log}
            </pre>
          </section>
        </div>
      )}
    </DashboardShell>
  );
}
