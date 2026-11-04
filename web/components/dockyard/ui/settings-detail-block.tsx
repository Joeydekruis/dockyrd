import type { ReactNode } from "react";
import { Button } from "@/components/dockyard/ui/button";
import { KeyValueRow } from "@/components/dockyard/ui/key-value-row";

export type SettingsDetailRow = {
  label: string;
  value: ReactNode;
};

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
      className="flex flex-wrap items-start justify-between gap-4 border-t py-5 first:border-t-0 first:pt-0"
      style={{ borderColor: "var(--border)" }}
    >
      <div className="min-w-0 flex-1">
        <h3 className="text-sm font-semibold tracking-tight">{title}</h3>
        {description ? (
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            {description}
          </p>
        ) : null}
        <dl className="mt-4">
          {rows.map((row) => (
            <KeyValueRow key={row.label} label={row.label} value={row.value} />
          ))}
        </dl>
      </div>
      {onAction ? (
        <Button
          type="button"
          variant="outline"
          disabled={actionDisabled}
          onClick={onAction}
        >
          {actionLabel}
        </Button>
      ) : null}
    </div>
  );
}
