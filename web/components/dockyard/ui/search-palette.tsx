"use client";

import type { LucideIcon } from "lucide-react";
import { FileText, Layers, Search } from "lucide-react";
import Link from "next/link";
import { forwardRef, type Ref, useMemo } from "react";
import { Kbd } from "@/components/dockyard/ui/kbd";
import { palettePages } from "@/lib/nav";

export type SearchResultRow = {
  kind: string;
  label: string;
  href: string;
  meta?: string;
};

type SearchPaletteProps = {
  query: string;
  onQueryChange: (value: string) => void;
  results: SearchResultRow[];
  activeIndex?: number;
  onSelect?: (href: string) => void;
  onClose?: () => void;
  emptyMessage?: string;
};

const sectionForKind: Record<string, string> = {
  page: "Pages",
  project: "Resources",
};

const hrefIcon = new Map(palettePages.map((p) => [p.href, p.icon]));

function iconForRow(row: SearchResultRow): LucideIcon {
  const fromNav = hrefIcon.get(row.href);
  if (fromNav) return fromNav;
  if (row.kind === "project") return Layers;
  return FileText;
}

function groupResults(results: SearchResultRow[]) {
  const order = ["page", "project"];
  const buckets = new Map<string, { index: number; row: SearchResultRow }[]>();
  results.forEach((row, index) => {
    const key = order.includes(row.kind) ? row.kind : "other";
    const list = buckets.get(key) ?? [];
    list.push({ index, row });
    buckets.set(key, list);
  });
  const sections: {
    title: string;
    items: { index: number; row: SearchResultRow }[];
  }[] = [];
  for (const kind of [...order, "other"]) {
    const items = buckets.get(kind);
    if (!items?.length) continue;
    sections.push({
      title:
        sectionForKind[kind] ?? kind.charAt(0).toUpperCase() + kind.slice(1),
      items,
    });
  }
  return sections;
}

export const SearchPalette = forwardRef(function SearchPalette(
  {
    query,
    onQueryChange,
    results,
    activeIndex = 0,
    onSelect,
    onClose,
    emptyMessage,
  }: SearchPaletteProps,
  ref: Ref<HTMLInputElement>,
) {
  const sections = useMemo(() => groupResults(results), [results]);
  const showEmpty = results.length === 0;
  const emptyCopy =
    emptyMessage ??
    (query.trim()
      ? `No results for “${query.trim()}”. Try a project, service or page name.`
      : "No matches");

  return (
    <div className="search-palette">
      <div
        className="search-palette-query-row"
        style={{ borderColor: "var(--border)" }}
      >
        <Search
          strokeWidth={1.8}
          className="size-[18px] shrink-0"
          style={{ color: "var(--muted-text)" }}
          aria-hidden
        />
        <input
          ref={ref}
          value={query}
          onChange={(e) => onQueryChange(e.target.value)}
          placeholder="Search pages, resources and actions…"
          className="search-palette-input"
          aria-label="Search query"
        />
        <button
          type="button"
          className="shrink-0 border-0 bg-transparent p-0"
          onClick={onClose}
          aria-label="Close search"
        >
          <Kbd>Esc</Kbd>
        </button>
      </div>

      {showEmpty ? (
        <p
          className="search-palette-empty"
          style={{ color: "var(--muted-text)" }}
        >
          {emptyCopy}
        </p>
      ) : (
        <div className="search-palette-results">
          {sections.map((section) => (
            <div key={section.title}>
              <p className="search-palette-section-title">{section.title}</p>
              <ul className="m-0 list-none p-0">
                {section.items.map(({ index, row }) => {
                  const Icon = iconForRow(row);
                  const active = index === activeIndex;
                  return (
                    <li key={`${row.kind}-${row.href}-${index}`}>
                      <Link
                        href={row.href}
                        onClick={() => onSelect?.(row.href)}
                        className={`search-palette-row${active ? " search-palette-row-active" : ""}`}
                        aria-current={active ? "true" : undefined}
                        data-active={active ? "true" : undefined}
                      >
                        <span
                          className="shrink-0"
                          style={{ color: "var(--secondary-text)" }}
                        >
                          <Icon
                            strokeWidth={1.8}
                            className="size-[18px]"
                            aria-hidden
                          />
                        </span>
                        <span className="min-w-0 flex-1 truncate font-medium">
                          {row.label}
                        </span>
                        {row.meta ? (
                          <span
                            className="hidden shrink-0 truncate text-[13px] sm:inline"
                            style={{ color: "var(--muted-text)" }}
                          >
                            {row.meta}
                          </span>
                        ) : null}
                        <span
                          className={`shrink-0 text-[13px]${active ? "" : " invisible"}`}
                          style={{ color: "var(--muted-text)" }}
                          aria-hidden
                        >
                          ↵
                        </span>
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </div>
      )}

      <div
        className="search-palette-footer"
        style={{ borderColor: "var(--border)", color: "var(--muted-text)" }}
      >
        <span className="inline-flex items-center gap-1.5">
          <Kbd>↑↓</Kbd>
          Navigate
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Kbd>↵</Kbd>
          Open
        </span>
        <span className="inline-flex items-center gap-1.5">
          <Kbd>Esc</Kbd>
          Close
        </span>
      </div>
    </div>
  );
});
