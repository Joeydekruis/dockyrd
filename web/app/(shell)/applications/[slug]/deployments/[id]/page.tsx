import { ApplicationDeploymentDetailPage } from "@/components/dockyard/application-deployment-detail-page";

export function generateStaticParams() {
  return [{ slug: "_", id: "_" }];
}

export default function ApplicationDeploymentDetailRoute() {
  return <ApplicationDeploymentDetailPage />;
}
