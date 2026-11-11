import type { ActiveDeployment } from "@/lib/api";

/** Dashboard path for a deployment detail screen (static export). */
export function deploymentDetailHref(d: {
  id: string;
  service_slug: string;
  service_kind: string;
}): string {
  const base =
    d.service_kind === "static_site"
      ? `/static-sites/${d.service_slug}`
      : `/applications/${d.service_slug}`;
  return `${base}/deployments/${d.id}`;
}

export function activeDeploymentHref(d: ActiveDeployment): string {
  return deploymentDetailHref(d);
}
