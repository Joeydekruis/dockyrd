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
  buildEnabled?: boolean;
  installCmd?: string;
  buildScript?: string;
  buildCmd?: string;
  outputDir?: string;
  nodeImage?: string;
  indexFile?: string;
  errorFile?: string;
  customDomain?: string;
  autoHttps?: boolean;
  spaFallback?: boolean;
};

function coalesceCmd(value: unknown): string {
  if (value === undefined || value === null) return "";
  if (typeof value === "boolean") return "";
  const s = String(value).trim();
  if (s === "false" || s === "true") return "";
  return s;
}

/** Normalize API config blobs (fixes legacy `"false"` / boolean noise). */
export function normalizeStaticSiteConfig(
  raw: Record<string, unknown>,
): StaticSiteConfig {
  const source =
    raw.source && typeof raw.source === "object"
      ? (raw.source as StaticSiteConfig["source"])
      : undefined;
  return {
    ...raw,
    source,
    buildEnabled:
      typeof raw.buildEnabled === "boolean" ? raw.buildEnabled : undefined,
    installCmd: coalesceCmd(raw.installCmd),
    buildScript: coalesceCmd(raw.buildScript),
    buildCmd: coalesceCmd(raw.buildCmd),
    outputDir: typeof raw.outputDir === "string" ? raw.outputDir : undefined,
    nodeImage: typeof raw.nodeImage === "string" ? raw.nodeImage : undefined,
    indexFile:
      typeof raw.indexFile === "string" ? raw.indexFile.trim() : undefined,
    errorFile:
      typeof raw.errorFile === "string" ? raw.errorFile.trim() : undefined,
    spaFallback:
      typeof raw.spaFallback === "boolean" ? raw.spaFallback : undefined,
    autoDeploy:
      typeof raw.autoDeploy === "boolean" ? raw.autoDeploy : undefined,
  };
}

export function parseStaticSiteConfig(service: Service): StaticSiteConfig {
  return normalizeStaticSiteConfig(
    (service.config ?? {}) as Record<string, unknown>,
  );
}

export function isBuildEnabled(cfg: StaticSiteConfig): boolean {
  if (cfg.buildEnabled === false) return false;
  if (cfg.buildEnabled === true) return true;
  const parts = staticSiteBuildParts(cfg);
  return parts.install !== "" || parts.build !== "";
}

export function nodeMajorFromImage(nodeImage?: string): string {
  const tag = (nodeImage ?? "node:22-alpine").split(":").pop() ?? "22-alpine";
  const major = tag.replace(/\D.*$/, "") || "22";
  return major;
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

/** Static site list/detail: live deployments read as “Deployed”. */
export function staticSiteDeploymentStatusDisplay(status: string): {
  label: string;
  variant: StatusVariant;
} {
  const s = status.trim().toLowerCase();
  if (s === "live") return { label: "Deployed", variant: "ok" };
  return deploymentStatusDisplay(status);
}

export function deploymentTriggerLabel(trigger: string): string {
  const t = trigger.trim().toLowerCase();
  if (t === "github_push" || t === "github") return "Automatic on commit";
  if (t === "manual") return "Manual";
  if (t === "redeploy") return "Redeploy";
  if (!t) return "—";
  return trigger.replace(/_/g, " ");
}

export function deploymentSourceTitle(d: Deployment): string {
  const msg = d.commit_message?.trim();
  if (msg) {
    const first = msg.split("\n")[0]?.trim();
    if (first) return first;
  }
  if (d.branch?.trim()) return `Deploy ${d.branch.trim()}`;
  return deploymentTriggerLabel(d.trigger);
}

export function deploymentAuthorLabel(d: Deployment): string {
  const author = d.commit_author?.trim();
  return author || "—";
}

export function deploymentCommitShaShort(d: Deployment): string {
  const sha = d.commit_sha?.trim();
  if (!sha) return "—";
  return sha.slice(0, 7);
}

export function formatDeploymentStartedAt(unixSec?: number): string {
  if (!unixSec) return "—";
  return new Date(unixSec * 1000).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
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

export function splitBuildCmd(buildCmd: string | undefined): {
  install: string;
  build: string;
} {
  const raw = typeof buildCmd === "string" ? buildCmd.trim() : "";
  if (!raw || raw === "false") {
    return { install: "", build: "" };
  }
  const parts = raw.split("&&").map((p) => p.trim());
  if (parts.length >= 2) {
    return {
      install: parts[0] === "false" ? "" : parts[0],
      build:
        parts.slice(1).join(" && ") === "false"
          ? ""
          : parts.slice(1).join(" && "),
    };
  }
  const single = raw === "false" ? "" : raw;
  return { install: "", build: single };
}

/** Read install/build for display from config (explicit fields first). */
export function staticSiteBuildParts(cfg: StaticSiteConfig): {
  install: string;
  build: string;
} {
  const hasExplicit =
    cfg.installCmd !== undefined || cfg.buildScript !== undefined;
  if (hasExplicit) {
    return {
      install: coalesceCmd(cfg.installCmd),
      build: coalesceCmd(cfg.buildScript),
    };
  }
  return splitBuildCmd(cfg.buildCmd);
}

export function joinBuildCmd(install: string, build: string): string {
  const i = install.trim();
  const b = build.trim();
  if (!i && !b) return "";
  if (!i) return b;
  if (!b) return i;
  return `${i} && ${b}`;
}

export function formatBuildPart(value: string): string {
  const v = value.trim();
  return v === "" ? "—" : v;
}

export { domainStatusDisplay } from "@/lib/domain-dns";

export function visitSiteHref(hostname?: string): string | undefined {
  if (!hostname) return undefined;
  return hostname.startsWith("http") ? hostname : `https://${hostname}`;
}
