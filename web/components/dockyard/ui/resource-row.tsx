import type { LucideIcon } from "lucide-react";
import { Box } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";
import { CommitChip } from "@/components/dockyard/ui/commit-chip";
import { StatusBadge } from "@/components/dockyard/ui/status-badge";

type ResourceRowProps = {
  name: string;
  href?: string;
  meta?: string;
  subMeta?: string;
  statusLabel: string;
  statusVariant?: "ok" | "warn" | "err" | "idle" | "info";
  sha?: string;
  commitMessage?: string;
  icon?: LucideIcon;
  trailing?: ReactNode;
};

function RowName({ name, href }: { name: string; href?: string }) {
  const className = "font-semibold hover:underline";
  const style = { color: "var(--ink-navy)" };
  if (href && href !== "#") {
    return (
      <Link href={href} className={className} style={style}>
        {name}
      </Link>
    );
  }
  return (
    <span className={className} style={style}>
      {name}
    </span>
  );
}

export function ResourceRow({
  name,
  href = "#",
  meta,
  subMeta,
  statusLabel,
  statusVariant = "ok",
  sha,
  commitMessage,
  icon: Icon = Box,
  trailing,
}: ResourceRowProps) {
  return (
    <div
      className="flex flex-wrap items-center gap-4 border-t py-4 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <div
        className="flex size-10 shrink-0 items-center justify-center rounded-xl"
        style={{ background: "var(--brand-soft)", color: "var(--link-blue)" }}
      >
        <Icon strokeWidth={1.8} className="size-5" aria-hidden />
      </div>
      <div className="min-w-0 flex-1">
        <RowName name={name} href={href} />
        {meta ? (
          <p
            className="mt-0.5 font-mono text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            {meta}
          </p>
        ) : null}
        {subMeta ? (
          <p
            className="mt-1 text-[13px]"
            style={{ color: "var(--muted-text)" }}
          >
            {subMeta}
          </p>
        ) : null}
      </div>
      {trailing}
      <StatusBadge label={statusLabel} variant={statusVariant} />
      {sha ? <CommitChip sha={sha} message={commitMessage} /> : null}
    </div>
  );
}
