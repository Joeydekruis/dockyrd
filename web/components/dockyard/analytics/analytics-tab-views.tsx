"use client";

import type { EChartsOption } from "echarts";
import {
  areaLineOption,
  dualAreaOption,
  horizontalStackOption,
  sparklineOption,
} from "@/components/dockyrd/analytics/chart-options";
import {
  AnalyticsChartCard,
  type AnalyticsTabId,
  ViewDetailsLink,
} from "@/components/dockyrd/analytics/use-analytics-data";
import { BuildCapacityCard } from "@/components/dockyrd/build-capacity-card";
import { Button } from "@/components/dockyrd/ui/button";
import { Chip } from "@/components/dockyrd/ui/chip";
import { EChart } from "@/components/dockyrd/ui/echart";
import { FilterChips } from "@/components/dockyrd/ui/filter-chips";
import { ProgressBar } from "@/components/dockyrd/ui/progress-bar";
import {
  ServerContainersTable,
  type ServerContainersTableRow,
} from "@/components/dockyrd/ui/server-containers-table";
import { StatCard } from "@/components/dockyrd/ui/stat-card";
import { analyticsChartColors as c } from "@/lib/analytics-chart-theme";
import {
  average,
  cpuPercentDisplay,
  cpuPercentFraction,
  diskUsedPercent,
  formatBytes,
  formatBytesPerSec,
  formatUptime,
  memUsedPercent,
  networkRatesFromPoints,
  peak,
  percentile,
  totalDelta,
} from "@/lib/analytics-stats";
import type {
  CleanupSummary,
  ContainerMetricSnapshot,
  dockyrdEvent,
  HostMetricPoint,
  ServerContainer,
  ServerInfo,
} from "@/lib/api";
import { useUserLocale } from "@/lib/user-locale-context";

type TabProps = {
  points: HostMetricPoint[];
  latest?: HostMetricPoint;
  serverInfo: ServerInfo | null;
  containerMetrics: ContainerMetricSnapshot[];
  rangeLabel: string;
  onSelectTab: (tab: AnalyticsTabId) => void;
};

function topByCpu(metrics: ContainerMetricSnapshot[], limit = 6) {
  return [...metrics]
    .sort((a, b) => b.cpu_percent - a.cpu_percent)
    .slice(0, limit);
}

function topByTraffic(metrics: ContainerMetricSnapshot[], limit = 6) {
  return [...metrics]
    .sort((a, b) => b.net_rx + b.net_tx - (a.net_rx + a.net_tx))
    .slice(0, limit);
}

function TopContainersTable({
  title,
  subtitle,
  rows,
  valueLabel,
  getValue,
  barPercent,
}: {
  title: string;
  subtitle: string;
  rows: ContainerMetricSnapshot[];
  valueLabel: string;
  getValue: (r: ContainerMetricSnapshot) => string;
  barPercent: (r: ContainerMetricSnapshot) => number;
}) {
  const maxBar = Math.max(...rows.map(barPercent), 1);
  return (
    <AnalyticsChartCard title={title} subtitle={subtitle}>
      {rows.length === 0 ? (
        <p className="text-sm" style={{ color: "var(--muted-text)" }}>
          No container metrics yet — wait ~1 minute after containers start.
        </p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)]">
              <th className="pb-2 pr-3">Container</th>
              <th className="pb-2 pr-3">Project</th>
              <th className="pb-2">{valueLabel}</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr
                key={row.service_id}
                className="border-t border-[var(--border)]"
              >
                <td className="py-2 pr-3 font-mono text-[13px]">
                  {row.name || row.service_id.slice(0, 8)}
                </td>
                <td className="py-2 pr-3">{row.project_label || "—"}</td>
                <td className="py-2">
                  <div className="flex items-center gap-2">
                    <span className="w-12 font-mono tabular-nums">
                      {getValue(row)}
                    </span>
                    <div className="min-w-[80px] flex-1">
                      <ProgressBar value={(barPercent(row) / maxBar) * 100} />
                    </div>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </AnalyticsChartCard>
  );
}

function rangeHint(rangeLabel: string): string {
  return `last ${rangeLabel}`;
}

export function OverviewTab({
  points,
  latest,
  serverInfo,
  rangeLabel,
  onSelectTab,
}: TabProps) {
  const { metricTimeLabels, prefs } = useUserLocale();
  const labels = metricTimeLabels(points);
  const cpuSeries = points.map((p) => cpuPercentFraction(p.cpu));
  const memSeries = points.map((p) => memUsedPercent(p));
  const diskSeries = points.map((p) => diskUsedPercent(p));
  const { inbound } = networkRatesFromPoints(points, prefs);
  const _netLabels =
    inbound.length > 0 ? labels.slice(-inbound.length) : labels;

  const cards = [
    {
      id: "cpu" as const,
      label: "CPU",
      value: latest ? cpuPercentDisplay(latest.cpu) : "—",
      hint: serverInfo?.cpu_count
        ? `${serverInfo.cpu_count} vCPU · ${rangeHint(rangeLabel)}`
        : rangeHint(rangeLabel),
      series: cpuSeries,
      yMax: 100,
    },
    {
      id: "memory" as const,
      label: "Memory",
      value: latest ? formatBytes(latest.mem_used) : "—",
      hint: latest?.mem_total
        ? `of ${formatBytes(latest.mem_total)} · ${rangeHint(rangeLabel)}`
        : rangeHint(rangeLabel),
      series: memSeries,
      yMax: 100,
    },
    {
      id: "disk" as const,
      label: "Disk",
      value: latest ? formatBytes(latest.disk_used) : "—",
      hint: latest?.disk_total
        ? `of ${formatBytes(latest.disk_total)} · ${rangeHint(rangeLabel)}`
        : rangeHint(rangeLabel),
      series: diskSeries,
      yMax: 100,
    },
    {
      id: "network" as const,
      label: "Network",
      value: inbound.length > 0 ? formatBytesPerSec(inbound.at(-1) ?? 0) : "—",
      hint: `inbound · ${rangeHint(rangeLabel)}`,
      series: inbound.length > 0 ? inbound : [0],
      yMax: Math.max(...inbound, 1),
      format: true,
    },
  ];

  return (
    <div className="space-y-4">
      <BuildCapacityCard />
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {cards.map((card) => (
          <article
            key={card.id}
            className="rounded-[20px] border p-5"
            style={{
              background: "var(--surface)",
              borderColor: "var(--border)",
            }}
          >
            <p
              className="font-mono text-[11px] uppercase tracking-widest"
              style={{ color: "var(--muted-text)" }}
            >
              {card.label}
            </p>
            <p className="mt-2 text-[32px] font-semibold leading-none tabular-nums">
              {card.value}
            </p>
            <p
              className="mt-2 text-[13px]"
              style={{ color: "var(--muted-text)" }}
            >
              {card.hint}
            </p>
            <EChart
              option={sparklineOption(
                card.series,
                card.format ? card.yMax : 100,
              )}
              height={48}
              ariaLabel={`${card.label} sparkline`}
            />
            <ViewDetailsLink tab={card.id} onSelectTab={onSelectTab} />
          </article>
        ))}
      </div>
    </div>
  );
}

export function CpuTab({
  points,
  latest,
  serverInfo,
  rangeLabel,
  containerMetrics,
}: TabProps) {
  const { metricTimeLabels } = useUserLocale();
  const labels = metricTimeLabels(points);
  const cpuSeries = points.map((p) => cpuPercentFraction(p.cpu));
  const now = latest ? cpuPercentFraction(latest.cpu) : 0;
  const chartOpt = areaLineOption({
    labels,
    values: cpuSeries,
    yMax: 100,
    unit: "%",
    yFormatter: (v) => `${v.toFixed(0)}%`,
    alertAt: 80,
    alertLabel: "Alert at 80%",
    seriesName: "Total usage",
  });

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Now"
          value={`${now.toFixed(0)}%`}
          hint={
            serverInfo?.cpu_count ? `${serverInfo.cpu_count} vCPU` : undefined
          }
        />
        <StatCard
          label="Average"
          value={`${average(cpuSeries).toFixed(0)}%`}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="Peak"
          value={`${peak(cpuSeries).toFixed(0)}%`}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="95th percentile"
          value={`${percentile(cpuSeries, 95).toFixed(0)}%`}
          hint="busy 5% of the time"
        />
      </div>
      <AnalyticsChartCard
        title="CPU usage"
        subtitle={
          serverInfo?.cpu_count
            ? `All ${serverInfo.cpu_count} vCPUs combined`
            : "Host CPU"
        }
      >
        <EChart option={chartOpt} height={280} ariaLabel="CPU usage chart" />
      </AnalyticsChartCard>
      <div className="grid gap-6 lg:grid-cols-2">
        <AnalyticsChartCard title="Per core" subtitle="Current load per vCPU">
          {serverInfo?.cores && serverInfo.cores.length > 0 ? (
            <div className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
              {serverInfo.cores.map((core) => (
                <div key={core.core} className="flex items-center gap-2">
                  <span className="w-14 text-[var(--muted-text)]">
                    Core {core.core}
                  </span>
                  <div className="flex-1">
                    <ProgressBar value={core.cpu * 100} />
                  </div>
                  <span className="w-10 font-mono tabular-nums text-right">
                    {(core.cpu * 100).toFixed(0)}%
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm" style={{ color: "var(--muted-text)" }}>
              Per-core usage appears after a second live sample (Linux hosts).
            </p>
          )}
        </AnalyticsChartCard>
        <TopContainersTable
          title="Top containers"
          subtitle="Share of total CPU right now"
          rows={topByCpu(containerMetrics)}
          valueLabel="CPU"
          getValue={(r) => `${(r.cpu_percent * 100).toFixed(1)}%`}
          barPercent={(r) => r.cpu_percent * 100}
        />
      </div>
    </div>
  );
}

export function MemoryTab({
  points,
  latest,
  rangeLabel,
  serverInfo,
  containerMetrics,
}: TabProps) {
  const { metricTimeLabels } = useUserLocale();
  const labels = metricTimeLabels(points);
  const usedSeries = points.map((p) => p.mem_used / (1024 * 1024 * 1024));
  const nowGb = latest ? latest.mem_used / (1024 * 1024 * 1024) : 0;
  const totalGb = latest?.mem_total
    ? latest.mem_total / (1024 * 1024 * 1024)
    : 0;

  const chartOpt = areaLineOption({
    labels,
    values: usedSeries,
    yMax: totalGb > 0 ? totalGb * 1.05 : undefined,
    yFormatter: (v) => `${v.toFixed(1)} GB`,
    seriesName: "Used",
  });

  const free = latest ? Math.max(0, latest.mem_total - latest.mem_used) : 0;

  const containersMem = containerMetrics.reduce(
    (s, m) => s + m.memory_bytes,
    0,
  );
  const systemMem = latest ? Math.max(0, latest.mem_used - containersMem) : 0;

  const stackOpt = horizontalStackOption([
    { name: "Containers", value: containersMem, color: c.primary },
    { name: "System", value: systemMem, color: c.secondary },
    { name: "Free", value: free, color: c.segmentFree },
  ]);

  const swapGb =
    serverInfo?.swap_total && serverInfo.swap_total > 0
      ? (serverInfo.swap_used ?? 0) / (1024 * 1024 * 1024)
      : 0;

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Used now"
          value={`${nowGb.toFixed(1)} GB`}
          hint={totalGb > 0 ? `of ${totalGb.toFixed(0)} GB` : undefined}
        />
        <StatCard
          label="Average"
          value={`${average(usedSeries).toFixed(1)} GB`}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="Peak"
          value={`${peak(usedSeries).toFixed(1)} GB`}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="Swap"
          value={serverInfo?.swap_total ? `${swapGb.toFixed(1)} GB` : "0.0 GB"}
          hint={
            serverInfo?.swap_total
              ? serverInfo.swap_used
                ? "in use"
                : "not in use"
              : "not available"
          }
        />
      </div>
      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <AnalyticsChartCard
            title="Memory usage"
            subtitle={
              totalGb > 0 ? `Out of ${totalGb.toFixed(0)} GB` : undefined
            }
          >
            <EChart option={chartOpt} height={280} ariaLabel="Memory chart" />
          </AnalyticsChartCard>
        </div>
        <AnalyticsChartCard
          title="Where the memory goes"
          subtitle="Host + containers"
        >
          <EChart option={stackOpt} height={24} ariaLabel="Memory breakdown" />
          <ul className="mt-4 space-y-2 text-sm">
            <li className="flex justify-between">
              <span>Containers</span>
              <span className="font-mono">{formatBytes(containersMem)}</span>
            </li>
            <li className="flex justify-between">
              <span>System and Dockyrd</span>
              <span className="font-mono">{formatBytes(systemMem)}</span>
            </li>
            <li className="flex justify-between">
              <span>Free</span>
              <span className="font-mono">{formatBytes(free)}</span>
            </li>
          </ul>
        </AnalyticsChartCard>
      </div>
      <TopContainersTable
        title="Top containers"
        subtitle="Memory in use right now"
        rows={[...containerMetrics]
          .sort((a, b) => b.memory_bytes - a.memory_bytes)
          .slice(0, 6)}
        valueLabel="Memory"
        getValue={(r) => formatBytes(r.memory_bytes)}
        barPercent={(r) => r.memory_bytes}
      />
    </div>
  );
}

export function DiskTab({
  points,
  latest,
  rangeLabel,
  cleanup,
}: TabProps & { cleanup: CleanupSummary | null }) {
  const { metricTimeLabels } = useUserLocale();
  const labels = metricTimeLabels(points);
  const usedGb = points.map((p) => p.disk_used / (1024 * 1024 * 1024));
  const totalGb = latest?.disk_total
    ? latest.disk_total / (1024 * 1024 * 1024)
    : 0;
  const used = latest?.disk_used ?? 0;
  const total = latest?.disk_total ?? 0;
  const free = Math.max(0, total - used);
  const reclaimable =
    (cleanup?.build_cache_bytes ?? 0) +
    (cleanup?.dangling_images_bytes ?? 0) +
    (cleanup?.old_releases_bytes ?? 0);
  const firstPoint = points[0];
  const lastPoint = points.at(-1);
  const growth =
    firstPoint && lastPoint && points.length >= 2
      ? (used - (firstPoint.disk_used ?? 0)) /
        Math.max(1, (lastPoint.ts - firstPoint.ts) / 86400)
      : 0;

  const chartOpt = areaLineOption({
    labels,
    values: usedGb,
    yMax: totalGb > 0 ? totalGb * 1.05 : undefined,
    yFormatter: (v) => `${v.toFixed(0)} GB`,
    alertAt: totalGb > 0 ? totalGb * 0.8 : undefined,
    alertLabel: "Alert at 80%",
    seriesName: "Used space",
  });

  const stackOpt = horizontalStackOption([
    {
      name: "Measured used",
      value: used,
      color: c.segmentVolume,
    },
    {
      name: "Build cache",
      value: cleanup?.build_cache_bytes ?? 0,
      color: c.segmentCache,
    },
    {
      name: "Free",
      value: free,
      color: c.segmentFree,
    },
  ]);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Used"
          value={formatBytes(used)}
          hint={
            total > 0 && latest
              ? `${diskUsedPercent(latest).toFixed(0)}% of ${formatBytes(total)}`
              : undefined
          }
        />
        <StatCard
          label="Free"
          value={formatBytes(free)}
          hint="on data volume"
        />
        <StatCard
          label="Growth"
          value={growth > 0 ? `+${formatBytes(growth)}/day` : "—"}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="Reclaimable"
          value={formatBytes(reclaimable)}
          hint="via Docker cleanup"
        />
      </div>
      <AnalyticsChartCard
        title="Disk usage"
        subtitle={
          total > 0 ? `${formatBytes(total)} volume` : "dockyrd data directory"
        }
      >
        <EChart option={chartOpt} height={280} ariaLabel="Disk usage chart" />
      </AnalyticsChartCard>
      <div className="grid gap-6 lg:grid-cols-2">
        <AnalyticsChartCard
          title="What is using the disk"
          subtitle="Best-effort breakdown"
        >
          <EChart option={stackOpt} height={24} ariaLabel="Disk breakdown" />
          <ul className="mt-4 space-y-2 text-sm">
            <li className="flex justify-between">
              <span>Data directory (measured)</span>
              <span className="font-mono">{formatBytes(used)}</span>
            </li>
            <li className="flex justify-between">
              <span>Build cache</span>
              <span className="font-mono">
                {formatBytes(cleanup?.build_cache_bytes ?? 0)}
              </span>
            </li>
            <li className="flex justify-between">
              <span>Free</span>
              <span className="font-mono">{formatBytes(free)}</span>
            </li>
          </ul>
        </AnalyticsChartCard>
        <UnavailablePanel title="Largest volumes">
          Per-volume sizes for managed databases are not exposed in the API yet.
          Use Docker cleanup below for reclaimable Docker layers.
        </UnavailablePanel>
      </div>
    </div>
  );
}

export function NetworkTab({ points, rangeLabel, containerMetrics }: TabProps) {
  const { prefs } = useUserLocale();
  const { inbound, outbound, labels } = networkRatesFromPoints(points, prefs);
  const hasNet = inbound.some((v) => v > 0) || outbound.some((v) => v > 0);
  const chartOpt: EChartsOption = dualAreaOption({
    labels,
    primary: inbound,
    secondary: outbound,
    primaryName: "Inbound",
    secondaryName: "Outbound",
    yFormatter: formatBytesPerSec,
  });

  const rxTotal = totalDelta(points, (p) => p.net_rx);
  const txTotal = totalDelta(points, (p) => p.net_tx);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Inbound now"
          value={inbound.length ? formatBytesPerSec(inbound.at(-1) ?? 0) : "—"}
          hint={
            inbound.length
              ? `peak ${formatBytesPerSec(peak(inbound))}`
              : undefined
          }
        />
        <StatCard
          label="Outbound now"
          value={
            outbound.length ? formatBytesPerSec(outbound.at(-1) ?? 0) : "—"
          }
          hint={
            outbound.length
              ? `peak ${formatBytesPerSec(peak(outbound))}`
              : undefined
          }
        />
        <StatCard
          label="Received"
          value={rxTotal > 0 ? formatBytes(rxTotal) : "—"}
          hint={rangeHint(rangeLabel)}
        />
        <StatCard
          label="Sent"
          value={txTotal > 0 ? formatBytes(txTotal) : "—"}
          hint={rangeHint(rangeLabel)}
        />
      </div>
      <AnalyticsChartCard
        title="Network traffic"
        subtitle="Inbound and outbound (all interfaces except loopback)"
      >
        {hasNet || points.length > 1 ? (
          <EChart option={chartOpt} height={280} ariaLabel="Network chart" />
        ) : (
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            Network counters appear after a few minutes on Linux hosts.
            Non-Linux dev builds do not sample network yet.
          </p>
        )}
      </AnalyticsChartCard>
      <TopContainersTable
        title="Top containers"
        subtitle="Cumulative rx+tx from latest sample"
        rows={topByTraffic(containerMetrics)}
        valueLabel="Traffic"
        getValue={(r) => formatBytes(r.net_rx + r.net_tx)}
        barPercent={(r) => r.net_rx + r.net_tx}
      />
    </div>
  );
}

function UnavailablePanel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div
      className="rounded-[20px] border p-5 text-sm"
      style={{ borderColor: "var(--border)", background: "var(--surface)" }}
    >
      <h3 className="font-semibold">{title}</h3>
      <p className="mt-2" style={{ color: "var(--muted-text)" }}>
        {children}
      </p>
    </div>
  );
}

export function EventsAndOpsSection({
  events,
  allEvents,
  eventFilter,
  setEventFilter,
  onAck,
  cleanup,
  isAdmin,
  cleanupBusy,
  onCleanup,
  containerRows,
}: {
  events: dockyrdEvent[];
  allEvents: dockyrdEvent[];
  eventFilter: string;
  setEventFilter: (v: string) => void;
  onAck: (id: string) => void;
  cleanup: CleanupSummary | null;
  isAdmin: boolean;
  cleanupBusy: boolean;
  onCleanup: () => void;
  containerRows: ServerContainersTableRow[];
}) {
  const needsAttention = allEvents.filter(
    (ev) =>
      !ev.acknowledged_at && (ev.level === "error" || ev.level === "warning"),
  );

  return (
    <div className="space-y-8">
      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-lg font-semibold">Important events</h2>
          <FilterChips
            options={[
              { value: "all", label: "All" },
              { value: "issues", label: "Issues" },
              { value: "deploys", label: "Deploys" },
              { value: "system", label: "System" },
            ]}
            active={eventFilter}
            onChange={setEventFilter}
          />
        </div>
        <ul
          className="divide-y divide-[var(--border)] rounded-[20px] border"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          {events.length === 0 ? (
            <li
              className="px-5 py-8 text-sm"
              style={{ color: "var(--muted-text)" }}
            >
              No events in this filter.
            </li>
          ) : (
            events.slice(0, 8).map((ev) => (
              <li
                key={ev.id}
                className="flex flex-wrap items-start justify-between gap-3 px-5 py-4"
              >
                <div>
                  <div className="mb-1 flex flex-wrap items-center gap-2">
                    <EventLevelChip level={ev.level} />
                    <span className="text-sm font-medium">{ev.title}</span>
                  </div>
                  {ev.detail ? (
                    <p
                      className="text-[13px]"
                      style={{ color: "var(--muted-text)" }}
                    >
                      {ev.detail}
                    </p>
                  ) : null}
                </div>
                <button
                  type="button"
                  className="rounded-md border px-2 py-1 text-xs font-medium"
                  style={{ borderColor: "var(--border)" }}
                  onClick={() => onAck(ev.id)}
                >
                  Dismiss
                </button>
              </li>
            ))
          )}
        </ul>
      </section>

      <div className="grid gap-6 lg:grid-cols-2">
        <div
          className="rounded-[20px] border p-5"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Needs attention</h2>
            <Chip>{needsAttention.length} open</Chip>
          </div>
          <ul className="mt-4 space-y-3">
            {needsAttention.length === 0 ? (
              <li className="text-sm" style={{ color: "var(--muted-text)" }}>
                Nothing flagged.
              </li>
            ) : (
              needsAttention.slice(0, 5).map((ev) => (
                <li key={ev.id} className="text-sm">
                  <p>{ev.title}</p>
                  {ev.action_href ? (
                    <Button
                      href={ev.action_href}
                      variant="link"
                      className="mt-1 text-[13px]"
                    >
                      Open service →
                    </Button>
                  ) : null}
                </li>
              ))
            )}
          </ul>
        </div>
        <div
          className="rounded-[20px] border p-5"
          style={{ borderColor: "var(--border)", background: "var(--surface)" }}
        >
          <h2 className="text-lg font-semibold">Docker cleanup</h2>
          {cleanup ? (
            <>
              <ul className="mt-3 space-y-2 text-sm">
                <li className="flex justify-between">
                  <span>Build cache</span>
                  <span className="font-mono">
                    {formatBytes(cleanup.build_cache_bytes)}
                  </span>
                </li>
                <li className="flex justify-between">
                  <span>Unused images</span>
                  <span className="font-mono">
                    {formatBytes(cleanup.dangling_images_bytes)}
                  </span>
                </li>
                <li className="flex justify-between">
                  <span>Stopped containers</span>
                  <span className="font-mono">
                    {cleanup.stopped_unmanaged_containers}
                  </span>
                </li>
              </ul>
              {isAdmin ? (
                <Button
                  type="button"
                  disabled={cleanupBusy}
                  className="mt-4"
                  variant="outline"
                  onClick={onCleanup}
                >
                  {cleanupBusy ? "Cleaning…" : "Review and clean up"}
                </Button>
              ) : null}
            </>
          ) : (
            <p className="mt-2 text-sm" style={{ color: "var(--muted-text)" }}>
              Docker is not available on this host.
            </p>
          )}
        </div>
      </div>

      <ServerContainersTable
        rows={containerRows}
        runningCount={containerRows.length}
      />
    </div>
  );
}

function EventLevelChip({ level }: { level: string }) {
  const tone =
    level === "error"
      ? { bg: "#fdecea", fg: "#b3261e", label: "ERROR" }
      : level === "warning"
        ? { bg: "#fff4e5", fg: "#b45309", label: "WARNING" }
        : level === "success"
          ? { bg: "#e6f4ea", fg: "#1a7f4b", label: "SUCCESS" }
          : { bg: "#e3f0ff", fg: "#0b57c9", label: "INFO" };
  return (
    <span
      className="rounded px-1.5 py-0.5 font-mono text-[10px] tracking-widest"
      style={{ background: tone.bg, color: tone.fg }}
    >
      {tone.label}
    </span>
  );
}

export function containerTableRows(
  containers: ServerContainer[],
): ServerContainersTableRow[] {
  const now = Math.floor(Date.now() / 1000);
  return containers.map((c) => {
    const uptimeSec = c.created_at ? Math.max(0, now - c.created_at) : 0;
    const statusVariant =
      c.state === "running" ? (c.kind === "system" ? "idle" : "ok") : "warn";
    return {
      id: c.id,
      name: c.name,
      projectLabel:
        c.project_label ?? (c.kind === "system" ? "dockyrd" : (c.kind ?? "—")),
      cpu: c.cpu_percent != null ? `${(c.cpu_percent * 100).toFixed(1)}%` : "—",
      memory: c.memory_bytes != null ? formatBytes(c.memory_bytes) : "—",
      uptime: c.created_at ? formatUptime(uptimeSec) : "—",
      statusLabel:
        c.state === "running"
          ? c.kind === "system"
            ? "SYSTEM"
            : "HEALTHY"
          : c.state.toUpperCase(),
      statusVariant,
    };
  });
}

export function rangeLabelFromRange(range: string): string {
  switch (range) {
    case "1h":
      return "1 hour";
    case "7d":
      return "7 days";
    case "30d":
      return "30 days";
    default:
      return "24 hours";
  }
}
