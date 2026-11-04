import type { StatusVariant } from "@/components/dockyard/ui/status-badge";
import type { Deployment, Service, ServiceDomain } from "@/lib/api";

export type StaticSiteConfig = {
  source?: {
    type?: string;
    repo?: string;
    branch?: string;
    rootDir?: string;
  };
  autoDeploy?: boolean;
  buildCmd?: string;
  outputDir?: string;
  nodeImage?: string;
  customDomain?: string;
  autoHttps?: boolean;
  spaFallback?: boolean;
};

export function parseStaticSiteConfig(service: Service): StaticSiteConfig {
  return (service.config ?? {}) as StaticSiteConfig;
}

export function primaryDomainHost(
  domains: ServiceDomain[],
): string | undefined {
  const primary =
    domains.find((d) => d.is_primary) ??
    domains.find((d) => !d.is_platform) ??
    domains[0];
  return primary?.hostname;
}

export function deploymentStatusDisplay(status: string): {
  label: string;
  variant: StatusVariant;
} {
  const s = status.trim().toLowerCase();
  if (s === "live") return { label: "Live", variant: "ok" };
  if (s === "failed" || s === "error")
    return { label: "Failed", variant: "err" };
  if (s === "superseded") return { label: "Superseded", variant: "idle" };
  if (s === "building" || s === "queued" || s === "running") {
    return { label: "Deploying", variant: "warn" };
  }
  if (s === "cancelled") return { label: "Cancelled", variant: "idle" };
  return {
    label: status.charAt(0).toUpperCase() + status.slice(1),
    variant: "info",
  };
}

export function deploymentDurationSec(d: Deployment): number | undefined {
  const start = d.started_at ?? d.queued_at;
  const end = d.finished_at;
  if (!start || !end || end < start) return undefined;
  return end - start;
}

export function formatDurationSec(sec: number): string {
  if (sec < 60) return `${sec}s`;
  const m = Math.floor(sec / 60);
  const s = sec % 60;
  return s > 0 ? `${m}m ${s}s` : `${m}m`;
}

export function formatDeploymentWhen(unixSec?: number): string {
  if (!unixSec) return "—";
  const d = new Date(unixSec * 1000);
  const now = new Date();
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate();
  const time = d.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
  if (sameDay) return `Today ${time}`;
  return d.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function splitBuildCmd(buildCmd: string): {
  install: string;
  build: string;
} {
  const parts = buildCmd.split("&&").map((p) => p.trim());
  if (parts.length >= 2) {
    return { install: parts[0], build: parts.slice(1).join(" && ") };
  }
  return { install: "npm ci", build: buildCmd.trim() || "npm run build" };
}

export function domainStatusDisplay(domain: ServiceDomain): {
  label: string;
  variant: StatusVariant;
} {
  const dns = (domain.dns_status ?? "").toLowerCase();
  if (dns === "active") return { label: "Active", variant: "ok" };
  if (dns === "pending") return { label: "Pending DNS", variant: "warn" };
  return { label: domain.dns_status ?? "Unknown", variant: "idle" };
}

export function visitSiteHref(hostname?: string): string | undefined {
  if (!hostname) return undefined;
  return hostname.startsWith("http") ? hostname : `https://${hostname}`;
}
