import { DatabaseDetailPage } from "@/components/dockyard/database-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function DatabaseDetailRoute() {
  return <DatabaseDetailPage />;
}
