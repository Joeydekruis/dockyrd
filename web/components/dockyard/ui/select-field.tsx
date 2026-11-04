"use client";

import { forwardRef, useId } from "react";
import { Select, type SelectOption } from "@/components/dockyard/ui/select";

const labelClass = "mb-1 block font-mono text-[11px] uppercase tracking-widest";

type SelectFieldProps = {
  label: string;
  options: SelectOption[];
  hint?: string;
  error?: string;
  placeholder?: string;
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  id?: string;
  name?: string;
  className?: string;
  defaultOpen?: boolean;
};

export const SelectField = forwardRef<HTMLButtonElement, SelectFieldProps>(
  function SelectField(
    {
      label,
      options,
      hint,
      error,
      placeholder,
      value,
      defaultValue,
      onChange,
      disabled,
      id,
      name,
      className = "",
      defaultOpen,
    },
    ref,
  ) {
    const autoId = useId();
    const fieldId = id ?? name ?? autoId;
    const hintId = hint ? `${fieldId}-hint` : undefined;
    const errorId = error ? `${fieldId}-error` : undefined;
    const labelId = `${fieldId}-label`;

    return (
      <div className={`block text-sm ${className}`.trim()}>
        <span
          id={labelId}
          className={labelClass}
          style={{ color: "var(--muted-text)" }}
        >
          {label}
        </span>
        <Select
          ref={ref}
          id={fieldId}
          name={name}
          options={options}
          value={value}
          defaultValue={defaultValue}
          onChange={onChange}
          placeholder={placeholder}
          disabled={disabled}
          invalid={Boolean(error)}
          aria-labelledby={labelId}
          aria-describedby={
            [hintId, errorId].filter(Boolean).join(" ") || undefined
          }
          defaultOpen={defaultOpen}
        />
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
  },
);
