import { deploymentIsInProgress } from "@/lib/static-site-display";

const STATUS_PROGRESS: Record<string, number> = {
  queued: 8,
  waiting: 18,
  building: 48,
  starting: 72,
  running: 80,
  healthchecking: 92,
  live: 100,
  failed: 100,
  cancelled: 100,
};

export function deploymentProgressPercent(status: string): number {
  const key = status.trim().toLowerCase();
  return STATUS_PROGRESS[key] ?? 0;
}

export function deploymentDeployDone(status: string): boolean {
  const key = status.trim().toLowerCase();
  return key === "live" || key === "failed" || key === "cancelled";
}

export function deploymentStageLabel(status: string): string {
  const key = status.trim().toLowerCase();
  switch (key) {
    case "queued":
      return "Queued…";
    case "waiting":
      return "Waiting for build slot…";
    case "building":
      return "Building with Railpack…";
    case "starting":
      return "Starting container…";
    case "running":
      return "Container running…";
    case "healthchecking":
      return "Health checks…";
    case "live":
      return "Live — traffic switched";
    case "cancelled":
      return "Deploy cancelled";
    case "failed":
      return "Deploy failed";
    default:
      return deploymentIsInProgress(status) ? "Deploying…" : "—";
  }
}
