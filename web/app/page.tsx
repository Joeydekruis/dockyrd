import { AppShell } from "@/components/dockyard/app-shell";

export default function HomePage() {
  return (
    <AppShell>
      <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
      <p className="mt-2 max-w-lg" style={{ color: "var(--muted-text)" }}>
        M0 foundation — projects and services will appear here after Sprint 4.
      </p>
    </AppShell>
  );
}
