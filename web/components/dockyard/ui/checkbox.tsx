import type { InputHTMLAttributes } from "react";

type CheckboxProps = {
  label: string;
} & Omit<InputHTMLAttributes<HTMLInputElement>, "type">;

export function Checkbox({
  label,
  id,
  className = "",
  ...props
}: CheckboxProps) {
  const inputId = id ?? props.name?.toString();
  return (
    <label
      htmlFor={inputId}
      className="inline-flex min-h-[44px] cursor-pointer items-center gap-3 text-sm"
    >
      <input
        id={inputId}
        type="checkbox"
        className={`size-4 shrink-0 rounded border accent-[var(--primary-button)] ${className}`}
        style={{ borderColor: "var(--input-border)" }}
        {...props}
      />
      <span>{label}</span>
    </label>
  );
}
