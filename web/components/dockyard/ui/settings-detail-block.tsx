"use client";

import type { LucideIcon } from "lucide-react";
import { Circle, CircleCheck, Copy, GitBranch } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import { useCopyToClipboard } from "@/lib/snackbar-context";

export type SettingsDetailRow = {
  label: string;
  value: ReactNode;
  /** Wider column span on large grids (e.g. repository). */
  span?: 2 | 3;
};

type SettingsGroupHeadingProps = {
  icon: LucideIcon;
  label: string;
  tone?: "default" | "danger";
};

export function SettingsGroupHeading({
  icon: Icon,
  label,
  tone = "default",
}: SettingsGroupHeadingProps) {
  const color = tone === "danger" ? "var(--danger)" : "var(--muted-text)";
  return (
    <div className="mb-3 flex items-center gap-2.5">
      <Icon className="size-4 shrink-0" strokeWidth={1.8} style={{ color }} />
      <span
        className="font-mono text-[11px] font-semibold uppercase tracking-[0.14em]"
        style={{ color }}
      >
        {label}
      </span>
    </div>
  );
}

export function SettingsMetaField({
  label,
  value,
  span,
}: {
  label: string;
  value: ReactNode;
  span?: 2 | 3;
}) {
  const spanClass =
    span === 3
      ? "sm:col-span-2 lg:col-span-3"
      : span === 2
        ? "sm:col-span-2"
        : "";
  return (
    <div className={spanClass}>
      <p
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </p>
      <div
        className="mt-2 text-sm font-medium"
        style={{ color: "var(--ink-navy)" }}
      >
        {value}
      </div>
    </div>
  );
}

export function SettingsMetaGrid({ children }: { children: ReactNode }) {
  return (
    <div className="mt-5 grid gap-x-8 gap-y-6 sm:grid-cols-2 lg:grid-cols-3">
      {children}
    </div>
  );
}

export function SettingsCopyPill({ value }: { value: string }) {
  const copyToClipboard = useCopyToClipboard();
  return (
    <span
      className="inline-flex max-w-full items-center gap-1 rounded-lg border py-1 pl-2.5 pr-1 font-mono text-[13px]"
      style={{
        borderColor: "var(--border)",
        background: "var(--hover-row)",
      }}
    >
      <span className="truncate">{value}</span>
      <IconButton
        icon={Copy}
        label="Copy"
        onClick={() => {
          void copyToClipboard(value);
        }}
      />
    </span>
  );
}

export function SettingsEnabledStatus({ enabled }: { enabled: boolean }) {
  const Icon = enabled ? CircleCheck : Circle;
  return (
    <span className="inline-flex items-center gap-2 text-sm font-medium">
      <Icon
        className="size-4 shrink-0"
        strokeWidth={1.8}
        style={{ color: enabled ? "#2d8a4e" : "var(--muted-text)" }}
        aria-hidden
      />
      {enabled ? "Enabled" : "Disabled"}
    </span>
  );
}

export function SettingsBranchPill({ branch }: { branch: string }) {
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 font-mono text-[13px]"
      style={{
        borderColor: "var(--border)",
        background: "var(--hover-row)",
      }}
    >
      <GitBranch className="size-3.5 shrink-0" strokeWidth={1.8} aria-hidden />
      {branch}
    </span>
  );
}

type SettingsDetailBlockProps = {
  title: string;
  description?: string;
  rows: SettingsDetailRow[];
  actionLabel?: string;
  onAction?: () => void;
  actionDisabled?: boolean;
};

export function SettingsDetailBlock({
  title,
  description,
  rows,
  actionLabel = "Edit",
  onAction,
  actionDisabled,
}: SettingsDetailBlockProps) {
  return (
    <div
      className="border-t py-6 first:border-t-0 first:pt-0"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="min-w-0 flex-1 max-w-3xl">
          <h3 className="text-base font-semibold tracking-tight">{title}</h3>
          {description ? (
            <p
              className="mt-1.5 text-sm leading-relaxed"
              style={{ color: "var(--muted-text)" }}
            >
              {description}
            </p>
          ) : null}
        </div>
        {onAction ? (
          <Button
            type="button"
            variant="outline"
            disabled={actionDisabled}
            onClick={onAction}
            className="shrink-0 font-mono text-[11px] uppercase tracking-widest"
          >
            {actionLabel}
          </Button>
        ) : null}
      </div>
      <SettingsMetaGrid>
        {rows.map((row) => (
          <SettingsMetaField
            key={row.label}
            label={row.label}
            value={row.value}
            span={row.span}
          />
        ))}
      </SettingsMetaGrid>
    </div>
  );
}
