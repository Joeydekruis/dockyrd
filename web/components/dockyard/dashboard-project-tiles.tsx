import Link from "next/link";
import type { ProjectOverview } from "@/lib/projects-overview";

const dotByVariant = {
  ok: "#1a7f4b",
  warn: "#d98a00",
  err: "#b3261e",
  idle: "#6b7686",
  info: "#3b93f5",
} as const;

type DashboardProjectTilesProps = {
  projects: ProjectOverview[];
};

export function DashboardProjectTiles({
  projects,
}: DashboardProjectTilesProps) {
  if (projects.length === 0) {
    return (
      <p className="text-sm" style={{ color: "var(--muted-text)" }}>
        No projects yet. Create one from the banner to get started.
      </p>
    );
  }

  const tiles = projects.slice(0, 8);

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {tiles.map((row) => {
        const resourceLabel =
          row.serviceCount === 1
            ? "1 resource"
            : `${row.serviceCount} resources`;
        return (
          <Link
            key={row.project.id}
            href={`/projects/${row.project.slug}`}
            className="dockyrd-surface-interactive flex min-w-0 flex-col gap-1.5 rounded-2xl border px-4 py-3.5"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
            }}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="truncate text-[15px] font-semibold tracking-tight">
                {row.project.name}
              </span>
              <span
                className="size-2 shrink-0 rounded-full"
                style={{ background: dotByVariant[row.status.variant] }}
                title={row.status.label}
              />
            </div>
            <span
              className="truncate text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {resourceLabel} · {row.status.label}
            </span>
          </Link>
        );
      })}
    </div>
  );
}
