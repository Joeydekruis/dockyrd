import { DatabaseDetailPage } from "@/components/dockyrd/database-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function DatabaseDetailRoute() {
  return <DatabaseDetailPage />;
}
