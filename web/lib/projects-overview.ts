import type { LucideIcon } from "lucide-react";
import { AppWindow, Database, Globe, HardDrive, Layers } from "lucide-react";
import type { StatusVariant } from "@/components/dockyard/ui/status-badge";
import type { Project, Service } from "@/lib/api";
import { formatTimeAgo, resourceStatusDisplay } from "@/lib/resource-dashboard";

export type ProjectResourceRow = {
  id: string;
  name: string;
  kindLabel: string;
  icon: LucideIcon;
  statusLabel: string;
  statusVariant: StatusVariant;
};

export const PROJECT_CARD_RESOURCE_LIMIT = 5;

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
  resources: ProjectResourceRow[];
  extraResourceCount: number;
};

function serviceResourceSurface(
  kind: string,
): "application" | "static" | "database" {
  if (kind === "static_site") return "static";
  if (kind === "postgres" || kind === "redis") return "database";
  return "application";
}

function serviceKindLabel(service: Service): string {
  switch (service.kind) {
    case "app":
      return "App";
    case "static_site":
      return "Static site";
    case "postgres": {
      const version = service.config?.version;
      return typeof version === "number" || typeof version === "string"
        ? `Postgres ${version}`
        : "Postgres";
    }
    case "redis": {
      const version = service.config?.version;
      return typeof version === "number" || typeof version === "string"
        ? `Redis ${version}`
        : "Redis";
    }
    case "bucket":
      return "Object storage";
    case "compose":
      return "Compose stack";
    default:
      return service.kind;
  }
}

function serviceResourceIcon(kind: string): LucideIcon {
  switch (kind) {
    case "static_site":
      return Globe;
    case "postgres":
    case "redis":
      return Database;
    case "bucket":
      return HardDrive;
    case "compose":
      return Layers;
    default:
      return AppWindow;
  }
}

function buildResourceRows(services: Service[]): ProjectResourceRow[] {
  const sorted = [...services].sort((a, b) => b.updated_at - a.updated_at);
  return sorted.map((service) => {
    const surface = serviceResourceSurface(service.kind);
    const status = resourceStatusDisplay(service.desired_state, surface);
    return {
      id: service.id,
      name: service.name,
      kindLabel: serviceKindLabel(service),
      icon: serviceResourceIcon(service.kind),
      statusLabel: status.label,
      statusVariant: status.variant,
    };
  });
}

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
  const resources = buildResourceRows(services);
  const extraResourceCount = Math.max(
    0,
    resources.length - PROJECT_CARD_RESOURCE_LIMIT,
  );

  return {
    project,
    ...counts,
    serviceCount,
    status,
    activityLine: projectActivityLine(status, latestUnix || undefined),
    environments: ["Production"],
    summaryLine: formatProjectServiceSummary(counts),
    resources,
    extraResourceCount,
  };
}
