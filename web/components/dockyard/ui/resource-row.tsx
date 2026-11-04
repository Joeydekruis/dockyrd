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
  timeAgo?: string;
  url?: string;
  urlHref?: string;
  internalOnly?: boolean;
  icon?: LucideIcon;
  trailing?: ReactNode;
};

function RowName({ name, navigable }: { name: string; navigable: boolean }) {
  const className = navigable
    ? "font-semibold"
    : "font-semibold hover:underline";
  const style = { color: "var(--ink-navy)" };
  return (
    <span className={className} style={style}>
      {name}
    </span>
  );
}

function UrlCell({
  url,
  urlHref,
  internalOnly,
}: {
  url?: string;
  urlHref?: string;
  internalOnly?: boolean;
}) {
  if (internalOnly) {
    return (
      <span
        className="text-[13px] font-medium"
        style={{ color: "var(--muted-text)" }}
      >
        Internal only
      </span>
    );
  }
  if (!url) {
    return null;
  }
  const href = urlHref ?? (url.startsWith("http") ? url : `https://${url}`);
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="relative z-[2] truncate text-[13px] font-medium hover:underline pointer-events-auto"
      style={{ color: "var(--link-blue)" }}
    >
      {url}
    </a>
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
  timeAgo,
  url,
  urlHref,
  internalOnly,
  icon: Icon = Box,
  trailing,
}: ResourceRowProps) {
  const showDeployMeta = Boolean(sha);
  const showUrl = Boolean(url || internalOnly);
  const navigable = Boolean(href && href !== "#");
  const cellClass = navigable
    ? "relative z-[1] pointer-events-none"
    : undefined;

  return (
    <div
      className={`resource-row relative -mx-6 grid grid-cols-1 items-center gap-4 border-t px-6 py-4 first:border-t-0 md:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)_auto_auto_minmax(0,1fr)]${navigable ? " transition-colors hover:bg-[var(--hover-row)]" : ""}`}
      style={{ borderColor: "var(--border)" }}
    >
      {navigable ? (
        <Link
          href={href}
          className="absolute inset-0 z-0"
          aria-label={`View ${name}`}
        />
      ) : null}
      <div
        className={`flex min-w-0 items-center gap-4 md:col-span-1 ${cellClass ?? ""}`}
      >
        <div
          className="flex size-10 shrink-0 items-center justify-center rounded-xl"
          style={{ background: "var(--brand-soft)", color: "var(--link-blue)" }}
        >
          <Icon strokeWidth={1.8} className="size-5" aria-hidden />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <RowName name={name} navigable={navigable} />
            <div className="shrink-0 md:hidden">
              <StatusBadge label={statusLabel} variant={statusVariant} />
            </div>
          </div>
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
              className="mt-1 text-[13px] md:hidden"
              style={{ color: "var(--muted-text)" }}
            >
              {subMeta}
            </p>
          ) : null}
        </div>
      </div>

      {showUrl ? (
        <div className={`min-w-0 md:text-center ${cellClass ?? ""}`}>
          <UrlCell url={url} urlHref={urlHref} internalOnly={internalOnly} />
        </div>
      ) : subMeta ? (
        <div
          className={`min-w-0 text-[13px] md:text-center ${cellClass ?? ""}`}
          style={{ color: "var(--muted-text)" }}
        >
          {subMeta}
        </div>
      ) : (
        <div className={`hidden md:block ${cellClass ?? ""}`} aria-hidden />
      )}

      <div
        className={`hidden flex-wrap items-center gap-2 md:flex ${cellClass ?? ""}`}
      >
        <StatusBadge label={statusLabel} variant={statusVariant} />
        {trailing}
      </div>

      {showDeployMeta ? (
        <div className={`justify-self-end ${cellClass ?? ""}`}>
          <CommitChip
            sha={sha ?? ""}
            message={commitMessage}
            timeAgo={timeAgo}
          />
        </div>
      ) : (
        <div className={`hidden md:block ${cellClass ?? ""}`} aria-hidden />
      )}
    </div>
  );
}
