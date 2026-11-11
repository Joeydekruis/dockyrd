/** Matches internal bucket project slug (`internal/project/unassigned.go`). */
export const UNASSIGNED_PROJECT_SLUG = "unassigned";

export function isUnassignedProject(slug?: string | null): boolean {
  return slug === UNASSIGNED_PROJECT_SLUG;
}

export function projectDisplayName(name: string, slug?: string | null): string {
  if (isUnassignedProject(slug) || !name.trim()) {
    return "No project";
  }
  return name;
}
