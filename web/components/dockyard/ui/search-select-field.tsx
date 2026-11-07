"use client";

import { forwardRef, useId } from "react";
import {
  SearchSelect,
  type SearchSelectOption,
} from "@/components/dockyrd/ui/search-select";

const labelClass = "mb-1 block font-mono text-[11px] uppercase tracking-widest";

type SearchSelectFieldProps = {
  label: string;
  options: SearchSelectOption[];
  hint?: string;
  error?: string;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  id?: string;
  name?: string;
  className?: string;
  defaultOpen?: boolean;
};

export const SearchSelectField = forwardRef<
  HTMLButtonElement,
  SearchSelectFieldProps
>(function SearchSelectField(
  {
    label,
    options,
    hint,
    error,
    placeholder,
    searchPlaceholder,
    emptyMessage,
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
      <SearchSelect
        ref={ref}
        id={fieldId}
        name={name}
        options={options}
        value={value}
        defaultValue={defaultValue}
        onChange={onChange}
        placeholder={placeholder}
        searchPlaceholder={searchPlaceholder}
        emptyMessage={emptyMessage}
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
});
