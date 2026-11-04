import {
  StatusBadge,
  type StatusVariant,
} from "@/components/dockyard/ui/status-badge";

type DomainRowProps = {
  hostname: string;
  statusLabel: string;
  statusVariant?: StatusVariant;
  primary?: boolean;
  href?: string;
};

export function DomainRow({
  hostname,
  statusLabel,
  statusVariant = "ok",
  primary,
  href = "#",
}: DomainRowProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-3 border-t py-3 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <a
        href={href}
        className="min-w-0 flex-1 font-mono text-sm hover:underline"
      >
        {hostname}
      </a>
      {primary ? (
        <span
          className="font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          Primary
        </span>
      ) : null}
      <StatusBadge label={statusLabel} variant={statusVariant} />
    </div>
  );
}
