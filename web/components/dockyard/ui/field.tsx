import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";

const labelClass = "mb-1 block font-mono text-[11px] uppercase tracking-widest";

type FieldProps = {
  label: string;
  error?: string;
  hint?: string;
  children?: ReactNode;
} & InputHTMLAttributes<HTMLInputElement>;

export const Field = forwardRef<HTMLInputElement, FieldProps>(function Field(
  { label, error, hint, children, className, id, ...inputProps },
  ref,
) {
  const inputId = id ?? inputProps.name?.toString();
  const hintId = hint ? `${inputId}-hint` : undefined;
  const errorId = error ? `${inputId}-error` : undefined;
  return (
    <div className="block text-sm">
      <label htmlFor={children ? undefined : inputId}>
        <span className={labelClass} style={{ color: "var(--muted-text)" }}>
          {label}
        </span>
        {children ?? (
          <input
            ref={ref}
            id={inputId}
            aria-invalid={error ? true : undefined}
            aria-describedby={
              [hintId, errorId].filter(Boolean).join(" ") || undefined
            }
            className={className ? `field-input ${className}` : "field-input"}
            {...inputProps}
          />
        )}
      </label>
      {hint && !error ? (
        <p
          id={hintId}
          className="mt-1 text-[13px]"
          style={{ color: "var(--muted-text)" }}
        >
          {hint}
        </p>
      ) : null}
      {error ? (
        <p
          id={errorId}
          className="mt-1 text-[13px]"
          style={{ color: "var(--danger)" }}
        >
          {error}
        </p>
      ) : null}
    </div>
  );
});
