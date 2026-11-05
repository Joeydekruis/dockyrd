"use client";

import { ExternalLink } from "lucide-react";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { DeploymentResourceStrip } from "@/components/dockyard/deployment-resource-strip";
import { DockyrdPageLoading } from "@/components/dockyard/dockyrd-page-loading";
import { Button } from "@/components/dockyard/ui/button";
import {
  SettingsBranchPill,
  SettingsCopyPill,
  SettingsMetaField,
  SettingsMetaGrid,
} from "@/components/dockyard/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import {
  type Deployment,
  fetchDeploymentLog,
  listDeployments,
  listProjects,
  listServiceDomains,
  listServices,
  type Project,
  type Service,
  type ServiceDomain,
} from "@/lib/api";
import {
  deploymentAuthorLabel,
  deploymentCommitShaShort,
  deploymentDurationSec,
  deploymentTriggerLabel,
  formatDeploymentStartedAt,
  formatDurationSec,
  primaryDomainHost,
  staticSiteDeploymentStatusDisplay,
  visitSiteHref,
} from "@/lib/static-site-display";
import { useRouteParam } from "@/lib/use-route-param";

type Row = {
  project: Project;
  service: Service;
  deployment: Deployment;
  domains: ServiceDomain[];
};

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
        const [{ deployments }, { domains }] = await Promise.all([
          listDeployments(service.id),
          listServiceDomains(service.id),
        ]);
        const deployment = deployments.find((d) => d.id === deploymentId);
        if (!deployment) continue;
        setRow({ project, service, deployment, domains });
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
      setLog(text || "");
      setLoadingLog(false);
    });
    return () => {
      cancelled = true;
    };
  }, [deploymentId]);

  const status = useMemo(
    () =>
      row ? staticSiteDeploymentStatusDisplay(row.deployment.status) : null,
    [row],
  );
  const duration = row ? deploymentDurationSec(row.deployment) : undefined;
  const sha = deploymentCommitShaShort(row?.deployment ?? ({} as Deployment));
  const visitHref = row
    ? visitSiteHref(primaryDomainHost(row.domains))
    : undefined;
  const branch = row?.deployment.branch?.trim() || "main";
  const startedAt = row?.deployment.started_at ?? row?.deployment.queued_at;
  const commitMessage = row?.deployment.commit_message?.trim();

  return (
    <DashboardShell title={sha} breadcrumb="Deployment">
      {loading ? (
        <DockyrdPageLoading />
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

          <section
            className="rounded-[20px] border px-6 py-5"
            style={{
              borderColor: "var(--border)",
              background: "var(--surface)",
            }}
            aria-label="Deployment summary"
          >
            <div className="flex flex-wrap items-start justify-between gap-4">
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2">
                  {status ? (
                    <StatusBadge
                      label={status.label}
                      variant={status.variant}
                    />
                  ) : null}
                  {row.deployment.status.trim().toLowerCase() === "live" ? (
                    <StatusBadge label="Production" variant="info" />
                  ) : null}
                </div>
                <SettingsBranchPill branch={branch} />
              </div>
              {visitHref ? (
                <Button
                  href={visitHref}
                  variant="outline"
                  target="_blank"
                  iconRight={ExternalLink}
                >
                  Visit site
                </Button>
              ) : null}
            </div>
          </section>

          <DeploymentResourceStrip deployment={row.deployment} />

          <section
            className="rounded-[20px] border px-6 py-5"
            style={{
              borderColor: "var(--border)",
              background: "var(--surface)",
            }}
            aria-label="Details"
          >
            <h2 className="text-lg font-semibold tracking-tight">Details</h2>
            <SettingsMetaGrid>
              <SettingsMetaField
                label="Deployed branch"
                value={<SettingsBranchPill branch={branch} />}
              />
              <SettingsMetaField
                label="Commit"
                value={
                  row.deployment.commit_sha ? (
                    <SettingsCopyPill value={sha} />
                  ) : (
                    "—"
                  )
                }
              />
              <SettingsMetaField
                label="By"
                value={deploymentAuthorLabel(row.deployment)}
              />
              <SettingsMetaField
                label="Trigger"
                value={deploymentTriggerLabel(row.deployment.trigger)}
              />
              <SettingsMetaField
                label="Started at"
                value={formatDeploymentStartedAt(startedAt)}
              />
            </SettingsMetaGrid>
            {commitMessage ? (
              <div
                className="mt-8 border-t pt-6"
                style={{ borderColor: "var(--border)" }}
              >
                <p
                  className="font-mono text-[11px] uppercase tracking-widest"
                  style={{ color: "var(--muted-text)" }}
                >
                  Commit message
                </p>
                <p
                  className="mt-2 whitespace-pre-wrap text-sm leading-relaxed"
                  style={{ color: "var(--ink-navy)" }}
                >
                  {commitMessage}
                </p>
              </div>
            ) : null}
          </section>

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
            aria-label="Deployment logs"
            className="overflow-hidden rounded-[20px] border"
            style={{
              borderColor: "var(--border)",
              background: "var(--surface)",
            }}
          >
            <div
              className="flex flex-wrap items-center justify-between gap-3 border-b px-6 py-4"
              style={{ borderColor: "var(--border)" }}
            >
              <div className="flex flex-wrap items-center gap-3">
                <h2 className="text-lg font-semibold tracking-tight">
                  Deployment logs
                </h2>
                {status ? (
                  <StatusBadge label={status.label} variant={status.variant} />
                ) : null}
                {duration !== undefined ? (
                  <span
                    className="font-mono text-[13px] tabular-nums"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {formatDurationSec(duration)}
                  </span>
                ) : null}
              </div>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setLoadingLog(true);
                  void fetchDeploymentLog(row.deployment.id).then((text) => {
                    setLog(text || "");
                    setLoadingLog(false);
                  });
                }}
              >
                Refresh
              </Button>
            </div>
            {!loadingLog && !log.trim() ? (
              <div className="px-6 py-10 text-center">
                <p className="text-sm font-medium">No logs found</p>
                <p
                  className="mt-2 text-sm"
                  style={{ color: "var(--muted-text)" }}
                >
                  Logs can take up to a minute to appear after a deploy
                  finishes.
                </p>
              </div>
            ) : (
              <pre
                className="max-h-[min(60vh,520px)] overflow-auto p-6 font-mono text-[12px] leading-relaxed whitespace-pre-wrap"
                style={{ color: "var(--ink-navy)" }}
              >
                {loadingLog ? "Loading log…" : log}
              </pre>
            )}
          </section>
        </div>
      )}
    </DashboardShell>
  );
}
