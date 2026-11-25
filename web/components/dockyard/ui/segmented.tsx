"use client";

type SegmentedOption = {
  value: string;
  label: string;
};

type SegmentedProps = {
  name?: string;
  legend?: string;
  value: string;
  options: SegmentedOption[];
  onChange: (value: string) => void;
  disabled?: boolean;
  "aria-label"?: string;
};

export function Segmented({
  name,
  legend,
  value,
  options,
  onChange,
  disabled,
  "aria-label": ariaLabel,
}: SegmentedProps) {
  return (
    <fieldset
      className="m-0 border-0 p-0"
      disabled={disabled}
      name={name}
    >
      {legend ? (
        <legend
          className="mb-2 block font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {legend}
        </legend>
      ) : null}
      <div
        className="inline-flex rounded-full border p-0.5"
        style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        role="radiogroup"
        aria-label={ariaLabel ?? legend}
      >
        {options.map((opt) => {
          const active = opt.value === value;
          return (
            <button
              key={opt.value}
              type="button"
              role="radio"
              aria-checked={active}
              disabled={disabled}
              className="dockyrd-segment-btn rounded-full px-3 py-1.5 text-sm"
              style={{
                background: active ? "var(--ink-navy)" : "transparent",
                color: active ? "#fff" : "var(--muted-text)",
                opacity: disabled ? 0.55 : 1,
              }}
              onClick={() => onChange(opt.value)}
            >
              {opt.label}
            </button>
          );
        })}
      </div>
    </fieldset>
  );
}
