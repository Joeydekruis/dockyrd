"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { SearchPalette } from "@/components/dockyard/ui/search-palette";
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
    <div className="fixed inset-0 z-50 flex items-start justify-center px-4 pt-[12vh]">
      <button
        type="button"
        className="absolute inset-0 border-0 bg-black/20 backdrop-blur-sm"
        aria-label="Close search"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Search"
        className="relative z-10 w-full max-w-xl"
      >
        <SearchPalette
          ref={searchInputRef}
          query={query}
          onQueryChange={setQuery}
          results={filtered}
          onSelect={() => onClose()}
        />
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
