"use client";

import { Check, ChevronDown } from "lucide-react";
import {
  forwardRef,
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import {
  isSelectMenuOutsideClick,
  useSelectMenuPosition,
} from "@/components/dockyrd/ui/use-select-menu-position";

export type SelectOption = {
  value: string;
  label: string;
};

type SelectProps = {
  options: SelectOption[];
  value?: string;
  defaultValue?: string;
  onChange?: (value: string) => void;
  placeholder?: string;
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

export const Select = forwardRef<HTMLButtonElement, SelectProps>(
  function Select(
    {
      options,
      value,
      defaultValue = "",
      onChange,
      placeholder = "Choose…",
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
    const rootRef = useRef<HTMLDivElement>(null);
    const [open, setOpen] = useState(defaultOpen);
    const { menuRef, menuStyle } = useSelectMenuPosition(open, rootRef);
    const [uncontrolled, setUncontrolled] = useState(defaultValue);
    const [activeIndex, setActiveIndex] = useState(0);

    const selectedValue = value !== undefined ? value : uncontrolled;
    const selectedOption = options.find((o) => o.value === selectedValue);
    const displayLabel = selectedOption?.label ?? placeholder;
    const hasValue = Boolean(selectedOption);

    const setValue = useCallback(
      (next: string) => {
        if (value === undefined) setUncontrolled(next);
        onChange?.(next);
        setOpen(false);
      },
      [onChange, value],
    );

    useEffect(() => {
      if (!open) return;
      const onDoc = (e: MouseEvent) => {
        if (isSelectMenuOutsideClick(e.target as Node, rootRef, menuRef)) {
          setOpen(false);
        }
      };
      document.addEventListener("mousedown", onDoc);
      return () => document.removeEventListener("mousedown", onDoc);
    }, [open, menuRef]);

    useEffect(() => {
      if (!open) return;
      const idx = options.findIndex((o) => o.value === selectedValue);
      setActiveIndex(Math.max(idx, 0));
    }, [open, options, selectedValue]);

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
            const next =
              e.key === "ArrowDown"
                ? Math.min(options.length - 1, i + 1)
                : Math.max(0, i - 1);
            return next;
          });
          break;
        case "Enter":
        case " ":
          e.preventDefault();
          if (open) setValue(options[activeIndex]?.value ?? "");
          else setOpen(true);
          break;
        case "Escape":
          e.preventDefault();
          setOpen(false);
          break;
        case "Home":
          if (open) {
            e.preventDefault();
            setActiveIndex(0);
          }
          break;
        case "End":
          if (open) {
            e.preventDefault();
            setActiveIndex(options.length - 1);
          }
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
            open && options[activeIndex]
              ? `${listId}-opt-${activeIndex}`
              : undefined
          }
          aria-invalid={invalid || undefined}
          className={`select-trigger field-input ${open ? "select-trigger-open" : ""}`.trim()}
          onClick={() => !disabled && setOpen((o) => !o)}
          onKeyDown={onTriggerKeyDown}
        >
          <span
            className="min-w-0 flex-1 truncate"
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
        {open && typeof document !== "undefined"
          ? createPortal(
              <div
                ref={menuRef}
                id={listId}
                role="listbox"
                className="select-menu"
                style={menuStyle}
              >
                {options.map((option, index) => {
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
                        <span>{option.label}</span>
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
                })}
              </div>,
              document.body,
            )
          : null}
      </div>
    );
  },
);
