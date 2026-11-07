import type { StatusVariant } from "@/components/dockyrd/ui/status-badge";
import type { Project, Service } from "@/lib/api";
import { formatTimeAgo } from "@/lib/resource-dashboard";

export type ProjectOverview = {
  project: Project;
  appCount: number;
  databaseCount: number;
  staticSiteCount: number;
  bucketCount: number;
  serviceCount: number;
  status: { label: string; variant: StatusVariant };
  activityLine: string;
  environments: string[];
  summaryLine: string;
};

function countServices(services: Service[]) {
  let appCount = 0;
  let databaseCount = 0;
  let staticSiteCount = 0;
  let bucketCount = 0;
  for (const s of services) {
    if (s.kind === "app") appCount += 1;
    else if (s.kind === "static_site") staticSiteCount += 1;
    else if (s.kind === "bucket") bucketCount += 1;
    else if (s.kind === "postgres" || s.kind === "redis") databaseCount += 1;
  }
  return { appCount, databaseCount, staticSiteCount, bucketCount };
}

export function formatProjectServiceSummary(counts: {
  appCount: number;
  databaseCount: number;
  staticSiteCount: number;
  bucketCount: number;
}): string {
  const parts: string[] = [];
  if (counts.appCount > 0) {
    parts.push(`${counts.appCount} app${counts.appCount === 1 ? "" : "s"}`);
  }
  if (counts.databaseCount > 0) {
    parts.push(
      `${counts.databaseCount} database${counts.databaseCount === 1 ? "" : "s"}`,
    );
  }
  if (counts.staticSiteCount > 0) {
    parts.push(
      `${counts.staticSiteCount} static site${counts.staticSiteCount === 1 ? "" : "s"}`,
    );
  }
  if (counts.bucketCount > 0) {
    parts.push(
      `${counts.bucketCount} bucket${counts.bucketCount === 1 ? "" : "s"}`,
    );
  }
  if (parts.length === 0) return "No services yet";
  return parts.join(" · ");
}

function aggregateProjectStatus(services: Service[]): {
  label: string;
  variant: StatusVariant;
} {
  if (services.length === 0) {
    return { label: "Healthy", variant: "ok" };
  }
  const normalized = services.map((s) => s.desired_state.trim().toLowerCase());
  if (normalized.some((s) => s === "failed" || s === "error")) {
    return { label: "Failing", variant: "err" };
  }
  if (normalized.some((s) => s === "deploying" || s === "pending")) {
    return { label: "Deploying", variant: "warn" };
  }
  if (normalized.every((s) => s === "stopped")) {
    return { label: "Stopped", variant: "idle" };
  }
  return { label: "Healthy", variant: "ok" };
}

function projectActivityLine(
  status: { label: string },
  latestUnix: number | undefined,
): string {
  const time = formatTimeAgo(latestUnix);
  if (!time) return "Not deployed yet";
  if (status.label === "Failing") return `Build failed ${time}`;
  if (status.label === "Deploying") return `Deploying · started ${time}`;
  if (status.label === "Stopped") return `Stopped · ${time}`;
  return `Deployed ${time}`;
}

export function buildProjectOverview(
  project: Project,
  services: Service[],
): ProjectOverview {
  const counts = countServices(services);
  const serviceCount =
    counts.appCount +
    counts.databaseCount +
    counts.staticSiteCount +
    counts.bucketCount;
  const status = aggregateProjectStatus(services);
  const latestUnix = Math.max(
    project.updated_at,
    ...services.map((s) => s.updated_at),
    0,
  );
  return {
    project,
    ...counts,
    serviceCount,
    status,
    activityLine: projectActivityLine(status, latestUnix || undefined),
    environments: ["Production"],
    summaryLine: formatProjectServiceSummary(counts),
  };
}
