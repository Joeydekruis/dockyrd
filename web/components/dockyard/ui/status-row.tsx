import type { ReactNode } from "react";
import { StatusBadge } from "@/components/dockyrd/ui/status-badge";

type StatusRowProps = {
  title: string;
  status: "ok" | "warn" | "fail";
  detail: string;
  actions?: ReactNode;
};

export function StatusRow({ title, status, detail, actions }: StatusRowProps) {
  const variant = status === "ok" ? "ok" : status === "warn" ? "warn" : "err";
  return (
    <div
      className="flex flex-col gap-3 border-t py-4 first:border-t-0 sm:flex-row sm:items-start sm:justify-between"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          <span
            className="text-sm font-medium"
            style={{ color: "var(--text)" }}
          >
            {title}
          </span>
          <StatusBadge variant={variant} label={status} />
        </div>
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          {detail}
        </p>
      </div>
      {actions ? (
        <div className="flex shrink-0 flex-wrap gap-2">{actions}</div>
      ) : null}
    </div>
  );
}
