import type { ReactNode } from "react";

type KeyValueRowProps = {
  label: string;
  value: ReactNode;
};

export function KeyValueRow({ label, value }: KeyValueRowProps) {
  return (
    <div
      className="flex items-start gap-3 border-t py-3 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <dt
        className="shrink-0 pt-0.5 font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </dt>
      <dd className="min-w-0 flex-1 text-right text-sm font-medium">{value}</dd>
    </div>
  );
}
