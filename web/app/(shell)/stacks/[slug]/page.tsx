import { ComposeDetailPage } from "@/components/dockyrd/compose-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ComposeDetailRoute() {
  return <ComposeDetailPage />;
}
