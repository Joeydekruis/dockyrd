import type { StatusVariant } from "@/components/dockyard/ui/status-badge";

export function serviceStatusDisplay(desiredState: string): {
  label: string;
  variant: StatusVariant;
} {
  const normalized = desiredState.trim().toLowerCase();
  if (normalized === "running") {
    return { label: "Running", variant: "ok" };
  }
  if (normalized === "stopped") {
    return { label: "Stopped", variant: "idle" };
  }
  if (normalized === "deploying" || normalized === "pending") {
    return { label: "Deploying", variant: "info" };
  }
  if (normalized === "failed" || normalized === "error") {
    return { label: "Failed", variant: "err" };
  }
  if (!normalized) {
    return { label: "Unknown", variant: "idle" };
  }
  return {
    label: desiredState.charAt(0).toUpperCase() + desiredState.slice(1),
    variant: "info",
  };
}

export function databaseEngineLabel(kind: string): string {
  if (kind === "postgres") return "Postgres";
  if (kind === "redis") return "Redis";
  return kind;
}

export function filterByQuery<T>(
  rows: T[],
  query: string,
  pickStrings: (row: T) => string[],
): T[] {
  const q = query.trim().toLowerCase();
  if (!q) return rows;
  return rows.filter((row) =>
    pickStrings(row).some((s) => s.toLowerCase().includes(q)),
  );
}
