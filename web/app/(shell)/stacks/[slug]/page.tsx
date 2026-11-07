import { ComposeDetailPage } from "@/components/dockyard/compose-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ComposeDetailRoute() {
  return <ComposeDetailPage />;
}
