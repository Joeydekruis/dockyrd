import { ObjectStorageDetailPage } from "@/components/dockyrd/object-storage-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ObjectStorageDetailRoute() {
  return <ObjectStorageDetailPage />;
}
