import { formatBytes } from "@/lib/analytics-stats";
import type { Deployment } from "@/lib/api";

export function formatBuildLocation(value: string | undefined): string {
  const v = (value ?? "local").trim().toLowerCase();
  if (v === "" || v === "local") return "On this server";
  if (v === "github_actions") return "GitHub Actions";
  if (v === "auto") return "Auto (remote when available)";
  if (v.startsWith("remote:")) {
    const id = v.slice("remote:".length);
    return id ? `Remote builder (${id.slice(0, 8)}…)` : "Remote builder";
  }
  return value ?? "On this server";
}

export function deploymentResourceStripRows(d: Deployment): {
  label: string;
  value: string;
}[] {
  const rows: { label: string; value: string }[] = [];
  if (d.budget_bytes != null && d.budget_bytes > 0) {
    rows.push({ label: "Build budget", value: formatBytes(d.budget_bytes) });
  }
  if (d.mem_peak_bytes != null && d.mem_peak_bytes > 0) {
    rows.push({ label: "Peak memory", value: formatBytes(d.mem_peak_bytes) });
  }
  if (d.swap_peak_bytes != null && d.swap_peak_bytes > 0) {
    rows.push({ label: "Peak swap", value: formatBytes(d.swap_peak_bytes) });
  }
  if (d.waited_ms != null && d.waited_ms > 0) {
    const sec = Math.round(d.waited_ms / 1000);
    rows.push({
      label: "Waited for capacity",
      value: sec < 60 ? `${sec}s` : `${Math.floor(sec / 60)}m ${sec % 60}s`,
    });
  }
  if (d.frozen_ms != null && d.frozen_ms > 0) {
    const sec = Math.round(d.frozen_ms / 1000);
    rows.push({
      label: "Paused (memory pressure)",
      value: sec < 60 ? `${sec}s` : `${Math.floor(sec / 60)}m ${sec % 60}s`,
    });
  }
  if (d.builder?.trim()) {
    rows.push({ label: "Builder", value: d.builder });
  }
  if (d.failure_kind?.trim()) {
    rows.push({ label: "Failure type", value: d.failure_kind });
  }
  return rows;
}

export function hasDeploymentResourceMetrics(d: Deployment): boolean {
  return deploymentResourceStripRows(d).length > 0;
}
