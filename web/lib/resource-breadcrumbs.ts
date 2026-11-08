import type { BreadcrumbItem } from "@/components/dockyard/ui/breadcrumbs";

type ProjectCrumb = { name: string; slug: string };

export function staticSiteDetailBreadcrumbs(
  project: ProjectCrumb | null | undefined,
  resourceName: string,
): BreadcrumbItem[] {
  const items: BreadcrumbItem[] = [
    { label: "All static sites", href: "/static-sites" },
  ];
  if (project) {
    items.push({
      label: project.name,
      href: `/projects/${project.slug}`,
    });
  }
  items.push({ label: resourceName });
  return items;
}
