"use client";

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
  listServices,
  type Project,
  type Service,
} from "@/lib/api";
import {
  deploymentAuthorLabel,
  deploymentCommitDescription,
  deploymentCommitShaShort,
  deploymentDisplayName,
  deploymentDurationSec,
  deploymentIsInProgress,
  deploymentReleaseTag,
  deploymentTriggerLabel,
  formatDeploymentStartedAt,
  formatDurationSec,
  staticSiteDeploymentStatusDisplay,
  truncateDeploymentTitle,
} from "@/lib/static-site-display";
import { useDeploymentPoll } from "@/lib/use-deployment-poll";
import { useRouteParam } from "@/lib/use-route-param";

type Row = {
  project: Project;
  service: Service;
  deployment: Deployment;
};

export function ApplicationDeploymentDetailPage() {
  const appSlug = useRouteParam("slug");
  const deploymentId = useRouteParam("deploymentId");
  const [row, setRow] = useState<Row | null>(null);
  const [log, setLog] = useState("");
  const [loadingLog, setLoadingLog] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(
    async (opts?: { silent?: boolean }) => {
      if (!appSlug || !deploymentId) {
        setLoading(false);
        return;
      }
      if (!opts?.silent) setLoading(true);
      try {
        const { projects } = await listProjects();
        for (const project of projects) {
          const { services } = await listServices(project.id, "app");
          const service = services.find((s) => s.slug === appSlug);
          if (!service) continue;
          const { deployments } = await listDeployments(service.id);
          const deployment = deployments.find((d) => d.id === deploymentId);
          if (!deployment) continue;
          setRow({ project, service, deployment });
          setError(null);
          if (!opts?.silent) setLoading(false);
          return;
        }
        setRow(null);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to load deployment");
      } finally {
        if (!opts?.silent) setLoading(false);
      }
    },
    [appSlug, deploymentId],
  );

  const refreshLog = useCallback(async () => {
    if (!deploymentId) return;
    try {
      const text = await fetchDeploymentLog(deploymentId);
      setLog(text || "");
      setLoadingLog(false);
    } catch {
      /* keep last log */
    }
  }, [deploymentId]);

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

  const deploymentPollActive = useMemo(
    () => (row ? deploymentIsInProgress(row.deployment.status) : false),
    [row],
  );

  const pollDeployment = useCallback(() => {
    void refresh({ silent: true });
    void refreshLog();
  }, [refresh, refreshLog]);

  useDeploymentPoll(deploymentPollActive, pollDeployment);

  const status = useMemo(
    () =>
      row ? staticSiteDeploymentStatusDisplay(row.deployment.status) : null,
    [row],
  );
  const duration = row ? deploymentDurationSec(row.deployment) : undefined;
  const sha = deploymentCommitShaShort(row?.deployment ?? ({} as Deployment));
  const branch = row?.deployment.branch?.trim() || "main";
  const startedAt = row?.deployment.started_at ?? row?.deployment.queued_at;
  const commitMessage = row?.deployment.commit_message?.trim();
  const deploymentName = row
    ? deploymentDisplayName(row.deployment)
    : "Deployment";
  const shellTitle = truncateDeploymentTitle(deploymentName);
  const commitDescription = row
    ? deploymentCommitDescription(row.deployment)
    : undefined;
  const releaseTag = row ? deploymentReleaseTag(row.deployment) : undefined;

  return (
    <DashboardShell title={shellTitle} breadcrumb="Deployment">
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
              href={`/applications/${row.service.slug}?tab=deployments`}
              className="hover:underline"
              style={{ color: "var(--link-blue)" }}
            >
              ← {row.service.name}
            </Link>
          </p>

          <section className="py-2" aria-label="Deployment summary">
            <div className="flex flex-wrap lg:flex-nowrap items-start justify-between gap-4">
              <div className="min-w-0 space-y-3">
                <div>
                  <h1 className="text-2xl font-semibold tracking-tight">
                    {deploymentName}
                  </h1>
                  {commitDescription ? (
                    <p
                      className="mt-2 whitespace-pre-wrap text-sm leading-relaxed"
                      style={{ color: "var(--muted-text)" }}
                    >
                      {commitDescription}
                    </p>
                  ) : null}
                  <p
                    className="mt-2 flex flex-wrap items-center gap-2 font-mono text-[12px] uppercase tracking-wide"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {sha !== "—" ? <SettingsCopyPill value={sha} /> : null}
                    {releaseTag ? (
                      <>
                        {sha !== "—" ? <span aria-hidden>·</span> : null}
                        <span>
                          Release{" "}
                          <span style={{ color: "var(--ink-navy)" }}>
                            {releaseTag}
                          </span>
                        </span>
                      </>
                    ) : null}
                  </p>
                </div>
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
                  <SettingsBranchPill branch={branch} />
                </div>
              </div>
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
              {releaseTag ? (
                <SettingsMetaField label="Release" value={releaseTag} />
              ) : null}
              <SettingsMetaField
                label="Started at"
                value={formatDeploymentStartedAt(startedAt)}
              />
              <SettingsMetaField
                label="Duration"
                value={
                  duration !== undefined ? formatDurationSec(duration) : "—"
                }
              />
            </SettingsMetaGrid>
            {commitMessage && !commitDescription ? (
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
