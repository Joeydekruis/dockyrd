import type { StatusVariant } from "@/components/dockyard/ui/status-badge";

export function formatResourceCount(count: number, noun: string): string {
  const label = count === 1 ? noun : `${noun}s`;
  return `${count} ${label}`.toUpperCase();
}

export function projectPathLabel(projectName: string): string {
  return `${projectName} / production`;
}

/** Project line for resource lists when a git branch is known (static sites). */
export function projectBranchLabel(
  projectName: string,
  branch?: string,
): string {
  const b = (branch ?? "main").trim() || "main";
  return `${projectName} / ${b}`;
}

export function resourceStatusDisplay(
  desiredState: string,
  surface: "application" | "static" | "database" = "application",
): { label: string; variant: StatusVariant } {
  const normalized = desiredState.trim().toLowerCase();
  if (normalized === "running") {
    if (surface === "static") return { label: "Live", variant: "ok" };
    if (surface === "database") return { label: "Healthy", variant: "ok" };
    return { label: "Healthy", variant: "ok" };
  }
  if (normalized === "stopped") {
    return { label: "Stopped", variant: "idle" };
  }
  if (normalized === "deploying" || normalized === "pending") {
    return { label: "Deploying", variant: "warn" };
  }
  if (normalized === "failed" || normalized === "error") {
    return { label: "Failing", variant: "err" };
  }
  if (!normalized) {
    return { label: "Unknown", variant: "idle" };
  }
  return {
    label: desiredState.charAt(0).toUpperCase() + desiredState.slice(1),
    variant: "info",
  };
}

export function summarizeApplicationStats(
  rows: { desired_state: string; project_id: string }[],
) {
  const total = rows.length;
  const projectCount = new Set(rows.map((r) => r.project_id)).size;
  let healthy = 0;
  let inProgress = 0;
  let failing = 0;
  let stopped = 0;
  for (const row of rows) {
    const s = row.desired_state.trim().toLowerCase();
    if (s === "running") healthy += 1;
    else if (s === "deploying" || s === "pending") inProgress += 1;
    else if (s === "failed" || s === "error") failing += 1;
    else if (s === "stopped") stopped += 1;
  }
  const needsAttention = failing + stopped;
  const attentionHint =
    needsAttention === 0
      ? "none need attention"
      : `${failing} failing · ${stopped} stopped`;
  return {
    total,
    projectCount,
    healthy,
    inProgress,
    needsAttention,
    attentionHint,
  };
}

export function formatTimeAgo(unixSec?: number): string | undefined {
  if (!unixSec) return undefined;
  const diffMs = Date.now() - unixSec * 1000;
  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes} min ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} h ago`;
  const days = Math.floor(hours / 24);
  if (days === 1) return "Yesterday";
  return `${days} days ago`;
}
