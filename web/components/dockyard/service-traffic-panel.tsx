"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  areaLineOption,
  stackedBarOption,
} from "@/components/dockyrd/analytics/chart-options";
import { Button } from "@/components/dockyrd/ui/button";
import { EChart } from "@/components/dockyrd/ui/echart";
import { Tabs } from "@/components/dockyrd/ui/tabs";
import { formatBytes } from "@/lib/analytics-stats";
import {
  fetchServiceLogs,
  getServiceAnalytics,
  type ServiceAnalytics,
} from "@/lib/api";
import {
  RESOURCE_DETAIL_CARD,
  resourceDetailCardStyle,
} from "@/lib/resource-detail-page";

type ServiceAccessLogsPanelProps = {
  serviceId: string;
};

export function ServiceAccessLogsPanel({
  serviceId,
}: ServiceAccessLogsPanelProps) {
  const [log, setLog] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const text = await fetchServiceLogs(serviceId, "200", "access");
      setLog(text);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load logs");
      setLog("");
    } finally {
      setLoading(false);
    }
  }, [serviceId]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  return (
    <section
      className={RESOURCE_DETAIL_CARD}
      style={resourceDetailCardStyle()}
      aria-label="Access logs"
    >
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold tracking-tight">Access logs</h2>
          <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
            HTTP requests served by Caddy for this site&apos;s domains.
          </p>
        </div>
        <Button type="button" variant="outline" onClick={() => void refresh()}>
          Refresh
        </Button>
      </div>
      {error ? (
        <p className="mt-4 text-sm text-red-600" role="alert">
          {error}
        </p>
      ) : null}
      <pre
        className="mt-4 max-h-[480px] overflow-auto rounded-xl border p-4 font-mono text-[12px] leading-relaxed whitespace-pre-wrap"
        style={{
          borderColor: "var(--border)",
          background: "var(--bg)",
          color: "var(--text)",
        }}
      >
        {loading
          ? "Loading…"
          : log.trim() ||
            "No access log lines yet. Traffic appears here after requests hit your domains."}
      </pre>
    </section>
  );
}

const RANGE_TABS = [
  { id: "1h", label: "1 hour" },
  { id: "24h", label: "24 hours" },
  { id: "7d", label: "7 days" },
] as const;

type RangeId = (typeof RANGE_TABS)[number]["id"];

function formatBucketLabel(ts: number, range: RangeId): string {
  const d = new Date(ts * 1000);
  if (range === "7d") {
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  }
  return d.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  });
}

type ServiceTrafficAnalyticsPanelProps = {
  serviceId: string;
};

export function ServiceTrafficAnalyticsPanel({
  serviceId,
}: ServiceTrafficAnalyticsPanelProps) {
  const [range, setRange] = useState<RangeId>("24h");
  const [data, setData] = useState<ServiceAnalytics | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getServiceAnalytics(serviceId, range);
      setData(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load analytics");
      setData(null);
    } finally {
      setLoading(false);
    }
  }, [serviceId, range]);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const labels = useMemo(
    () => (data?.series ?? []).map((b) => formatBucketLabel(b.ts, range)),
    [data, range],
  );

  const requestsOpt = useMemo(
    () =>
      areaLineOption({
        labels,
        values: (data?.series ?? []).map((b) => b.requests),
        seriesName: "Requests",
        unit: "",
      }),
    [data, labels],
  );

  const latencyOpt = useMemo(
    () =>
      areaLineOption({
        labels,
        values: (data?.series ?? []).map((b) => Math.round(b.latency_avg_ms)),
        seriesName: "Avg response",
        unit: " ms",
      }),
    [data, labels],
  );

  const statusOpt = useMemo(
    () =>
      stackedBarOption({
        labels,
        series: [
          {
            name: "2xx",
            values: (data?.series ?? []).map((b) => b.status_2xx),
          },
          {
            name: "3xx",
            values: (data?.series ?? []).map((b) => b.status_3xx),
          },
          {
            name: "4xx",
            values: (data?.series ?? []).map((b) => b.status_4xx),
          },
          {
            name: "5xx",
            values: (data?.series ?? []).map((b) => b.status_5xx),
          },
        ],
      }),
    [data, labels],
  );

  const bytesOpt = useMemo(
    () =>
      areaLineOption({
        labels,
        values: (data?.series ?? []).map((b) => b.bytes),
        yFormatter: (v) => formatBytes(v),
        seriesName: "Bandwidth",
      }),
    [data, labels],
  );

  return (
    <div className="space-y-6">
      <section
        className={RESOURCE_DETAIL_CARD}
        style={resourceDetailCardStyle()}
      >
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold tracking-tight">Analytics</h2>
            <p className="mt-1 text-sm" style={{ color: "var(--muted-text)" }}>
              Traffic from Caddy access logs — no cookies, no stored IPs.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Tabs
              tabs={[...RANGE_TABS]}
              activeId={range}
              onChange={(id) => setRange(id as RangeId)}
            />
            <Button
              type="button"
              variant="outline"
              onClick={() => void refresh()}
            >
              Refresh
            </Button>
          </div>
        </div>
        {error ? (
          <p className="mt-4 text-sm text-red-600" role="alert">
            {error}
          </p>
        ) : null}
        <dl className="mt-6 grid gap-4 sm:grid-cols-3">
          <div>
            <dt className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)]">
              Requests
            </dt>
            <dd className="mt-1 text-2xl font-semibold tabular-nums">
              {loading ? "…" : (data?.requests_total ?? 0).toLocaleString()}
            </dd>
          </div>
          <div>
            <dt className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)]">
              Bandwidth
            </dt>
            <dd className="mt-1 text-2xl font-semibold tabular-nums">
              {loading ? "…" : formatBytes(data?.bytes_total ?? 0)}
            </dd>
          </div>
          <div>
            <dt className="font-mono text-[11px] uppercase tracking-widest text-[var(--muted-text)]">
              Errors (4xx/5xx)
            </dt>
            <dd className="mt-1 text-2xl font-semibold tabular-nums">
              {loading
                ? "…"
                : (
                    (data?.status_4xx ?? 0) + (data?.status_5xx ?? 0)
                  ).toLocaleString()}
            </dd>
          </div>
        </dl>
      </section>

      {!loading && (data?.series.length ?? 0) === 0 ? (
        <section
          className={RESOURCE_DETAIL_CARD}
          style={resourceDetailCardStyle()}
        >
          <p className="text-sm" style={{ color: "var(--muted-text)" }}>
            No traffic recorded in this range yet. Open the site in a browser or
            check that domains route through dockyrd.
          </p>
        </section>
      ) : (
        <>
          <ChartCard title="Requests per minute" option={requestsOpt} />
          <ChartCard title="Response time (avg)" option={latencyOpt} />
          <ChartCard title="Status codes" option={statusOpt} height={220} />
          <ChartCard title="Bandwidth" option={bytesOpt} />
        </>
      )}
    </div>
  );
}

function ChartCard({
  title,
  option,
  height = 200,
}: {
  title: string;
  option: ReturnType<typeof areaLineOption>;
  height?: number;
}) {
  return (
    <section className={RESOURCE_DETAIL_CARD} style={resourceDetailCardStyle()}>
      <h3 className="text-base font-semibold tracking-tight">{title}</h3>
      <div className="mt-4">
        <EChart option={option} height={height} ariaLabel={title} />
      </div>
    </section>
  );
}
