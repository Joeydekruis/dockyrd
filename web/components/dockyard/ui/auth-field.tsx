import { forwardRef, type InputHTMLAttributes, type ReactNode } from "react";

type AuthFieldProps = {
  label: string;
  children?: ReactNode;
} & InputHTMLAttributes<HTMLInputElement>;

export const AuthField = forwardRef<HTMLInputElement, AuthFieldProps>(
  function AuthField({ label, children, className, id, ...inputProps }, ref) {
    const inputId = id ?? inputProps.name?.toString();
    return (
      <label className="block text-sm" htmlFor={children ? undefined : inputId}>
        <span
          className="mb-1 block font-mono text-[11px] uppercase tracking-widest"
          style={{ color: "var(--muted-text)" }}
        >
          {label}
        </span>
        {children ?? (
          <input
            ref={ref}
            id={inputId}
            className={className ? `field-input ${className}` : "field-input"}
            {...inputProps}
          />
        )}
      </label>
    );
  },
);
