"use client";

import { MoreHorizontal } from "lucide-react";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { IconButton } from "@/components/dockyard/ui/icon-button";
import {
  isSelectMenuOutsideClick,
  useSelectMenuPosition,
} from "@/components/dockyard/ui/use-select-menu-position";

export type PopoverMenuItem = {
  id: string;
  label: string;
  onSelect?: () => void;
  destructive?: boolean;
  disabled?: boolean;
  dividerBefore?: boolean;
};

const POPOVER_MENU_POSITION = {
  align: "end" as const,
  minWidth: 160,
};

type PopoverMenuProps = {
  items: PopoverMenuItem[];
  trigger?: ReactNode;
  label?: string;
};

export function PopoverMenu({
  items,
  trigger,
  label = "Open menu",
}: PopoverMenuProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const { menuRef, menuStyle } = useSelectMenuPosition(
    open,
    rootRef,
    POPOVER_MENU_POSITION,
  );

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (isSelectMenuOutsideClick(e.target as Node, rootRef, menuRef)) {
        setOpen(false);
      }
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onDoc);
      document.removeEventListener("keydown", onKey);
    };
  }, [open, menuRef]);

  const menu = open ? (
    <div
      ref={menuRef}
      role="menu"
      className="rounded-xl border py-1 shadow-lg"
      style={{
        ...menuStyle,
        background: "var(--surface)",
        borderColor: "var(--border)",
      }}
    >
      {items.map((item) => (
        <div key={item.id}>
          {item.dividerBefore ? (
            <hr className="my-1" style={{ borderColor: "var(--border)" }} />
          ) : null}
          <button
            type="button"
            role="menuitem"
            disabled={item.disabled}
            className="block w-full px-3 py-2 text-left text-sm hover:bg-[var(--hover-row)] disabled:cursor-not-allowed disabled:opacity-45 disabled:hover:bg-transparent"
            style={{
              color: item.destructive ? "var(--danger)" : "var(--ink-navy)",
            }}
            onClick={() => {
              if (item.disabled) return;
              item.onSelect?.();
              setOpen(false);
            }}
          >
            {item.label}
          </button>
        </div>
      ))}
    </div>
  ) : null;

  return (
    <div ref={rootRef} className="relative inline-flex">
      {trigger ?? (
        <IconButton
          icon={MoreHorizontal}
          label={label}
          aria-expanded={open}
          onClick={() => setOpen((v) => !v)}
        />
      )}
      {menu && typeof document !== "undefined"
        ? createPortal(menu, document.body)
        : null}
    </div>
  );
}
