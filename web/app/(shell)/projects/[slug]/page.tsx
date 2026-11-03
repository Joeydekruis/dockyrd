import { ProjectDetailPage } from "@/components/dockyard/project-detail-page";

export function generateStaticParams() {
  return [{ slug: "_" }];
}

export default function ProjectDetailRoute() {
  return <ProjectDetailPage />;
}
