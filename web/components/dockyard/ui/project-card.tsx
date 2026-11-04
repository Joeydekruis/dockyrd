import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

type ProjectCardProps = {
  name: string;
  repo?: string;
  servicesCount?: number;
  statusLabel?: string;
  statusVariant?: StatusVariant;
  href?: string;
};

export function ProjectCard({
  name,
  repo,
  servicesCount,
  statusLabel,
  statusVariant = "ok",
  href = "#",
}: ProjectCardProps) {
  return (
    <a
      href={href}
      className="block rounded-[20px] border p-5 transition-colors hover:bg-[var(--hover-row)]"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <p className="text-lg font-semibold">{name}</p>
      {repo ? (
        <p
          className="mt-1 font-mono text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {repo}
        </p>
      ) : null}
      <div className="mt-4 flex flex-wrap items-center gap-3">
        {statusLabel ? (
          <StatusBadge label={statusLabel} variant={statusVariant} />
        ) : null}
        {servicesCount != null ? (
          <span className="text-[13px]" style={{ color: "var(--muted-text)" }}>
            {servicesCount} services
          </span>
        ) : null}
      </div>
    </a>
  );
}
