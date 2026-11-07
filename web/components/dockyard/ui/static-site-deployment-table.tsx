import { CircleCheck, CircleX, LoaderCircle } from "lucide-react";
import Link from "next/link";
import { SettingsBranchPill } from "@/components/dockyrd/ui/settings-detail-block";
import { StatusBadge } from "@/components/dockyrd/ui/status-badge";
import type { Deployment } from "@/lib/api";
import { formatTimeAgo } from "@/lib/resource-dashboard";
import {
  deploymentCommitShaShort,
  deploymentDurationSec,
  deploymentSourceTitle,
  deploymentTriggerLabel,
  formatDurationSec,
  staticSiteDeploymentStatusDisplay,
} from "@/lib/static-site-display";

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

type StaticSiteDeploymentTableProps = {
  deployments: Deployment[];
  siteSlug: string;
  sourceRepo?: string;
  defaultBranch?: string;
  compact?: boolean;
};

export function StaticSiteDeploymentTable({
  deployments,
  siteSlug,
  sourceRepo,
  defaultBranch,
  compact = false,
}: StaticSiteDeploymentTableProps) {
  const showSourceHeader = !compact && sourceRepo;

  return (
    <div>
      {showSourceHeader ? (
        <p
          className="border-b pb-3 font-mono text-[12px] uppercase tracking-wide"
          style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
        >
          {sourceRepo}
          {defaultBranch ? ` · ${defaultBranch}` : null}
        </p>
      ) : null}
      {!compact ? (
        <div
          className="hidden border-b py-2.5 text-[11px] font-medium uppercase tracking-widest md:grid md:grid-cols-[minmax(0,1fr)_160px_140px]"
          style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
        >
          <span>Source</span>
          <span>Trigger</span>
          <span className="text-right">Date · duration</span>
        </div>
      ) : null}
      <ul className="list-none p-0">
        {deployments.map((deployment) => (
          <StaticSiteDeploymentTableRow
            key={deployment.id}
            deployment={deployment}
            siteSlug={siteSlug}
            compact={compact}
          />
        ))}
      </ul>
    </div>
  );
}

function StaticSiteDeploymentTableRow({
  deployment,
  siteSlug,
  compact,
}: {
  deployment: Deployment;
  siteSlug: string;
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
  const trigger = deploymentTriggerLabel(deployment.trigger);
  const href = `/static-sites/${siteSlug}/deployments/${deployment.id}`;

  return (
    <li
      className="border-t first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <Link
        href={href}
        className="block py-4 transition-colors hover:bg-[var(--hover-row)] md:grid md:grid-cols-[minmax(0,1fr)_160px_140px] md:items-center md:gap-4"
      >
        <div className="flex min-w-0 gap-3">
          <DeploymentStatusIcon status={deployment.status} />
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="min-w-0 truncate text-[15px] font-semibold tracking-tight">
                {title}
              </span>
              <StatusBadge label={st.label} variant={st.variant} />
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-2">
              <SettingsBranchPill branch={branch} />
              <code
                className="rounded-md px-2 py-0.5 font-mono text-[11px] uppercase"
                style={{
                  background: "var(--hover-row)",
                  color: "var(--ink-navy)",
                }}
              >
                {sha}
              </code>
            </div>
            {compact ? (
              <p
                className="mt-2 text-[13px] md:hidden"
                style={{ color: "var(--muted-text)" }}
              >
                {trigger}
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
              className="mt-3 text-[13px] md:mt-0"
              style={{ color: "var(--muted-text)" }}
            >
              {trigger}
            </p>
            <div className="mt-2 text-right md:mt-0">
              <p className="text-[13px]" style={{ color: "var(--muted-text)" }}>
                {timeAgo}
              </p>
              {dur !== undefined ? (
                <p
                  className="font-mono text-[12px] tabular-nums"
                  style={{ color: "var(--muted-text)" }}
                >
                  {formatDurationSec(dur)}
                </p>
              ) : null}
            </div>
          </>
        ) : null}
      </Link>
    </li>
  );
}
