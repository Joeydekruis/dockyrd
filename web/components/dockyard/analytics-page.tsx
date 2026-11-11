"use client";

import { useMemo, useState } from "react";
import {
  CpuTab,
  containerTableRows,
  DiskTab,
  EventsAndOpsSection,
  MemoryTab,
  NetworkTab,
  OverviewTab,
  rangeLabelFromRange,
} from "@/components/dockyard/analytics/analytics-tab-views";
import {
  type AnalyticsTabId,
  useAnalyticsData,
} from "@/components/dockyard/analytics/use-analytics-data";
import { DashboardShell } from "@/components/dockyard/dashboard-shell";
import { Button } from "@/components/dockyard/ui/button";
import { Chip } from "@/components/dockyard/ui/chip";
import { PageHeader } from "@/components/dockyard/ui/page-header";
import { Tabs } from "@/components/dockyard/ui/tabs";
import { TimeRangeToggle } from "@/components/dockyard/ui/time-range-toggle";
import { formatBytes, formatUptime } from "@/lib/analytics-stats";
import { useUserLocale } from "@/lib/user-locale-context";

const analyticsTabs = [
  { id: "overview", label: "Overview" },
  { id: "cpu", label: "CPU" },
  { id: "memory", label: "Memory" },
  { id: "disk", label: "Disk" },
  { id: "network", label: "Network" },
] as const;

export function AnalyticsPage() {
  const data = useAnalyticsData();
  const { formatDateTime } = useUserLocale();
  const [tab, setTab] = useState<AnalyticsTabId>("overview");

  const points = data.metrics?.points ?? [];
  const latest = points.at(-1);
  const rangeLabel = rangeLabelFromRange(data.range);
  const containerRows = useMemo(
    () => containerTableRows(data.containers),
    [data.containers],
  );

  const tabProps = {
    points,
    latest,
    serverInfo: data.serverInfo,
    containerMetrics: data.containerMetrics,
    rangeLabel,
    onSelectTab: setTab,
  };

  const metaChips = (
    <div className="mb-4 flex flex-wrap gap-2">
      {data.serverInfo?.hostname ? (
        <Chip>{data.serverInfo.hostname.toUpperCase()}</Chip>
      ) : null}
      {data.serverInfo?.cpu_count ? (
        <Chip>{data.serverInfo.cpu_count} VCPU</Chip>
      ) : null}
      {latest?.mem_total ? (
        <Chip>{formatBytes(latest.mem_total)} RAM</Chip>
      ) : null}
      {data.serverInfo?.uptime_sec ? (
        <Chip>UP {formatUptime(data.serverInfo.uptime_sec)}</Chip>
      ) : null}
    </div>
  );

  const lastUpdatedLine =
    data.lastRefreshedAt != null ? (
      <p
        className="shrink-0 pb-3 text-right text-[13px]"
        style={{ color: "var(--muted-text)" }}
      >
        Last refreshed:{" "}
        <time
          dateTime={new Date(data.lastRefreshedAt * 1000).toISOString()}
          className="font-mono tabular-nums"
        >
          {formatDateTime(data.lastRefreshedAt)}
        </time>
        {data.loading ? " · updating…" : null}
      </p>
    ) : data.loading ? (
      <p
        className="shrink-0 pb-3 text-right text-[13px]"
        style={{ color: "var(--muted-text)" }}
      >
        Loading metrics…
      </p>
    ) : null;

  return (
    <DashboardShell title="Analytics">
      {metaChips}
      <PageHeader
        title="Analytics"
        subtitle="Monitor your server, from the key numbers to the detailed charts. Refreshes every minute while you stay on this page."
        action={
          <div className="flex flex-wrap items-center gap-2">
            <TimeRangeToggle value={data.range} onChange={data.setRange} />
            <Button
              type="button"
              variant="outline"
              onClick={() => void data.refresh()}
              loading={data.loading}
            >
              Refresh
            </Button>
          </div>
        }
      />
      <div
        className="flex flex-wrap items-end justify-between gap-x-6 gap-y-2 border-b"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="min-w-0 [&>div]:border-b-0">
          <Tabs
            tabs={[...analyticsTabs]}
            activeId={tab}
            onChange={(id) => setTab(id as AnalyticsTabId)}
          />
        </div>
        {lastUpdatedLine}
      </div>

      <div className="mt-8 space-y-10">
        {data.error ? (
          <p
            className="rounded-lg border px-3 py-2 text-sm"
            style={{
              borderColor: "rgba(179,38,30,0.3)",
              background: "rgba(179,38,30,0.08)",
              color: "#b3261e",
            }}
          >
            {data.error}
          </p>
        ) : null}

        {data.loading && !data.metrics ? (
          <p className="sr-only" aria-live="polite">
            Loading metrics…
          </p>
        ) : null}

        {tab === "overview" ? <OverviewTab {...tabProps} /> : null}
        {tab === "cpu" ? <CpuTab {...tabProps} /> : null}
        {tab === "memory" ? <MemoryTab {...tabProps} /> : null}
        {tab === "disk" ? (
          <DiskTab
            {...tabProps}
            cleanup={data.cleanup}
            diskUsage={data.diskUsage}
          />
        ) : null}
        {tab === "network" ? <NetworkTab {...tabProps} /> : null}

        {tab === "overview" ? (
          <EventsAndOpsSection
            events={data.events}
            allEvents={data.allEvents}
            eventFilter={data.eventFilter}
            setEventFilter={data.setEventFilter}
            onAck={(id) => void data.ack(id)}
            cleanup={data.cleanup}
            isAdmin={data.isAdmin}
            cleanupBusy={data.cleanupBusy}
            onCleanup={() =>
              void data.runCleanup([
                "dangling_images",
                "build_cache",
                "stopped_unmanaged",
                "old_releases",
              ])
            }
            containerRows={containerRows}
          />
        ) : null}
      </div>
    </DashboardShell>
  );
}
