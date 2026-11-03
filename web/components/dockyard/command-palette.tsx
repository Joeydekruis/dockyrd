"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { palettePages } from "@/lib/nav";

type CommandPaletteProps = {
  open: boolean;
  onClose: () => void;
};

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState("");

  useEffect(() => {
    if (!open) {
      setQuery("");
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return palettePages;
    return palettePages.filter((p) => p.label.toLowerCase().includes(q));
  }, [query]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-black/20 px-4 pt-[12vh] backdrop-blur-sm">
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Search"
        className="w-full max-w-xl overflow-hidden rounded-[20px] border shadow-lg"
        style={{ background: "var(--surface)", borderColor: "var(--border)" }}
      >
        <input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search pages…"
          className="w-full border-b px-4 py-3 text-sm outline-none"
          style={{ borderColor: "var(--border)" }}
        />
        <ul className="max-h-80 overflow-auto p-2">
          {filtered.map((page) => (
            <li key={page.href}>
              <Link
                href={page.href}
                onClick={onClose}
                className="block rounded-lg px-3 py-2 text-sm hover:bg-[var(--hover-row)]"
              >
                {page.label}
              </Link>
            </li>
          ))}
          {filtered.length === 0 ? (
            <li className="px-3 py-6 text-center text-sm" style={{ color: "var(--muted-text)" }}>
              No matches (M0 skeleton)
            </li>
          ) : null}
        </ul>
      </div>
    </div>
  );
}

export function useCommandPalette() {
  const [open, setOpen] = useState(false);
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);
  return { open, setOpen, close: () => setOpen(false) };
}
