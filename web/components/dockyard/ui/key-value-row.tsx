import type { ReactNode } from "react";

type KeyValueRowProps = {
  label: string;
  value: ReactNode;
};

export function KeyValueRow({ label, value }: KeyValueRowProps) {
  return (
    <div
      className="flex flex-wrap items-baseline justify-between gap-2 border-t py-3 first:border-t-0"
      style={{ borderColor: "var(--border)" }}
    >
      <dt
        className="font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {label}
      </dt>
      <dd className="text-sm font-medium text-right">{value}</dd>
    </div>
  );
}
