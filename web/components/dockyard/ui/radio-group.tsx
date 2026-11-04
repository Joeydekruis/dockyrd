type RadioOption = {
  value: string;
  label: string;
  description?: string;
};

type RadioGroupProps = {
  name: string;
  legend: string;
  options: RadioOption[];
  value: string;
  onChange: (value: string) => void;
};

export function RadioGroup({
  name,
  legend,
  options,
  value,
  onChange,
}: RadioGroupProps) {
  return (
    <fieldset className="border-0 p-0">
      <legend
        className="mb-2 block font-mono text-[11px] uppercase tracking-widest"
        style={{ color: "var(--muted-text)" }}
      >
        {legend}
      </legend>
      <div className="flex flex-col gap-2">
        {options.map((opt) => {
          const id = `${name}-${opt.value}`;
          return (
            <label
              key={opt.value}
              htmlFor={id}
              className="flex min-h-[44px] cursor-pointer items-start gap-3 rounded-xl border px-4 py-3"
              style={{
                borderColor:
                  value === opt.value ? "var(--accent-blue)" : "var(--border)",
                background:
                  value === opt.value ? "var(--brand-soft)" : "var(--surface)",
              }}
            >
              <input
                id={id}
                type="radio"
                name={name}
                value={opt.value}
                checked={value === opt.value}
                onChange={() => onChange(opt.value)}
                className="mt-1 size-4 shrink-0"
              />
              <span>
                <span className="block text-sm font-medium">{opt.label}</span>
                {opt.description ? (
                  <span
                    className="mt-0.5 block text-[13px]"
                    style={{ color: "var(--muted-text)" }}
                  >
                    {opt.description}
                  </span>
                ) : null}
              </span>
            </label>
          );
        })}
      </div>
    </fieldset>
  );
}
