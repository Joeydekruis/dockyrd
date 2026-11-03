"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  ackEvent,
  type CleanupSummary,
  type DockyardEvent,
  getMe,
  getServerCleanup,
  getServerContainers,
  getServerMetrics,
  listEvents,
  type MetricRange,
  postServerCleanup,
  type ServerContainer,
  type ServerMetricsResponse,
} from "@/lib/api";

const ranges: { id: MetricRange; label: string }[] = [
  { id: "1h", label: "1 hour" },
  { id: "24h", label: "24 hours" },
  { id: "7d", label: "7 days" },
];

const eventFilters = [
  { id: "all", label: "All" },
  { id: "deploys", label: "Deploys" },
] as const;

function formatBytes(n: number): string {
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

function Sparkline({
  values,
  stroke = "var(--primary)",
}: {
  values: number[];
  stroke?: string;
}) {
  if (values.length < 2) {
    return (
      <div className="h-8 w-full rounded bg-[var(--muted)]/30" aria-hidden />
    );
  }
  const min = Math.min(...values);
  const max = Math.max(...values);
  const span = max - min || 1;
  const w = 160;
  const h = 32;
  const points = values
    .map((v, i) => {
      const x = (i / (values.length - 1)) * w;
      const y = h - ((v - min) / span) * (h - 4) - 2;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");
  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      className="h-8 w-full text-[var(--primary)]"
      role="img"
      aria-hidden
    >
      <polyline
        fill="none"
        stroke={stroke}
        strokeWidth="1.5"
        strokeLinejoin="round"
        strokeLinecap="round"
        points={points}
      />
    </svg>
  );
}

function MetricTile({
  label,
  value,
  sub,
  series,
}: {
  label: string;
  value: string;
  sub?: string;
  series: number[];
}) {
  return (
    <article className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
      <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-foreground)]">
        {label}
      </p>
      <p className="mt-1 text-2xl font-semibold tabular-nums text-[var(--foreground)]">
        {value}
      </p>
      {sub ? (
        <p className="text-xs text-[var(--muted-foreground)]">{sub}</p>
      ) : null}
      <div className="mt-3">
        <Sparkline values={series} />
      </div>
    </article>
  );
}

export function AnalyticsPage() {
  const [range, setRange] = useState<MetricRange>("24h");
  const [metrics, setMetrics] = useState<ServerMetricsResponse | null>(null);
  const [events, setEvents] = useState<DockyardEvent[]>([]);
  const [eventFilter, setEventFilter] =
    useState<(typeof eventFilters)[number]["id"]>("all");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [containers, setContainers] = useState<ServerContainer[]>([]);
  const [cleanup, setCleanup] = useState<CleanupSummary | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [cleanupBusy, setCleanupBusy] = useState(false);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [m, ev] = await Promise.all([
        getServerMetrics(range),
        listEvents({
          limit: 50,
          category: eventFilter === "all" ? undefined : eventFilter,
        }),
      ]);
      setMetrics(m);
      setEvents(ev.events);
      try {
        const [{ containers: cs }, sum] = await Promise.all([
          getServerContainers(),
          getServerCleanup(),
        ]);
        setContainers(cs.filter((c) => c.state === "running"));
        setCleanup(sum);
      } catch {
        setContainers([]);
        setCleanup(null);
      }
      try {
        const me = await getMe();
        setIsAdmin(me.role === "owner" || me.role === "admin");
      } catch {
        setIsAdmin(false);
      }
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load analytics");
    } finally {
      setLoading(false);
    }
  }, [range, eventFilter]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const latest = metrics?.points.at(-1);
  const cpuSeries = useMemo(
    () => (metrics?.points ?? []).map((p) => p.cpu * 100),
    [metrics],
  );
  const memSeries = useMemo(
    () =>
      (metrics?.points ?? []).map((p) =>
        p.mem_total > 0 ? (p.mem_used / p.mem_total) * 100 : 0,
      ),
    [metrics],
  );
  const diskSeries = useMemo(
    () =>
      (metrics?.points ?? []).map((p) =>
        p.disk_total > 0 ? (p.disk_used / p.disk_total) * 100 : 0,
      ),
    [metrics],
  );
  const loadSeries = useMemo(
    () => (metrics?.points ?? []).map((p) => p.load1),
    [metrics],
  );

  async function onAck(id: string) {
    try {
      await ackEvent(id);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ack failed");
    }
  }

  const visibleEvents = events.filter((ev) => !ev.acknowledged_at);
  const needsAttention = events.filter(
    (ev) =>
      !ev.acknowledged_at && (ev.level === "error" || ev.level === "warning"),
  );

  async function onCleanup(targets: string[]) {
    setCleanupBusy(true);
    try {
      await postServerCleanup(targets);
      await refresh();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Cleanup failed");
    } finally {
      setCleanupBusy(false);
    }
  }

  return (
    <DashboardShell title="Analytics">
      <div className="space-y-8">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="text-sm text-[var(--muted-foreground)]">
            Host metrics sampled every minute.
          </p>
          <div
            className="flex flex-wrap gap-1 rounded-lg border border-[var(--border)] bg-[var(--surface)] p-1"
            role="tablist"
            aria-label="Metric range"
          >
            {ranges.map((r) => (
              <button
                key={r.id}
                type="button"
                role="tab"
                aria-selected={range === r.id}
                className={`rounded-md px-3 py-1.5 text-sm ${
                  range === r.id
                    ? "bg-[var(--primary)] text-[var(--primary-foreground)]"
                    : "text-[var(--muted-foreground)] hover:text-[var(--foreground)]"
                }`}
                onClick={() => setRange(r.id)}
              >
                {r.label}
              </button>
            ))}
          </div>
        </div>

        {error ? (
          <p className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-700">
            {error}
          </p>
        ) : null}

        {loading && !metrics ? (
          <p className="text-sm text-[var(--muted-foreground)]">
            Loading metrics…
          </p>
        ) : (
          <section
            className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4"
            aria-label="Host metrics"
          >
            <MetricTile
              label="CPU"
              value={latest ? `${(latest.cpu * 100).toFixed(1)}%` : "—"}
              series={cpuSeries}
            />
            <MetricTile
              label="Memory"
              value={
                latest && latest.mem_total > 0
                  ? `${((latest.mem_used / latest.mem_total) * 100).toFixed(1)}%`
                  : "—"
              }
              sub={
                latest
                  ? `${formatBytes(latest.mem_used)} / ${formatBytes(latest.mem_total)}`
                  : undefined
              }
              series={memSeries}
            />
            <MetricTile
              label="Disk"
              value={
                latest && latest.disk_total > 0
                  ? `${((latest.disk_used / latest.disk_total) * 100).toFixed(1)}%`
                  : "—"
              }
              sub={
                latest
                  ? `${formatBytes(latest.disk_used)} / ${formatBytes(latest.disk_total)}`
                  : undefined
              }
              series={diskSeries}
            />
            <MetricTile
              label="Load (1m)"
              value={latest ? latest.load1.toFixed(2) : "—"}
              series={loadSeries}
            />
          </section>
        )}

        <section className="space-y-3" aria-labelledby="events-heading">
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h2 id="events-heading" className="text-lg font-semibold">
              Important events
            </h2>
            <div className="flex flex-wrap gap-1">
              {eventFilters.map((f) => (
                <button
                  key={f.id}
                  type="button"
                  className={`rounded-full border px-3 py-1 text-xs font-medium ${
                    eventFilter === f.id
                      ? "border-[var(--primary)] bg-[var(--primary)]/10 text-[var(--foreground)]"
                      : "border-[var(--border)] text-[var(--muted-foreground)]"
                  }`}
                  onClick={() => setEventFilter(f.id)}
                >
                  {f.label}
                </button>
              ))}
            </div>
          </div>
          <ul className="divide-y divide-[var(--border)] rounded-xl border border-[var(--border)] bg-[var(--surface)]">
            {visibleEvents.length === 0 ? (
              <li className="px-4 py-6 text-sm text-[var(--muted-foreground)]">
                No unacknowledged events in this range.
              </li>
            ) : (
              visibleEvents.map((ev) => (
                <li
                  key={ev.id}
                  className="flex flex-wrap items-start justify-between gap-3 px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-medium text-[var(--foreground)]">
                      {ev.title}
                    </p>
                    {ev.detail ? (
                      <p className="mt-0.5 text-sm text-[var(--muted-foreground)]">
                        {ev.detail}
                      </p>
                    ) : null}
                    <p className="mt-1 text-xs text-[var(--muted-foreground)]">
                      {ev.level} · {ev.category || "general"}
                    </p>
                  </div>
                  <button
                    type="button"
                    className="rounded-md border border-[var(--border)] px-2 py-1 text-xs font-medium hover:bg-[var(--muted)]/40"
                    onClick={() => void onAck(ev.id)}
                  >
                    Dismiss
                  </button>
                </li>
              ))
            )}
          </ul>
        </section>

        <section className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
            <h2 className="text-lg font-semibold">Needs attention</h2>
            <ul className="mt-3 space-y-2">
              {needsAttention.length === 0 ? (
                <li className="text-sm text-[var(--muted-foreground)]">Nothing flagged.</li>
              ) : (
                needsAttention.slice(0, 5).map((ev) => (
                  <li key={ev.id} className="flex items-center justify-between gap-2 text-sm">
                    <span>{ev.title}</span>
                    {ev.action_href ? (
                      <Link href={ev.action_href} className="text-xs underline">
                        Open service
                      </Link>
                    ) : null}
                  </li>
                ))
              )}
            </ul>
          </div>
          <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
            <h2 className="text-lg font-semibold">Docker cleanup</h2>
            {cleanup ? (
              <ul className="mt-3 space-y-2 text-sm">
                <li>Build cache: {formatBytes(cleanup.build_cache_bytes)}</li>
                <li>
                  Dangling images: {formatBytes(cleanup.dangling_images_bytes)}
                </li>
                <li>
                  Stopped unmanaged containers:{" "}
                  {cleanup.stopped_unmanaged_containers}
                </li>
                <li>
                  Old static releases: {formatBytes(cleanup.old_releases_bytes)}
                </li>
              </ul>
            ) : (
              <p className="mt-2 text-sm text-[var(--muted-foreground)]">
                Docker is not available on this host.
              </p>
            )}
            {isAdmin && cleanup ? (
              <button
                type="button"
                disabled={cleanupBusy}
                className="mt-4 rounded-md bg-[var(--primary)] px-3 py-1.5 text-xs text-[var(--primary-foreground)]"
                onClick={() =>
                  void onCleanup([
                    "dangling_images",
                    "build_cache",
                    "stopped_unmanaged",
                    "old_releases",
                  ])
                }
              >
                {cleanupBusy ? "Cleaning…" : "Review and clean up"}
              </button>
            ) : null}
          </div>
        </section>

        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
          <h2 className="text-lg font-semibold">Running containers</h2>
          <div className="mt-3 overflow-x-auto">
            {containers.length === 0 ? (
              <p className="text-sm text-[var(--muted-foreground)]">
                No running Dockyard containers.
              </p>
            ) : (
              <table
                className="w-full text-left text-sm"
                aria-label="Running containers"
              >
                <thead>
                  <tr className="text-[var(--muted-foreground)]">
                    <th className="pb-2 pr-4 font-medium">Name</th>
                    <th className="pb-2 pr-4 font-medium">Kind</th>
                    <th className="pb-2 font-medium">State</th>
                  </tr>
                </thead>
                <tbody>
                  {containers.map((c) => (
                    <tr key={c.id} className="border-t border-[var(--border)]">
                      <td className="py-2 pr-4 font-mono text-xs">{c.name}</td>
                      <td className="py-2 pr-4">{c.kind || "—"}</td>
                      <td className="py-2">{c.state}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </section>
      </div>
    </DashboardShell>
  );
}
