import { CircleCheck, CircleX, LoaderCircle } from "lucide-react";
import Link from "next/link";
import { DeploymentAuthorAvatar } from "@/components/dockyard/ui/deployment-author-avatar";
import { GitHubRepoLink } from "@/components/dockyard/ui/github-repo-link";
import { SettingsBranchPill } from "@/components/dockyard/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";
import type { Deployment } from "@/lib/api";
import { githubRepoWebUrl } from "@/lib/github-url";
import { formatTimeAgo } from "@/lib/resource-dashboard";
import {
  deploymentCommitShaShort,
  deploymentDurationSec,
  deploymentSourceTitle,
  deploymentTriggerLabel,
  formatDurationSec,
  staticSiteDeploymentStatusDisplay,
} from "@/lib/static-site-display";

const TABLE_GRID =
  "md:grid md:grid-cols-[minmax(0,1fr)_minmax(120px,1fr)_96px_72px] md:items-center md:gap-4";

function DeploymentStatusIcon({ status }: { status: string }) {
  const s = status.trim().toLowerCase();
  if (s === "live") {
    return (
      <CircleCheck
        className="size-5 shrink-0"
        strokeWidth={1.8}
        style={{ color: "#1a7f4b" }}
        aria-hidden
      />
    );
  }
  if (s === "failed" || s === "error") {
    return (
      <CircleX
        className="size-5 shrink-0"
        strokeWidth={1.8}
        style={{ color: "var(--danger)" }}
        aria-hidden
      />
    );
  }
  if (
    s === "building" ||
    s === "queued" ||
    s === "running" ||
    s === "starting"
  ) {
    return (
      <LoaderCircle
        className="size-5 shrink-0 animate-spin"
        strokeWidth={1.8}
        style={{ color: "#d98a00" }}
        aria-hidden
      />
    );
  }
  return (
    <span
      className="inline-block size-2 shrink-0 rounded-full"
      style={{ background: "var(--muted-text)" }}
      aria-hidden
    />
  );
}

function deploymentFullCommitMessage(d: Deployment): string | undefined {
  const msg = d.commit_message?.trim();
  return msg || undefined;
}

function isManualTrigger(trigger: string): boolean {
  const t = trigger.trim().toLowerCase();
  return t === "manual" || t === "redeploy";
}

export type ServiceDeploymentTableProps = {
  deployments: Deployment[];
  resourceSlug: string;
  resourceKind: "applications" | "static-sites";
  sourceRepo?: string;
  defaultBranch?: string;
  compact?: boolean;
};

export function ServiceDeploymentTable({
  deployments,
  resourceSlug,
  resourceKind,
  sourceRepo,
  defaultBranch,
  compact = false,
}: ServiceDeploymentTableProps) {
  const repoHref = sourceRepo ? githubRepoWebUrl(sourceRepo) : undefined;

  return (
    <div>
      {!compact && sourceRepo && repoHref ? (
        <div
          className="flex flex-wrap items-center gap-3 border-b pb-3"
          style={{ borderColor: "var(--border)" }}
        >
          <GitHubRepoLink fullName={sourceRepo} href={repoHref} />
          {defaultBranch ? <SettingsBranchPill branch={defaultBranch} /> : null}
        </div>
      ) : null}
      {!compact ? (
        <div
          className={`hidden border-b py-2.5 text-[11px] font-medium uppercase tracking-widest ${TABLE_GRID}`}
          style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
        >
          <span>Source</span>
          <span>Trigger</span>
          <span className="text-right">Deployed</span>
          <span className="text-right">Duration</span>
        </div>
      ) : null}
      <ul className="list-none p-0">
        {deployments.map((deployment) => (
          <ServiceDeploymentTableRow
            key={deployment.id}
            deployment={deployment}
            resourceSlug={resourceSlug}
            resourceKind={resourceKind}
            compact={compact}
          />
        ))}
      </ul>
    </div>
  );
}

function ServiceDeploymentTableRow({
  deployment,
  resourceSlug,
  resourceKind,
  compact,
}: {
  deployment: Deployment;
  resourceSlug: string;
  resourceKind: "applications" | "static-sites";
  compact?: boolean;
}) {
  const st = staticSiteDeploymentStatusDisplay(deployment.status);
  const dur = deploymentDurationSec(deployment);
  const whenTs =
    deployment.finished_at ?? deployment.started_at ?? deployment.queued_at;
  const timeAgo = formatTimeAgo(whenTs) ?? "—";
  const branch = deployment.branch?.trim() || "main";
  const sha = deploymentCommitShaShort(deployment);
  const title = deploymentSourceTitle(deployment);
  const fullMessage = deploymentFullCommitMessage(deployment);
  const triggerBase = deploymentTriggerLabel(deployment.trigger);
  const manual = isManualTrigger(deployment.trigger);
  const triggeredBy = deployment.triggered_by_name?.trim();
  const commitAuthor = deployment.commit_author?.trim();
  const href = `/${resourceKind}/${resourceSlug}/deployments/${deployment.id}`;

  const metadataAuthor = manual
    ? triggeredBy || commitAuthor
    : commitAuthor || triggeredBy;

  return (
    <li
      className="border-t first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <Link
        href={href}
        className={`block py-4 transition-colors hover:bg-[var(--hover-row)] ${TABLE_GRID}`}
      >
        <div className="flex min-w-0 gap-3">
          <DeploymentStatusIcon status={deployment.status} />
          <div className="min-w-0 flex-1">
            <p
              className="line-clamp-2 text-[15px] font-semibold leading-snug tracking-tight"
              title={fullMessage ?? title}
            >
              {title}
            </p>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <StatusBadge label={st.label} variant={st.variant} />
              {metadataAuthor ? (
                <DeploymentAuthorAvatar name={metadataAuthor} />
              ) : null}
              <code
                className="rounded-md px-2 py-0.5 font-mono text-[11px] uppercase"
                style={{
                  background: "var(--hover-row)",
                  color: "var(--ink-navy)",
                }}
              >
                {sha}
              </code>
              <SettingsBranchPill branch={branch} />
            </div>
            {compact ? (
              <p
                className="mt-2 text-[13px] md:hidden"
                style={{ color: "var(--muted-text)" }}
              >
                {manual && triggeredBy
                  ? `Manual · ${triggeredBy}`
                  : triggerBase}
                {dur !== undefined ? ` · ${formatDurationSec(dur)}` : null}
                {" · "}
                {timeAgo}
              </p>
            ) : null}
          </div>
        </div>
        {!compact ? (
          <>
            <p
              className="mt-3 flex flex-wrap items-center gap-2 text-[13px] md:mt-0"
              style={{ color: "var(--muted-text)" }}
            >
              <span>{manual ? "Manual" : triggerBase}</span>
              {manual && triggeredBy ? (
                <>
                  <DeploymentAuthorAvatar name={triggeredBy} />
                  <span
                    className="font-medium"
                    style={{ color: "var(--ink-navy)" }}
                  >
                    {triggeredBy}
                  </span>
                </>
              ) : null}
            </p>
            <p
              className="mt-2 text-right text-[13px] md:mt-0"
              style={{ color: "var(--muted-text)" }}
            >
              {timeAgo}
            </p>
            <p
              className="mt-1 text-right font-mono text-[12px] tabular-nums md:mt-0"
              style={{ color: "var(--muted-text)" }}
            >
              {dur !== undefined ? formatDurationSec(dur) : "—"}
            </p>
          </>
        ) : null}
      </Link>
    </li>
  );
}
