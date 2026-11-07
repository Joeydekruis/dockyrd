"use client";

import { Button } from "@/components/dockyard/ui/button";
import {
  TERMINAL_THEMES,
  type TerminalTheme,
  type TerminalThemeId,
} from "@/lib/terminal-themes";

type Props = {
  open: boolean;
  themeId: TerminalThemeId;
  onPick: (id: TerminalThemeId) => void;
  onClose: () => void;
};

function ThemeCard({
  theme,
  selected,
  onPick,
}: {
  theme: TerminalTheme;
  selected: boolean;
  onPick: () => void;
}) {
  const p = theme.preview;
  return (
    <button
      type="button"
      aria-pressed={selected}
      onClick={onPick}
      className="flex flex-col overflow-hidden rounded-[14px] border-2 text-left"
      style={{
        borderColor: selected ? "var(--accent)" : "var(--border)",
        background: "var(--surface)",
        color: "var(--text)",
      }}
    >
      <span
        className="block border-b p-3.5 font-mono text-xs leading-relaxed"
        style={{ background: p.bg, borderColor: p.border }}
      >
        <span className="block truncate">
          <span style={{ color: p.prompt, fontWeight: 600 }}>
            root@host:~#{" "}
          </span>
          <span style={{ color: p.cmd }}>dockyard status</span>
        </span>
        <span className="block truncate" style={{ color: p.text }}>
          Server healthy
        </span>
        <span className="block truncate" style={{ color: p.ok }}>
          ✓ web-api is healthy
        </span>
        <span className="block truncate" style={{ color: p.warn }}>
          ! Build cache is large
        </span>
      </span>
      <span className="flex items-center gap-2 p-3">
        <span className="flex-1 font-semibold">{theme.name}</span>
        <span className="rounded-md bg-[var(--surface-2)] px-2 py-0.5 font-mono text-[11px] uppercase tracking-wide text-[var(--muted-foreground)]">
          {theme.mode}
        </span>
        {selected ? (
          <span
            className="inline-flex h-[22px] w-[22px] items-center justify-center rounded-full text-white"
            style={{ background: "var(--accent)" }}
            aria-hidden
          >
            ✓
          </span>
        ) : null}
      </span>
    </button>
  );
}

export function TerminalSettingsDialog({
  open,
  themeId,
  onPick,
  onClose,
}: Props) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-[60] flex items-start justify-center p-4 pt-24">
      <button
        type="button"
        aria-label="Close settings"
        className="absolute inset-0 bg-[rgba(11,31,58,0.45)]"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal
        aria-label="Terminal settings"
        className="relative w-full max-w-[640px] rounded-[20px] border border-[var(--border)] bg-[var(--surface)] p-6 shadow-lg"
      >
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-xl font-semibold tracking-tight">
            Terminal settings
          </h2>
          <button
            type="button"
            aria-label="Close"
            onClick={onClose}
            className="flex h-11 w-11 items-center justify-center rounded-[10px] text-[var(--muted-foreground)]"
          >
            ×
          </button>
        </div>
        <div className="mb-4 space-y-1">
          <p className="font-mono text-xs uppercase tracking-wide text-[var(--muted-foreground)]">
            Theme
          </p>
          <p className="text-sm text-[var(--muted-foreground)]">
            Colors for the terminal. Applies to every session.
          </p>
        </div>
        <div className="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-2">
          {TERMINAL_THEMES.map((t) => (
            <ThemeCard
              key={t.id}
              theme={t}
              selected={t.id === themeId}
              onPick={() => onPick(t.id)}
            />
          ))}
        </div>
        <div className="flex justify-end">
          <Button type="button" variant="primary" onClick={onClose}>
            Done
          </Button>
        </div>
      </div>
    </div>
  );
}
