"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";
import { SearchPalette } from "@/components/dockyard/ui/search-palette";
import { globalSearch, type SearchResult } from "@/lib/api";
import { palettePages } from "@/lib/nav";

type CommandPaletteProps = {
  open: boolean;
  onClose: () => void;
};

export function CommandPalette({ open, onClose }: CommandPaletteProps) {
  const router = useRouter();
  const [query, setQuery] = useState("");
  const [remote, setRemote] = useState<SearchResult[] | null>(null);
  const [activeIndex, setActiveIndex] = useState(0);
  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!open) {
      setQuery("");
      setRemote(null);
      setActiveIndex(0);
      return;
    }
    searchInputRef.current?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
      }
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
    const base = remote
      ? remote.flatMap((r) => {
          if (r.kind === "service" && r.href.startsWith("/applications/")) {
            return [
              r,
              {
                kind: "action",
                label: `Show connections · ${r.label}`,
                href: `${r.href}?tab=overview`,
              },
            ];
          }
          return [r];
        })
      : null;
    if (base) return base;
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

  function openResult(href: string) {
    onClose();
    router.push(href);
  }

  function onPaletteKeyDown(e: React.KeyboardEvent) {
    if (filtered.length === 0) return;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActiveIndex((i) => Math.min(i + 1, filtered.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActiveIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === "Enter") {
      e.preventDefault();
      const target = filtered[activeIndex];
      if (target) openResult(target.href);
    }
  }

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
        className="relative z-10 w-full max-w-[740px]"
        onKeyDown={onPaletteKeyDown}
      >
        <SearchPalette
          ref={searchInputRef}
          query={query}
          onQueryChange={(value) => {
            setQuery(value);
            setActiveIndex(0);
          }}
          results={filtered}
          activeIndex={activeIndex}
          onSelect={openResult}
          onClose={onClose}
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
