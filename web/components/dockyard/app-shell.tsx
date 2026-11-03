import type { ReactNode } from "react";

type AppShellProps = {
  children: ReactNode;
};

export function AppShell({ children }: AppShellProps) {
  return (
    <div className="flex min-h-dvh">
      <aside
        className="flex w-64 flex-col border-r px-4 py-6"
        style={{
          background: "var(--surface)",
          borderColor: "var(--border)",
        }}
      >
        <a
          href="/"
          className="mb-8 text-lg font-semibold tracking-tight"
          aria-label="Dockyrd home"
        >
          Dockyrd
        </a>
        <nav className="text-sm" style={{ color: "var(--muted-text)" }}>
          Dashboard (M0 shell)
        </nav>
      </aside>
      <main className="flex flex-1 flex-col p-8">{children}</main>
    </div>
  );
}
