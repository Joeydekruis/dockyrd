"use client";

type AppHeaderProps = {
  title: string;
  breadcrumb?: string;
  onOpenSearch: () => void;
};

export function AppHeader({ title, breadcrumb = "Dashboard", onOpenSearch }: AppHeaderProps) {
  return (
    <header className="flex items-center justify-between gap-4 border-b px-8 py-4" style={{ borderColor: "var(--border)" }}>
      <nav aria-label="Breadcrumb" className="text-sm" style={{ color: "var(--muted-text)" }}>
        <span>{breadcrumb}</span>
        <span className="mx-2" aria-hidden>
          /
        </span>
        <span style={{ color: "var(--ink-navy)" }}>{title}</span>
      </nav>
      <div className="flex items-center gap-3">
        <button
          type="button"
          onClick={onOpenSearch}
          className="flex min-h-11 items-center gap-2 rounded-[10px] border px-3 text-sm"
          style={{ borderColor: "var(--input-border)", color: "var(--muted-text)" }}
          aria-label="Search pages, resources and actions"
        >
          <span>Search…</span>
          <kbd className="font-mono text-[11px]">⌘K</kbd>
        </button>
        <a
          href="https://github.com/Joeydekruis/dockyrd"
          className="inline-flex min-h-11 items-center rounded-[10px] border px-4 font-mono text-xs uppercase tracking-wide"
          style={{ borderColor: "var(--secondary-button-border)", color: "var(--ink-navy)" }}
        >
          Docs
        </a>
      </div>
    </header>
  );
}
