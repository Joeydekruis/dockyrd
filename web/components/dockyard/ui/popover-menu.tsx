"use client";

import { MoreHorizontal } from "lucide-react";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { IconButton } from "@/components/dockyard/ui/icon-button";

export type PopoverMenuItem = {
  id: string;
  label: string;
  onSelect?: () => void;
  destructive?: boolean;
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

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, [open]);

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
      {open ? (
        <div
          role="menu"
          className="absolute right-0 top-full z-30 mt-1 min-w-[160px] rounded-xl border py-1 shadow-lg"
          style={{ background: "var(--surface)", borderColor: "var(--border)" }}
        >
          {items.map((item) => (
            <button
              key={item.id}
              type="button"
              role="menuitem"
              className="block w-full px-3 py-2 text-left text-sm hover:bg-[var(--hover-row)]"
              style={{
                color: item.destructive ? "var(--danger)" : "var(--ink-navy)",
              }}
              onClick={() => {
                item.onSelect?.();
                setOpen(false);
              }}
            >
              {item.label}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
