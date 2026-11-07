"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { TimeRange } from "@/components/dockyard/ui/time-range-toggle";
import {
  ackEvent,
  type CleanupSummary,
  type ContainerMetricSnapshot,
  type dockyrdEvent,
  getMe,
  getServerCleanup,
  getServerContainerMetrics,
  getServerContainers,
  getServerInfo,
  getServerMetrics,
  listEvents,
  type MetricRange,
  postServerCleanup,
  type ServerContainer,
  type ServerInfo,
  type ServerMetricsResponse,
} from "@/lib/api";

export type AnalyticsTabId = "overview" | "cpu" | "memory" | "disk" | "network";

export type AnalyticsData = {
  range: TimeRange;
  metrics: ServerMetricsResponse | null;
  serverInfo: ServerInfo | null;
  events: dockyrdEvent[];
  allEvents: dockyrdEvent[];
  containers: ServerContainer[];
  containerMetrics: ContainerMetricSnapshot[];
  cleanup: CleanupSummary | null;
  isAdmin: boolean;
  loading: boolean;
  error: string | null;
  refresh: (opts?: { silent?: boolean }) => Promise<void>;
  setRange: (r: TimeRange) => void;
  eventFilter: string;
  setEventFilter: (f: string) => void;
  cleanupBusy: boolean;
  runCleanup: (targets: string[]) => Promise<void>;
  ack: (id: string) => Promise<void>;
  lastRefreshedAt: number | null;
};

function metricRangeFromTimeRange(range: TimeRange): MetricRange {
  return range;
}

export function useAnalyticsData(): AnalyticsData {
  const [range, setRange] = useState<TimeRange>("24h");
  const [metrics, setMetrics] = useState<ServerMetricsResponse | null>(null);
  const [serverInfo, setServerInfo] = useState<ServerInfo | null>(null);
  const [events, setEvents] = useState<dockyrdEvent[]>([]);
  const [eventFilter, setEventFilter] = useState("all");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [containers, setContainers] = useState<ServerContainer[]>([]);
  const [containerMetrics, setContainerMetrics] = useState<
    ContainerMetricSnapshot[]
  >([]);
  const [cleanup, setCleanup] = useState<CleanupSummary | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [cleanupBusy, setCleanupBusy] = useState(false);
  const [lastRefreshedAt, setLastRefreshedAt] = useState<number | null>(null);

  const refresh = useCallback(
    async (opts?: { silent?: boolean }) => {
      const silent = opts?.silent === true;
      if (!silent) {
        setLoading(true);
      }
      try {
        const apiRange = metricRangeFromTimeRange(range);
        const [m, ev, info, cm] = await Promise.all([
          getServerMetrics(apiRange),
          listEvents({ limit: 50 }),
          getServerInfo(),
          getServerContainerMetrics(apiRange).catch(() => ({
            range: apiRange,
            metrics: [],
          })),
        ]);
        setMetrics(m);
        setServerInfo(info);
        setContainerMetrics(cm.metrics);
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
        setLastRefreshedAt(Math.floor(Date.now() / 1000));
      } catch (e) {
        if (!silent) {
          setError(e instanceof Error ? e.message : "Failed to load analytics");
        }
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [range],
  );

  useEffect(() => {
    void refresh();
  }, [refresh]);

  useEffect(() => {
    const id = window.setInterval(() => {
      if (document.visibilityState === "visible") {
        void refresh({ silent: true });
      }
    }, 60_000);
    return () => window.clearInterval(id);
  }, [refresh]);

  const filteredEvents = useMemo(() => {
    return events.filter((ev) => {
      if (ev.acknowledged_at) return false;
      if (eventFilter === "all") return true;
      if (eventFilter === "issues") {
        return ev.level === "error" || ev.level === "warning";
      }
      if (eventFilter === "deploys") return ev.category === "deploys";
      if (eventFilter === "system") return ev.category === "system";
      return ev.category === eventFilter;
    });
  }, [events, eventFilter]);

  const runCleanup = useCallback(
    async (targets: string[]) => {
      setCleanupBusy(true);
      try {
        await postServerCleanup(targets);
        await refresh();
      } finally {
        setCleanupBusy(false);
      }
    },
    [refresh],
  );

  const ack = useCallback(
    async (id: string) => {
      await ackEvent(id);
      await refresh();
    },
    [refresh],
  );

  return {
    range,
    metrics,
    serverInfo,
    events: filteredEvents,
    allEvents: events,
    containers,
    containerMetrics,
    cleanup,
    isAdmin,
    loading,
    error,
    refresh,
    setRange,
    eventFilter,
    setEventFilter,
    cleanupBusy,
    runCleanup,
    ack,
    lastRefreshedAt,
  };
}

export function ViewDetailsLink({
  tab,
  onSelectTab,
}: {
  tab: AnalyticsTabId;
  onSelectTab: (id: AnalyticsTabId) => void;
}) {
  return (
    <button
      type="button"
      className="mt-3 text-[13px] font-medium underline-offset-2 hover:underline"
      style={{ color: "var(--link-blue)" }}
      onClick={() => onSelectTab(tab)}
    >
      View details →
    </button>
  );
}

export function AnalyticsChartCard({
  title,
  subtitle,
  children,
}: {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
}) {
  return (
    <article
      className="rounded-[20px] border p-5"
      style={{ background: "var(--surface)", borderColor: "var(--border)" }}
    >
      <p className="font-semibold">{title}</p>
      {subtitle ? (
        <p className="mt-1 text-[13px]" style={{ color: "var(--muted-text)" }}>
          {subtitle}
        </p>
      ) : null}
      <div className="mt-4">{children}</div>
    </article>
  );
}
