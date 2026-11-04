import { StaticSiteDeploymentDetailPage } from "@/components/dockyard/static-site-deployment-detail-page";

export function generateStaticParams() {
  return [{ slug: "_", id: "_" }];
}

export default function StaticSiteDeploymentDetailRoute() {
  return <StaticSiteDeploymentDetailPage />;
}
