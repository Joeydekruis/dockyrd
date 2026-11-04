"use client";

import Link from "next/link";
import { forwardRef, type Ref } from "react";

export type SearchResultRow = {
  kind: string;
  label: string;
  href: string;
};

type SearchPaletteProps = {
  query: string;
  onQueryChange: (value: string) => void;
  results: SearchResultRow[];
  onSelect?: (href: string) => void;
  emptyMessage?: string;
};

export const SearchPalette = forwardRef(function SearchPalette(
  {
    query,
    onQueryChange,
    results,
    onSelect,
    emptyMessage = "No matches",
  }: SearchPaletteProps,
  ref: Ref<HTMLInputElement>,
) {
  return (
    <div
      className="w-full max-w-xl rounded-[20px] border shadow-lg"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <input
        ref={ref}
        value={query}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="Search pages and projects…"
        className="search-palette-query field-input rounded-none rounded-t-[20px] border-0 border-b"
        style={{ borderColor: "var(--border)" }}
        aria-label="Search query"
      />
      <ul className="max-h-80 overflow-auto rounded-b-[20px] p-2">
        {results.map((row) => (
          <li key={`${row.kind}-${row.href}`}>
            <Link
              href={row.href}
              onClick={() => onSelect?.(row.href)}
              className="flex items-center justify-between rounded-lg px-3 py-2 text-sm hover:bg-[var(--hover-row)]"
            >
              <span>{row.label}</span>
              <span
                className="text-xs capitalize"
                style={{ color: "var(--muted-text)" }}
              >
                {row.kind}
              </span>
            </Link>
          </li>
        ))}
        {results.length === 0 ? (
          <li
            className="px-3 py-6 text-center text-sm"
            style={{ color: "var(--muted-text)" }}
          >
            {emptyMessage}
          </li>
        ) : null}
      </ul>
    </div>
  );
});
