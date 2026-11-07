import { ApplicationDetailPage } from "@/components/dockyrd/application-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ApplicationDetailRoute() {
  return <ApplicationDetailPage />;
}
