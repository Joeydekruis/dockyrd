import Link from "next/link";
import { Chip } from "@/components/dockyrd/ui/chip";
import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyrd/ui/status-badge";

type ProjectCardProps = {
  name: string;
  href?: string;
  statusLabel?: string;
  statusVariant?: StatusVariant;
  environments?: string[];
  summaryLine?: string;
  activityLine?: string;
  /** @deprecated Use summaryLine for the projects overview layout. */
  repo?: string;
  /** @deprecated Use summaryLine for the projects overview layout. */
  servicesCount?: number;
};

export function ProjectCard({
  name,
  href = "#",
  statusLabel,
  statusVariant = "ok",
  environments,
  summaryLine,
  activityLine,
  repo,
  servicesCount,
}: ProjectCardProps) {
  const overviewLayout =
    environments != null || summaryLine != null || activityLine != null;

  return (
    <Link
      href={href}
      className="block rounded-[20px] border p-5 transition-colors hover:bg-[var(--hover-row)]"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <div className="flex items-start justify-between gap-3">
        <p className="text-lg font-semibold">{name}</p>
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
          {summaryLine ? (
            <p
              className="mt-3 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {summaryLine}
            </p>
          ) : null}
          {activityLine ? (
            <p
              className="mt-4 text-[13px]"
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
