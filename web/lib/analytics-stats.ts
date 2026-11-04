import type { HostMetricPoint } from "@/lib/api";
import { formatTimeAxisLabel, type UserLocalePrefs } from "@/lib/user-locale";

export function formatBytes(n: number): string {
  if (n <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let v = n;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i += 1;
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export function formatBytesPerSec(n: number): string {
  if (n <= 0) return "0 B/s";
  return `${formatBytes(n)}/s`;
}

export function cpuPercentFraction(cpu: number): number {
  if (cpu <= 1) return cpu * 100;
  return cpu;
}

export function cpuPercentDisplay(cpu: number): string {
  return `${cpuPercentFraction(cpu).toFixed(1)}%`;
}

export function memUsedPercent(point: HostMetricPoint): number {
  if (point.mem_total <= 0) return 0;
  return (point.mem_used / point.mem_total) * 100;
}

export function diskUsedPercent(point: HostMetricPoint): number {
  if (point.disk_total <= 0) return 0;
  return (point.disk_used / point.disk_total) * 100;
}

export function average(values: number[]): number {
  if (values.length === 0) return 0;
  return values.reduce((a, b) => a + b, 0) / values.length;
}

export function peak(values: number[]): number {
  if (values.length === 0) return 0;
  return Math.max(...values);
}

export function percentile(values: number[], p: number): number {
  if (values.length === 0) return 0;
  const sorted = [...values].sort((a, b) => a - b);
  const idx = Math.min(
    sorted.length - 1,
    Math.max(0, Math.ceil((p / 100) * sorted.length) - 1),
  );
  return sorted[idx] ?? 0;
}

export function formatTimeAxis(
  ts: number,
  index: number,
  total: number,
  prefs: UserLocalePrefs,
): string {
  return formatTimeAxisLabel(ts, index, total, prefs);
}

export function formatDuration(sec: number): string {
  if (sec < 60) return `${sec}s`;
  if (sec < 3600) return `${Math.floor(sec / 60)} min`;
  if (sec < 86400) return `${Math.floor(sec / 3600)} h`;
  const days = Math.floor(sec / 86400);
  return `${days} d`;
}

export function formatUptime(sec: number): string {
  return formatDuration(sec);
}

/** Net throughput (bytes/sec) between consecutive host metric points. */
export function networkRatesFromPoints(
  points: HostMetricPoint[],
  prefs: UserLocalePrefs,
): {
  inbound: number[];
  outbound: number[];
  labels: string[];
} {
  const inbound: number[] = [];
  const outbound: number[] = [];
  const labels: string[] = [];
  for (let i = 1; i < points.length; i += 1) {
    const prev = points[i - 1];
    const cur = points[i];
    const dt = cur.ts - prev.ts;
    if (dt <= 0) continue;
    const rx = ((cur.net_rx ?? 0) - (prev.net_rx ?? 0)) / dt;
    const tx = ((cur.net_tx ?? 0) - (prev.net_tx ?? 0)) / dt;
    inbound.push(Math.max(0, rx));
    outbound.push(Math.max(0, tx));
    labels.push(formatTimeAxis(cur.ts, i, points.length, prefs));
  }
  return { inbound, outbound, labels };
}

export function totalDelta(
  points: HostMetricPoint[],
  pick: (p: HostMetricPoint) => number,
): number {
  if (points.length < 2) return 0;
  const first = points[0];
  const last = points.at(-1);
  if (!first || !last) return 0;
  return Math.max(0, pick(last) - pick(first));
}
