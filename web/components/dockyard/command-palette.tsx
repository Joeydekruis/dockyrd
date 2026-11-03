"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef, useState } from "react";
import { globalSearch, type SearchResult } from "@/lib/api";
import { palettePages } from "@/lib/nav";

type CommandPaletteProps = {
  open: boolean;
  onClose: () => void;
};

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState("");
  const [remote, setRemote] = useState<SearchResult[] | null>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open) {
      setQuery("");
      setRemote(null);
      return;
    }
    searchInputRef.current?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [open, onClose]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    const q = query.trim();
    if (q === "") {
      setRemote(null);
      return;
    }
    const timer = window.setTimeout(() => {
      globalSearch(q)
        .then((res) => {
          if (!cancelled) setRemote(res.results);
        })
        .catch(() => {
          if (!cancelled) setRemote([]);
        });
    }, 200);
    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, [open, query]);

  const filtered = useMemo(() => {
    if (remote) return remote;
    const q = query.trim().toLowerCase();
    if (!q) {
      return palettePages.map((p) => ({
        kind: "page",
        label: p.label,
        href: p.href,
      }));
    }
    return palettePages
      .filter((p) => p.label.toLowerCase().includes(q))
      .map((p) => ({ kind: "page", label: p.label, href: p.href }));
  }, [query, remote]);

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
          ref={searchInputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search pages and projects…"
          className="w-full border-b px-4 py-3 text-sm outline-none"
          style={{ borderColor: "var(--border)" }}
        />
        <ul className="max-h-80 overflow-auto p-2">
          {filtered.map((page) => (
            <li key={`${page.kind}-${page.href}`}>
              <Link
                href={page.href}
                onClick={onClose}
                className="flex items-center justify-between rounded-lg px-3 py-2 text-sm hover:bg-[var(--hover-row)]"
              >
                <span>{page.label}</span>
                <span className="text-xs capitalize text-[var(--muted-foreground)]">
                  {page.kind}
                </span>
              </Link>
            </li>
          ))}
          {filtered.length === 0 ? (
            <li
              className="px-3 py-6 text-center text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              No matches
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
  return {
    open,
    setOpen,
    close: () => setOpen(false),
  };
}
