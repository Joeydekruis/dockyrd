import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { EmptyState } from "@/components/dockyard/empty-state";

export default function HomePage() {
  return (
    <DashboardShell title="Dashboard" breadcrumb="Home">
      <div className="mb-8">
        <h1 className="text-4xl font-semibold tracking-tight">Dashboard</h1>
        <p className="mt-2" style={{ color: "var(--muted-text)" }}>
          Your instance is ready. Create resources to get started.
        </p>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <EmptyState
          title="No applications yet"
          description="Deploy from GitHub or a Dockerfile when M1 lands."
          actionLabel="Create application"
          actionHref="/applications"
        />
        <EmptyState
          title="No databases yet"
          description="Provision Postgres or Valkey from the dashboard soon."
          actionLabel="Create database"
          actionHref="/databases"
        />
      </div>
    </DashboardShell>
  );
}
