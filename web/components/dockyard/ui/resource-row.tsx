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
      className="truncate text-[13px] font-medium hover:underline"
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

  return (
    <div
      className="resource-row grid grid-cols-1 items-center gap-4 border-t py-4 first:border-t-0 md:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)_auto_auto_minmax(0,1fr)]"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="flex min-w-0 items-center gap-4 md:col-span-1">
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
              className="mt-1 text-[13px] md:hidden"
              style={{ color: "var(--muted-text)" }}
            >
              {subMeta}
            </p>
          ) : null}
        </div>
      </div>

      {showUrl ? (
        <div className="min-w-0 md:text-center">
          <UrlCell url={url} urlHref={urlHref} internalOnly={internalOnly} />
        </div>
      ) : subMeta ? (
        <div
          className="min-w-0 text-[13px] md:text-center"
          style={{ color: "var(--muted-text)" }}
        >
          {subMeta}
        </div>
      ) : (
        <div className="hidden md:block" aria-hidden />
      )}

      <div className="flex flex-wrap items-center gap-2">
        <StatusBadge label={statusLabel} variant={statusVariant} />
        {trailing}
      </div>

      {showDeployMeta ? (
        <div className="justify-self-end">
          <CommitChip sha={sha ?? ""} message={commitMessage} timeAgo={timeAgo} />
        </div>
      ) : (
        <div className="hidden md:block" aria-hidden />
      )}
    </div>
  );
}
