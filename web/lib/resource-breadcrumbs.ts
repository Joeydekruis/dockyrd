import type { BreadcrumbItem } from "@/components/dockyard/ui/breadcrumbs";

type ProjectCrumb = { name: string; slug: string };

function projectResourceBreadcrumbs(
  list: { label: string; href: string },
  project: ProjectCrumb | null | undefined,
  resourceName: string,
): BreadcrumbItem[] {
  const items: BreadcrumbItem[] = [{ label: list.label, href: list.href }];
  if (project) {
    items.push({
      label: project.name,
      href: `/projects/${project.slug}`,
    });
  }
  items.push({ label: resourceName });
  return items;
}

export function applicationDetailBreadcrumbs(
  project: ProjectCrumb | null | undefined,
  resourceName: string,
): BreadcrumbItem[] {
  return projectResourceBreadcrumbs(
    { label: "All applications", href: "/applications" },
    project,
    resourceName,
  );
}

export function databaseDetailBreadcrumbs(
  project: ProjectCrumb | null | undefined,
  resourceName: string,
): BreadcrumbItem[] {
  return projectResourceBreadcrumbs(
    { label: "All databases", href: "/databases" },
    project,
    resourceName,
  );
}

export function staticSiteDetailBreadcrumbs(
  project: ProjectCrumb | null | undefined,
  resourceName: string,
): BreadcrumbItem[] {
  return projectResourceBreadcrumbs(
    { label: "All static sites", href: "/static-sites" },
    project,
    resourceName,
  );
}
