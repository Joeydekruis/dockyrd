"use client";

import { ChevronDown, type LucideIcon } from "lucide-react";
import {
  type ComponentPropsWithoutRef,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import { createPortal } from "react-dom";
import {
  isSelectMenuOutsideClick,
  useSelectMenuPosition,
} from "@/components/dockyard/ui/use-select-menu-position";

/** `secondary` = inverse (white) — use only on gradient/dark backgrounds; prefer `outline` on cards and page chrome. */
export type SplitButtonVariant = "primary" | "outline" | "secondary";

export type SplitButtonMenuItem = {
  id: string;
  label: string;
  icon?: LucideIcon;
  onSelect?: () => void;
  destructive?: boolean;
};

const variantClass: Record<SplitButtonVariant, string> = {
  primary: "btn-primary",
  outline: "btn-outline",
  secondary: "btn-inverse",
};

type SplitButtonProps = {
  variant?: SplitButtonVariant;
  iconLeft?: LucideIcon;
  menuLabel?: string;
  menuItems?: SplitButtonMenuItem[];
  onMenuClick?: () => void;
  /** Storybook / design board (open menu) */
  defaultMenuOpen?: boolean;
  menuWidth?: number;
  /** When true, clicking the main segment toggles the menu (same as the chevron). */
  primaryOpensMenu?: boolean;
  children: React.ReactNode;
} & Omit<ComponentPropsWithoutRef<"button">, "children">;

export function SplitButton({
  variant = "primary",
  iconLeft: IconLeft,
  menuLabel = "More options",
  menuItems,
  onMenuClick,
  defaultMenuOpen = false,
  menuWidth = 260,
  primaryOpensMenu = false,
  children,
  className = "",
  type = "button",
  disabled,
  onClick: primaryOnClick,
  ...primaryProps
}: SplitButtonProps) {
  const [menuOpen, setMenuOpen] = useState(defaultMenuOpen);
  const rootRef = useRef<HTMLDivElement>(null);
  const anchorRef = useRef<HTMLDivElement>(null);
  const menuId = useId();
  const btnClass = variantClass[variant];
  const groupClass =
    `split-button split-button-${variant === "secondary" ? "inverse" : variant}`.trim();
  const hasMenu = menuItems != null && menuItems.length > 0;
  const showMenu = hasMenu && (defaultMenuOpen || menuOpen);
  const usePortal = !defaultMenuOpen;
  const { menuRef, menuStyle } = useSelectMenuPosition(
    showMenu && usePortal,
    anchorRef,
    { width: menuWidth },
  );

  useEffect(() => {
    if (!menuOpen || defaultMenuOpen) return;
    const onDoc = (e: MouseEvent) => {
      if (isSelectMenuOutsideClick(e.target as Node, rootRef, menuRef)) {
        setMenuOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setMenuOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, [menuOpen, defaultMenuOpen, menuRef]);

  const onChevronClick = () => {
    if (disabled) return;
    if (hasMenu) {
      setMenuOpen((v) => !v);
      return;
    }
    onMenuClick?.();
  };

  return (
    <div
      ref={rootRef}
      className={`relative inline-flex flex-col items-start ${className}`.trim()}
    >
      <div ref={anchorRef} className={groupClass}>
        <button
          type={type}
          disabled={disabled}
          className={`${btnClass} split-button-main`}
          onClick={(e) => {
            primaryOnClick?.(e);
            if (primaryOpensMenu && hasMenu && !disabled) {
              setMenuOpen((v) => !v);
            }
          }}
          {...primaryProps}
        >
          {IconLeft ? (
            <IconLeft strokeWidth={2} className="size-4 shrink-0" aria-hidden />
          ) : null}
          <span>{children}</span>
        </button>
        <button
          type="button"
          disabled={disabled}
          className={`${btnClass} split-button-menu${showMenu ? " split-button-menu-open" : ""}`}
          aria-label={menuLabel}
          aria-haspopup={hasMenu ? "menu" : undefined}
          aria-expanded={hasMenu ? showMenu : undefined}
          aria-controls={hasMenu ? menuId : undefined}
          onClick={onChevronClick}
        >
          <ChevronDown
            strokeWidth={2}
            className={`size-4 transition-transform${showMenu ? " rotate-180" : ""}`}
            aria-hidden
          />
        </button>
      </div>
      {showMenu ? (
        usePortal && typeof document !== "undefined" ? (
          createPortal(
            <div
              ref={menuRef}
              id={menuId}
              role="menu"
              className="dockyrd-menu-panel"
              style={{ ...menuStyle, width: menuWidth }}
            >
              {menuItems?.map((item) => {
                const ItemIcon = item.icon;
                return (
                  <button
                    key={item.id}
                    type="button"
                    role="menuitem"
                    className="dockyrd-menu-item"
                    style={{
                      color: item.destructive
                        ? "var(--danger)"
                        : "var(--ink-navy)",
                    }}
                    onClick={() => {
                      item.onSelect?.();
                      setMenuOpen(false);
                    }}
                  >
                    {ItemIcon ? (
                      <ItemIcon
                        strokeWidth={1.8}
                        className="size-4 shrink-0"
                        aria-hidden
                      />
                    ) : null}
                    <span className="min-w-0 flex-1 text-left">
                      {item.label}
                    </span>
                  </button>
                );
              })}
            </div>,
            document.body,
          )
        ) : (
          <div
            id={menuId}
            role="menu"
            className="dockyrd-menu-panel"
            style={{ width: menuWidth }}
          >
            {menuItems?.map((item) => {
              const ItemIcon = item.icon;
              return (
                <button
                  key={item.id}
                  type="button"
                  role="menuitem"
                  className="dockyrd-menu-item"
                  style={{
                    color: item.destructive
                      ? "var(--danger)"
                      : "var(--ink-navy)",
                  }}
                  onClick={() => {
                    item.onSelect?.();
                    if (!defaultMenuOpen) setMenuOpen(false);
                  }}
                >
                  {ItemIcon ? (
                    <ItemIcon
                      strokeWidth={1.8}
                      className="size-4 shrink-0"
                      aria-hidden
                    />
                  ) : null}
                  <span className="min-w-0 flex-1 text-left">{item.label}</span>
                </button>
              );
            })}
          </div>
        )
      ) : null}
    </div>
  );
}
