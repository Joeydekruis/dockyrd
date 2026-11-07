import { ApplicationDetailPage } from "@/components/dockyard/application-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ApplicationDetailRoute() {
  return <ApplicationDetailPage />;
}
