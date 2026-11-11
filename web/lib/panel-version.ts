/** Display label for the running dockyrd binary version (sidebar, health card). */
export function formatPanelVersion(version: string): string {
  const trimmed = version.trim();
  if (!trimmed) return "—";
  if (trimmed.startsWith("v")) return trimmed;
  return `v${trimmed}`;
}
