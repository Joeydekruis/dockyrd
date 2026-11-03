import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { EmptyState } from "@/components/dockyard/empty-state";

export default function SettingsProfilePage() {
  return (
    <DashboardShell title="Profile & security" breadcrumb="Settings">
      <EmptyState
        title="Profile & security"
        description="Account settings and 2FA arrive in M4."
      />
    </DashboardShell>
  );
}
