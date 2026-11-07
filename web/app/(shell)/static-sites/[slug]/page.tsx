import { StaticSiteDetailPage } from "@/components/dockyrd/static-site-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function StaticSiteDetailRoute() {
  return <StaticSiteDetailPage />;
}
