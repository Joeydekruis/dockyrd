import { ServiceDeploymentTable } from "@/components/dockyard/ui/service-deployment-table";
import type { Deployment } from "@/lib/api";

type StaticSiteDeploymentTableProps = {
  deployments: Deployment[];
  siteSlug: string;
  sourceRepo?: string;
  defaultBranch?: string;
  compact?: boolean;
};

export function StaticSiteDeploymentTable({
  deployments,
  siteSlug,
  sourceRepo,
  defaultBranch,
  compact = false,
}: StaticSiteDeploymentTableProps) {
  return (
    <ServiceDeploymentTable
      deployments={deployments}
      resourceSlug={siteSlug}
      resourceKind="static-sites"
      sourceRepo={sourceRepo}
      defaultBranch={defaultBranch}
      compact={compact}
    />
  );
}
