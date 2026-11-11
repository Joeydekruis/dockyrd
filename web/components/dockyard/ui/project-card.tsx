import Link from "next/link";
import { Chip } from "@/components/dockyard/ui/chip";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";
import type { ProjectResourceRow } from "@/lib/projects-overview";

const statusDot: Record<StatusVariant, string> = {
  ok: "#1a7f4b",
  warn: "#d98a00",
  err: "#b3261e",
  idle: "#6b7686",
  info: "#3b93f5",
};

const statusInk: Record<StatusVariant, string> = {
  ok: "#176b3d",
  warn: "#8a5200",
  err: "#b3261e",
  idle: "#4a5361",
  info: "#0b3d91",
};

type ProjectCardProps = {
  name: string;
  href?: string;
  statusLabel?: string;
  statusVariant?: StatusVariant;
  environments?: string[];
  summaryLine?: string;
  activityLine?: string;
  resources?: ProjectResourceRow[];
  extraResourceCount?: number;
  /** @deprecated Use summaryLine for the projects overview layout. */
  repo?: string;
  /** @deprecated Use summaryLine for the projects overview layout. */
  servicesCount?: number;
};

function ResourceStatusPill({
  label,
  variant,
}: {
  label: string;
  variant: StatusVariant;
}) {
  return (
    <span
      className="inline-flex shrink-0 items-center gap-1.5 text-xs font-medium"
      style={{ color: statusInk[variant] }}
    >
      <span
        className="size-1.5 rounded-full"
        style={{ background: statusDot[variant] }}
        aria-hidden
      />
      {label}
    </span>
  );
}

function ResourceRow({ row }: { row: ProjectResourceRow }) {
  const Icon = row.icon;
  return (
    <div
      className="flex items-center gap-2.5 border-b py-2.5 last:border-b-0"
      style={{ borderColor: "#f1f0ed" }}
    >
      <span
        className="inline-flex shrink-0"
        style={{ color: "var(--muted-text)" }}
        aria-hidden
      >
        <Icon strokeWidth={1.8} className="size-4" />
      </span>
      <span className="flex min-w-0 flex-1 flex-wrap items-baseline gap-x-2 gap-y-0.5">
        <span className="font-semibold">{row.name}</span>
        <span className="text-xs" style={{ color: "var(--muted-text)" }}>
          {row.kindLabel}
        </span>
      </span>
      <ResourceStatusPill label={row.statusLabel} variant={row.statusVariant} />
    </div>
  );
}

export function ProjectCard({
  name,
  href = "#",
  statusLabel,
  statusVariant = "ok",
  environments,
  summaryLine,
  activityLine,
  resources,
  extraResourceCount = 0,
  repo,
  servicesCount,
}: ProjectCardProps) {
  const overviewLayout =
    environments != null ||
    summaryLine != null ||
    activityLine != null ||
    (resources != null && resources.length > 0);

  const visibleResources = resources?.slice(0, 5) ?? [];

  return (
    <Link
      href={href}
      className="dockyrd-surface-interactive flex h-full flex-col rounded-[20px] border p-5"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-lg font-semibold tracking-tight">{name}</p>
        {statusLabel ? (
          <StatusBadge label={statusLabel} variant={statusVariant} />
        ) : null}
      </div>

      {overviewLayout ? (
        <>
          {environments && environments.length > 0 ? (
            <div className="mt-3 flex flex-wrap gap-2">
              {environments.map((env) => (
                <Chip key={env}>{env}</Chip>
              ))}
            </div>
          ) : null}
          {visibleResources.length > 0 ? (
            <div
              className="mt-3 flex flex-col border-t"
              style={{ borderColor: "var(--border)" }}
            >
              {visibleResources.map((row) => (
                <ResourceRow key={row.id} row={row} />
              ))}
              {extraResourceCount > 0 ? (
                <p
                  className="pt-2 text-[13px] font-medium"
                  style={{ color: "var(--muted-text)" }}
                >
                  + {extraResourceCount} resource
                  {extraResourceCount === 1 ? "" : "s"}
                </p>
              ) : null}
            </div>
          ) : summaryLine ? (
            <p
              className="mt-3 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {summaryLine}
            </p>
          ) : null}
          {activityLine ? (
            <p
              className="mt-auto pt-4 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {activityLine}
            </p>
          ) : null}
        </>
      ) : (
        <>
          {repo ? (
            <p
              className="mt-1 font-mono text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {repo}
            </p>
          ) : null}
          {servicesCount != null ? (
            <p
              className="mt-4 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {servicesCount} services
            </p>
          ) : null}
        </>
      )}
    </Link>
  );
}
