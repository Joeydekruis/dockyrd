import { ServiceDeploymentTable } from "@/components/dockyard/ui/service-deployment-table";
import type { Deployment } from "@/lib/api";

type ApplicationDeploymentTableProps = {
  deployments: Deployment[];
  appSlug: string;
  sourceRepo?: string;
  defaultBranch?: string;
  compact?: boolean;
};

export function ApplicationDeploymentTable({
  deployments,
  appSlug,
  sourceRepo,
  defaultBranch,
  compact = false,
}: ApplicationDeploymentTableProps) {
  return (
    <ServiceDeploymentTable
      deployments={deployments}
      resourceSlug={appSlug}
      resourceKind="applications"
      sourceRepo={sourceRepo}
      defaultBranch={defaultBranch}
      compact={compact}
    />
  );
}
