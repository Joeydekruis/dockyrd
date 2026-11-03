import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { EmptyState } from "@/components/dockyard/empty-state";
import { DatabasesPage } from "@/components/dockyard/databases-page";
import { ObjectStoragePage } from "@/components/dockyard/object-storage-page";
import { ProjectsPage } from "@/components/dockyard/projects-page";
import { StaticSitesPage } from "@/components/dockyard/static-sites-page";

const pages = [
  { segment: "applications", title: "Applications", description: "Application list arrives in M1." },
  { segment: "projects", title: "Projects", description: "Project list arrives in M1." },
  { segment: "databases", title: "Databases", description: "Managed Postgres and Redis." },
  { segment: "static-sites", title: "Static sites", description: "Static site hosting." },
  { segment: "object-storage", title: "Object storage", description: "Garage S3 buckets." },
  { segment: "analytics", title: "Analytics", description: "Server analytics arrive in M4." },
] as const;

export function generateStaticParams() {
  return pages.map((p) => ({ segment: p.segment }));
}

export default async function SegmentPlaceholder({
  params,
}: {
  params: Promise<{ segment: string }>;
}) {
  const { segment } = await params;
  if (segment === "projects") {
    return <ProjectsPage />;
  }
  if (segment === "static-sites") {
    return <StaticSitesPage />;
  }
  if (segment === "databases") {
    return <DatabasesPage />;
  }
  if (segment === "object-storage") {
    return <ObjectStoragePage />;
  }
  const page = pages.find((p) => p.segment === segment) ?? pages[0];
  return (
    <DashboardShell title={page.title}>
      <EmptyState title={page.title} description={page.description} />
    </DashboardShell>
  );
}
