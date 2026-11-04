"use client";

const ranges = ["1h", "24h", "7d", "30d"] as const;

export type TimeRange = (typeof ranges)[number];

type TimeRangeToggleProps = {
  value: TimeRange;
  onChange: (value: TimeRange) => void;
};

export function TimeRangeToggle({ value, onChange }: TimeRangeToggleProps) {
  return (
    <fieldset
      className="inline-flex rounded-full border p-0.5 m-0"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
      aria-label="Time range"
    >
      {ranges.map((range) => {
        const active = range === value;
        return (
          <button
            key={range}
            type="button"
            className="rounded-full px-3 py-1.5 font-mono text-[11px] uppercase tracking-widest transition-colors"
            style={{
              background: active ? "var(--ink-navy)" : "transparent",
              color: active ? "#fff" : "var(--muted-text)",
            }}
            aria-pressed={active}
            onClick={() => onChange(range)}
          >
            {range}
          </button>
        );
      })}
    </fieldset>
  );
}
