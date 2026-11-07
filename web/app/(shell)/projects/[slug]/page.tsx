import { ProjectDetailPage } from "@/components/dockyrd/project-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ProjectDetailRoute() {
  return <ProjectDetailPage />;
}
