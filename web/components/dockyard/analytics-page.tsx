"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import {
  ackEvent,
  type DockyardEvent,
  getServerMetrics,
  listEvents,
  type MetricRange,
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

function Sparkline({ values, stroke = "var(--primary)" }: { values: number[]; stroke?: string }) {
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
      <p className="mt-1 text-2xl font-semibold tabular-nums text-[var(--foreground)]">{value}</p>
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
  const [eventFilter, setEventFilter] = useState<(typeof eventFilters)[number]["id"]>("all");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const [m, ev] = await Promise.all([
        getServerMetrics(range),
        listEvents({ limit: 50, category: eventFilter === "all" ? undefined : eventFilter }),
      ]);
      setMetrics(m);
      setEvents(ev.events);
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
    () => (metrics?.points ?? []).map((p) => (p.mem_total > 0 ? (p.mem_used / p.mem_total) * 100 : 0)),
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
          <p className="text-sm text-[var(--muted-foreground)]">Loading metrics…</p>
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
                <li key={ev.id} className="flex flex-wrap items-start justify-between gap-3 px-4 py-3">
                  <div>
                    <p className="text-sm font-medium text-[var(--foreground)]">{ev.title}</p>
                    {ev.detail ? (
                      <p className="mt-0.5 text-sm text-[var(--muted-foreground)]">{ev.detail}</p>
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
            <p className="mt-2 text-sm text-[var(--muted-foreground)]">
              Service health alerts arrive in a later M4 slice.
            </p>
          </div>
          <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
            <h2 className="text-lg font-semibold">Docker cleanup</h2>
            <p className="mt-2 text-sm text-[var(--muted-foreground)]">
              Prune recommendations connect to the reconciler prune job next.
            </p>
          </div>
        </section>

        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4">
          <h2 className="text-lg font-semibold">Running containers</h2>
          <p className="mt-2 text-sm text-[var(--muted-foreground)]">
            Container metrics table ships with Docker stats API wiring.
          </p>
        </section>
      </div>
    </DashboardShell>
  );
}
