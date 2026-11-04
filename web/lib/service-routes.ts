/** Dashboard path for a service detail view (static export `_` placeholder on server). */
export function serviceDetailHref(kind: string, slug: string): string {
  switch (kind) {
    case "static_site":
      return `/static-sites/${slug}`;
    case "app":
      return `/applications/${slug}`;
    case "postgres":
    case "redis":
      return `/databases/${slug}`;
    case "bucket":
      return `/object-storage/${slug}`;
    default:
      return `/projects/_`;
  }
}
