import { forwardRef, type TextareaHTMLAttributes } from "react";

type TextareaFieldProps = {
  label: string;
  maxLength?: number;
  value?: string;
  hint?: string;
  error?: string;
} & TextareaHTMLAttributes<HTMLTextAreaElement>;

export const TextareaField = forwardRef<
  HTMLTextAreaElement,
  TextareaFieldProps
>(function TextareaField(
  { label, maxLength, value = "", hint, error, className, id, ...props },
  ref,
) {
  const inputId = id ?? props.name?.toString();
  const len = typeof value === "string" ? value.length : 0;
  return (
    <div className="block text-sm">
      <label htmlFor={inputId}>
        <span
          className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {label}
        </span>
        <textarea
          ref={ref}
          id={inputId}
          value={value}
          maxLength={maxLength}
          className={
            className
              ? `field-input field-textarea ${className}`
              : "field-input field-textarea"
          }
          {...props}
        />
      </label>
      <div className="mt-1 flex justify-between gap-2 text-[13px]">
        {error ? (
          <span style={{ color: "var(--danger)" }}>{error}</span>
        ) : hint ? (
          <span style={{ color: "var(--muted-text)" }}>{hint}</span>
        ) : (
          <span />
        )}
        {maxLength != null ? (
          <span
            className="font-mono text-xs tabular-nums"
            style={{ color: "var(--muted-text)" }}
          >
            {len}/{maxLength}
          </span>
        ) : null}
      </div>
    </div>
  );
});
