"use client";

import { Check, ChevronDown, Search } from "lucide-react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SelectOption } from "@/components/dockyard/ui/select";

export type SearchSelectOption = SelectOption;

type SearchSelectProps = {
  options: SearchSelectOption[];
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  id?: string;
  name?: string;
  "aria-label"?: string;
  "aria-labelledby"?: string;
  "aria-describedby"?: string;
  className?: string;
  invalid?: boolean;
  defaultOpen?: boolean;
};

function filterOptions(
  options: SearchSelectOption[],
  query: string,
): SearchSelectOption[] {
  const q = query.trim().toLowerCase();
  if (!q) return options;
  return options.filter(
    (o) =>
      o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q),
  );
}

export const SearchSelect = forwardRef<HTMLButtonElement, SearchSelectProps>(
  function SearchSelect(
    {
      options,
      value,
      defaultValue = "",
      onChange,
      placeholder = "Choose…",
      searchPlaceholder = "Search…",
      emptyMessage = "No matches.",
      disabled = false,
      id,
      name,
      "aria-label": ariaLabel,
      "aria-labelledby": ariaLabelledBy,
      "aria-describedby": ariaDescribedBy,
      className = "",
      invalid,
      defaultOpen = false,
    },
    ref,
  ) {
    const listId = useId();
    const searchId = `${listId}-search`;
    const rootRef = useRef<HTMLDivElement>(null);
    const searchRef = useRef<HTMLInputElement>(null);
    const [open, setOpen] = useState(defaultOpen);
    const [uncontrolled, setUncontrolled] = useState(defaultValue);
    const [activeIndex, setActiveIndex] = useState(0);
    const [searchQuery, setSearchQuery] = useState("");

    const selectedValue = value !== undefined ? value : uncontrolled;
    const selectedOption = options.find((o) => o.value === selectedValue);
    const displayLabel = selectedOption?.label ?? placeholder;
    const hasValue = Boolean(selectedOption);

    const filteredOptions = useMemo(
      () => filterOptions(options, searchQuery),
      [options, searchQuery],
    );

    const setValue = useCallback(
      (next: string) => {
        if (value === undefined) setUncontrolled(next);
        onChange?.(next);
        setOpen(false);
        setSearchQuery("");
      },
      [onChange, value],
    );

    useEffect(() => {
      if (!open) return;
      const onDoc = (e: MouseEvent) => {
        if (!rootRef.current?.contains(e.target as Node)) {
          setOpen(false);
          setSearchQuery("");
        }
      };
      document.addEventListener("mousedown", onDoc);
      return () => document.removeEventListener("mousedown", onDoc);
    }, [open]);

    useEffect(() => {
      if (!open) return;
      const idx = filteredOptions.findIndex((o) => o.value === selectedValue);
      setActiveIndex(Math.max(idx, 0));
      const t = window.setTimeout(() => searchRef.current?.focus(), 0);
      return () => window.clearTimeout(t);
    }, [open, filteredOptions, selectedValue]);

    useEffect(() => {
      setActiveIndex((i) =>
        filteredOptions.length === 0
          ? 0
          : Math.min(i, filteredOptions.length - 1),
      );
    }, [filteredOptions.length]);

    const onTriggerKeyDown = (e: React.KeyboardEvent) => {
      if (disabled) return;
      switch (e.key) {
        case "ArrowDown":
        case "ArrowUp":
          e.preventDefault();
          if (!open) {
            setOpen(true);
            return;
          }
          setActiveIndex((i) => {
            if (filteredOptions.length === 0) return 0;
            return e.key === "ArrowDown"
              ? Math.min(filteredOptions.length - 1, i + 1)
              : Math.max(0, i - 1);
          });
          break;
        case "Enter":
        case " ":
          e.preventDefault();
          if (open && filteredOptions[activeIndex]) {
            setValue(filteredOptions[activeIndex].value);
          } else if (!open) setOpen(true);
          break;
        case "Escape":
          e.preventDefault();
          setOpen(false);
          setSearchQuery("");
          break;
        default:
          break;
      }
    };

    const onSearchKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          if (filteredOptions.length > 0) {
            setActiveIndex((i) => Math.min(filteredOptions.length - 1, i + 1));
          }
          break;
        case "ArrowUp":
          e.preventDefault();
          if (filteredOptions.length > 0) {
            setActiveIndex((i) => Math.max(0, i - 1));
          }
          break;
        case "Enter":
          e.preventDefault();
          if (filteredOptions[activeIndex]) {
            setValue(filteredOptions[activeIndex].value);
          }
          break;
        case "Escape":
          e.preventDefault();
          setOpen(false);
          setSearchQuery("");
          break;
        default:
          break;
      }
    };

    return (
      <div ref={rootRef} className={`relative ${className}`.trim()}>
        {name ? (
          <input type="hidden" name={name} value={selectedValue} readOnly />
        ) : null}
        <button
          ref={ref}
          type="button"
          id={id}
          role="combobox"
          disabled={disabled}
          aria-label={ariaLabel}
          aria-labelledby={ariaLabelledBy}
          aria-describedby={ariaDescribedBy}
          aria-haspopup="listbox"
          aria-expanded={open}
          aria-controls={open ? listId : undefined}
          aria-activedescendant={
            open && filteredOptions[activeIndex]
              ? `${listId}-opt-${activeIndex}`
              : undefined
          }
          aria-invalid={invalid || undefined}
          className={`select-trigger field-input ${open ? "select-trigger-open" : ""}`.trim()}
          onClick={() => !disabled && setOpen((o) => !o)}
          onKeyDown={onTriggerKeyDown}
        >
          <span
            className="min-w-0 flex-1 truncate text-left"
            style={{
              color: hasValue ? "var(--ink-navy)" : "var(--muted-text)",
            }}
          >
            {displayLabel}
          </span>
          <ChevronDown
            strokeWidth={2}
            className={`size-4 shrink-0 transition-transform ${open ? "rotate-180" : ""}`}
            style={{ color: "var(--muted-text)" }}
            aria-hidden
          />
        </button>
        {open ? (
          <div id={listId} role="listbox" className="select-menu">
            <div className="select-menu-search-wrap">
              <Search
                strokeWidth={2}
                className="size-4 shrink-0"
                style={{ color: "var(--muted-text)" }}
                aria-hidden
              />
              <input
                ref={searchRef}
                id={searchId}
                type="search"
                aria-controls={listId}
                aria-label={searchPlaceholder}
                placeholder={searchPlaceholder}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={onSearchKeyDown}
                className="select-menu-search"
                autoComplete="off"
              />
            </div>
            {filteredOptions.length === 0 ? (
              <p
                className="select-menu-empty"
                style={{ color: "var(--muted-text)" }}
              >
                {emptyMessage}
              </p>
            ) : (
              filteredOptions.map((option, index) => {
                const isSelected = option.value === selectedValue;
                const isActive = index === activeIndex;
                return (
                  <div key={option.value} role="none">
                    <button
                      type="button"
                      id={`${listId}-opt-${index}`}
                      role="option"
                      aria-selected={isSelected}
                      tabIndex={-1}
                      className="select-option"
                      data-active={isActive ? "true" : undefined}
                      onMouseEnter={() => setActiveIndex(index)}
                      onClick={() => setValue(option.value)}
                    >
                      <span className="min-w-0 truncate">{option.label}</span>
                      {isSelected ? (
                        <Check
                          strokeWidth={2}
                          className="size-4 shrink-0"
                          style={{ color: "var(--ink-navy)" }}
                          aria-hidden
                        />
                      ) : (
                        <span className="size-4 shrink-0" aria-hidden />
                      )}
                    </button>
                  </div>
                );
              })
            )}
          </div>
        ) : null}
      </div>
    );
  },
);
