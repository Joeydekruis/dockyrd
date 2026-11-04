type SelectButtonOption = {
  value: string;
  label: string;
};

type SelectButtonProps = {
  options: SelectButtonOption[];
  value: string | string[];
  onChange: (value: string | string[]) => void;
  multi?: boolean;
  ariaLabel: string;
};

export function SelectButton({
  options,
  value,
  onChange,
  multi = false,
  ariaLabel,
}: SelectButtonProps) {
  const selected = Array.isArray(value) ? value : [value];
  const toggle = (v: string) => {
    if (!multi) {
      onChange(v);
      return;
    }
    const next = selected.includes(v)
      ? selected.filter((x) => x !== v)
      : [...selected, v];
    onChange(next);
  };
  return (
    <fieldset className="inline-flex flex-wrap gap-2 border-0 p-0">
      <legend className="sr-only">{ariaLabel}</legend>
      {options.map((opt) => {
        const active = selected.includes(opt.value);
        return (
          <button
            key={opt.value}
            type="button"
            aria-pressed={active}
            onClick={() => toggle(opt.value)}
            className="rounded-full border px-3 py-1.5 font-mono text-[11px] uppercase tracking-widest transition-colors"
            style={{
              borderColor: active
                ? "var(--ink-navy)"
                : "var(--secondary-button-border)",
              background: active ? "var(--ink-navy)" : "transparent",
              color: active ? "#fff" : "var(--secondary-text)",
            }}
          >
            {opt.label}
          </button>
        );
      })}
    </fieldset>
  );
}
