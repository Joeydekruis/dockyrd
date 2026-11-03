import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { EmptyState } from "@/components/dockyard/empty-state";

export default function SettingsProfilePage() {
  return (
    <DashboardShell title="Profile & security" breadcrumb="Settings">
      <p className="mb-4 text-sm">
        <a
          href="/settings/team"
          className="text-[var(--muted-foreground)] underline"
        >
          Team
        </a>
        {" · "}
        <a
          href="/settings/instance"
          className="text-[var(--muted-foreground)] underline"
        >
          Instance settings (GitHub App)
        </a>
      </p>
      <EmptyState
        title="Profile & security"
        description="Account settings and 2FA arrive in M4."
      />
    </DashboardShell>
  );
}
