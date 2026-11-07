type SwitchProps = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  id?: string;
  disabled?: boolean;
};

export function Switch({
  checked,
  onChange,
  label,
  id,
  disabled,
}: SwitchProps) {
  const switchId =
    id ?? (label ? `switch-${label.replace(/\s+/g, "-")}` : undefined);
  return (
    <div
      className="inline-flex min-h-[44px] items-center gap-3"
      style={{ opacity: disabled ? 0.5 : 1 }}
    >
      <button
        id={switchId}
        type="button"
        role="switch"
        aria-checked={checked}
        aria-label={label}
        disabled={disabled}
        onClick={() => onChange(!checked)}
        className="dockyrd-switch-track relative inline-flex h-6 w-11 shrink-0 rounded-full border-0 p-0"
        style={{
          background: checked ? "var(--primary-button)" : "var(--input-border)",
        }}
      >
        <span
          className="dockyrd-switch-thumb absolute top-0.5 left-0.5 size-5 rounded-full bg-white shadow-sm"
          style={{ transform: checked ? "translateX(20px)" : "translateX(0)" }}
          aria-hidden
        />
      </button>
      {label ? (
        <label
          htmlFor={switchId}
          className="cursor-pointer text-sm font-medium"
        >
          {label}
        </label>
      ) : null}
    </div>
  );
}
