"use client";

import {
  ChevronDown,
  FileText,
  LogOut,
  type LucideIcon,
  SlidersHorizontal,
  User,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useId, useRef, useState } from "react";
import { Avatar } from "@/components/dockyard/ui/avatar";
import { logout, type MeResponse } from "@/lib/api";

export function initialsFromUser(name: string, email: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0][0] ?? ""}${parts[parts.length - 1][0] ?? ""}`.toUpperCase();
  }
  if (parts.length === 1 && parts[0].length >= 2) {
    return parts[0].slice(0, 2).toUpperCase();
  }
  return email.slice(0, 2).toUpperCase();
}

type MenuLink = {
  id: string;
  label: string;
  href: string;
  icon: LucideIcon;
  external?: boolean;
};

const menuLinks: MenuLink[] = [
  {
    id: "profile",
    label: "Profile & security",
    href: "/settings/profile",
    icon: User,
  },
  {
    id: "settings",
    label: "Settings",
    href: "/settings/instance",
    icon: SlidersHorizontal,
  },
  {
    id: "docs",
    label: "Documentation",
    href: "https://github.com/Joeydekruis/dockyrd",
    icon: FileText,
    external: true,
  },
];

type UserMenuProps = {
  user: MeResponse;
  /** Storybook */
  defaultOpen?: boolean;
};

export function UserMenu({ user, defaultOpen = false }: UserMenuProps) {
  const router = useRouter();
  const [open, setOpen] = useState(defaultOpen);
  const rootRef = useRef<HTMLDivElement>(null);
  const menuId = useId();
  const initials = initialsFromUser(user.name, user.email);

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
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
  }, [open]);

  async function onSignOut() {
    try {
      await logout();
    } finally {
      setOpen(false);
      router.replace("/sign-in");
    }
  }

  return (
    <div ref={rootRef} className="relative shrink-0">
      <button
        type="button"
        className="user-menu-trigger"
        aria-label={`Account menu, ${user.name}`}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={menuId}
        onClick={() => setOpen((v) => !v)}
      >
        <Avatar initials={initials} size="sm" />
        <ChevronDown
          strokeWidth={2}
          className={`size-4 shrink-0 transition-transform${open ? " rotate-180" : ""}`}
          aria-hidden
        />
      </button>
      {open ? (
        <div
          id={menuId}
          role="menu"
          className="dockyard-menu-panel user-menu-panel"
        >
          <div className="user-menu-header">
            <Avatar initials={initials} size="md" />
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">{user.name}</p>
              <p
                className="truncate text-[13px]"
                style={{ color: "var(--muted-text)" }}
              >
                {user.email}
              </p>
            </div>
          </div>
          <hr
            className="my-1 border-0 border-t"
            style={{ borderColor: "var(--border)" }}
          />
          {menuLinks.map((item) => {
            const Icon = item.icon;
            const className = "dockyard-menu-item";
            if (item.external) {
              return (
                <a
                  key={item.id}
                  role="menuitem"
                  href={item.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={className}
                  onClick={() => setOpen(false)}
                >
                  <Icon
                    strokeWidth={1.8}
                    className="size-4 shrink-0"
                    aria-hidden
                  />
                  <span className="min-w-0 flex-1 text-left">{item.label}</span>
                </a>
              );
            }
            return (
              <Link
                key={item.id}
                role="menuitem"
                href={item.href}
                className={className}
                onClick={() => setOpen(false)}
              >
                <Icon
                  strokeWidth={1.8}
                  className="size-4 shrink-0"
                  aria-hidden
                />
                <span className="min-w-0 flex-1 text-left">{item.label}</span>
              </Link>
            );
          })}
          <hr
            className="my-1 border-0 border-t"
            style={{ borderColor: "var(--border)" }}
          />
          <button
            type="button"
            role="menuitem"
            className="dockyard-menu-item dockyard-menu-item-danger"
            onClick={() => void onSignOut()}
          >
            <LogOut strokeWidth={1.8} className="size-4 shrink-0" aria-hidden />
            <span className="min-w-0 flex-1 text-left">Sign out</span>
          </button>
        </div>
      ) : null}
    </div>
  );
}
